package game

import (
	"testing"

	"contractor/foundation"
)

func TestStackedViewDoesNotMutate(t *testing.T) {
	a := &GenericItem{InternalName: "gold", Category: foundation.ItemCategoryGold, StackSize: 5}
	b := &GenericItem{InternalName: "gold", Category: foundation.ItemCategoryGold, StackSize: 3}
	c := &Container{ContainedItems: []foundation.Item{a, b}}
	for range 3 {
		StackedFilteredAndSortedItems(c.GetItems(), func(foundation.Item) bool { return true })
		c.ItemsFiltered(func(foundation.Item) bool { return true })
	}
	if c.ItemCount("gold") != 8 || len(c.ContainedItems) != 1 {
		t.Fatalf("want one stack of 8, got %d items, count %d", len(c.ContainedItems), c.ItemCount("gold"))
	}
}
