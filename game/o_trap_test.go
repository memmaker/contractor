package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/recfile"
	"testing"
)

func TestTrapWalkOverTriggers(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	trap := g.NewTrap(recfile.Record{{Name: "Name", Value: "frag_mine"}, {Name: "ZapEffect", Value: "explode"}}, nil)
	if !trap.IsWalkable(nil) {
		t.Fatal("mine must be walkable")
	}
	trap.OnWalkOver(nil)
	if trap.state != TrapTriggered {
		t.Fatalf("expected triggered, got %s", trap.state)
	}
}
