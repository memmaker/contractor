package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
	"testing"
	"time"
)

// After Harker dies (Drake not freed) the EBI medic takes the clinic and the
// cleaner walks from his room in the west to the hospital, scrubs every stain and walks home.
func TestClinicTakeoverAndCleanup(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom", SimulateAllLoadedMaps: true})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	g.mapContainsPlayer = false
	g.ui = testUI{}

	hospital := g.ensureMapIsLoaded("zone_residential_south")
	var stains []geometry.Point
	for y := 0; y < hospital.GetHeight() && len(stains) < 5; y++ {
		for x := 0; x < hospital.GetWidth() && len(stains) < 5; x++ {
			if p := (geometry.Point{X: x, Y: y}); hospital.IsZoneAt(p, "public_hospital") && hospital.IsTileWalkable(p) {
				hospital.StainTile(p, color.RGBA{R: 200, A: 255}, color.RGBA{R: 100, A: 255})
				stains = append(stains, p)
			}
		}
	}
	if len(stains) == 0 {
		t.Fatal("no hospital floor found")
	}
	g.gameFlags.Set("Killed(daniel_harker)", 1)
	g.Scripts.RunScriptByName("../data_atom", "clinic_takeover", g.GetScriptFuncs())
	g.Scripts.CheckAndRunFrames()
	g.gameTime = g.gameTime.AddDuration(49 * time.Hour)

	cleanerAtHospital, cleaned := false, false
	for turn := 0; turn < 3000 && !(cleaned && !g.Scripts.IsScriptRunning("clinic_cleanup")); turn++ {
		for _, name := range []string{"ebi_medic", "cleaner"} {
			if a := g.actorNamed(name); a != nil {
				g.ExecuteOnMap(a.currentMapName, func() {
					a.AddTimeEnergy(100)
					for tu := g.TryAIAction(a); tu > 0; tu = g.TryAIAction(a) {
						a.SpendTimeEnergy(tu)
					}
				})
			}
		}
		g.advanceTimeAndTurn(10 * time.Second)
		for _, action := range g.afterAnimationActions {
			action()
		}
		g.afterAnimationActions = nil
		if c := g.actorNamed("cleaner"); c != nil && c.currentMapName == "zone_residential_south" {
			cleanerAtHospital = true
		}
		cleaned = cleaned || g.gameFlags.HasFlag("clinic_cleaned")
	}
	if !g.gameFlags.HasFlag("clinic(ebi)") || !cleanerAtHospital || !cleaned {
		t.Fatalf("ebi=%v cleanerAtHospital=%v cleaned=%v", g.gameFlags.HasFlag("clinic(ebi)"), cleanerAtHospital, cleaned)
	}
	for _, p := range stains {
		if hospital.IsStained(p) {
			t.Errorf("stain left at %v", p)
		}
	}
	if g.Scripts.IsScriptRunning("clinic_cleanup") {
		t.Error("cleanup never ended")
	}
	c := g.actorNamed("cleaner")
	if c.currentMapName != "zone_residential_west" || c.Position() != g.ensureMapIsLoaded("zone_residential_west").GetNamedLocation("cleaner_room") {
		t.Errorf("cleaner did not go home: %s %v", c.currentMapName, c.Position())
	}
}

func (g *GameState) actorNamed(name string) (found *Actor) {
	g.IterateAllActors(func(_ string, a *Actor) bool {
		if a.GetInternalName() == name {
			found = a
		}
		return found == nil
	})
	return found
}

// testUI swallows animations; any other UI call panics and shows up in the test.
type testUI struct{ foundation.GameUI }

func (testUI) AddAnimations([]foundation.Animation) {}
func (testUI) UpdateLogWindow()                     {}
func (testUI) GetAnimBackgroundColor(geometry.Point, string, int, func()) foundation.Animation {
	return nil
}
