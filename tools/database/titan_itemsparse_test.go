package database

import (
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func TestLoadTitanItemSparseNorthrendHelm(t *testing.T) {
	items, err := LoadTitanItemSparse(
		"../../assets/database/dbfilesclient/ItemSparse.csv",
		"../../assets/database/dbfilesclient/Item.csv",
		"../../assets/db_inputs/titan_icon_names.csv",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 100 {
		t.Fatalf("expected hundreds of Titan items, got %d", len(items))
	}

	var helm *proto.UIItem
	for _, it := range items {
		if it.Id == 257631 {
			helm = it
			break
		}
	}
	if helm == nil {
		t.Fatal("missing 257631 北境战盔")
	}
	if helm.Name != "北境战盔" {
		t.Errorf("name=%q", helm.Name)
	}
	if helm.Icon != "inv_helmet_06" {
		t.Errorf("icon=%q", helm.Icon)
	}
	if helm.Type != proto.ItemType_ItemTypeHead || helm.ArmorType != proto.ArmorType_ArmorTypePlate {
		t.Errorf("slot/type = %v %v", helm.Type, helm.ArmorType)
	}
	if helm.Ilvl != 213 || helm.Quality != proto.ItemQuality_ItemQualityEpic {
		t.Errorf("ilvl/quality = %d %v", helm.Ilvl, helm.Quality)
	}
	if helm.Phase != 1 {
		t.Errorf("phase=%d want 1 (ilvl 213 → 时光 P1)", helm.Phase)
	}
	if helm.SetName != "北境" {
		t.Errorf("set=%q", helm.SetName)
	}
	if len(helm.ClassAllowlist) != 1 || helm.ClassAllowlist[0] != proto.Class_ClassDeathknight {
		t.Errorf("class allowlist=%v", helm.ClassAllowlist)
	}
	if len(helm.GemSockets) != 2 || helm.GemSockets[0] != proto.GemColor_GemColorMeta || helm.GemSockets[1] != proto.GemColor_GemColorYellow {
		t.Errorf("sockets=%v", helm.GemSockets)
	}

	st := stats.FromFloatArray(helm.Stats)
	if st[stats.Strength] != 85 || st[stats.Stamina] != 98 || st[stats.MeleeHit] != 86 || st[stats.MeleeCrit] != 32 || st[stats.Armor] != 1867 {
		t.Errorf("stats=%v", st)
	}
	bonus := stats.FromFloatArray(helm.SocketBonus)
	if bonus[stats.Strength] != 8 {
		t.Errorf("socket bonus=%v", bonus)
	}
}

func TestTitanPhaseFromIlvl(t *testing.T) {
	cases := []struct {
		ilvl, want int32
	}{
		{166, 1},
		{200, 1},
		{213, 1},
		{219, 2},
		{226, 3},
		{232, 4},
		{238, 5},
		{245, 6},
		{251, 7},
		{258, 8},
		{264, 9},
		{272, 10},
		{280, 11},
		{296, 11},
	}
	for _, c := range cases {
		got := titanPhaseFromIlvl(c.ilvl)
		if got != c.want {
			t.Errorf("ilvl=%d → phase %d, want %d", c.ilvl, got, c.want)
		}
	}
}

func TestLoadTitanItemSparseClassicRaidItem(t *testing.T) {
	items, err := LoadTitanItemSparse(
		"../../assets/database/dbfilesclient/ItemSparse.csv",
		"../../assets/database/dbfilesclient/Item.csv",
		"../../assets/db_inputs/titan_icon_names.csv",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 3000 {
		t.Fatalf("expected thousands of Titan items including original IDs, got %d", len(items))
	}

	var fang *proto.UIItem
	original := 0
	for _, it := range items {
		if it.Id < 200000 {
			original++
		}
		if it.Id == 17070 {
			fang = it
		}
	}
	if original < 1000 {
		t.Fatalf("expected original-ID Titan loot, got %d", original)
	}
	if fang == nil {
		t.Fatal("missing 17070 秘法之牙 (MC weapon reused under original ID)")
	}
	if fang.Name != "秘法之牙" {
		t.Errorf("name=%q", fang.Name)
	}
	if fang.Ilvl != 213 || fang.Phase != 1 {
		t.Errorf("ilvl/phase = %d %d, want 213 / P1", fang.Ilvl, fang.Phase)
	}
	if fang.Type != proto.ItemType_ItemTypeWeapon {
		t.Errorf("type=%v", fang.Type)
	}
}

func TestFormatSpellDescription(t *testing.T) {
	points := map[int32][]int32{
		1262348: {6},
		1262346: {1000},
		1262095: {50},
	}
	got := formatSpellDescription("你的十字军打击所造成的伤害提高$s1%。", 1262348, points, nil)
	if got != "你的十字军打击所造成的伤害提高6%。" {
		t.Errorf("2pc: %q", got)
	}
	got = formatSpellDescription("你的神圣风暴技能的冷却时间缩短$/1000;s1秒。", 1262346, points, nil)
	if got != "你的神圣风暴技能的冷却时间缩短1秒。" {
		t.Errorf("4pc: %q", got)
	}
	got = formatSpellDescription("你的寒冬号角在使用时额外产生$/10;1262095s1点符文能量。", 1262093, points, nil)
	if got != "你的寒冬号角在使用时额外产生5点符文能量。" {
		t.Errorf("other-spell: %q", got)
	}
}

func TestTitanTrinketEffects(t *testing.T) {
	items, err := LoadTitanItemSparse(
		"../../assets/database/dbfilesclient/ItemSparse.csv",
		"../../assets/database/dbfilesclient/Item.csv",
		"../../assets/db_inputs/titan_icon_names.csv",
	)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]int32, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.Id)
	}
	effects, err := ExportTitanItemEffects("../../assets/database/dbfilesclient", ids)
	if err != nil {
		t.Fatal(err)
	}
	lines := effects[47080]
	if len(lines) == 0 {
		t.Fatal("missing effects for 47080 萨崔娜的抗阻甲虫")
	}
	if !strings.Contains(lines[0], "使用：") || !strings.Contains(lines[0], "生命值") {
		t.Errorf("normal use effect=%q", lines[0])
	}
	heroic := effects[47088]
	if len(heroic) == 0 {
		t.Fatal("missing effects for 47088 heroic scarab")
	}
	if lines[0] == heroic[0] {
		t.Fatalf("normal and heroic use effects should differ: %q", lines[0])
	}
	verdict := effects[47115]
	if len(verdict) == 0 || !strings.Contains(verdict[0], "装备：") || !strings.Contains(verdict[0], "力量或敏捷") {
		t.Errorf("death's verdict=%v", verdict)
	}
}

