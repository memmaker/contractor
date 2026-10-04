package game

import (
	"contractor/d100"
	"contractor/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
	"image/color"
)

// nullUI is the headless GameUI: it shows nothing and animates nothing.
type nullUI struct{}

func (nullUI) GetKeybindingsAsString(command string) string { return "" }
func (nullUI) QuitGame()                                    {}
func (nullUI) UpdateStats()                                 {}
func (nullUI) UpdateInventory()                             {}
func (nullUI) UpdateLogWindow()                             {}
func (nullUI) UpdateVisibleActors()                         {}
func (nullUI) SelectTarget(getCth func(target foundation.ActorForUI) foundation.AttackInfo, onSelected func(targetPos geometry.Point)) {
}
func (nullUI) SelectDirection(onSelected func(direction geometry.CompassDirection)) {}
func (nullUI) SelectBodyPart(previousAim d100.BodyPart, onSelected func(victim foundation.ActorForUI, hitZone d100.BodyPart)) {
}
func (nullUI) AskForString(prompt string, prefill string, result func(entered string))          {}
func (nullUI) AskForConfirmation(title string, message string, onConfirm func(didConfirm bool)) {}
func (nullUI) OpenInventoryForManagement()                                                      {}
func (nullUI) OpenInventoryForSelection(stack []foundation.Item, prompt string, onSelected func(item foundation.Item)) {
}
func (nullUI) OpenInventoryForSelectionWithClose(stack []foundation.Item, prompt string, onSelected func(item foundation.Item), done func()) {
}
func (nullUI) OpenTextWindow(description string)                             {}
func (nullUI) ShowTextFileFullscreen(filename string, onClose func())        {}
func (nullUI) OpenMenu(actions []foundation.MenuItem)                        {}
func (nullUI) OpenMenuWithTitle(title string, actions []foundation.MenuItem) {}
func (nullUI) OpenMenuWithTitleAndClose(title string, actions []foundation.MenuItem, onClose func()) {
}
func (nullUI) OpenKeypad(specialAction string, correctSequence []rune, onSpecialAction func() bool, onCompletion func(success bool)) {
}
func (nullUI) OpenVendorMenu(title string, itemsForSale []foundation.Item, buyItem func(ui foundation.Item, amount int, price int), inspect func(item foundation.Item), onClose func()) {
}
func (nullUI) ShowGameOver(score foundation.ScoreInfo, highScores []foundation.ScoreInfo) {}
func (nullUI) ShowTakeOnlyContainer(name string, containedItems []foundation.Item, transfer func(ui foundation.Item)) {
}
func (nullUI) ShowGiveAndTakeContainer(leftName string, leftItems []foundation.Item, rightName string, rightItems []foundation.Item, transferToLeft func(itemTaken foundation.Item, amount int), transferToRight func(itemTaken foundation.Item, amount int), takeAll func()) {
}
func (nullUI) OpenAimedShotPicker(actorAt foundation.ActorForUI, previousAim d100.BodyPart, onSelected func(victim foundation.ActorForUI, hitZone d100.BodyPart)) {
}
func (nullUI) SelectSaveName()                               {}
func (nullUI) SelectLoadName()                               {}
func (nullUI) AfterPlayerMoved(moveInfo foundation.MoveInfo) {}
func (nullUI) SetConversationState(text string, options []foundation.MenuItem, conversationPartner foundation.ChatterSource, isTerminal bool) {
}
func (nullUI) CloseConversation()                                              {}
func (nullUI) TryAddChatter(victim foundation.ChatterSource, text string) bool { return false }
func (nullUI) SetColors(palette textiles.ColorPalette, colors map[foundation.ItemCategory]color.RGBA) {
}
func (nullUI) ClearOverlays()                                                                   {}
func (nullUI) PlayMusic(fileName string)                                                        {}
func (nullUI) PlayCue(cue string)                                                               {}
func (nullUI) SetSneakOverlay(overlay map[geometry.Point]fxtools.HDRColor)                      {}
func (nullUI) FadeToBlack()                                                                     {}
func (nullUI) FadeFromBlack()                                                                   {}
func (nullUI) ForceUIRedraw()                                                                   {}
func (nullUI) IndicateConversationStartByNPC(partner foundation.ChatterSource, done func())     {}
func (nullUI) ForceListening(partner foundation.ChatterSource, monologue []string, done func()) {}
func (nullUI) AddAnimations(animations []foundation.Animation) {
	for _, a := range animations {
		if n, ok := a.(*headlessAnim); ok {
			n.finish()
		}
	}
}
func (nullUI) AnimatePending() (cancelled bool) { return false }
func (nullUI) SkipAnimations()                  {}
func (nullUI) GetAnimThrow(item foundation.Item, origin geometry.Point, target geometry.Point) (foundation.Animation, int) {
	return nopAnim(nil), 0
}
func (nullUI) GetAnimDamage(spreadBlood func(mapPos geometry.Point), actorPos geometry.Point, damage int, bullets int) foundation.Animation {
	return nopAnim(nil)
}
func (nullUI) GetAnimMove(actor foundation.ActorForUI, old geometry.Point, new geometry.Point) foundation.Animation {
	return nopAnim(nil)
}
func (nullUI) GetAnimQuickMove(actor foundation.ActorForUI, path []geometry.Point) foundation.Animation {
	return nopAnim(nil)
}
func (nullUI) GetAnimAttack(attacker, defender foundation.ActorForUI) foundation.Animation {
	return nopAnim(nil)
}
func (nullUI) GetAnimMuzzleFlash(position geometry.Point, flashColor fxtools.HDRColor, radius int, bulletCount int, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimProjectile(icon rune, colorName string, origin geometry.Point, dest geometry.Point, done func()) (foundation.Animation, int) {
	return nopAnim(done), 0
}
func (nullUI) GetAnimProjectileWithTrail(leadIcon rune, colorNames []string, path []geometry.Point, done func()) (foundation.Animation, int) {
	return nopAnim(done), 0
}
func (nullUI) GetAnimProjectileWithLight(leadIcon rune, lightColorName string, pathOfFlight []geometry.Point, done func()) (foundation.Animation, int) {
	return nopAnim(done), 0
}
func (nullUI) GetAnimTiles(positions []geometry.Point, frames []textiles.TextIcon, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimTeleport(actor foundation.ActorForUI, origin geometry.Point, targetPos geometry.Point, appearOnMap func()) (vanishAnim, appearAnim foundation.Animation) {
	return nopAnim(nil), nopAnim(appearOnMap)
}
func (nullUI) GetAnimRadialReveal(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimRadialAlert(position geometry.Point, dijkstra map[geometry.Point]int, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimUncloakAtPosition(actor foundation.ActorForUI, position geometry.Point) (foundation.Animation, int) {
	return nopAnim(nil), 0
}
func (nullUI) GetAnimExplosion(points []geometry.Point, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimRadialExplosion(points map[geometry.Point]int, lightColor fxtools.HDRColor, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimEnchantArmor(actor foundation.ActorForUI, position geometry.Point, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimEnchantWeapon(actor foundation.ActorForUI, position geometry.Point, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimVorpalizeWeapon(origin geometry.Point, done func()) []foundation.Animation {
	return nil
}
func (nullUI) GetAnimConfuse(position geometry.Point, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimEnterCombat(actor foundation.ActorForUI, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimSuspicious(actor foundation.ActorForUI, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimBreath(flight []geometry.Point, done func()) []foundation.Animation { return nil }
func (nullUI) GetAnimBackgroundColor(position geometry.Point, colorName string, frameCount int, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimAppearance(actor foundation.ActorForUI, position geometry.Point, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimWakeUp(position geometry.Point, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimEvade(defender foundation.ActorForUI, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) GetAnimLaser(path []geometry.Point, lightColor fxtools.HDRColor, done func()) foundation.Animation {
	return nopAnim(done)
}
func (nullUI) StartGameLoop()  {}
func (nullUI) StartWithIntro() {}

// headlessAnim stands in for every animation headless: it finishes as soon as it is added.
type headlessAnim struct {
	done     []func()
	followUp []foundation.Animation
	finished bool
}

func nopAnim(done func()) *headlessAnim { return &headlessAnim{done: []func(){done}} }

func (n *headlessAnim) finish() {
	if n.finished {
		return
	}
	n.finished = true
	for _, d := range n.done {
		if d != nil {
			d()
		}
	}
	nullUI{}.AddAnimations(n.followUp)
}
func (n *headlessAnim) IsDone() bool                         { return n.finished }
func (n *headlessAnim) SetFollowUp(f []foundation.Animation) { n.followUp = f }
func (n *headlessAnim) SetDoneCallback(d func())             { n.done = append(n.done, d) }
func (n *headlessAnim) RequestMapUpdateOnFinish()            {}
func (n *headlessAnim) SetAudioCue(string)                   {}
