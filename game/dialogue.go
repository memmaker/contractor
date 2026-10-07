package game

import (
	"contractor/foundation"
	"fmt"
	"github.com/memmaker/go/convo"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"path/filepath"
	"strings"
)

func (g *GameState) PlayerStartDialogue(dialogueFile string, partner foundation.ChatterSource) {
	conversationFilename := filepath.Join(g.config.DataRootDir, "dialogues", dialogueFile+".rec")
	if !fxtools.FileExists(conversationFilename) {
		g.msg(foundation.HiLite("%s has nothing to say.", partner.Name()))
		return
	}
	conversation, err := convo.ParseConversation(conversationFilename, g)
	if err != nil {
		g.msg(foundation.Msg(fmt.Sprintf("Dialogue error in %s: %v", dialogueFile, err)))
		return
	}

	var npcName string
	params := make(map[string]interface{})
	if actor, isActor := partner.(*Actor); isActor {
		npcName = actor.GetInternalName()
		talkedFlagName := fmt.Sprintf("TalkedTo(%s)", npcName)
		g.gameFlags.Increment(talkedFlagName)
		params["NPC"] = actor
	} else if named, ok := partner.(interface{ GetInternalName() string }); ok {
		npcName = named.GetInternalName()
	} else {
		npcName = partner.Name()
	}
	params["NPC_NAME"] = npcName

	conversation.MergeVariables(params)
	rootNode := conversation.GetRootNode()

	g.dialogueOptionIDs = g.loadDialogueOptionIDs(conversationFilename)
	state := conversation.OpenDialogueNode(rootNode)
	g.updateDialogueState(conversation, state, partner)
}

// loadDialogueOptionIDs maps "node|option text" to the option's o_id.
// convo ignores unknown o_ fields, so we read them here.
func (g *GameState) loadDialogueOptionIDs(filename string) map[string]string {
	ids := make(map[string]string)
	records, _ := recfile.ReadMultiAndClose(fxtools.MustOpen(filename))
	for _, node := range records["Nodes"] {
		nodeName, optionText := "", ""
		for _, field := range node {
			switch strings.ToLower(field.Name) {
			case "name":
				nodeName = field.Value
			case "o_text":
				optionText = g.FillTemplatedText(strings.TrimSpace(field.Value))
			case "o_id":
				ids[nodeName+"|"+optionText] = field.Value
			}
		}
	}
	return ids
}

func (g *GameState) NPCStartDialogue(dialogueFile string, partner foundation.ChatterSource, tryInitiateWith string, isTerminal bool) bool {
	conversationFilename := filepath.Join(g.config.DataRootDir, "dialogues", dialogueFile+".rec")
	if !fxtools.FileExists(conversationFilename) {
		g.msg(foundation.HiLite("%s has nothing to say.", partner.Name()))
		return false
	}
	conversation, err := convo.ParseConversation(conversationFilename, g)
	if err != nil {
		g.msg(foundation.Msg(fmt.Sprintf("Dialogue error in %s: %v", dialogueFile, err)))
		return false
	}

	var npcName string
	params := make(map[string]interface{})
	if actor, isActor := partner.(*Actor); isActor {
		npcName = actor.GetInternalName()
		talkedFlagName := fmt.Sprintf("TalkedTo(%s)", npcName)
		g.gameFlags.Increment(talkedFlagName)
		params["NPC"] = actor
	} else {
		npcName = partner.Name()
	}
	params["NPC_NAME"] = npcName

	conversation.MergeVariables(params)

	opening := conversation.GetOpeningBranchByName(tryInitiateWith)

	openingCondition, openingErr := opening.BranchCondition.Evaluate(conversation.Variables)
	asBool, isBool := openingCondition.(bool)
	if asBool && isBool && openingErr == nil {
		optionIDs := g.loadDialogueOptionIDs(conversationFilename)
		g.QueueActionAfterAnimation(func() {
			g.dialogueOptionIDs = optionIDs
			g.msg(foundation.HiLite("%s is addressing you.", partner.Name()))
			g.ui.IndicateConversationStartByNPC(partner, func() {
				state := conversation.OpenDialogueNode(opening.BranchName)
				g.updateDialogueState(conversation, state, partner)
			})
		})
		return true
	}

	return false
}

func (g *GameState) updateDialogueState(conversation *convo.Conversation, state convo.ConversationState, partner foundation.ChatterSource) {
	for _, effect := range state.Effects {
		if g.ApplyNodeEffect(effect, partner) {
			state.Flow = convo.ConversationEndInstantlyWithChatter
		}
	}
	if g.pendingEnding != "" { // EndGame('<id>') ran as a node effect
		ending := g.pendingEnding
		g.pendingEnding = ""
		g.ui.CloseConversation()
		g.endGame(ending)
		return
	}
	var menuItems []foundation.MenuItem
	switch state.Flow {
	case convo.ConversationEndInstantlyWithChatter:
		g.ui.CloseConversation()
		if actor, isActor := partner.(*Actor); isActor {
			g.tryAddChatter(actor, state.NPCText)
		}
		return
	case convo.ConversationEndNormal:
		menuItems = append(menuItems, foundation.MenuItem{
			Name:       "<Leave>",
			Action:     g.ui.CloseConversation,
			CloseMenus: true,
		})
	default:
		for _, option := range state.PlayerOptions {
			menuItems = append(menuItems, foundation.MenuItem{
				ID:   g.dialogueOptionIDs[state.NodeName+"|"+option.PlayerText],
				Name: option.PlayerText,
				Action: func() {
					nextState := conversation.OpenDialogueNode(option.FollowUpBranch)
					openNext := func() {
						g.updateDialogueState(conversation, nextState, partner)
					}
					if option.Effect == "" {
						openNext()
					} else {
						g.ApplyOptionEffect(option.Effect, partner, openNext)
					}
				},
				CloseMenus: true,
			})
		}
	}

	isTerminal := false
	if _, ok := partner.(*Actor); !ok {
		isTerminal = true
	}

	g.ui.SetConversationState(state.NPCText, menuItems, partner, isTerminal)
}

// ApplyNodeEffect handles the parameterless game effects of a dialogue node.
// Returns true if the effect ends the conversation.
func (g *GameState) ApplyNodeEffect(effect string, conversationPartner convo.ConversationPartner) (endsConversation bool) {
	actor, isActor := conversationPartner.(*Actor)
	if !isActor {
		return false
	}
	switch effect {
	case "StartCombat":
		actor.FSM.SendEvent(NewProvokedEvent(g.Player))
		return true
	case "EndHostility":
		actor.FSM.SendEvent(NewCalmedEvent(g.Player))
	case "HealPlayer":
		g.Player.Heal(g.Player.GetHitPointsMax())
		g.msg(foundation.Msg("You have been healed."))
		g.updateUIStatus()
	}
	return false
}

func (g *GameState) ApplyOptionEffect(effect string, conversationPartner convo.ConversationPartner, followUp func()) {
	actor, isActor := conversationPartner.(*Actor)
	if !isActor {
		return
	}
	switch effect {
	case "StartTrading":
		g.openVendorMenu(actor, followUp)
	case "StartRepair":
		g.openNPCRepairMenu(actor, followUp)
	case "StartCyberware":
		g.openCyberwareMenu(actor, followUp)
	}
}
