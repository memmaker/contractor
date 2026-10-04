package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/textiles"
	"testing"
)

// The player carries at most one stack per letter; extra stacks go to onFull.
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
}
