package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"maps"
	"path/filepath"
	"slices"
	"time"
)

// endPlayerTurn is called by game actions that end the player's turn.
// It will
// - animate the player's actions
// - run the AI for all enemies
// - advance time and turn counter
// - trigger "before enemy turn" effects
// - animate the enemies' actions
// - trigger "after enemy turn" effects
// - simulate all other active maps
// - execute any actions that were queued to be executed after animations
// - update the UI status
// - check if the player can act
func (g *GameState) endPlayerTurn(playerTimeTakenForTurn int) {
	// player has changed the game state..
	g.actorsComputed = 0

	didCancel := g.ui.AnimatePending() // animate player actions..

	g.tickWorld(time.Second*time.Duration(float64(playerTimeTakenForTurn)/10), false, func() {
		if didCancel {
			g.ui.SkipAnimations()
		} else {
			didCancel = g.ui.AnimatePending() // animate enemy actions
		}
	})

	if didCancel {
		g.ui.SkipAnimations()
	} else {
		g.ui.AnimatePending() // animate afterTurn effects
	}

	// actors moved and doors changed this turn; shadows must follow
	if len(g.currentMap().DynamicLights) > 0 {
		g.currentMap().UpdateDynamicLights()
	}

	g.checkPlayerCanAct()

	g.gameFlags.Set("ActorsComputed", g.actorsComputed)

	if g.Player.HasFlag(foundation.FlagSneaking) {
		g.ui.SetSneakOverlay(g.createSneakOverlay())
	} else {
		g.ui.SetSneakOverlay(nil)
	}

	g.updateUIStatus()
}

func (g *GameState) onOtherMaps(call func(string)) {
	for _, mapName := range slices.Sorted(maps.Keys(g.activeMaps)) { // fixed order keeps seeded runs reproducible
		if mapName == g.currentMapName {
			continue
		}
		g.ExecuteOnMap(mapName, func() {
			call(mapName)
		})
	}
}

// tickWorld is the one place where game time passes: a player turn, or one minute of a rest (the same simulation, headless).
// Everything that triggers on time or turns runs here, so events fire the same way whether the player acts or rests.
func (g *GameState) tickWorld(duration time.Duration, resting bool, afterActorsActed func()) {
	energy := int(duration.Seconds()) * 10

	// AI Actions (incl. Behaviours, Goals, Schedules) and State Changes happen here
	g.actorsOnMapAct(energy)
	afterActorsActed()
	if g.config.SimulateAllLoadedMaps {
		g.onOtherMaps(func(string) {
			g.actorsOnMapAct(energy)
			g.ui.SkipAnimations()
		})
	}

	// advancing the time will run scripts
	if resting {
		g.gameTime = g.gameTime.AddDuration(duration)
		g.Scripts.CheckAndRunFrames()
	} else {
		g.advanceTimeAndTurn(duration)
	}
	if g.config.SimulateAllLoadedMaps {
		g.onOtherMaps(func(string) { g.afterTurnEffectsForActors(int(duration.Seconds())) })
	}
	g.ui.SkipAnimations()

	// trigger turn based events; timers count turns, a minute of rest is worth that many
	for i := 1; i < int(duration.Seconds()) && len(g.metronome.timed) > 0; i++ {
		g.metronome.Tick(g)
	}
	g.afterTurn(duration)

	// This is where level transitions are handled
	for _, action := range g.afterAnimationActions {
		action()
	}
	g.afterAnimationActions = nil
}

// actorsOnMapAct hands out the time in rounds: all of it at once while the map is at peace, one turn (10) at a time
// once anyone on it fights, so a rest minute never lets one side empty its clips before the other side may answer.
func (g *GameState) actorsOnMapAct(timeSpent int) {
	for timeSpent > 0 {
		step := timeSpent
		if g.currentMapHasCombat() {
			step = min(10, timeSpent)
		}
		timeSpent -= step
		g.actorsOnMapActRound(step)
	}
}

func (g *GameState) currentMapHasCombat() bool {
	for _, actor := range g.currentMap().Actors() {
		if actor != g.Player && actor.IsAlive() && actor.IsInCombat() {
			return true
		}
	}
	return false
}

