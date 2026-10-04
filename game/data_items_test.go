package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"os"
	"testing"
)

// Weapons without weapon_skill_used used to fall back to the first skill in rules.rec,
// so map-placed guns rolled (and showed) Melee Combat.
func TestWeaponsGetSkillMatchingAttackMode(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()

	var weapons []*Weapon
	records, _ := recfile.ReadAndClose(fxtools.MustOpen("../data_atom/definitions/weapons.rec"))
	for _, rec := range records {
		if name, ok := rec.FindFieldIgnoreCase("name"); ok {
			if w, isWeapon := g.NewItemFromString(name.Value).(*Weapon); isWeapon {
				weapons = append(weapons, w)
			}
		}
	}
	entries, _ := os.ReadDir("../data_atom/maps")
	for _, e := range entries {
		if m := g.ensureMapIsLoaded(e.Name()); e.IsDir() && m != nil {
			for _, item := range m.Items() {
				if w, isWeapon := item.(*Weapon); isWeapon {
					weapons = append(weapons, w)
				}
			}
		}
	}

	if len(weapons) == 0 {
		t.Fatal("no weapons loaded")
	}
	for _, w := range weapons {
		mode := w.GetCurrentAttackMode().Mode
		skill := w.GetSkillUsed()
		if mode.IsMelee() && mode != TargetingModeThrow && !skill.IsMeleeAttackSkill() {
			t.Errorf("%s: melee weapon uses %s", w.GetInternalName(), skill)
		}
		if !mode.IsMelee() && !skill.IsRangedAttackSkill() {
			t.Errorf("%s: ranged weapon uses %s", w.GetInternalName(), skill)
		}
	}
}
