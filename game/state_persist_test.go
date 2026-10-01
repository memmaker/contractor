package game

import (
	"contractor/foundation"
	"os"
	"path/filepath"
	"testing"
)

// HasOutfit('business') in the dialogues relies on armor_style being loaded from the item records.
func TestArmorStyleIsLoaded(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	armor, isArmor := g.NewItemFromString("ebi_officer_uniform").(*Armor)
	if !isArmor {
		t.Fatal("ebi_officer_uniform is not an armor")
	}
	if armor.Style != "business" {
		t.Fatalf("expected style 'business', got %q", armor.Style)
	}
}

func TestSaveReplacesOldSaveOnlyOnSuccess(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	base := t.TempDir()
	saveDir := filepath.Join(base, "slot")

	if err := g.Save(saveDir); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(saveDir, "marker")
	os.WriteFile(marker, nil, 0644)

	// overwriting produces a fresh directory, no leftovers from the old save
	if err := g.Save(saveDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("old save contents survived an overwrite")
	}
	os.WriteFile(marker, nil, 0644)

	// a failing save must report an error and keep the previous save intact
	os.Chmod(base, 0555)
	defer os.Chmod(base, 0755)
	if err := g.Save(saveDir); err == nil {
		t.Fatal("expected save into read-only directory to fail")
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("failed save destroyed the previous save")
	}

	entries, _ := os.ReadDir(base)
	if len(entries) != 1 {
		t.Fatalf("expected only the save slot to remain, got %d entries", len(entries))
	}
}