func TestTitanLawbringerSet(t *testing.T) {
	items, err := LoadTitanItemSparse(
		"../../assets/database/dbfilesclient/ItemSparse.csv",
		"../../assets/database/dbfilesclient/Item.csv",
		"../../assets/db_inputs/titan_icon_names.csv",
	)
	if err != nil {
		t.Fatal(err)
	}
	var bracer *proto.UIItem
	for _, it := range items {
		if it.Id == 257640 {
			bracer = it
			break
		}
	}
	if bracer == nil {
		t.Fatal("missing 257640 秩序之源腕铠")
	}
	if bracer.SetName != "秩序之源战装" {
		t.Errorf("set=%q", bracer.SetName)
	}

	sets, err := ExportTitanSets("../../assets/database/dbfilesclient", items)
	if err != nil {
		t.Fatal(err)
	}
	var set *TitanItemSet
	for i := range sets {
		if sets[i].ID == 2003 {
			set = &sets[i]
			break
		}
	}
	if set == nil {
		t.Fatal("missing set 2003")
	}
	if len(set.Items) < 8 {
		t.Errorf("pieces=%d", len(set.Items))
	}
	if len(set.Bonuses) != 2 {
		t.Fatalf("bonuses=%v", set.Bonuses)
	}
	if set.Bonuses[0].Threshold != 2 || !strings.Contains(set.Bonuses[0].Description, "十字军打击") {
		t.Errorf("2pc=%+v", set.Bonuses[0])
	}
	if set.Bonuses[1].Threshold != 4 || !strings.Contains(set.Bonuses[1].Description, "神圣风暴") {
		t.Errorf("4pc=%+v", set.Bonuses[1])
	}
}

