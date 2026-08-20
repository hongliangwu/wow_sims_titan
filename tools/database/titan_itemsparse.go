package database

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

const (
	titanCustomItemMinID     = 200000
	titanOriginalIlvlMin     = 166
	itemFlagUniqueEquippable = 0x80000
)

// titanPhaseFromIlvl is the fallback when AtlasLoot has no mapped drop source
// (embers, world drops, unlisted loot). Raid/dungeon items use TitanPhaseFromSources.
func titanPhaseFromIlvl(ilvl int32) int32 {
	switch {
	case ilvl <= 213:
		return 1
	case ilvl <= 219:
		return 2
	case ilvl <= 226:
		return 3
	case ilvl <= 232:
		return 4
	case ilvl <= 238:
		return 5
	case ilvl <= 245:
		return 6
	case ilvl <= 251:
		return 7
	case ilvl <= 258:
		return 8
	case ilvl <= 264:
		return 9
	case ilvl <= 272:
		return 10
	default:
		return 11
	}
}

type itemClassRow struct {
	classID        int
	subclassID     int
	inventory      int
	iconFileDataID int32
}

// LoadTitanItemSparse reads a DBC2CSV dump of the Titan Time ItemSparse table
// (plus Item.csv for class/subclass) and returns equippable custom items.
func LoadTitanItemSparse(sparsePath, itemPath, iconMapPath string) ([]*proto.UIItem, error) {
	itemRows, err := loadItemClassRows(itemPath)
	if err != nil {
		return nil, err
	}
	iconNames, err := loadTitanIconNames(iconMapPath)
	if err != nil {
		log.Printf("Titan icon map not loaded (%v); item icons will be empty", err)
		iconNames = map[int32]string{}
	}

	f, err := os.Open(sparsePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read ItemSparse header: %w", err)
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	for _, required := range []string{"ID", "Display_lang", "ItemLevel", "OverallQualityID", "InventoryType"} {
		if _, ok := col[required]; !ok {
			return nil, fmt.Errorf("ItemSparse.csv missing column %s", required)
		}
	}

	type pending struct {
		item  *proto.UIItem
		setID int32
		name  string
	}
	var rows []pending
	setNames := map[int32][]string{}
	unknownSocketBonus := map[int32]int{}

	catalog, catErr := loadItemSetCatalog(filepath.Dir(sparsePath))
	if catErr != nil {
		log.Printf("Titan ItemSet catalog not loaded (%v); set names will use item-name prefixes", catErr)
		catalog = &itemSetCatalog{names: map[int32]string{}}
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read ItemSparse row: %w", err)
		}

		id := csvInt(rec, col, "ID")
		ilvl := csvInt(rec, col, "ItemLevel")
		// Custom Titan IDs are always imported; original IDs keep raid/dungeon loot
		// that the timewalking client still uses (MC/SSC/SWP etc. are not remapped to 200000+).
		if id < titanCustomItemMinID && ilvl < titanOriginalIlvlMin {
			continue
		}
		inv := csvInt(rec, col, "InventoryType")
		itemType, handType := inventoryToItemType(inv)
		if itemType == proto.ItemType_ItemTypeUnknown {
			continue
		}

		name := csvStr(rec, col, "Display_lang")
		if name == "" || titanSkipItemName(name) {
			continue
		}

		quality := proto.ItemQuality(csvInt(rec, col, "OverallQualityID"))
		if quality < proto.ItemQuality_ItemQualityUncommon {
			continue
		}

		st := stats.Stats{}
		applyResistances(&st, rec, col, itemType, handType)
		hasBonusStat := false
		for i := 0; i < 10; i++ {
			mod := csvInt(rec, col, fmt.Sprintf("StatModifier_bonusStat[%d]", i))
			amount := csvFloat(rec, col, fmt.Sprintf("StatModifier_bonusAmount[%d]", i))
			if mod < 0 || amount == 0 {
				continue
			}
			if applyItemMod(&st, mod, amount) {
				hasBonusStat = true
			}
		}

		minDmg := csvFloat(rec, col, "MinDamage[0]")
		maxDmg := csvFloat(rec, col, "MaxDamage[0]")
		speed := csvFloat(rec, col, "ItemDelay") / 1000
		if !hasBonusStat && minDmg == 0 && st[stats.Armor] == 0 && itemType != proto.ItemType_ItemTypeRanged {
			continue
		}

		classInfo := itemRows[id]
		armorType, weaponType, rangedType := classToTypes(classInfo.classID, classInfo.subclassID, int(inv), itemType)

		gemSockets := parseGemSockets(rec, col)
		socketBonusID := csvInt(rec, col, "Socket_match_enchantment_ID")
		socketBonus, ok := titanSocketBonus(socketBonusID)
		if socketBonusID != 0 && !ok {
			unknownSocketBonus[socketBonusID]++
		}

		flags0 := csvInt(rec, col, "Flags[0]")
		ui := &proto.UIItem{
			Id:               id,
			Name:             name,
			Icon:             iconNames[classInfo.iconFileDataID],
			Type:             itemType,
			ArmorType:        armorType,
			WeaponType:       weaponType,
			HandType:         handType,
			RangedWeaponType: rangedType,
			Stats:            st.ToFloatArray(),
			GemSockets:       gemSockets,
			SocketBonus:      socketBonus.ToFloatArray(),
			WeaponDamageMin:  minDmg,
			WeaponDamageMax:  maxDmg,
			WeaponSpeed:      speed,
			Ilvl:             ilvl,
			Phase:            titanPhaseFromIlvl(ilvl),
			Quality:          quality,
			Unique:           csvInt(rec, col, "MaxCount") == 1 || flags0&itemFlagUniqueEquippable != 0,
			// Titan Time raids do not use heroic difficulty; the DBC heroic bit
			// still appears on remapped loot and would show a leftover [H] badge.
			Heroic:             false,
			ClassAllowlist:     classMaskToAllowlist(csvInt(rec, col, "AllowableClass")),
			RequiredProfession: skillToProfession(csvInt(rec, col, "RequiredSkill")),
			Expansion:          proto.Expansion_ExpansionWotlk,
		}

		setID := csvInt(rec, col, "ItemSet")
		if setID != 0 {
			setNames[setID] = append(setNames[setID], name)
		}
		rows = append(rows, pending{item: ui, setID: setID, name: name})
	}

	resolvedSets := map[int32]string{}
	for setID, names := range setNames {
		if official, ok := catalog.names[setID]; ok && official != "" {
			resolvedSets[setID] = official
			continue
		}
		resolvedSets[setID] = commonSetName(names, setID)
	}

	out := make([]*proto.UIItem, 0, len(rows))
	for _, row := range rows {
		if row.setID != 0 {
			row.item.SetName = resolvedSets[row.setID]
		}
		out = append(out, row.item)
	}

	if len(unknownSocketBonus) > 0 {
		log.Printf("Titan ItemSparse: %d socket-bonus enchant IDs have no stat map (first: %v)", len(unknownSocketBonus), unknownSocketBonus)
	}
	return out, nil
}

