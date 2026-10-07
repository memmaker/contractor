package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/textiles"
	"testing"
)

// The chapter 2 and 3 maps carry every location, object and actor the outline names.
func TestChapterMapsContent(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	for _, item := range []string{"mansion_key", "faust_safe_key", "cryo_keycard", "cryo_fuse"} {
		if g.NewItemFromString(item) == nil {
			t.Errorf("item %s undefined", item)
		}
	}
	if g.NewActorFromName("ebi_trooper", "mansion") == nil {
		t.Error("ebi_trooper undefined")
	}
	for _, c := range []struct {
		mapName          string
		locations        []string
		objects          []string
		actors           map[string]int
		transition, dest string
	}{
		{"mansion",
			[]string{"taxi_stand", "taxi_driver", "mansion_gate", "mansion_hall", "faust_study", "mansion_kitchen", "service_door", "cellar_stairs"},
			[]string{"Taxi", "mansion_front_door", "mansion_service_door", "faust_terminal", "faust_safe", "cellar_door"},
			map[string]int{"logan_faust": 1, "mira_sykes": 1, "mansion_guard": 3},
			"cellar_stairs", "cryo_lab"},
		{"cryo_lab",
			[]string{"lab_entrance", "pod_room", "lab_console", "generator_room", "pod_1", "pod_2", "pod_3", "pod_4"},
			[]string{"ark_terminal", "cryo_pod_1", "cryo_pod_2", "cryo_pod_3", "cryo_pod_4", "generator_door", "spare_parts"},
			map[string]int{},
			"lab_entrance", "mansion"},
	} {
		gMap := g.ensureMapIsLoaded(c.mapName)
		if gMap == nil {
			t.Fatalf("%s did not load", c.mapName)
		}
		for _, loc := range c.locations {
			if _, ok := gMap.TryGetNamedLocation(loc); !ok {
				t.Errorf("%s: location %s missing", c.mapName, loc)
			}
		}
		for _, name := range c.objects {
			found := false
			for _, obj := range gMap.Objects() {
				found = found || obj.GetInternalName() == name
			}
			if !found {
				t.Errorf("%s: object %s missing", c.mapName, name)
			}
		}
		for name, want := range c.actors {
			got := 0
			for _, actor := range gMap.Actors() {
				if actor.GetInternalName() == name {
					got++
				}
			}
			if got != want {
				t.Errorf("%s: %d %s, want %d", c.mapName, got, name, want)
			}
		}
		pos := gMap.GetNamedLocation(c.transition)
		if tr, ok := gMap.GetTransitionAt(pos); !ok || tr.TargetMap != c.dest {
			t.Errorf("%s: transition at %s does not lead to %s", c.mapName, c.transition, c.dest)
		}
	}
}
