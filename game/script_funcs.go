package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/geometry"
	"strconv"
	"strings"
	"time"
)

func (g *GameState) GetScriptFuncs() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		// Player Only
		"IsWounded": func(args ...interface{}) (interface{}, error) {
			return g.Player.IsWounded(), nil
		},
		"Skill": func(args ...interface{}) (interface{}, error) {
			skillName := args[0].(string)
			skillValue := g.Player.GetCharSheet().GetSkill(d100.SkillFromString(skillName))
			return (float64)(skillValue), nil
		},
		"RollSkill": func(args ...interface{}) (interface{}, error) {
			skillName := args[0].(string)
			diff := d100.Medium
			if len(args) > 1 {
				diff = d100.DifficultyFromString(args[1].(string))
			}
			result := g.Player.GetCharSheet().SkillRollVsDiff(d100.SkillFromString(skillName), diff)
			return g.forcedCheck(result.Success), nil
		},
		// o_test: DialogueCheck(NPC, 'intimidate')
		"DialogueCheck": func(args ...interface{}) (interface{}, error) {
			opponent := args[0].(*Actor)
			skillName := args[1].(string)
			skill := d100.SkillFromString(skillName)
			skillValue := g.Player.GetCharSheet().GetSkill(skill)
			mods := g.getDialogueCheckMods(g.Player, opponent, skill, nil)
			chance := mods.Apply(skillValue)
			result := d100.SuccessRoll(d100.Percentage(chance), 0)
			return g.forcedCheck(result.Success), nil
		},

		// Player Inventory & Equipment
		"RemoveItem": func(args ...interface{}) (interface{}, error) {
			count := 1
			itemName := args[0].(string)
			if len(args) > 1 {
				count = int(args[1].(float64))
			}
			removedItem := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			if len(removedItem) > 0 {
				g.msg(foundation.HiLite("%s removed.", removedItem[0].Name()))
			}
			return nil, nil
		},
		// eg. StackTransForTo(NPC, 'gold', 500)
		"ContainerStackTransferTo": func(args ...interface{}) (interface{}, error) {
			count := 1
			sourceContainer := args[0].(*Container)
			targetActor := args[1].(*Actor)
			itemName := args[2].(string)
			if len(args) > 3 {
				count = int(args[3].(float64))
			}
			removedItem := sourceContainer.RemoveItemsWithName(itemName, count)
			if len(removedItem) > 0 {
				targetActor.GetInventory().AddItems(removedItem)
				if targetActor == g.Player {
					g.msg(foundation.HiLite("%s received.", removedItem[0].Name()))
				} else if g.Player.CanSee(targetActor.Position()) && g.Player.CanSee(sourceContainer.Position()) {
					g.msg(foundation.HiLite("%s took something from %s.", targetActor.Name(), sourceContainer.Name()))
				}
			}
			return nil, nil
		},
		"StackTransferTo": func(args ...interface{}) (interface{}, error) {
			count := 1
			targetActor := args[0].(*Actor)
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItem := g.Player.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			if len(removedItem) > 0 {
				targetActor.GetInventory().AddItems(removedItem)
				g.msg(foundation.HiLite("%s removed.", removedItem[0].Name()))
			}
			return nil, nil
		},
		"StackTransferFrom": func(args ...interface{}) (interface{}, error) {
			count := 1
			sourceActor := args[0].(*Actor)
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItem := sourceActor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			if len(removedItem) == 0 && sourceActor.GetVendorInventory() != nil { // shopkeepers sell from their stock
				removedItem = sourceActor.GetVendorInventory().RemoveItemsByNameAndCount(itemName, count)
			}
			if len(removedItem) > 0 {
				g.Player.GetInventory().AddItems(removedItem)
				g.msg(foundation.HiLite("%s received.", removedItem[0].Name()))
			}
			return nil, nil
		},

		"HasItem": func(args ...interface{}) (interface{}, error) {
			itemName := args[0].(string)
			count := 1
			if len(args) > 1 {
				count = int(args[1].(float64))
			}
			return g.Player.GetInventory().HasItemWithNameAndCount(itemName, count), nil
		},
		"HasArmorEquipped": func(args ...interface{}) (interface{}, error) {
			return g.Player.GetInventory().HasArmorEquipped(), nil
		},
		// eg. HasOutfit('business'), checks the armor_style of the worn body armor
		"HasOutfit": func(args ...interface{}) (interface{}, error) {
			armor := g.Player.GetInventory().GetArmor()
			return armor != nil && armor.Style == args[0].(string), nil
		},
		"HasVisibleWeapon": func(args ...interface{}) (interface{}, error) {
			return g.Player.IsOpenCarryWeapon(), nil
		},
		"HasArmorEquippedWithName": func(args ...interface{}) (interface{}, error) {
			armorName := args[0].(string)
			return g.Player.GetInventory().HasArmorWithNameEquipped(armorName), nil
		},
		"HasWeaponEquippedWithName": func(args ...interface{}) (interface{}, error) {
			weaponName := args[0].(string)
			mainHandItem, hasMainHandItem := g.Player.GetInventory().GetMainHandItem()
			if !hasMainHandItem {
				return false, nil
			}
			return mainHandItem.GetInternalName() == weaponName, nil
		},

		// Global Queries & Actions
		"UserFunc": func(args ...interface{}) (interface{}, error) {
			funcName := args[0].(string)
			expression, exists := g.userFunctions[funcName]
			if !exists {
				return nil, nil
			}
			return expression.Evaluate(nil)
		},
		// Flags
		"HasFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			return (bool)(g.gameFlags.HasFlag(flagName)), nil
		},
		"SetFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			g.gameFlags.SetFlag(flagName)
			return nil, nil
		},
		"ClearFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			g.gameFlags.ClearFlag(flagName)
			return nil, nil
		},
		"GetFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			value := g.gameFlags.Get(flagName)
			return (float64)(value), nil
		},
		"IsMap": func(args ...interface{}) (interface{}, error) {
			mapName := args[0].(string)
			return g.currentMap().GetName() == mapName, nil
		},
		"RandomNumberCode": func(args ...interface{}) (interface{}, error) {
			nameOfDoor := args[0].(string)
			code := g.getRandomNumberCode(nameOfDoor)
			return string(code), nil
		},
		"LocationInZone": func(args ...interface{}) (interface{}, error) {
			mapName := args[0].(string)
			zoneName := args[1].(string)
			locIndex := int(args[2].(float64))
			gMap := g.ensureMapIsLoaded(mapName)
			loc := gMap.GetZoneLocationByIndex(zoneName, locIndex)
			return loc, nil
		},
		// Time / Turns
		"Turns": func(args ...interface{}) (interface{}, error) {
			return (float64)(g.TurnsTaken()), nil
		},
		"IsTurnsAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			turns := args[1].(float64)
			return g.IsTurnsAfter(namedTime, int(turns)), nil
		},
		"IsMinutesAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			minutes := args[1].(float64)
			return g.IsMinutesAfter(namedTime, int(minutes)), nil
		},
		"IsHoursAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			hours := args[1].(float64)
			return g.IsHoursAfter(namedTime, int(hours)), nil
		},
		// IsNight(): 22:00 to 05:00 game time
		"IsNight": func(args ...interface{}) (interface{}, error) {
			h := g.gameTime.Time.Hour()
			return h >= 22 || h < 5, nil
		},
		"IsDaysAfter": func(args ...interface{}) (interface{}, error) {
			namedTime := args[0].(string)
			days := args[1].(float64)
			return g.IsDaysAfter(namedTime, int(days)), nil
		},

		// Scripts
		"IsScriptRunning": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			return g.Scripts.IsScriptRunning(scriptName), nil
		},
		"RunScript": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			g.RunScriptByName(scriptName)
			return nil, nil
		},
		"StopScript": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			g.Scripts.StopScript(scriptName)
			return nil, nil
		},
		"RestartScript": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			g.Scripts.StopScript(scriptName)
			g.RunScriptByName(scriptName)
			return nil, nil
		},
		"AllFramesPlayed": func(args ...interface{}) (interface{}, error) {
			scriptName := args[0].(string)
			return g.Scripts.AllFramesPlayed(scriptName), nil
		},
		"RunScriptKill": func(args ...interface{}) (interface{}, error) {
			killerName := args[0].(string)
			victimName := args[1].(string)
			killer := g.actorWithName(killerName)
			victim := g.actorWithName(victimName)
			g.msg(foundation.HiLite("%s kills %s.", killerName, victimName))
			if victim == nil && len(args) > 2 { // victim lives on another map: the killer walks over there
				if otherMap := g.ensureMapIsLoaded(args[2].(string)); otherMap != nil {
					for _, actor := range otherMap.Actors() {
						if actor.GetInternalName() == victimName {
							victim = actor
						}
					}
				}
			}
			killScript := g.NewScriptKill(killer, victim)
			g.Scripts.Run(killScript)
			return nil, nil
		},
		"RunScriptLeave": func(args ...interface{}) (interface{}, error) {
			argZero := args[0]
			var leaver *Actor
			if nameOfLeaver, isString := argZero.(string); isString {
				leaver = g.actorWithName(nameOfLeaver)
			} else if actor, isActor := argZero.(*Actor); isActor {
				leaver = actor
			}

			isRunning := false
			if len(args) > 1 {
				isRunning = args[1].(bool)
			}

			killScript := g.NewScriptLeaveMap(leaver, isRunning)
			g.Scripts.Run(killScript)
			return nil, nil
		},
		"RunScriptLeaveAt": func(args ...interface{}) (interface{}, error) {
			argZero := args[0]
			var leaver *Actor
			if nameOfLeaver, isString := argZero.(string); isString {
				leaver = g.actorWithName(nameOfLeaver)
			} else if actor, isActor := argZero.(*Actor); isActor {
				leaver = actor
			}

			transitionLocation := args[1].(string)
			isRunning := false
			if len(args) > 2 {
				isRunning = args[2].(bool)
			}

			killScript := g.NewScriptLeaveMapAtLocation(leaver, isRunning, transitionLocation)
			g.Scripts.Run(killScript)
			return nil, nil
		},

		// Query Containers
		"ContainerWithName": func(args ...interface{}) (interface{}, error) {
			containerName := args[0].(string)
			var container *Container
			g.IterateAllObjects(func(mapName string, object Object) bool {
				if object.GetInternalName() == containerName {
					ctn, isContainer := object.(*Container)
					if isContainer {
						container = ctn
						return false
					}
				}
				return true
			})
			return container, nil
		},
		"IsItemInContainer": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			count := 1
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			return container.HasItemsWithName(itemName, count), nil
		},

		// Query Actors
		"PlayerInZone": func(args ...interface{}) (interface{}, error) {
			zoneName := args[0].(string)
			return g.currentMap().IsZoneAt(g.Player.Position(), zoneName), nil
		},
		"ActorWithName": func(args ...interface{}) (interface{}, error) {
			actorName := args[0].(string)
			var foundActor *Actor
			g.IterateAllActors(func(mapName string, actor *Actor) bool {
				if actor.GetInternalName() == actorName {
					foundActor = actor
					return false
				}
				return true
			})
			return foundActor, nil
		},
		"HasActorFlag": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			flagName := args[1].(string)
			return actor.HasFlag(foundation.ActorFlagFromString(flagName)), nil
		},
		"ItemCountZoneRecursiveSinglePrefix": func(args ...interface{}) (interface{}, error) {
			itemPrefix := args[0].(string)
			mapName := args[1].(string)
			zoneName := args[2].(string)
			count := 0
			g.IterateItemsOnMapZoneRecursively(mapName, zoneName, func(item foundation.Item) {
				if item.GetStackSize() == 1 &&
					strings.HasPrefix(item.GetInternalName(), itemPrefix) {
					count++
				}
			})
			return (float64)(count), nil
		},
		"ItemCountFactionInventorySinglePrefix": func(args ...interface{}) (interface{}, error) {
			itemPrefix := args[0].(string)
			mapName := args[1].(string)
			factionName := args[2].(string)
			count := 0
			g.IterateItemsInAllInventories(mapName, func(owner *Actor, item foundation.Item) {
				if owner.GetTeam() == factionName &&
					item.GetStackSize() == 1 &&
					strings.HasPrefix(item.GetInternalName(), itemPrefix) {
					count++
				}
			})
			return (float64)(count), nil
		},
		"IsActorWounded": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			return actor.IsWounded(), nil
		},
		"IsActorInShootingRange": func(args ...interface{}) (interface{}, error) {
			attacker := args[0].(*Actor)
			defender := args[1].(*Actor)
			return g.IsInShootingRange(attacker, defender), nil
		},
		"IsActorNextTo": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			mapName := args[1].(string)
			target := args[2].(geometry.Point)
			gMap := g.ensureMapIsLoaded(mapName)
			moveDist := gMap.MoveDistance(actor.Position(), target)
			return moveDist <= 1, nil
		},
		"IsActorAtNamedLocation": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			mapName := args[1].(string)
			locName := args[2].(string)
			gMap := g.ensureMapIsLoaded(mapName)
			loc := gMap.GetNamedLocation(locName)
			return actor.Position() == loc, nil
		},
		"IsActorAbleToReach": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			mapName := args[1].(string)
			locName := args[2].(string)
			gMap := g.ensureMapIsLoaded(mapName)
			loc := gMap.GetNamedLocation(locName)
			if loc == actor.Position() { // TODO: check for the actor being on the right map..
				return true, nil
			}
			pathToDest := gMap.GetJPSPath(actor.Position(), loc, func(point geometry.Point) bool {
				return gMap.IsWalkableFor(point, actor)
			})
			if len(pathToDest) == 0 {
				return false, nil
			}

			lastPos := pathToDest[len(pathToDest)-1]
			return lastPos == loc, nil
		},
		"IsActorDead": func(args ...interface{}) (interface{}, error) {
			actor, ok := args[0].(*Actor)
			return ok && actor != nil && !actor.IsAlive(), nil
		},
		"IsActorInCombat": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if actor.IsInCombat() {
				return true, nil
			}
			return false, nil
		},
		"IsActorInTalkingRange": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			return g.IsInTalkingRange(actor, target), nil
		},
		"IsActorInCombatWithPlayer": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if actor.IsHostileTowards(g.Player) {
				return true, nil
			}
			return false, nil
		},
		// Actor Actions,
		"ActorDropItem": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			count := 1
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			for _, item := range removedItems {
				g.ensureMapIsLoaded(actor.currentMapName).AddItemWithDisplacement(item, actor.Position())
			}
			if len(removedItems) > 0 {
				first := removedItems[0]
				displayName := first.Name()
				g.msg(foundation.HiLite("%s dropped %s.", actor.Name(), displayName))
			}
			return nil, nil
		},
		"ActorDropItemAt": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			count := 1
			targetPos := args[1].(geometry.Point)
			itemName := args[2].(string)
			if len(args) > 3 {
				count = int(args[3].(float64))
			}
			removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			for _, item := range removedItems {
				g.ensureMapIsLoaded(actor.currentMapName).AddItemWithDisplacement(item, targetPos)
			}
			if len(removedItems) > 0 {
				first := removedItems[0]
				displayName := first.Name()
				if actor != g.Player && g.Player.CanSee(actor.Position()) {
					g.msg(foundation.HiLite("%s dropped %s.", actor.Name(), displayName))
				}
			}
			return nil, nil
		},
		"ActorTransition": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			currentMap := g.currentMap()
			transition, exists := currentMap.GetTransitionAt(actor.Position())
			if !exists {
				return nil, nil
			}
			g.actorTransition(currentMap, actor, transition)
			return nil, nil
		},
		// Dialogue travel, eg. Transition('zone_commerce', 'taxi_stand'), followed by EndWithChatter
		"Transition": func(args ...interface{}) (interface{}, error) {
			g.ui.FadeToBlack()
			g.transitionToMapLocation(args[0].(string), args[1].(string))
			g.ui.FadeFromBlack()
			return nil, nil
		},
		// eg. TransitionWithDriver(NPC, 'zone_commerce', 'taxi_stand'), the driver will be placed at 'taxi_driver'
		"TransitionWithDriver": func(args ...interface{}) (interface{}, error) {
			driver := args[0].(*Actor)
			g.ui.FadeToBlack()
			g.currentMap().RemoveActor(driver)
			g.transitionToMapLocation(args[1].(string), args[2].(string))
			g.currentMap().AddActor(driver, g.currentMap().GetNamedLocation("taxi_driver"))
			driver.SpawnMapName, driver.SpawnPosition = g.currentMapName, driver.Position() // idle actors walk back to spawn: the new stand is home now
			g.ui.FadeFromBlack()
			return nil, nil
		},
		"ActorSleep": func(args ...interface{}) (interface{}, error) {
			args[0].(*Actor).SetSleeping()
			return nil, nil
		},
		"ActorDie": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			damage := SourcedDamage{
				NameOfThing:  "self-destruct",
				Attacker:     actor,
				DamageType:   DamageTypeNormal,
				DamageAmount: actor.GetHitPoints() + actor.GetHitPointsMax(),
				BodyPart:     d100.Body,
			}
			g.ui.AddAnimations(OneAnimation(g.damageActor(damage, actor)))
			return nil, nil
		},
		// ActorsDie('rat'): every living actor with that internal name dies, on every loaded map (poison, gas).
		"ActorsDie": func(args ...interface{}) (interface{}, error) {
			name := args[0].(string)
			var doomed []*Actor
			g.IterateAllActors(func(mapName string, actor *Actor) bool {
				if actor.GetInternalName() == name && actor.IsAlive() {
					doomed = append(doomed, actor)
				}
				return true
			})
			for _, actor := range doomed {
				g.damageActor(SourcedDamage{NameOfThing: "poison", Attacker: actor, DamageType: DamageTypeNormal, DamageAmount: actor.GetHitPoints() + actor.GetHitPointsMax(), BodyPart: d100.Body}, actor)
			}
			return nil, nil
		},
		// ActorsLeave('rat'): every living actor with that internal name leaves the game, on every loaded map.
		"ActorsLeave": func(args ...interface{}) (interface{}, error) {
			name := args[0].(string)
			for _, m := range g.activeMaps {
				var leaving []*Actor
				for _, actor := range m.Actors() {
					if actor.GetInternalName() == name && actor.IsAlive() && actor != g.Player {
						leaving = append(leaving, actor)
					}
				}
				for _, actor := range leaving {
					m.RemoveActor(actor)
				}
			}
			return nil, nil
		},
		"PlayerAddCyberware": func(args ...interface{}) (interface{}, error) {
			cyberwareName := args[0].(string)
			g.playerAddCyberware(NewCyberWareFromString(cyberwareName))
			return nil, nil
		},
		// Global Actions
		"SaveTimeNow": func(args ...interface{}) (interface{}, error) {
			nameForTime := args[0].(string)
			g.SaveTimeNow(nameForTime)
			return nil, nil
		},
		"AdvanceTimeByMinutes": func(args ...interface{}) (interface{}, error) {
			minutes := int(args[0].(float64))
			g.advanceTime(time.Minute * time.Duration(minutes))
			return nil, nil
		},
		"AddChatter": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			chatter := args[1].(string)
			g.tryAddChatter(actor, chatter)
			return nil, nil
		},
		"Hilite": func(args ...interface{}) (interface{}, error) {
			text := args[0].(string)
			g.msg(foundation.HiLite(text))
			return nil, nil
		},

		// Actors Goals
		"MoveTo": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			mapName := args[1].(string)
			locName := args[2].(string)
			gMap := g.ensureMapIsLoaded(mapName)
			pos := gMap.GetNamedLocation(locName)
			actor.FSM.SetState(StateScripted, LocationEvent{Event: EventNone, Location: MapPosition{
				MapName:      mapName,
				LocationName: locName,
				Position:     pos,
			}})
			return nil, nil
		},
		"MoveNextTo": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			mapName := args[1].(string)
			location := args[2].(geometry.Point)
			targetMap := g.ensureMapIsLoaded(mapName)
			targetLocation := location
			neighbors := targetMap.GetFreeCardinalNeighbors(location)
			if len(neighbors) > 0 {
				targetLocation = neighbors[0]
			}
			actor.FSM.SetState(StateScripted, LocationEvent{Event: EventNone, Location: MapPosition{
				MapName:  mapName,
				Position: targetLocation,
			}})
			return nil, nil
		},
		"Approach": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			target := args[1].(*Actor)
			actor.FSM.SetState(StateScripted, ActorEvent{Event: EventApproach, Actor: target})
			return nil, nil
		},
		"SetIdle": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			actor.FSM.SetState(StateIdle, NoEvent)
			return nil, nil
		},
		// Container Actions
		"ContainerRemoveItem": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			itemName := args[1].(string)
			count := 1
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			return container.RemoveItemsWithName(itemName, count), nil
		},
		"ContainerAddItem": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			if item, isItem := args[1].(foundation.Item); isItem {
				container.AddItem(item)
				return nil, nil
			} else if items, isItems := args[1].([]foundation.Item); isItems {
				container.AddItems(items)
				return nil, nil
			}
			newItem := g.NewItemFromString(args[1].(string))
			container.AddItem(newItem)
			return nil, nil
		},
		"ContainerEnsure": func(args ...interface{}) (interface{}, error) {
			container := args[0].(*Container)
			itemName := args[1].(string)
			itemCount := int(args[2].(float64))

			itemCountInContainer := container.ItemCount(itemName)

			if itemCountInContainer >= itemCount {
				return nil, nil
			}
			diff := itemCount - itemCountInContainer
			newItem := g.NewItemFromString(itemName)
			newItem.SetStackSize(diff)
			container.AddItem(newItem)
			return nil, nil
		},
		"ActorRemoveItem": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			count := 1
			itemName := args[1].(string)
			if len(args) > 2 {
				count = int(args[2].(float64))
			}
			removedItems := actor.GetInventory().RemoveItemsByNameAndCount(itemName, count)
			return removedItems, nil
		},
		"ActorAddItem": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			if item, isItem := args[1].(foundation.Item); isItem {
				actor.GetInventory().AddItem(item)
				return nil, nil
			} else if items, isItems := args[1].([]foundation.Item); isItems {
				actor.GetInventory().AddItems(items)
				return nil, nil
			}
			newItem := g.NewItemFromString(args[1].(string))
			actor.GetInventory().AddItem(newItem)
			return nil, nil
		},
		"PlayerInTalkRange": func(args ...interface{}) (interface{}, error) {
			actor := args[0].(*Actor)
			return g.IsInTalkingRange(g.Player, actor), nil
		},
		"PlayerAddItem": func(args ...interface{}) (interface{}, error) {
			newItem := g.NewItemFromString(args[0].(string))
			g.Player.GetInventory().AddItem(newItem)
			g.msg(foundation.HiLite("%s received.", newItem.Name()))
			return nil, nil
		},
		// CleanStains(actor, mapName, zone...) is called every turn; true once the zones are clean (or the rest is out of reach).
		"CleanStains": func(args ...interface{}) (interface{}, error) {
			actor, ok := args[0].(*Actor)
			if !ok || actor == nil {
				return false, nil
			}
			var zones []string
			for _, zone := range args[2:] {
				zones = append(zones, zone.(string))
			}
			return g.cleanStainsStep(actor, args[1].(string), zones), nil
		},
		"IsActorAt": func(args ...interface{}) (interface{}, error) {
			actor, ok := args[0].(*Actor)
			if !ok || actor == nil {
				return false, nil
			}
			mapName, locName := args[1].(string), args[2].(string)
			return actor.currentMapName == mapName && actor.Position() == g.ensureMapIsLoaded(mapName).GetNamedLocation(locName), nil
		},
		"LoadMap": func(args ...interface{}) (interface{}, error) {
			g.ensureMapIsLoaded(args[0].(string))
			return true, nil
		},
		"ProvokeFaction": func(args ...interface{}) (interface{}, error) {
			faction := args[0].(string)
			for _, actor := range g.currentMap().Actors() {
				if actor.Faction == faction && actor.IsAlive() {
					actor.FSM.SendEvent(NewProvokedEvent(g.Player))
				}
			}
			return nil, nil
		},
		// EndGame('cryo_sleep'): sets Ending(<id>) and shows the game over screen once the node's effects ran.
		"EndGame": func(args ...interface{}) (interface{}, error) {
			id := args[0].(string)
			g.gameFlags.SetFlag(fmt.Sprintf("Ending(%s)", id))
			g.pendingEnding = id
			return nil, nil
		},
		// SpawnTeamHunting('ebi_strike', 'mansion', 'taxi_stand', 'logan_faust'): the team appears at the location,
		// its leader hunts the named actor on that map (the player if he is dead or gone), the members follow.
		"SpawnTeamHunting": func(args ...interface{}) (interface{}, error) {
			teamName, mapName, locName, victimName := args[0].(string), args[1].(string), args[2].(string), args[3].(string)
			targetMap := g.ensureMapIsLoaded(mapName)
			leader, _ := g.SpawnTeam(teamName, MapPosition{MapName: mapName, LocationName: locName, Position: targetMap.GetNamedLocation(locName)})
			if leader == nil {
				g.msg(foundation.HiLite("SpawnTeamHunting: no team %s", teamName))
				return nil, nil
			}
			for _, actor := range targetMap.Actors() {
				if actor.GetInternalName() == victimName && actor.IsAlive() {
					g.Scripts.Run(g.NewScriptKill(leader, actor))
					return nil, nil
				}
			}
			leader.GetFlags().Set(foundation.FlagRelentless)
			leader.FSM.SetState(StateHunt, ActorEvent{Event: EventProvoked, Actor: g.Player})
			return nil, nil
		},
		// SpawnActor('def_name', 'map', 'location'): a new actor from the definitions appears at a named location.
		"SpawnActor": func(args ...interface{}) (interface{}, error) {
			defName, mapName, locName := args[0].(string), args[1].(string), args[2].(string)
			targetMap := g.ensureMapIsLoaded(mapName)
			actor := g.NewActorFromName(defName, mapName)
			if actor == nil {
				g.msg(foundation.HiLite("SpawnActor: no actor %s", defName))
				return nil, nil
			}
			targetMap.AddActorWithDisplacement(actor, targetMap.GetNamedLocation(locName))
			return nil, nil
		},
		// RemoveActor('internal_name'): the actor leaves the world, from whatever loaded map they are on.
		"RemoveActor": func(args ...interface{}) (interface{}, error) {
			name := args[0].(string)
			for _, m := range g.activeMaps {
				for _, actor := range m.Actors() {
					if actor.GetInternalName() == name && actor != g.Player {
						m.RemoveActor(actor)
						return nil, nil
					}
				}
			}
			return nil, nil
		},
		"PlayerAddGold": func(args ...interface{}) (interface{}, error) {
			goldAmount := int(args[0].(float64))
			g.Player.GetInventory().AddItem(g.NewGold(goldAmount))
			g.msg(foundation.HiLite("%s sat received.", strconv.Itoa(goldAmount)))
			return nil, nil
		},
	}
}

func (g *GameState) forcedCheck(rolled bool) bool {
	switch g.ForcedChecks {
	case "success":
		return true
	case "fail":
		return false
	}
	return rolled
}
