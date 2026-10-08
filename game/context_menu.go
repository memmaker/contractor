package game

import (
	"contractor/d100"
	"contractor/foundation"
	"contractor/gridmap"
	"fmt"
)

func (g *GameState) animatedActionFromMenu(action func()) func() {
	return func() {
		g.ui.ForceUIRedraw()
		action()
	}
}

func (g *GameState) canPlayerTalkToActor(actor *Actor) bool {
	distance := g.currentMap().MoveDistance(g.Player.Position(), actor.Position())
	return g.Player.CanSee(actor.Position()) && actor.HasDialogue() && !actor.IsSleeping() && distance <= 6
}

func (g *GameState) appendContextActionsForActor(buffer []foundation.MenuItem, actor *Actor) []foundation.MenuItem {
	distance := g.currentMap().MoveDistance(g.Player.Position(), actor.Position())

	if g.canPlayerTalkToActor(actor) {
		buffer = append(buffer, foundation.MenuItem{
			Name:       "[white]Talk To[-]",
			Action:     func() { g.PlayerStartDialogue(actor.GetDialogueFile(), actor) },
			CloseMenus: true,
		})
	}
	buffer = append(buffer, foundation.MenuItem{
		Name:       "Look at",
		Action:     func() { g.ui.OpenTextWindow(actor.GetDetailInfo()) },
		CloseMenus: true,
	})

	if g.Player.HasPerk(d100.PerkDisarm) && actor.GetInventory().HasWeaponEquipped() && distance == 1 {
		chances := formatContest(g.Player.unarmedChance(), actor.unarmedChance())
		label := fmt.Sprintf("Disarm (%s)", chances)
		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: g.animatedActionFromMenu(func() {
				g.actorDisarm(g.Player, actor)
			}),
			CloseMenus: true,
		})
	}

	if actor.IsHostileTowards(g.Player) || distance > 1 {
		return buffer
	}
	if g.Player.HasPerk(d100.PerkPickpocket) && !actor.HasFlag(foundation.FlagAnimal) {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Pickpocket",
			Action: g.animatedActionFromMenu(func() {
				g.StartPickpocket(actor)
			}),
			CloseMenus: true,
		})
	}

	if !actor.IsSleeping() {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Melee Attack",
			Action: g.animatedActionFromMenu(func() {
				g.playerMeleeAttack(actor)
			}),
			CloseMenus: true,
		})
		if g.Player.HasPerk(d100.PerkNonLethalTakeDown) {
			nonLethalChanceString := formatContest(g.Player.unarmedChance(), actor.unarmedChance())
			buffer = append(buffer, foundation.MenuItem{
				Name: fmt.Sprintf("Non-Lethal Takedown (%s)", nonLethalChanceString),
				Action: g.animatedActionFromMenu(func() {
					g.playerNonLethalTakedown(actor)
				}),
				CloseMenus: true,
			})
		}
	} else {
		buffer = append(buffer, foundation.MenuItem{
			Name: "Wake Up",
			Action: func() {
				g.msg(foundation.Msg("You shake the sleeping figure awake."))
				actor.WakeUp()
				g.endPlayerTurn(g.Player.TimeNeededForActions())
			},
			CloseMenus: true,
		})
	}

	if g.currentMap().IsPositionNextToTileWithFlag(actor.Position(), gridmap.TileFlagWater) {
		label := "Drown"
		if !actor.IsSleeping() {
			drownChanceString := formatContest(g.Player.unarmedChance(), actor.unarmedChance())
			label = fmt.Sprintf("Drown (%s)", drownChanceString)
		}

		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: g.animatedActionFromMenu(func() {
				g.playerDrown(actor)
			}),
			CloseMenus: true,
		})
	}

	if g.Player.HasPerk(d100.PerkBackstab) && g.Player.GetInventory().HasMeleeWeaponEquipped() {
		label := "Backstab"
		if !actor.IsSleeping() {
			stabChanceString := formatContest(d100.Percentage(g.Player.GetCharSheet().GetSkill(d100.SkillForBackstabbing)), actor.awarenessChance())
			label = fmt.Sprintf("Backstab (%s)", stabChanceString)
		}
		buffer = append(buffer, foundation.MenuItem{
			Name: label,
			Action: g.animatedActionFromMenu(func() {
				g.playerBackstab(actor)
			}),
			CloseMenus: true,
		})
	}
	return buffer
}

func formatContest(one, two d100.Percentage) string {
	return fmt.Sprintf("%d%% vs %d%%", int(one), int(two))
}

func (g *GameState) OpenContextMenuForItem(uiItem foundation.Item, done func()) {
	item := uiItem
	contextActions := []foundation.MenuItem{
		{Name: "Inspect", Action: func() {
			g.PlayerExamineItem(item)
		}},
	}

	if len(contextActions) == 0 {
		return
	}
	if len(contextActions) == 1 {
		contextActions[0].Action()
		return
	}
	g.ui.OpenMenu(contextActions)
}
