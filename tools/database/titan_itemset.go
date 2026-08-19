package database

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/wotlk/sim/core/proto"
)

// TitanItemSet is the JSON payload used by the local item tooltip.
type TitanItemSet struct {
	ID      int32            `json:"id"`
	Name    string           `json:"name"`
	Items   []TitanSetPiece  `json:"items"`
	Bonuses []TitanSetBonus  `json:"bonuses"`
}

type TitanSetPiece struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type TitanSetBonus struct {
	Threshold   int32  `json:"threshold"`
	Description string `json:"description"`
}

type itemSetCatalog struct {
	names   map[int32]string
	itemIDs map[int32][]int32
	bonuses map[int32][]TitanSetBonus
}

var (
	reSlashOtherSpell = regexp.MustCompile(`\$/(\d+);(\d+)s(\d+)`)
	reSlashSameSpell  = regexp.MustCompile(`\$/(\d+);s(\d+)`)
	reSPoint          = regexp.MustCompile(`\$s(\d+)`)
)

func loadItemSetCatalog(dir string) (*itemSetCatalog, error) {
	cat := &itemSetCatalog{
		names:   map[int32]string{},
		itemIDs: map[int32][]int32{},
		bonuses: map[int32][]TitanSetBonus{},
	}
	setPath := filepath.Join(dir, "ItemSet.csv")
	if err := loadItemSetNames(setPath, cat); err != nil {
		return nil, err
	}
	spellPath := filepath.Join(dir, "Spell.csv")
	effectPath := filepath.Join(dir, "SpellEffect.csv")
	bonusPath := filepath.Join(dir, "ItemSetSpell.csv")
	descs, err := loadSpellDescriptions(spellPath)
	if err != nil {
		return cat, nil
	}
	points, err := loadSpellEffectPoints(effectPath)
	if err != nil {
		return cat, nil
	}
	if err := loadItemSetSpells(bonusPath, cat, descs, points); err != nil {
		return cat, nil
	}
	return cat, nil
}

func loadItemSetNames(path string, cat *itemSetCatalog) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("read ItemSet header: %w", err)
	}
	col := csvHeader(header)
	if _, ok := col["ID"]; !ok {
		return fmt.Errorf("ItemSet.csv missing ID")
	}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		id := csvInt(rec, col, "ID")
		if id == 0 {
			continue
		}
		name := csvString(rec, col, "Name_lang")
		if name != "" {
			cat.names[id] = name
		}
		var pieces []int32
		for i := 0; i < 17; i++ {
			itemID := csvInt(rec, col, fmt.Sprintf("ItemID_%d", i))
			if itemID != 0 {
				pieces = append(pieces, itemID)
			}
		}
		if len(pieces) > 0 {
			cat.itemIDs[id] = pieces
		}
	}
	return nil
}

func loadItemSetSpells(path string, cat *itemSetCatalog, descs map[int32]string, points map[int32][]int32) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("read ItemSetSpell header: %w", err)
	}
	col := csvHeader(header)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		setID := csvInt(rec, col, "ItemSetID")
		spellID := csvInt(rec, col, "SpellID")
		threshold := csvInt(rec, col, "Threshold")
		if setID == 0 || spellID == 0 || threshold == 0 {
			continue
		}
		raw := descs[spellID]
		text := formatSpellDescription(raw, spellID, points)
		if text == "" {
			continue
		}
		cat.bonuses[setID] = append(cat.bonuses[setID], TitanSetBonus{
			Threshold:   threshold,
			Description: text,
		})
	}
	for setID, list := range cat.bonuses {
		sort.Slice(list, func(i, j int) bool { return list[i].Threshold < list[j].Threshold })
		cat.bonuses[setID] = list
	}
	return nil
}

func loadSpellDescriptions(path string) (map[int32]string, error) {
	out := map[int32]string{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	col := csvHeader(header)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		id := csvInt(rec, col, "ID")
		desc := csvString(rec, col, "Description_lang")
		if desc == "" {
			desc = csvString(rec, col, "AuraDescription_lang")
		}
		if id != 0 && desc != "" {
			out[id] = desc
		}
	}
	return out, nil
}

