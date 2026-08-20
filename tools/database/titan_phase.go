package database

import (
	"github.com/wowsims/wotlk/sim/core/proto"
)

// Titan Time raid/dungeon phase map. ItemSparse has no phase field; ilvl bands
// collide (ZG and ZA are both 238). Drop zone / world-boss NPC is authoritative.
//
//	P1  Molten Core, Kazzak, Azuregos, all level-80 5-man, P1 embers (ilvl fallback)
//	P2  Tempest Keep, Serpentshrine, Outland world bosses
//	P3  Naxxramas, Obsidian Sanctum, Eye of Eternity
//	P4  Zul'Gurub, Trial of the Crusader, Vault of Archavon
//	P5  Sunwell Plateau, Zul'Aman, P5 embers (ilvl fallback)
//	P6  Ulduar
//	P7  Gruul's Lair, Karazhan, Magtheridon's Lair
//	P8  Blackwing Lair, Ruins of Ahn'Qiraj, Onyxia
//	P9  Ahn'Qiraj, four Dragons of Nightmare
//	P10 Icecrown Citadel, Ruby Sanctum
//	P11 Hyjal Summit, Black Temple
var titanZonePhase = map[int32]int32{
	// P1
	2717: 1, // Molten Core
	206:  1, // Utgarde Keep
	1196: 1, // Utgarde Pinnacle
	4265: 1, // The Nexus
	4228: 1, // The Oculus
	4415: 1, // The Violet Hold
	4196: 1, // Drak'Tharon Keep
	4416: 1, // Gundrak
	4277: 1, // Azjol-Nerub
	4494: 1, // Ahn'kahet: The Old Kingdom
	4100: 1, // The Culling of Stratholme
	4264: 1, // Halls of Stone
	4272: 1, // Halls of Lightning
	4809: 1, // The Forge of Souls
	4813: 1, // Pit of Saron
	4820: 1, // Halls of Reflection
	4723: 1, // Trial of the Champion

	// P2
	3845: 2, // Tempest Keep (The Eye)
	3607: 2, // Serpentshrine Cavern

	// P3
	3456: 3, // Naxxramas
	4493: 3, // The Obsidian Sanctum
	4500: 3, // The Eye of Eternity

	// P4
	1977: 4, // Zul'Gurub
	4722: 4, // Trial of the Crusader
	4603: 4, // Vault of Archavon

	// P5
	4075: 5, // Sunwell Plateau
	3805: 5, // Zul'Aman

	// P6
	4273: 6, // Ulduar

	// P7
	3923: 7, // Gruul's Lair
	3457: 7, // Karazhan
	3836: 7, // Magtheridon's Lair

	// P8
	2677: 8, // Blackwing Lair
	3429: 8, // Ruins of Ahn'Qiraj
	2159: 8, // Onyxia's Lair

	// P9
	3428: 9, // Ahn'Qiraj

	// P10
	4812: 10, // Icecrown Citadel
	4987: 10, // The Ruby Sanctum

	// P11
	3606: 11, // Hyjal Summit
	3959: 11, // Black Temple
}

// Outdoor world bosses are not in AtlasLoot dungeon tables; map by NPC.
var titanNpcPhase = map[int32]int32{
	6109:  1, // Azuregos
	12397: 1, // Lord Kazzak
	18728: 2, // Doom Lord Kazzak
	17711: 2, // Doomwalker
	14887: 9, // Ysondre
	14888: 9, // Lethon
	14889: 9, // Emeriss
	14890: 9, // Taerar
}

// TitanPhaseFromSources returns the Titan Time phase implied by drop sources,
// or 0 if none of the sources map to a raid/dungeon/world boss.
func TitanPhaseFromSources(sources []*proto.UIItemSource) int32 {
	var phase int32
	for _, src := range sources {
		drop := src.GetDrop()
		if drop == nil {
			continue
		}
		if p, ok := titanNpcPhase[drop.NpcId]; ok {
			phase = pickTitanDropPhase(phase, p)
			continue
		}
		if p, ok := titanZonePhase[drop.ZoneId]; ok {
			phase = pickTitanDropPhase(phase, p)
		}
	}
	return phase
}

// Prefer a later raid over a 5-man when an item is listed in both.
func pickTitanDropPhase(current, next int32) int32 {
	if current == 0 || next > current {
		return next
	}
	return current
}

// ApplyTitanDropPhase copies AtlasLoot / existing drop sources onto a Titan item
// and overrides the ilvl-estimated phase when a mapped raid or dungeon is found.
func ApplyTitanDropPhase(item *proto.UIItem, existing *proto.UIItem, atlasloot *WowDatabase) {
	if item == nil {
		return
	}
	var sources []*proto.UIItemSource
	if atlasloot != nil {
		if a, ok := atlasloot.Items[item.Id]; ok && len(a.Sources) > 0 {
			sources = a.Sources
		}
	}
	if len(sources) == 0 && existing != nil {
		sources = existing.Sources
	}
	if len(sources) == 0 {
		return
	}
	item.Sources = sources
	if p := TitanPhaseFromSources(sources); p != 0 {
		item.Phase = p
	}
}
