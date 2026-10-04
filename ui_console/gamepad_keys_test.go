package ui_console

import (
	"github.com/gdamore/tcell/v2"
	"path/filepath"
	"testing"
)

// Every key the gamepad emits (tcell_ebiten pollGamepads) must be bound in each keymap.
func TestGamepadKeysBound(t *testing.T) {
	maps, _ := filepath.Glob("../data_atom/keymaps/*.txt")
	for _, m := range maps {
		u := &UI{}
		u.loadKeyMap(m)
		for _, k := range []tcell.Key{tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight, tcell.KeyHome, tcell.KeyEnd,
			tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyEnter, tcell.KeyEscape, tcell.KeyTab,
			tcell.KeyF13, tcell.KeyF14, tcell.KeyF15, tcell.KeyF16, tcell.KeyF17, tcell.KeyF18, tcell.KeyF19, tcell.KeyF20} {
			if u.getCommandForKey(toUIKey(tcell.NewEventKey(k, 0, tcell.ModNone))) == "" {
				t.Errorf("%s: %s unbound", m, tcell.KeyNames[k])
			}
		}
		if u.getCommandForKey(toUIKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModShift))) == "" {
			t.Errorf("%s: Shift+Tab unbound", m)
		}
	}
}
