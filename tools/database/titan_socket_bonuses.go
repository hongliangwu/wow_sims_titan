package database

import "github.com/wowsims/wotlk/sim/core/stats"

func titanSocketBonus(id int32) (stats.Stats, bool) {
	if id == 0 {
		return stats.Stats{}, true
	}
	s, ok := titanSocketBonusByEnchantID[id]
	return s, ok
}

func hit(n float64) stats.Stats {
	return stats.Stats{stats.MeleeHit: n, stats.SpellHit: n}
}
func crit(n float64) stats.Stats {
	return stats.Stats{stats.MeleeCrit: n, stats.SpellCrit: n}
}
func haste(n float64) stats.Stats {
	return stats.Stats{stats.MeleeHaste: n, stats.SpellHaste: n}
}
func ap(n float64) stats.Stats {
	return stats.Stats{stats.AttackPower: n, stats.RangedAttackPower: n}
}

// SpellItemEnchantment IDs used as ItemSparse.Socket_match_enchantment_ID.
var titanSocketBonusByEnchantID = map[int32]stats.Stats{
	1338: {stats.FireResistance: 4},
	1340: {stats.FireResistance: 6},
	1342: {stats.FireResistance: 8},
	2787: crit(8),
	2843: crit(8),
	2844: hit(8),
	2854: {stats.MP5: 3},
	2864: crit(4),
	2865: {stats.MP5: 2},
	2868: {stats.Stamina: 6},
	2869: {stats.Intellect: 4},
	2871: {stats.Dodge: 4},
	2873: hit(4),
	2874: crit(4),
	2877: {stats.Agility: 4},
	2878: {stats.Resilience: 4},
	2882: {stats.Stamina: 6},
	2890: {stats.Spirit: 4},
	2892: {stats.Strength: 4},
	2908: hit(4),
	2927: {stats.Strength: 4},
	2932: {stats.Defense: 4},
	2936: ap(8),
	2952: crit(4),
	3094: {stats.Expertise: 4},
	3263: crit(4),
	3267: haste(4),
	3301: crit(4),
	3302: {stats.Defense: 8},
	3303: haste(8),
	3304: {stats.Dodge: 8},
	3305: {stats.Stamina: 12},
	3306: {stats.MP5: 2},
	3307: {stats.Stamina: 9},
	3308: haste(4),
	3309: haste(6),
	3310: {stats.Intellect: 6},
	3311: {stats.Spirit: 6},
	3312: {stats.Strength: 8},
	3313: {stats.Agility: 8},
	3314: crit(8),
	3316: crit(6),
	3351: hit(6),
	3352: {stats.Spirit: 8},
	3353: {stats.Intellect: 8},
	3354: {stats.Stamina: 12},
	3355: {stats.Agility: 6},
	3356: ap(12),
	3357: {stats.Strength: 6},
	3358: {stats.Dodge: 6},
	3359: {stats.Parry: 4},
	3360: {stats.Parry: 8},
	3361: {stats.Block: 6},
	3362: {stats.Expertise: 6},
	3596: {stats.SpellPower: 5},
	3600: {stats.Resilience: 6},
	3602: {stats.SpellPower: 7},
	3752: {stats.SpellPower: 5},
	3753: {stats.SpellPower: 9},
	3765: {stats.ArmorPenetration: 4},
	3766: {stats.Stamina: 12},
	3778: {stats.Expertise: 8},
	3871: {stats.Parry: 6},
	3877: ap(16),
}
