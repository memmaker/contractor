package validation

import (
	"contractor/d100"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var skillCallPattern = regexp.MustCompile(`\b(?:Roll)?Skill\('([^']+)'`)

// ValidateSkillNames reports skill names used in dialogues and scripts that the rules don't define.
// The game panics on an unknown skill name, so these crash the game once the line is reached.
func ValidateSkillNames(rootDir string) int {
	defined := make(map[string]bool)
	for i := 0; i < d100.SkillCount(); i++ {
		defined[normalizeSkillName(d100.Skill(i).String())] = true
	}

	errorCount := 0
	for _, subDir := range []string{"dialogues", "scripts"} {
		dir := filepath.Join(rootDir, subDir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".rec" {
				continue
			}
			content, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
			if readErr != nil {
				continue
			}
			reported := make(map[string]bool)
			for _, match := range skillCallPattern.FindAllStringSubmatch(string(content), -1) {
				name := match[1]
				if defined[normalizeSkillName(name)] || reported[name] {
					continue
				}
				reported[name] = true
				errorCount++
				println(fmt.Sprintf("ERR:  %s/%s :  Unknown skill %s", subDir, entry.Name(), name))
			}
		}
	}
	return errorCount
}

func normalizeSkillName(name string) string {
	return strings.ToLower(strings.NewReplacer(" ", "", "_", "", "-", "").Replace(name))
}
