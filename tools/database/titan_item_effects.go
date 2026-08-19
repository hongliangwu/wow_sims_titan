package database

import (
	"path/filepath"
	"sort"
)

const (
	itemEffectTriggerUse   = 0
	itemEffectTriggerEquip = 1
	itemEffectTriggerProc  = 2
)

type titanItemEffectRow struct {
	itemID   int32
	trigger  int32
	spellID  int32
	cooldown int32
	slot     int32
}

// ExportTitanItemEffects builds Use/Equip tooltip lines for imported Titan items.
func ExportTitanItemEffects(dbDir string, itemIDs []int32) (map[int32][]string, error) {
	wanted := map[int32]struct{}{}
	for _, id := range itemIDs {
		wanted[id] = struct{}{}
	}
	descs, err := loadSpellDescriptions(filepath.Join(dbDir, "Spell.csv"))
	if err != nil {
		return nil, err
	}
	points, err := loadSpellEffectPoints(filepath.Join(dbDir, "SpellEffect.csv"))
	if err != nil {
		return nil, err
	}
	durations := loadSpellDurations(dbDir)
	var rows []titanItemEffectRow
	err = loadCSVFile(filepath.Join(dbDir, "ItemEffect.csv"), func(rec []string, col map[string]int) {
		itemID := csvInt(rec, col, "ParentItemID")
		if _, ok := wanted[itemID]; !ok {
			return
		}
		trigger := csvInt(rec, col, "TriggerType")
		if trigger != itemEffectTriggerUse && trigger != itemEffectTriggerEquip && trigger != itemEffectTriggerProc {
			return
		}
		spellID := csvInt(rec, col, "SpellID")
		if spellID == 0 {
			return
		}
		cd := csvInt(rec, col, "CoolDownMSec")
		if cd < 0 {
			cd = 0
		}
		rows = append(rows, titanItemEffectRow{
			itemID:   itemID,
			trigger:  trigger,
			spellID:  spellID,
			cooldown: cd,
			slot:     csvInt(rec, col, "LegacySlotIndex"),
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].itemID != rows[j].itemID {
			return rows[i].itemID < rows[j].itemID
		}
		return rows[i].slot < rows[j].slot
	})

	out := map[int32][]string{}
	for _, row := range rows {
		raw := descs[row.spellID]
		text := formatSpellDescription(raw, row.spellID, points, durations)
		if text == "" {
			continue
		}
		line := itemEffectPrefix(row.trigger) + text
		if suffix := formatCooldownSuffix(row.cooldown); suffix != "" {
			line += suffix
		}
		out[row.itemID] = append(out[row.itemID], line)
	}
	return out, nil
}

func itemEffectPrefix(trigger int32) string {
	switch trigger {
	case itemEffectTriggerUse:
		return "使用："
	default:
		return "装备："
	}
}

func formatCooldownSuffix(ms int32) string {
	if ms <= 0 {
		return ""
	}
	return "（" + formatDurationMS(ms) + "冷却）"
}
