package game

import "testing"

func TestTakeAllLootTakesEveryStack(t *testing.T) {
	g := &GameState{}
	from, to := NewInventory(10), NewInventory(10)
	for i := 0; i < 3; i++ {
		gold := g.NewGold(5)
		from.items[gold.ID()] = gold
	}
	takeAllLoot(from, to)
	if !from.IsEmpty() {
		t.Fatalf("source still has %d items", len(from.items))
	}
}