func TestTuralyonBattlegearSetMatchesTriumphantVariant(t *testing.T) {
	items, err := LoadTitanItemSparse(
		"../../assets/database/dbfilesclient/ItemSparse.csv",
		"../../assets/database/dbfilesclient/Item.csv",
		"../../assets/db_inputs/titan_icon_names.csv",
	)
	if err != nil {
		t.Fatal(err)
	}
	var triumphant, canonical *proto.UIItem
	for _, it := range items {
		if it.Id == 48614 {
			triumphant = it
		}
		if it.Id == 48917 {
			canonical = it
		}
	}
	if triumphant == nil {
		t.Fatal("missing 48614 图拉扬的凯旋头盔")
	}
	if canonical == nil {
		t.Fatal("missing 48917 图拉扬的头盔")
	}
	if triumphant.SetName != "图拉扬的战甲" || canonical.SetName != triumphant.SetName {
		t.Fatalf("set names triumphant=%q canonical=%q", triumphant.SetName, canonical.SetName)
	}
	if triumphant.Type != proto.ItemType_ItemTypeHead || canonical.Type != proto.ItemType_ItemTypeHead {
		t.Fatalf("types triumphant=%v canonical=%v", triumphant.Type, canonical.Type)
	}

	sets, err := ExportTitanSets("../../assets/database/dbfilesclient", items)
	if err != nil {
		t.Fatal(err)
	}
	var set *TitanItemSet
	for i := range sets {
		if sets[i].ID == 877 {
			set = &sets[i]
			break
		}
	}
	if set == nil {
		t.Fatal("missing set 877 图拉扬的战甲")
	}
	var helm *TitanSetPiece
	for i := range set.Items {
		if set.Items[i].ID == 48917 {
			helm = &set.Items[i]
			break
		}
	}
	if helm == nil {
		t.Fatal("set 877 missing canonical helm 48917")
	}
	if helm.Type != int32(proto.ItemType_ItemTypeHead) {
		t.Fatalf("canonical helm type=%d, want head so tooltip can match 凯旋/征服 variants by slot", helm.Type)
	}
}

func TestLoadTitanDisplayNamesMetaGem(t *testing.T) {
	names, err := LoadTitanDisplayNames("../../assets/database/dbfilesclient/ItemSparse.csv")
	if err != nil {
		t.Fatal(err)
	}
	got := names[41398]
	if got != "残酷之大地侵攻钻石" {
		t.Fatalf("41398 name=%q", got)
	}
}

func TestDropTitanLowerIlvlDuplicates_DeathsVerdict(t *testing.T) {
	db := NewWowDatabase()
	db.Items[47115] = &proto.UIItem{
		Id: 47115, Name: "死亡的裁决", Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 238, Phase: 4, FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
	}
	db.Items[47131] = &proto.UIItem{
		Id: 47131, Name: "死亡的裁决", Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 232, Phase: 4, FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
	}
	db.Items[47303] = &proto.UIItem{
		Id: 47303, Name: "死亡的选择", Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 238, Phase: 4, FactionRestriction: proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY,
	}
	db.Items[1] = &proto.UIItem{
		Id: 1, Name: "死亡的裁决", Type: proto.ItemType_ItemTypeTrinket,
		Ilvl: 200, Phase: 1, FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY,
	}

	got := DropTitanLowerIlvlDuplicates(db)
	if got != 1 {
		t.Fatalf("dropped %d, want 1", got)
	}
	if _, ok := db.Items[47115]; !ok {
		t.Fatal("missing 238 死亡的裁决")
	}
	if _, ok := db.Items[47131]; ok {
		t.Fatal("232 死亡的裁决 should be removed (Titan Time only has 238)")
	}
	if _, ok := db.Items[47303]; !ok {
		t.Fatal("missing Horde 238 counterpart")
	}
	if _, ok := db.Items[1]; !ok {
		t.Fatal("different-phase item should be kept")
	}
}
