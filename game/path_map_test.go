package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/textiles"
	"os"
	"testing"
	"time"
)

// Cross-map pathfinding must terminate for every actor and every named target on the real maps.
// It used to spin forever backtracking over one-way transitions, freezing the game.
func TestFindPathTerminatesOnRealMaps(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	g.mapContainsPlayer = false
	entries, _ := os.ReadDir("../data_atom/maps")
	var names []string
	for _, e := range entries {
		if e.IsDir() && g.ensureMapIsLoaded(e.Name()) != nil {
			names = append(names, e.Name())
		}
	}

	done := make(chan int)
	go func() {
		calls := 0
		for _, from := range names {
			g.ExecuteOnMap(from, func() { // as the game does for actors on other maps
				for _, actor := range g.activeMaps[from].Actors() {
					for _, to := range names {
						for locName, pos := range g.activeMaps[to].GetNamedLocations() {
							g.pathfinder.FindPath(actor, from, MapPosition{MapName: to, LocationName: locName, Position: pos})
							calls++
						}
					}
				}
			})
		}
		done <- calls
	}()
	select {
	case calls := <-done:
		if calls == 0 {
			t.Fatal("no paths searched; maps or actors failed to load")
		}
	case <-time.After(60 * time.Second):
		t.Fatal("FindPath did not terminate")
	}
}
