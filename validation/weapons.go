package validation

import (
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"path/filepath"
)

type WeaponChecker struct {
	rootDir string
}

// PrintReport prints the caliber pairings and returns the number of errors found.
func (c WeaponChecker) PrintReport() int {
	errorCount := 0
	ammoPath := filepath.Join(c.rootDir, "definitions", "ammo.rec")
	weaponPath := filepath.Join(c.rootDir, "definitions", "weapons.rec")

	ammoRecords, _ := recfile.ReadAndClose(fxtools.MustOpen(ammoPath))
	weaponRecords, _ := recfile.ReadAndClose(fxtools.MustOpen(weaponPath))

	caliberToAmmo := make(map[int][]string)
	maxCaliberIndex := 0

	for _, record := range ammoRecords {
		ammoIdentifier := record.FindValueForKeyIgnoreCase("name")
		caliberField, hasCaliber := record.FindFieldIgnoreCase("ammo_caliber_index")
		if hasCaliber {
			caliberIndex := caliberField.AsInt()
			maxCaliberIndex = max(maxCaliberIndex, caliberIndex)
			caliberToAmmo[caliberIndex] = append(caliberToAmmo[caliberIndex], ammoIdentifier)
		}
	}
	caliberToWeapon := make(map[int][]string)
	weaponsWithoutCaliber := make([]string, 0)
	for _, record := range weaponRecords {
		weaponIdentifier := record.FindValueForKeyIgnoreCase("name")
		caliberField, hasCaliber := record.FindFieldIgnoreCase("weapon_caliber_index")
		caliberIndex := caliberField.AsInt()
		maxCaliberIndex = max(maxCaliberIndex, caliberIndex)
		if hasCaliber {
			caliberToWeapon[caliberIndex] = append(caliberToWeapon[caliberIndex], weaponIdentifier)
			ammoTypes, hasAmmoForIndex := caliberToAmmo[caliberIndex]
			if !hasAmmoForIndex || len(ammoTypes) == 0 {
				errorCount++
				println(fmt.Sprintf("ERR: Weapon %s has no ammo defined for caliber index %d", weaponIdentifier, caliberIndex))
				continue
			}
		} else {
			weaponsWithoutCaliber = append(weaponsWithoutCaliber, weaponIdentifier)
		}
	}

	println("Weapons without ammo:")
	for _, weapon := range weaponsWithoutCaliber {
		println("  ", weapon)
	}
	println()

	for caliberIndex := 0; caliberIndex <= maxCaliberIndex; caliberIndex++ {
		ammoTypes, hasAmmoForIndex := caliberToAmmo[caliberIndex]
		weaponTypes, hasWeaponsForIndex := caliberToWeapon[caliberIndex]

		if !hasAmmoForIndex && !hasWeaponsForIndex {
			continue // unused caliber index
		}
		if !hasAmmoForIndex {
			errorCount++
			println(fmt.Sprintf("ERR: No ammo defined for caliber index %d", caliberIndex))
			continue
		}
		if !hasWeaponsForIndex {
			errorCount++
			println(fmt.Sprintf("ERR: No weapons defined for caliber index %d", caliberIndex))
			continue
		}

		println(fmt.Sprintf("Caliber index %d", caliberIndex))
		println("Ammo:")
		for _, ammoType := range ammoTypes {
			println("  ", ammoType)
		}
		println("Weapons:")
		for _, weaponType := range weaponTypes {
			println("  ", weaponType)
		}
		println()

	}
	return errorCount
}

func ValidateWeaponAndAmmoPairings(rootDir string) int {
	weaponChecker := NewWeaponChecker(rootDir)
	return weaponChecker.PrintReport()
}

func NewWeaponChecker(dir string) *WeaponChecker {
	return &WeaponChecker{rootDir: dir}
}