// LoadTitanDisplayNames maps ItemSparse ID → zhCN Display_lang for overlaying
// names onto gems (and leftover English items) that are not re-imported as gear.
func LoadTitanDisplayNames(sparsePath string) (map[int32]string, error) {
	f, err := os.Open(sparsePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read ItemSparse header: %w", err)
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	if _, ok := col["ID"]; !ok {
		return nil, fmt.Errorf("ItemSparse.csv missing column ID")
	}
	if _, ok := col["Display_lang"]; !ok {
		return nil, fmt.Errorf("ItemSparse.csv missing column Display_lang")
	}

	out := map[int32]string{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read ItemSparse row: %w", err)
		}
		id := csvInt(rec, col, "ID")
		name := csvStr(rec, col, "Display_lang")
		if id == 0 || name == "" || titanSkipItemName(name) {
			continue
		}
		out[id] = name
	}
	return out, nil
}

type titanDedupeKey struct {
	name  string
	typ   proto.ItemType
	hand  proto.HandType
	fact  proto.UIItem_FactionRestriction
	phase int32
}

// DropTitanLowerIlvlDuplicates removes leftover difficulty variants that share a
// name/slot/faction/phase with a higher-ilvl counterpart (e.g. ToC 死亡的裁决
// 232 vs 238). Same-ilvl copies are kept.
func DropTitanLowerIlvlDuplicates(db *WowDatabase) int {
	if db == nil {
		return 0
	}
	groups := map[titanDedupeKey][]*proto.UIItem{}
	for _, item := range db.Items {
		if item == nil || item.Name == "" {
			continue
		}
		k := titanDedupeKey{
			name:  item.Name,
			typ:   item.Type,
			hand:  item.HandType,
			fact:  item.FactionRestriction,
			phase: item.Phase,
		}
		groups[k] = append(groups[k], item)
	}
	dropped := 0
	for _, items := range groups {
		if len(items) < 2 {
			continue
		}
		var maxIlvl int32
		for _, item := range items {
			if item.Ilvl > maxIlvl {
				maxIlvl = item.Ilvl
			}
		}
		for _, item := range items {
			if item.Ilvl < maxIlvl {
				delete(db.Items, item.Id)
				dropped++
			}
		}
	}
	return dropped
}

