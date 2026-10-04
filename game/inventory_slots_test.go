package game

import (
	"contractor/foundation"
	"contractor/gridmap"
	"github.com/memmaker/go/textiles"
	"testing"
)

// The player carries at most one stack per letter, items and ammo each; extra stacks go to onFull.
func TestInventoryHoldsTwentySixStacks(t *testing.T) {
	inv := NewInventory(InventorySlots)
	var overflow []foundation.Item
	inv.SetOnFull(func(item foundation.Item) { overflow = append(overflow, item) })

	for i := 0; i < InventorySlots+3; i++ {
		inv.AddItem(NewKey("key_"+string(rune('a'+i)), "a key", textiles.TextIcon{}))
	}
	if n := len(inv.GetItems()); n != InventorySlots || len(overflow) != 3 {
		t.Fatalf("%d stacks carried, %d overflowed", n, len(overflow))
	}
	// ammo is a second list with its own 26 slots
	for c := 0; c < InventorySlots; c++ {
		inv.AddItem(&Ammo{GenericItem: &GenericItem{UID: gridmap.NextItemID(), InternalName: "ammo"}, CaliberIndex: c})
	}
	if n, extra := len(inv.GetItems()), len(overflow)-3; n != 2*InventorySlots || extra != 0 {
		t.Fatalf("with ammo: %d stacks carried, %d ammo overflowed", n, extra)
	}
}