func (g *GameState) actorsOnMapActRound(playerTimeSpent int) {
	gridMap := g.currentMap()
	allEnemies := gridMap.Actors()
	for _, enemy := range allEnemies {
		if enemy == g.Player {
			continue
		}
		if !enemy.IsAlive() {
			continue
		}

		// TODO: Fix the time energy system
		// Currently actors go into negative energy, which is not intended
		// Also taxi driver skips a turn and then moves twice, when chasing the player

		// IMPORTANT:
		// Actions of enemies should never remove actors from the game directly
		g.actorsComputed++
		enemy.AddTimeEnergy(playerTimeSpent)
		hasActions := true
		for hasActions {
			tuSpent := g.TryAIAction(enemy)
			if tuSpent == 0 {
				hasActions = false
			} else {
				enemy.SpendTimeEnergy(tuSpent)
				//g.msg(foundation.HiLite("%s spent %s time units", enemy.Name(), strconv.Itoa(tuSpent)))
			}
		}
	}
}

func (g *GameState) isPlayerHungry() bool {
	lastEaten := g.GetNamedTime("PlayerLastAteAt")
	hours := 24
	if g.Player.HasFlag(foundation.FlagSlowDigestion) {
		hours = 48
	}
	// more than 24 hours since last meal
	if g.gameTime.Time.Sub(lastEaten.Time) > time.Duration(hours)*time.Hour {
		return true
	}
	return false
}

func (g *GameState) isPlayerStarving() bool {
	lastEaten := g.GetNamedTime("PlayerLastAteAt")
	hours := 72
	if g.Player.HasFlag(foundation.FlagSlowDigestion) {
		hours = 144
	}
	if g.gameTime.Time.Sub(lastEaten.Time) > time.Duration(hours)*time.Hour {
		return true
	}
	return false
}

func (g *GameState) afterTurn(duration time.Duration) {
	g.metronome.Tick(g)

	g.checkJournal()

	g.afterTurnEffectsForActors(int(duration.Seconds()))

	// handle player hunger
	if g.isPlayerStarving() {
		g.Player.SetFlag(foundation.FlagStarving)
		g.Player.GetFlags().Unset(foundation.FlagHunger)

		if g.Player.HasActionPoints() {
			g.Player.GetCharSheet().LooseActionPoints(1)
		} else if g.Player.GetHitPoints() > 1 {
			g.Player.GetCharSheet().TakeRawDamage(1)
		}
	} else if g.isPlayerHungry() {
		g.Player.SetFlag(foundation.FlagHunger)
		g.Player.GetFlags().Unset(foundation.FlagStarving)
	}
}

func (g *GameState) afterTurnEffectsForActors(seconds int) {
	// apply after turn effects for all actors
	currentMap := g.currentMap()
	allActorsOnThisMap := currentMap.Actors()
	for i := len(allActorsOnThisMap) - 1; i >= 0; i-- {
		actor := allActorsOnThisMap[i]
		if actor.IsAlive() && actor.HasFlag(foundation.FlagBleeding) {
			g.bleed(actor, seconds)
		}
		wornOff := actor.AfterTurn()
		if actor == g.Player && len(wornOff) > 0 {
			for _, effect := range wornOff {
				g.msg(effect)
			}
		}
		if actor.HasFlag(foundation.FlagRegenerating) && actor.IsWounded() {
			actor.Heal(1)
		}
	}
}

func (g *GameState) isAtScheduledLocation(actor *Actor, slot TimeSlot) bool {
	location := slot.Location
	if location == "" {
		return false
	}
	return actor.Position() == g.currentMap().GetNamedLocation(location)
}

func (g *GameState) loadSchedule(actor *Actor) {
	schedulePath := filepath.Join(g.config.DataRootDir, "schedules", actor.GetInternalName()+".rec")
	if fxtools.FileExists(schedulePath) {
		actor.Schedule = NewScheduleFromFile(schedulePath, func() time.Time {
			return g.gameTime.Time
		})
	}
}

// bleed costs one hit point per 10 seconds of game time, whether they pass as player turns or rest minutes.
func (g *GameState) bleed(actor *Actor, seconds int) {
	flags := actor.GetFlags()
	accrued := flags.Get(foundation.FlagBleeding) + seconds
	flags.SetFlagTo(foundation.FlagBleeding, accrued%10)
	if lost := accrued / 10; lost > 0 {
		if actor.GetHitPoints() < lost { // bleeding out, not a hit: no overkill splatter
			lost = actor.GetHitPoints()
		}
		g.msg(foundation.Msg(fmt.Sprintf("%s is bleeding", actor.Name())))
		g.ui.AddAnimations(OneAnimation(g.damageActor(SourcedDamage{NameOfThing: "blood loss", DamageType: DamageTypeNormal, DamageAmount: lost, BodyPart: d100.Body}, actor)))
	}
}
