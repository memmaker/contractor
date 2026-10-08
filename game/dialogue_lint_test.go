package game

import (
	"contractor/foundation"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/convo"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Every dialogue parses, and every cond/effect compiles against the real script functions (or is a known bare effect).
func TestDialogueExpressions(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	bare := map[string]bool{"EndConversation": true, "EndWithChatter": true, "ReturnToPreviousNode": true,
		"StartCombat": true, "EndHostility": true, "HealPlayer": true,
		"StartTrading": true, "StartRepair": true, "StartCyberware": true}
	field := regexp.MustCompile(`(?m)^(cond|effect|o_cond|o_test|o_effect):\s*(.*?)\s*$`)
	files, _ := filepath.Glob("../data_atom/dialogues/*.rec")
	for _, f := range files {
		name := filepath.Base(f)
		if strings.HasPrefix(name, "_") {
			continue
		}
		if _, err := convo.ParseConversation(f, g); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		data, _ := os.ReadFile(f)
		for _, m := range field.FindAllStringSubmatch(string(data), -1) {
			if bare[m[2]] {
				continue
			}
			if expr, err := govaluate.NewEvaluableExpressionWithFunctions(m[2], g.GetScriptFuncs()); err != nil {
				t.Errorf("%s: %s: %s: %v", name, m[1], m[2], err)
			} else if v := strings.Join(expr.Vars(), ""); v != "" && strings.ReplaceAll(strings.ReplaceAll(v, "NPC_NAME", ""), "NPC", "") != "" {
				t.Errorf("%s: %s: %s: unknown name %v", name, m[1], m[2], expr.Vars())
			}
		}
	}
}

// Every script and playtest line compiles against the real script functions (playtests also get the autoplay verbs),
// and every name it uses is a var defined in the file.
func TestScriptExpressions(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	field := regexp.MustCompile(`(?m)^(if|do|set|var):\s*(.*?)\s*$`)
	for _, dir := range []string{"scripts", "playtests"} {
		funcs := g.GetScriptFuncs()
		if dir == "playtests" {
			funcs = mergeMaps(funcs, (&Autoplay{g: g}).verbs())
		}
		files, _ := filepath.Glob("../data_atom/" + dir + "/*.rec")
		for _, f := range files {
			name := dir + "/" + filepath.Base(f)
			data, _ := os.ReadFile(f)
			vars := map[string]bool{}
			for _, m := range field.FindAllStringSubmatch(string(data), -1) {
				if m[1] == "var" {
					vars[m[2]] = true
				}
			}
			for _, m := range field.FindAllStringSubmatch(string(data), -1) {
				if m[1] == "var" {
					continue
				}
				expr, err := govaluate.NewEvaluableExpressionWithFunctions(m[2], funcs)
				if err != nil {
					t.Errorf("%s: %s: %s: %v", name, m[1], m[2], err)
					continue
				}
				for _, v := range expr.Vars() {
					if !vars[v] {
						t.Errorf("%s: %s: %s: undefined name %s", name, m[1], m[2], v)
					}
				}
			}
		}
	}
}
