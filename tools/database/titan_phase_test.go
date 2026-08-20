package database

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/tools"
)

func TestTitanPhaseFromSources(t *testing.T) {
	drop := func(zone, npc int32) []*proto.UIItemSource {
		return []*proto.UIItemSource{{
			Source: &proto.UIItemSource_Drop{
				Drop: &proto.DropSource{ZoneId: zone, NpcId: npc},
			},
		}}
	}

	cases := []struct {
		name    string
		sources []*proto.UIItemSource
		want    int32
	}{
		{"ZG Zin'rokh", drop(1977, 14834), 4},
		{"ZA Jin'rokh", drop(3805, 23863), 5},
		{"MC", drop(2717, 0), 1},
		{"BWL", drop(2677, 11583), 8},
		{"ToC", drop(4722, 0), 4},
		{"Ulduar", drop(4273, 0), 6},
		{"ICC", drop(4812, 0), 10},
		{"BT", drop(3959, 0), 11},
		{"Azuregos NPC", drop(16, 6109), 1},
		{"Ysondre NPC", drop(0, 14887), 9},
		{"crafted only", []*proto.UIItemSource{{
			Source: &proto.UIItemSource_Crafted{Crafted: &proto.CraftedSource{SpellId: 1}},
		}}, 0},
		{"empty", nil, 0},
	}
	for _, c := range cases {
		if got := TitanPhaseFromSources(c.sources); got != c.want {
			t.Errorf("%s: phase %d, want %d", c.name, got, c.want)
		}
	}
}

func TestApplyTitanDropPhaseOverridesIlvl(t *testing.T) {
	item := &proto.UIItem{Id: 19854, Ilvl: 238, Phase: titanPhaseFromIlvl(238)}
	if item.Phase != 5 {
		t.Fatalf("precondition: ilvl 238 maps to P%d, want P5", item.Phase)
	}
	atlas := NewWowDatabase()
	atlas.Items[19854] = &proto.UIItem{
		Id: 19854,
		Sources: []*proto.UIItemSource{{
			Source: &proto.UIItemSource_Drop{
				Drop: &proto.DropSource{ZoneId: 1977, NpcId: 14834},
			},
		}},
	}
	ApplyTitanDropPhase(item, nil, atlas)
	if item.Phase != 4 {
		t.Errorf("ZG item phase=%d, want 4", item.Phase)
	}
}

func TestApplyTitanDropPhaseKeepsIlvlWithoutSources(t *testing.T) {
	item := &proto.UIItem{Id: 1, Ilvl: 238, Phase: titanPhaseFromIlvl(238)}
	ApplyTitanDropPhase(item, nil, NewWowDatabase())
	if item.Phase != 5 {
		t.Errorf("phase=%d, want ilvl fallback 5", item.Phase)
	}
}

func TestTitanPhaseFromAtlasLootKnownRaids(t *testing.T) {
	db := ReadDatabaseFromJson(tools.ReadFile("../../assets/db_inputs/atlasloot_db.json"))
	cases := []struct {
		id   int32
		name string
		want int32
	}{
		{19854, "辛洛斯 / Zul'Gurub", 4},
		{33478, "金洛斯 / Zul'Aman", 5},
		{19364, "Ashkandi / BWL", 8},
	}
	for _, c := range cases {
		item, ok := db.Items[c.id]
		if !ok {
			t.Errorf("missing atlasloot item %d (%s)", c.id, c.name)
			continue
		}
		if got := TitanPhaseFromSources(item.Sources); got != c.want {
			t.Errorf("%s id=%d phase=%d want %d sources=%v", c.name, c.id, got, c.want, item.Sources)
		}
	}
}
