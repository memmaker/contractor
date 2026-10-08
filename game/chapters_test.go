package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/convo"
	"github.com/memmaker/go/textiles"
	"testing"
)

// Every ending of chapters 2 and 3 resolves from flags with the right skill points.
func TestChaptersQuestOutcomes(t *testing.T) {
	cases := []struct {
		quest, name, outcome string
		xp                   int
		flags                []string
	}{
		{"recluse", "The Recluse", "faust_dead", 10, []string{"TalkedTo(logan_faust)", "Killed(logan_faust)"}},
		{"recluse", "The Recluse", "sold_out", 10, []string{"faust_sold_out"}},
		{"recluse", "The Recluse", "allied", 10, []string{"TalkedTo(logan_faust)", "faust_allied"}},
		{"recluse", "The Recluse", "extorted", 10, []string{"TalkedTo(logan_faust)", "faust_extorted"}},
		{"cryo", "The Long Sleep", "cryo_sleep", 20, []string{"KnowsAbout(cryo_lab)", "PodFree(4)", "Ending(cryo_sleep)"}},
		{"cryo", "The Long Sleep", "against_all_odds", 20, []string{"KnowsAbout(cryo_lab)", "Ending(against_all_odds)"}},
		{"cryo", "The Long Sleep", "sold_lab", 20, []string{"KnowsAbout(cryo_lab)", "Ending(sold_lab)"}},
		{"cryo", "The Long Sleep", "gave_up", 0, []string{"KnowsAbout(cryo_lab)", "Ending(gave_up)"}},
		{"to_the_stars", "To The Stars", "scam", 0, []string{"KnowsAbout(golden_ticket)", "KnowsAbout(ticket_scam)"}},
	}
	for _, c := range cases {
		g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
		g.init()
		g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
		g.journal.SetChangeHandler(func() {})
		var rewards []Reward
		for _, flag := range c.flags {
			g.gameFlags.Set(flag, 1)
			rewards = append(rewards, g.journal.Update()...)
		}
		for i := 0; i < 3; i++ { // a quest moves one state per update: started, in progress, completed
			rewards = append(rewards, g.journal.Update()...)
		}
		if got := g.journal.OutcomeText(c.quest); got == "" {
			t.Fatalf("%v: quest %s not completed", c.flags, c.quest)
		}
		for _, q := range g.journal.quests["default"] {
			if q.Identifier == c.quest && q.Outcome != c.outcome {
				t.Errorf("%v: outcome %q, want %q", c.flags, q.Outcome, c.outcome)
			}
		}
		xp := -1
		for _, r := range rewards {
			if r.Text == c.name {
				xp = r.SkillPoints
			}
		}
		if xp != c.xp {
			t.Errorf("%v: xp %d, want %d", c.flags, xp, c.xp)
		}
	}
}

// The new dialogues parse, EndGame ends the game with the cryo outcome text, and ebi_strike loads.
func TestChaptersContentLoads(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	g.journal.SetChangeHandler(func() {})
	g.ui = nullUI{}
	g.mapContainsPlayer = false // no UI in tests: keep log messages out
	for _, name := range []string{"logan_faust", "mira_sykes", "mansion_guard", "faust_terminal", "ark_terminal", "cryo_pod"} {
		if _, err := convo.ParseConversation("../data_atom/dialogues/"+name+".rec", g); err != nil {
			t.Errorf("dialogue %s: %v", name, err)
		}
	}
	if LoadScript("../data_atom", "ebi_strike", g.GetScriptFuncs()).IsEmpty() {
		t.Error("script ebi_strike has no frames")
	}
	if _, ok := g.globalTeamTemplates["ebi_strike"]; !ok {
		t.Error("team ebi_strike is not defined in spawned_teams.rec")
	}

	g.gameFlags.SetFlag("KnowsAbout(cryo_lab)")
	if _, err := g.GetScriptFuncs()["EndGame"]("cryo_sleep"); err != nil {
		t.Fatal(err)
	}
	if !g.gameFlags.HasFlag("Ending(cryo_sleep)") || g.pendingEnding != "cryo_sleep" {
		t.Fatalf("EndGame did not flag the ending: %v", g.pendingEnding)
	}
	g.endGame(g.pendingEnding)
	if got := g.journal.OutcomeText("cryo"); got == "" || g.journal.quests["default"] == nil {
		t.Errorf("cryo quest not resolved by EndGame: %q", got)
	}
}
