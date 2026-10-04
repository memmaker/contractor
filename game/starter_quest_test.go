package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/textiles"
	"testing"
	"time"
)

// Every ending of the starter quest "The Debt" resolves from its flags, with the right XP.
func TestStarterQuestOutcomes(t *testing.T) {
	cases := []struct {
		outcome string
		xp      int
		flags   []string
	}{
		{"reward_received", 250, []string{"JobAccepted(starter)", "starter(money_collected)", "ClientRewardReceived(starter)"}},
		{"reward_received", 250, []string{"JobAccepted(starter)", "ContainerRemoved(ripperdoc_stash, gold)", "ClientRewardReceived(starter)"}},
		{"reward_received", 250, []string{"JobAccepted(starter)", "Killed(daniel_harker)", "ClientRewardReceived(starter)"}},
		{"paid_with_clients_money", 500, []string{"JobAccepted(starter)", "ContainerRemoved(jacob_safe, gold)", "ClientRewardReceived(starter)"}},
		{"killed_client_for_daniel", 250, []string{"JobAccepted(starter)", "WorkFor(daniel_harker)", "Killed(jacob_thorne)"}},
		{"killed_client", 250, []string{"JobAccepted(starter)", "Killed(jacob_thorne)"}},
		{"client_backed_off", 250, []string{"JobAccepted(starter)", "jacob_backed_off"}},
		{"client_extorted", 250, []string{"JobAccepted(starter)", "starter(client_extorted)"}},
		{"job_lost", 0, []string{"JobDeclined(starter)", "starter(other_assassin_done)"}},
		{"reward_received", 250, []string{"TalkedTo(drake_gallows)", "drake_freed", "ClientRewardReceived(starter)"}},
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
		if got := g.journal.OutcomeText("starter"); got == "" {
			t.Fatalf("%v: quest not completed", c.flags)
		}
		quest := g.journal.quests["default"]
		for _, q := range quest {
			if q.Identifier == "starter" && q.Outcome != c.outcome {
				t.Errorf("%v: outcome %q, want %q", c.flags, q.Outcome, c.outcome)
			}
		}
		xp := -1
		for _, r := range rewards {
			if r.Text == "Medical Extortion" {
				xp = r.XP
			}
		}
		if xp != c.xp {
			t.Errorf("%v: xp %d, want %d", c.flags, xp, c.xp)
		}
	}
}

// The quest scripts parse and Jacob can be found on his map.
func TestStarterQuestContentLoads(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	for _, name := range []string{"other_assassin_gets_job", "job_offer_timeout", "clinic_takeover", "dead_drop_01", "drakes_escape"} {
		if LoadScript("../data_atom", name, g.GetScriptFuncs()).IsEmpty() {
			t.Errorf("script %s has no frames", name)
		}
	}
}

func TestStarterQuestActorsArePlaced(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	for actorName, mapName := range map[string]string{"jacob_thorne": "zone_residential_east", "beggar": "zone_residential_east", "ebi_medic": "hq_ebi", "cleaner": "zone_residential_west", "quinn_rix": "zone_residential_south"} {
		found := false
		for _, actor := range g.ensureMapIsLoaded(mapName).Actors() {
			found = found || actor.GetInternalName() == actorName
		}
		if !found {
			t.Errorf("%s not found on %s", actorName, mapName)
		}
	}
}

// Leaving Jacob's offer unanswered for 48 hours counts as declining it.
func TestStarterJobOfferTimesOut(t *testing.T) {
	g := NewGameState(&foundation.Configuration{DataRootDir: "../data_atom"})
	g.init()
	g.Player = NewPlayer("tester", textiles.TextIcon{}, d100.NewCharSheet())
	g.mapContainsPlayer = false // no UI in tests: keep log messages out
	g.Scripts.RunScriptByName("../data_atom", "job_offer_timeout", g.GetScriptFuncs())
	g.Scripts.CheckAndRunFrames() // offer time saved

	g.gameTime = g.gameTime.AddDuration(47 * time.Hour)
	g.Scripts.CheckAndRunFrames()
	if g.gameFlags.HasFlag("JobDeclined(starter)") {
		t.Fatal("declined before 48 hours")
	}

	g.gameTime = g.gameTime.AddDuration(2 * time.Hour)
	g.Scripts.CheckAndRunFrames()
	if !g.gameFlags.HasFlag("JobDeclined(starter)") || !g.Scripts.IsScriptRunning("other_assassin_gets_job") {
		t.Fatal("offer did not time out into Quinn taking the job")
	}
}