func titanSkipItemName(name string) bool {
	if strings.HasPrefix(name, "Monster -") {
		return true
	}
	lower := strings.ToLower(name)
	if strings.Contains(lower, "qa test") || strings.Contains(lower, "obsolete") {
		return true
	}
	return false
}

func loadItemClassRows(path string) (map[int32]itemClassRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read Item.csv header: %w", err)
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	for _, required := range []string{"ID", "ClassID", "SubclassID"} {
		if _, ok := col[required]; !ok {
			return nil, fmt.Errorf("Item.csv missing column %s", required)
		}
	}

	out := map[int32]itemClassRow{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read Item.csv row: %w", err)
		}
		id := csvInt(rec, col, "ID")
		out[id] = itemClassRow{
			classID:        int(csvInt(rec, col, "ClassID")),
			subclassID:     int(csvInt(rec, col, "SubclassID")),
			inventory:      int(csvInt(rec, col, "InventoryType")),
			iconFileDataID: csvInt(rec, col, "IconFileDataID"),
		}
	}
	return out, nil
}

func loadTitanIconNames(path string) (map[int32]string, error) {
	if path == "" {
		return map[int32]string{}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	if _, ok := col["FileDataID"]; !ok {
		return nil, fmt.Errorf("icon map missing FileDataID")
	}
	if _, ok := col["Icon"]; !ok {
		return nil, fmt.Errorf("icon map missing Icon")
	}

	out := map[int32]string{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		id := csvInt(rec, col, "FileDataID")
		name := strings.TrimSpace(csvStr(rec, col, "Icon"))
		if id != 0 && name != "" {
			out[id] = name
		}
	}
	return out, nil
}

func csvStr(rec []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i >= len(rec) {
		return ""
	}
	return rec[i]
}

func csvInt(rec []string, col map[string]int, name string) int32 {
	s := strings.TrimSpace(csvStr(rec, col, name))
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}

func csvFloat(rec []string, col map[string]int, name string) float64 {
	s := strings.TrimSpace(csvStr(rec, col, name))
	if s == "" {
		return 0
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return n
}

func inventoryToItemType(inv int32) (proto.ItemType, proto.HandType) {
	switch inv {
	case 1:
		return proto.ItemType_ItemTypeHead, proto.HandType_HandTypeUnknown
	case 2:
		return proto.ItemType_ItemTypeNeck, proto.HandType_HandTypeUnknown
	case 3:
		return proto.ItemType_ItemTypeShoulder, proto.HandType_HandTypeUnknown
	case 5, 20:
		return proto.ItemType_ItemTypeChest, proto.HandType_HandTypeUnknown
	case 6:
		return proto.ItemType_ItemTypeWaist, proto.HandType_HandTypeUnknown
	case 7:
		return proto.ItemType_ItemTypeLegs, proto.HandType_HandTypeUnknown
	case 8:
		return proto.ItemType_ItemTypeFeet, proto.HandType_HandTypeUnknown
	case 9:
		return proto.ItemType_ItemTypeWrist, proto.HandType_HandTypeUnknown
	case 10:
		return proto.ItemType_ItemTypeHands, proto.HandType_HandTypeUnknown
	case 11:
		return proto.ItemType_ItemTypeFinger, proto.HandType_HandTypeUnknown
	case 12:
		return proto.ItemType_ItemTypeTrinket, proto.HandType_HandTypeUnknown
	case 13:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeOneHand
	case 14, 22, 23:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeOffHand
	case 15, 25, 26, 28:
		return proto.ItemType_ItemTypeRanged, proto.HandType_HandTypeUnknown
	case 16:
		return proto.ItemType_ItemTypeBack, proto.HandType_HandTypeUnknown
	case 17:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeTwoHand
	case 21:
		return proto.ItemType_ItemTypeWeapon, proto.HandType_HandTypeMainHand
	default:
		return proto.ItemType_ItemTypeUnknown, proto.HandType_HandTypeUnknown
	}
}

func classToTypes(classID, subclassID, inv int, itemType proto.ItemType) (proto.ArmorType, proto.WeaponType, proto.RangedWeaponType) {
	if classID == 2 {
		switch subclassID {
		case 0, 1:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeAxe, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 2:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeBow
		case 3:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeGun
		case 4, 5:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeMace, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 6:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypePolearm, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 7, 8:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeSword, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 10:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeStaff, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 13:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeFist, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 15:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeDagger, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 16:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeThrown
		case 18:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeCrossbow
		case 19:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeWand
		}
	}

	if classID == 4 {
		switch subclassID {
		case 1:
			return proto.ArmorType_ArmorTypeCloth, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 2:
			return proto.ArmorType_ArmorTypeLeather, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 3:
			return proto.ArmorType_ArmorTypeMail, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 4:
			return proto.ArmorType_ArmorTypePlate, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 6:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeShield, proto.RangedWeaponType_RangedWeaponTypeUnknown
		case 7:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeLibram
		case 8:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeIdol
		case 9:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeTotem
		case 10:
			return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeSigil
		}
	}

	if inv == 16 {
		return proto.ArmorType_ArmorTypeCloth, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeUnknown
	}
	if inv == 23 {
		return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeOffHand, proto.RangedWeaponType_RangedWeaponTypeUnknown
	}
	if inv == 14 {
		return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeShield, proto.RangedWeaponType_RangedWeaponTypeUnknown
	}
	_ = itemType
	return proto.ArmorType_ArmorTypeUnknown, proto.WeaponType_WeaponTypeUnknown, proto.RangedWeaponType_RangedWeaponTypeUnknown
}

func parseGemSockets(rec []string, col map[string]int) []proto.GemColor {
	var out []proto.GemColor
	for i := 0; i < 3; i++ {
		t := csvInt(rec, col, fmt.Sprintf("SocketType[%d]", i))
		color := socketTypeToColor(t)
		if color == proto.GemColor_GemColorUnknown {
			continue
		}
		out = append(out, color)
	}
	return out
}

func socketTypeToColor(t int32) proto.GemColor {
	switch t {
	case 1:
		return proto.GemColor_GemColorMeta
	case 2:
		return proto.GemColor_GemColorRed
	case 3:
		return proto.GemColor_GemColorYellow
	case 4:
		return proto.GemColor_GemColorBlue
	case 8, 14:
		return proto.GemColor_GemColorPrismatic
	default:
		return proto.GemColor_GemColorUnknown
	}
}

func applyItemMod(st *stats.Stats, mod int32, amount float64) bool {
	switch mod {
	case 0:
		st[stats.Mana] += amount
	case 1:
		st[stats.Health] += amount
	case 3:
		st[stats.Agility] += amount
	case 4:
		st[stats.Strength] += amount
	case 5:
		st[stats.Intellect] += amount
	case 6:
		st[stats.Spirit] += amount
	case 7:
		st[stats.Stamina] += amount
	case 12:
		st[stats.Defense] += amount
	case 13:
		st[stats.Dodge] += amount
	case 14:
		st[stats.Parry] += amount
	case 15:
		st[stats.Block] += amount
	case 16, 17, 18, 31:
		st[stats.MeleeHit] += amount
		st[stats.SpellHit] += amount
	case 19, 20, 21, 32:
		st[stats.MeleeCrit] += amount
		st[stats.SpellCrit] += amount
	case 35:
		st[stats.Resilience] += amount
	case 36:
		st[stats.MeleeHaste] += amount
		st[stats.SpellHaste] += amount
	case 37:
		st[stats.Expertise] += amount
	case 38:
		st[stats.AttackPower] += amount
		st[stats.RangedAttackPower] += amount
	case 39:
		st[stats.RangedAttackPower] += amount
	case 41, 42, 45:
		st[stats.SpellPower] += amount
	case 43:
		st[stats.MP5] += amount
	case 44:
		st[stats.ArmorPenetration] += amount
	case 47:
		st[stats.SpellPenetration] += amount
	case 48:
		st[stats.BlockValue] += amount
	default:
		return false
	}
	return true
}

func applyResistances(st *stats.Stats, rec []string, col map[string]int, itemType proto.ItemType, handType proto.HandType) {
	armor := csvFloat(rec, col, "Resistances[0]")
	if armor != 0 {
		bonusArmorSlot := itemType == proto.ItemType_ItemTypeNeck ||
			itemType == proto.ItemType_ItemTypeFinger ||
			itemType == proto.ItemType_ItemTypeTrinket ||
			(itemType == proto.ItemType_ItemTypeWeapon && handType != proto.HandType_HandTypeOffHand)
		if bonusArmorSlot {
			st[stats.BonusArmor] += armor
		} else {
			st[stats.Armor] += armor
		}
	}

	st[stats.FireResistance] += csvFloat(rec, col, "Resistances[2]")
	st[stats.NatureResistance] += csvFloat(rec, col, "Resistances[3]")
	st[stats.FrostResistance] += csvFloat(rec, col, "Resistances[4]")
	st[stats.ShadowResistance] += csvFloat(rec, col, "Resistances[5]")
	st[stats.ArcaneResistance] += csvFloat(rec, col, "Resistances[6]")
}

func classMaskToAllowlist(mask int32) []proto.Class {
	if mask <= 0 || uint32(mask) == 0xFFFFFFFF {
		return nil
	}
	type bitClass struct {
		bit   int32
		class proto.Class
	}
	bits := []bitClass{
		{1, proto.Class_ClassWarrior},
		{2, proto.Class_ClassPaladin},
		{4, proto.Class_ClassHunter},
		{8, proto.Class_ClassRogue},
		{16, proto.Class_ClassPriest},
		{32, proto.Class_ClassDeathknight},
		{64, proto.Class_ClassShaman},
		{128, proto.Class_ClassMage},
		{256, proto.Class_ClassWarlock},
		{1024, proto.Class_ClassDruid},
	}
	var out []proto.Class
	for _, b := range bits {
		if mask&b.bit != 0 {
			out = append(out, b.class)
		}
	}
	if len(out) == 0 || len(out) == len(bits) {
		return nil
	}
	return out
}

func skillToProfession(skill int32) proto.Profession {
	switch skill {
	case 164:
		return proto.Profession_Blacksmithing
	case 165:
		return proto.Profession_Leatherworking
	case 171:
		return proto.Profession_Alchemy
	case 182:
		return proto.Profession_Herbalism
	case 186:
		return proto.Profession_Mining
	case 197:
		return proto.Profession_Tailoring
	case 202:
		return proto.Profession_Engineering
	case 333:
		return proto.Profession_Enchanting
	case 393:
		return proto.Profession_Skinning
	case 755:
		return proto.Profession_Jewelcrafting
	case 773:
		return proto.Profession_Inscription
	default:
		return proto.Profession_ProfessionUnknown
	}
}

func commonSetName(names []string, setID int32) string {
	if len(names) == 0 {
		return fmt.Sprintf("套装%d", setID)
	}
	prefix := names[0]
	for _, n := range names[1:] {
		prefix = utf8CommonPrefix(prefix, n)
		if prefix == "" {
			break
		}
	}
	prefix = strings.TrimRight(prefix, "的之 ·-")
	if utf8.RuneCountInString(prefix) < 2 {
		return fmt.Sprintf("套装%d", setID)
	}
	return prefix
}

func utf8CommonPrefix(a, b string) string {
	ar, br := []rune(a), []rune(b)
	n := len(ar)
	if len(br) < n {
		n = len(br)
	}
	i := 0
	for i < n && ar[i] == br[i] {
		i++
	}
	return string(ar[:i])
}
