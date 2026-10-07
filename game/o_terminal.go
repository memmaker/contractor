package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"strings"
)

type Terminal struct {
	*BaseObject
	isPlayer          func(*Actor) bool
	startDialogue     func()
	DeclareAsTerminal bool
	DialogueFile      string // defaults to the internal name
}

func (g *GameState) NewTerminal(rec recfile.Record, resolver func(objType string) textiles.TextIcon) *Terminal {
	terminal := &Terminal{BaseObject: NewObject(foundation.ObjectTerminal, resolver), DeclareAsTerminal: true}
	terminal.SetWalkable(false)
	terminal.SetHidden(false)
	terminal.SetTransparent(true)
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			terminal.InternalName = field.Value
		case "dialogue":
			terminal.DialogueFile = field.Value
			if terminal.InternalName == "" {
				terminal.InternalName = field.Value
			}
		case "iconoverride":
			terminal.CustomIcon = terminal.iconForObject(field.Value)
			terminal.UseCustomIcon = true
		case "description":
			terminal.DisplayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			terminal.SetPosition(spawnPos)
		case "tags":
			if strings.ToLower(field.Value) == "no_sound" {
				terminal.DeclareAsTerminal = false
			}
		case "declared_as_terminal":
			terminal.DeclareAsTerminal = recfile.StrBool(field.Value)
		}
	}
	if terminal.DialogueFile == "" {
		terminal.DialogueFile = terminal.InternalName
	}
	terminal.InitWithGameState(g)
	return terminal
}

func (t *Terminal) AppendContextActions(actions []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	return append(actions, foundation.MenuItem{
		Name:       "Interact",
		Action:     t.startDialogue,
		CloseMenus: true,
	})
}
func (t *Terminal) InitWithGameState(g *GameState) {
	t.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	t.startDialogue = func() { g.PlayerStartDialogue(t.DialogueFile, t) }
}

func (t *Terminal) OnBump(actor *Actor) {
	if t.isPlayer(actor) {
		t.startDialogue()
	}
}

func (t *Terminal) ToRecord() recfile.Record {
	return recfile.Record{
		{Name: "category", Value: t.Category.String()},
		{Name: "description", Value: t.DisplayName},
		{Name: "name", Value: t.InternalName},
		{Name: "dialogue", Value: t.DialogueFile},
		{Name: "position", Value: t.RawPosition.Encode()},
		{Name: "icon", Value: string(t.CustomIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(t.CustomIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(t.CustomIcon.Bg)},
		{Name: "declared_as_terminal", Value: recfile.BoolStr(t.DeclareAsTerminal)},
	}
}
