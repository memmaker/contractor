package ui_console

import (
	"contractor/foundation"
	"github.com/gdamore/tcell/v2"
	"testing"
)

type fakeItem struct {
	foundation.Item
	key rune
}

func (f fakeItem) Shortcut() rune                        { return f.key }
func (f fakeItem) InventoryNameWithColors(string) string { return "thing" }
func (f fakeItem) GetCarryWeight() int                   { return 1 }
func (f fakeItem) GetCategory() foundation.ItemCategory  { return 0 }
func (f fakeItem) IsEquippable() bool                    { return false }
func (f fakeItem) IsUsableOrZappable() bool              { return false }
func (f fakeItem) IsConsumable() bool                    { return true }
func (f fakeItem) IsReadable() bool                      { return false }

// Feeds the inventory the events tcell_ebiten actually posts.
func TestInventoryKeys(t *testing.T) {
	for _, m := range []string{"wasd", "numpad"} {
		t.Run(m, func(t *testing.T) { inventoryKeys(t, "../data_atom/keymaps/"+m+".txt") })
	}
}

func inventoryKeys(t *testing.T, keymap string) {
	u := &UI{}
	u.loadKeyMap(keymap)
	var got []string
	inv := NewTextInventory(80, 25, func() bool { return false })
	inv.SetItems([]foundation.Item{fakeItem{key: 'a'}, fakeItem{key: 'b'}, fakeItem{key: 'c'}, fakeItem{key: 'd'}, fakeItem{key: 'e'}, fakeItem{key: 'f'}, fakeItem{key: 'g'}})
	inv.SetDefaultSelection(func(i foundation.Item) { got = append(got, "use "+string(i.Shortcut())) })
	inv.SetShiftSelection(func(i foundation.Item) { got = append(got, "drop "+string(i.Shortcut())) })
	inv.SetControlSelection(func(i foundation.Item) { got = append(got, "examine "+string(i.Shortcut())) })
	inv.SetContextMenu(func(i foundation.Item, done func()) { got = append(got, "menu "+string(i.Shortcut())) })
	in := u.directionalWrapperWithoutAlphabet(inv.handleInput)

	check := func(name string, ev *tcell.EventKey, want string) {
		got = nil
		in(ev)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s: got %v, want %s", name, got, want)
		}
	}
	check("shift+g", tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone), "drop g")
	check("ctrl+b", tcell.NewEventKey(tcell.KeyCtrlB, 0, tcell.ModNone), "examine b")
	check("ctrl+g", tcell.NewEventKey(tcell.KeyCtrlG, 0, tcell.ModCtrl), "examine g")
	items := []foundation.Item{}
	for r := 'a'; r <= 'm'; r++ {
		items = append(items, fakeItem{key: r})
	}
	inv.SetItems(items)
	check("ctrl+m", tcell.NewEventKey(tcell.KeyCtrlM, 0, tcell.ModCtrl), "examine m")
	check("ctrl+i", tcell.NewEventKey(tcell.KeyCtrlI, 0, tcell.ModCtrl), "examine i")
	check("enter", tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), "use a")
	in(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	check("gamepad X", tcell.NewEventKey(tcell.KeyF13, 0, tcell.ModNone), "menu b")

	// The letters belong to the rows: other inventory listings renumber the items' own index.
	stale := []foundation.Item{fakeItem{key: 'x'}, fakeItem{key: 'x'}}
	inv.SetItems(stale)
	check("shift+b on renumbered items", tcell.NewEventKey(tcell.KeyRune, 'B', tcell.ModNone), "drop x")
}