func loadSpellEffectPoints(path string) (map[int32][]int32, error) {
	// points[spellID][effectIndex] = displayed points
	raw := map[int32]map[int32]int32{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	col := csvHeader(header)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		spellID := csvInt(rec, col, "SpellID")
		idx := csvInt(rec, col, "EffectIndex")
		bp := csvInt(rec, col, "EffectBasePoints")
		die := csvInt(rec, col, "EffectDieSides")
		if spellID == 0 {
			continue
		}
		if raw[spellID] == nil {
			raw[spellID] = map[int32]int32{}
		}
		raw[spellID][idx] = spellEffectPoints(bp, die)
	}
	out := map[int32][]int32{}
	for spellID, byIdx := range raw {
		max := int32(-1)
		for idx := range byIdx {
			if idx > max {
				max = idx
			}
		}
		arr := make([]int32, max+1)
		for idx, v := range byIdx {
			arr[idx] = v
		}
		out[spellID] = arr
	}
	return out, nil
}

func spellEffectPoints(basePoints, dieSides int32) int32 {
	v := basePoints
	if dieSides > 0 {
		v += dieSides
	}
	if v < 0 {
		return -v
	}
	return v
}

func formatSpellDescription(desc string, spellID int32, points map[int32][]int32) string {
	if desc == "" {
		return ""
	}
	desc = reSlashOtherSpell.ReplaceAllStringFunc(desc, func(m string) string {
		sub := reSlashOtherSpell.FindStringSubmatch(m)
		div, _ := strconv.Atoi(sub[1])
		otherID, _ := strconv.Atoi(sub[2])
		idx, _ := strconv.Atoi(sub[3])
		v := spellPointValue(points, int32(otherID), idx)
		if div > 0 {
			return strconv.Itoa(int(v) / div)
		}
		return m
	})
	desc = reSlashSameSpell.ReplaceAllStringFunc(desc, func(m string) string {
		sub := reSlashSameSpell.FindStringSubmatch(m)
		div, _ := strconv.Atoi(sub[1])
		idx, _ := strconv.Atoi(sub[2])
		v := spellPointValue(points, spellID, idx)
		if div > 0 {
			return strconv.Itoa(int(v) / div)
		}
		return m
	})
	desc = reSPoint.ReplaceAllStringFunc(desc, func(m string) string {
		sub := reSPoint.FindStringSubmatch(m)
		idx, _ := strconv.Atoi(sub[1])
		v := spellPointValue(points, spellID, idx)
		return strconv.Itoa(int(v))
	})
	return strings.TrimSpace(desc)
}

func spellPointValue(points map[int32][]int32, spellID int32, oneBased int) int32 {
	arr := points[spellID]
	idx := oneBased - 1
	if idx < 0 || idx >= len(arr) {
		return 0
	}
	return arr[idx]
}

func csvHeader(header []string) map[string]int {
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	return col
}

func csvString(rec []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i < 0 || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

func ExportTitanSets(dbDir string, items []*proto.UIItem) ([]TitanItemSet, error) {
	cat, err := loadItemSetCatalog(dbDir)
	if err != nil {
		return nil, err
	}
	return buildTitanItemSets(cat, items), nil
}

func buildTitanItemSets(cat *itemSetCatalog, items []*proto.UIItem) []TitanItemSet {
	byID := map[int32]string{}
	for _, it := range items {
		byID[it.Id] = it.Name
	}
	used := map[int32]struct{}{}
	for setID, pieceIDs := range cat.itemIDs {
		for _, id := range pieceIDs {
			if _, ok := byID[id]; ok {
				used[setID] = struct{}{}
				break
			}
		}
	}
	out := make([]TitanItemSet, 0, len(used))
	for setID := range used {
		name := cat.names[setID]
		if name == "" {
			continue
		}
		seen := map[int32]struct{}{}
		var pieces []TitanSetPiece
		add := func(id int32) {
			if _, ok := seen[id]; ok {
				return
			}
			n := byID[id]
			if n == "" {
				return
			}
			seen[id] = struct{}{}
			pieces = append(pieces, TitanSetPiece{ID: id, Name: n})
		}
		for _, id := range cat.itemIDs[setID] {
			add(id)
		}
		sort.Slice(pieces, func(i, j int) bool { return pieces[i].ID < pieces[j].ID })
		if len(pieces) == 0 {
			continue
		}
		out = append(out, TitanItemSet{
			ID:      setID,
			Name:    name,
			Items:   pieces,
			Bonuses: append([]TitanSetBonus(nil), cat.bonuses[setID]...),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// TitanSetsPayload is written to titan_sets.json for the local tooltip and item picker.
type TitanSetsPayload struct {
	ItemIDs []int32        `json:"itemIds"`
	Sets    []TitanItemSet `json:"sets"`
}

func WriteTitanSetsJSON(path string, sets []TitanItemSet, itemIDs []int32) error {
	ids := append([]int32(nil), itemIDs...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	data, err := json.MarshalIndent(TitanSetsPayload{ItemIDs: ids, Sets: sets}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
