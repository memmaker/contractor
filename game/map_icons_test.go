package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/textiles"
	"os"
	"testing"
)

// Every actor, item, container (in every state) and visible object on every map has an icon (a zero icon renders as a grey '?').
func TestMapIcons(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	dirs, _ := os.ReadDir("../data_atom/maps")
	for _, dir := range dirs {
		gMap := g.ensureMapIsLoaded(dir.Name())
		if gMap == nil {
			t.Errorf("%s did not load", dir.Name())
			continue
		}
		for _, a := range append(gMap.Actors(), gMap.DownedActors()...) {
			if a.GetIcon().Char == 0 {
				t.Errorf("%s: actor %s has no icon", dir.Name(), a.GetInternalName())
			}
		}
		for _, i := range gMap.Items() {
			if i.GetIcon().Char == 0 {
				t.Errorf("%s: item %s has no icon", dir.Name(), i.GetInternalName())
			}
		}
		for _, o := range gMap.Objects() {
			if c, isContainer := o.(*Container); isContainer { // a container's icon changes as you open and empty it
				for _, state := range []foundation.ObjectCategory{foundation.ObjectUnknownContainer, foundation.ObjectKnownContainer, foundation.ObjectKnownEmptyContainer} {
					if c.iconForObject(state.LowerString()).Char == 0 {
						t.Errorf("%s: container %s has no icon when %s", dir.Name(), c.GetInternalName(), state)
					}
				}
				continue
			}
			if !o.IsHidden() && o.GetIcon().Char == 0 {
				t.Errorf("%s: object %s (%s) has no icon", dir.Name(), o.GetInternalName(), o.GetCategory())
			}
		}
	}
}
