package game

import (
	"contractor/d100"
	"contractor/foundation"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

// Autoplay drives the player from a playtest script (data/playtests/<name>.rec).
// It uses the same frames/outcomes format as NPC scripts. Frames run top to bottom:
// a frame's do: verbs queue intents (actions that may take many turns, like GoTo)
// and answers (picks for the menus, dialogues and containers the game opens).
// The first matching outcome ends the run: passed, unless its do: calls Fail().
type Autoplay struct {
	Name      string
	MaxTurns  int
	Out       io.Writer
	Result    string // empty while running
	Passed    bool
	lastTrace string

	g        *GameState
	file     string
	script   ActionScript
	frame    int
	intents  []intent
	answers  []answer
	modal    *modal
	steps    int
	stuck    int
	tries    int
	logSeen  int
	gameOver bool
}

type intent struct {
	desc string
	step func() (done bool)
}

type answer struct {
	kind   string // choose, take, put, target
	label  string
	count  int
	target *Actor // the victim a target answer points at
}

func (a answer) String() string {
	if a.count > 0 {
		return fmt.Sprintf("%s(%s, %d)", a.kind, a.label, a.count)
	}
	return fmt.Sprintf("%s(%s)", a.kind, a.label)
}

type modal struct {
	kind         string // menu, conversation, container, confirm, target, wait
	title        string
	items        []foundation.MenuItem
	onClose      func()
	confirm      func(bool)
	target       func(geometry.Point)
	done         func()
	mine, theirs []foundation.Item
	take, put    func(foundation.Item, int)
	buy          func(foundation.Item, int, int)
	text         func(string)
}

// NewAutoplay prepares a run. The first Step (after the game's UIReady) hooks into the UI and loads the script.
func NewAutoplay(g *GameState, scriptFile string, out io.Writer) *Autoplay {
	return &Autoplay{g: g, Out: out, MaxTurns: 5000, file: scriptFile, Name: strings.TrimSuffix(filepath.Base(scriptFile), ".rec")}
}

// NewHeadlessUI is a UI that draws nothing; the autoplayer answers its modals.
func NewHeadlessUI() foundation.GameUI { return nullUI{} }

// Step advances the run by one decision. Returns true once the run is over.
func (a *Autoplay) Step() (over bool) {
	if a.Result != "" {
		return true
	}
	defer func() {
		if r := recover(); r != nil {
			a.finish(false, fmt.Sprintf("panic: %v\n%s", r, debug.Stack()))
			over = true
		}
		a.flushGameLog()
	}()
	if a.steps == 0 {
		a.g.ui = autoplayUI{GameUI: a.g.ui, a: a}
		records, _ := recfile.ReadMultiAndClose(fxtools.MustOpen(a.file))
		a.script = NewActionScript(a.Name, records, mergeMaps(a.g.GetScriptFuncs(), a.verbs()))
	}
	a.steps++
	a.trace()
	vars := a.script.Variables
	for _, outcome := range a.script.Outcomes {
		if outcome.Condition(vars) {
			outcome.ExecuteActions(vars)
			if a.Result == "" {
				a.finish(true, "outcome: "+outcome.String())
			}
			return true
		}
	}
	switch {
	case a.gameOver || !a.g.Player.IsAlive():
		a.finish(false, "game over before any outcome")
	case a.g.TurnCount() > a.MaxTurns || a.steps > a.MaxTurns*4:
		a.finish(false, fmt.Sprintf("turn cap %d reached (frame %d/%d, %d intents left)", a.MaxTurns, a.frame, len(a.script.Frames), len(a.intents)))
	case a.modal != nil:
		a.answerModal()
	case a.defend():
	case len(a.intents) > 0:
		if a.intents[0].step() {
			a.intents = a.intents[1:]
			a.stuck, a.tries = 0, 0
		}
	case a.frame < len(a.script.Frames) && a.script.Frames[a.frame].Condition(vars):
		a.logf("frame %d: %s", a.frame+1, a.script.Frames[a.frame])
		a.script.Frames[a.frame].ExecuteActions(vars)
		a.frame++
	default:
		a.g.PlayerRest(false, time.Minute) // frame condition not met yet, or out of frames: wait, like a player would
	}
	return a.Result != ""
}

// Reading is true while a dialogue is waiting for its answer (visual mode lingers on those).
func (a *Autoplay) Reading() bool { return a.modal != nil && a.modal.kind == "conversation" }

func (a *Autoplay) finish(passed bool, why string) {
	if a.Result != "" {
		return
	}
	a.Passed, a.Result = passed, why
	if len(a.answers) > 0 {
		a.Result += fmt.Sprintf(" [unused answers: %v]", a.answers)
	}
	status := "FAIL"
	if passed {
		status = "PASS"
	}
	a.logf("%s %s after %d turns, %s game time: %s", status, a.Name, a.g.TurnCount(), a.g.gameTime.Time.Format("Mon 15:04"), a.Result)
}

func (a *Autoplay) fail(format string, args ...interface{}) bool {
	a.finish(false, fmt.Sprintf(format, args...))
	return true
}

func (a *Autoplay) logf(format string, args ...interface{}) {
	if a.Out != nil {
		fmt.Fprintf(a.Out, "[%5d] %s\n", a.g.TurnCount(), fmt.Sprintf(format, args...))
	}
}

func (a *Autoplay) flushGameLog() {
	for ; a.logSeen < len(a.g.logBuffer); a.logSeen++ {
		if text := a.g.logBuffer[a.logSeen].ToPlainText(); a.Out != nil {
			fmt.Fprintf(a.Out, "        . %s\n", text)
		}
	}
}

func (a *Autoplay) answerModal() {
	m := a.modal
	var next *answer
	if len(a.answers) > 0 {
		next = &a.answers[0]
	}
	pop := func() { a.answers = a.answers[1:] }
	switch m.kind {
	case "wait":
		a.modal = nil
		if m.done != nil {
			m.done()
		}
	case "confirm":
		a.modal = nil
		a.logf("confirm %q: yes", m.title)
		m.confirm(true)
	case "target":
		if next == nil || next.kind != "target" {
			a.fail("game asks for a target, script has no Kill() pending")
			return
		}
		pop()
		a.modal = nil
		m.target(next.target.Position())
	case "container":
		if next == nil || (next.kind != "take" && next.kind != "put") {
			a.modal = nil // done with this container
			return
		}
		pop()
		from, transfer := m.theirs, m.take
		if next.kind == "put" {
			from, transfer = m.mine, m.put
		}
		if transfer == nil {
			a.fail("%s: cannot put items into %s", next, m.title)
			return
		}
		for _, item := range from {
			if item.GetInternalName() == next.label {
				amount := next.count
				if amount <= 0 || amount > item.GetStackSize() {
					amount = item.GetStackSize()
				}
				a.logf("%s: %s x%d", m.title, next.kind, amount)
				a.modal = nil
				transfer(item, amount) // the game reopens the container
				return
			}
		}
		a.fail("%s: no %q to %s (container: %s, player: %s)", m.title, next.label, next.kind, itemNames(m.theirs), itemNames(m.mine))
	case "vendor":
		if next == nil || next.kind != "buy" {
			a.modal = nil
			if m.onClose != nil {
				m.onClose()
			}
			return
		}
		pop()
		for _, item := range m.theirs {
			if item.GetInternalName() == next.label {
				amount := max(1, next.count)
				a.logf("%s: buy %s x%d", m.title, next.label, amount)
				a.modal = nil
				m.buy(item, amount, item.GetPrice()*amount) // the game reopens the vendor menu
				return
			}
		}
		a.fail("%s: no %q for sale (%s)", m.title, next.label, itemNames(m.theirs))
	case "string":
		if next == nil || next.kind != "text" {
			a.fail("game asks %q, script has no answer", m.title)
			return
		}
		pop()
		a.logf("%s: %s", m.title, next.label)
		a.modal = nil
		m.text(next.label)
	default: // menu, conversation
		if next != nil && next.kind == "leave" && m.kind == "conversation" {
			pop()
			a.logf("  > <escape>")
			a.g.ui.CloseConversation()
			return
		}
		for _, item := range m.items {
			if next != nil && next.kind == "choose" && (item.ID == next.label || (m.kind == "menu" && strings.Contains(item.Name, next.label))) {
				pop()
				a.pick(item)
				return
			}
		}
		if len(m.items) == 1 && m.items[0].Name == "<Leave>" {
			a.pick(m.items[0])
			return
		}
		a.fail("unanswered %s %q, next answer %v, options:\n%s", m.kind, m.title, next, optionList(m.items))
	}
}

func (a *Autoplay) pick(item foundation.MenuItem) {
	m := a.modal
	a.modal = nil
	a.logf("  > %s", item.Name)
	item.Action()
	if m.onClose != nil && item.CloseMenus && a.modal == nil {
		m.onClose()
	}
}

func optionList(items []foundation.MenuItem) string {
	var lines []string
	for _, item := range items {
		id := item.ID
		if id == "" {
			id = "<no o_id>"
		}
		lines = append(lines, fmt.Sprintf("    %-20s %s", id, item.Name))
	}
	return strings.Join(lines, "\n")
}

func itemNames(items []foundation.Item) string {
	var names []string
	for _, item := range items {
		names = append(names, fmt.Sprintf("%s(%d)", item.GetInternalName(), item.GetStackSize()))
	}
	return strings.Join(names, ", ")
}

// withAnswer queues an answer in front, runs a game action, and drops the answer if no modal opened.
func (a *Autoplay) withAnswer(ans answer, action func()) {
	a.answers = append([]answer{ans}, a.answers...)
	action()
	wants := map[string][]string{"target": {"target"}, "choose": {"menu", "conversation"}, "text": {"string"}, "buy": {"vendor"}, "take": {"container"}}[ans.kind]
	wrongModal := a.modal != nil && !slices.Contains(wants, a.modal.kind) // e.g. a jam-confirm opened instead of the target picker
	if (a.modal == nil || wrongModal) && len(a.answers) > 0 && a.answers[0] == ans {
		a.answers = a.answers[1:]
	}
}

// --- locating things by name, never by position ---

// nearestActorWithName picks the closest living actor of that name we can get next to (there are many rats, some swim).
func (a *Autoplay) nearestActorWithName(name string) *Actor {
	m := a.g.currentMap()
	reach := m.GetDijkstraMapWithActorsNotBlocking(a.g.Player, 400)
	reachable := func(p geometry.Point) bool {
		return len(m.NeighborsAll(p, func(n geometry.Point) bool { _, ok := reach[n]; return ok || n == a.g.Player.Position() })) > 0
	}
	var best, fallback *Actor
	for _, actor := range m.Actors() {
		if actor.GetInternalName() != name || !actor.IsAlive() {
			continue
		}
		fallback = actor
		if reachable(actor.Position()) &&
			(best == nil || geometry.Distance(a.g.Player.Position(), actor.Position()) < geometry.Distance(a.g.Player.Position(), best.Position())) {
			best = actor
		}
	}
	if best == nil {
		return fallback // caged or across water: shoot from where we can
	}
	return best
}

func (a *Autoplay) object(name string) Object {
	for _, obj := range a.g.currentMap().Objects() {
		if obj.GetInternalName() == name {
			return obj
		}
	}
	return nil
}

func (a *Autoplay) locate(name string) (geometry.Point, bool) {
	if actor := a.g.actorWithName(name); actor != nil {
		return actor.Position(), true
	}
	if obj := a.object(name); obj != nil {
		return obj.Position(), true
	}
	return a.g.currentMap().TryGetNamedLocation(name)
}

// walkStep moves the player one step towards target (or next to it).
// Map transitions on the way are avoided unless they are the target.
func (a *Autoplay) walkStep(target geometry.Point, adjacent bool) (arrived bool) {
	g := a.g
	m := g.currentMap()
	pos := g.Player.Position()
	dist := m.MoveDistance(pos, target)
	if dist == 0 || (dist == 1 && (adjacent || m.IsActorAt(target))) {
		return true
	}
	guarded := func(zone string) bool {
		return zone != "" && slices.ContainsFunc(m.Actors(), func(actor *Actor) bool { return actor.IsAlive() && actor.IsGuarding(zone) })
	}
	// guarded zones are only entered on purpose: when the script names a spot or object inside one. People in there are talked to across the counter.
	deliberate := m.FirstZoneAt(target) == m.FirstZoneAt(pos) || !m.IsActorAt(target)
	trespass := func(p geometry.Point) bool {
		zone := m.FirstZoneAt(p)
		return guarded(zone) && zone != m.FirstZoneAt(pos) && !(deliberate && zone == m.FirstZoneAt(target))
	}
	path := m.GetAStarPath(pos, target, func(p geometry.Point) bool {
		_, corpse := m.TryGetDownedActorAt(p) // stepping on a corpse opens its loot and would eat a pending answer
		obj, hasObj := m.TryGetObjectAt(p)    // bumping a bed or terminal on the way opens its menu, doors are fine
		furniture := hasObj && !(obj.GetCategory() >= foundation.ObjectLockedDoor && obj.GetCategory() <= foundation.ObjectBrokenDoor)
		return p == target || (!m.IsTransitionAt(p) && !corpse && !furniture && !trespass(p) && m.IsWalkableFor(p, g.Player)) // the target may be an occupied or caged tile
	})
	if len(path) == 0 && adjacent { // unreachable (caged, in water): get as close as the map allows
		reach := m.GetDijkstraMapWithActorsNotBlocking(g.Player, 400)
		// a spot with a line of sight beats any closer spot behind a wall (talking across a counter)
		score := func(p geometry.Point) int {
			d := m.MoveDistance(p, target)
			if !m.IsLineOfSightClear(p, target, func(q geometry.Point) bool { return !m.IsTransparent(q) }) {
				d += 100
			}
			return d
		}
		closest, best := pos, score(pos)
		for p, cost := range reach {
			better := func(d int) bool { // full ordering: map iteration must not decide ties, runs stay reproducible
				c := reach[closest]
				return d < best || (d == best && (cost < c || (cost == c && (p.Y < closest.Y || (p.Y == closest.Y && p.X < closest.X)))))
			}
			if d := score(p); cost > 0 && better(d) && !m.IsActorAt(p) && !m.IsTransitionAt(p) && !trespass(p) {
				closest, best = p, d
			}
		}
		if closest == pos {
			return true
		}
		path = m.GetAStarPath(pos, closest, func(p geometry.Point) bool { return !m.IsTransitionAt(p) && m.IsWalkableFor(p, g.Player) })
	}
	// ponytail: replans every step, cache the path if long walks get slow
	for _, n := range m.NeighborsAll(pos, func(p geometry.Point) bool {
		return m.IsActorAt(p) && m.ActorAt(p).IsAlive() && (m.ActorAt(p).IsHostileTowards(g.Player) || m.ActorAt(p).Aggressive)
	}) {
		a.bump(n) // a hostile next to us gets attacked first, like a player would
		return false
	}
	if len(path) == 0 || m.IsActorAt(path[0]) {
		a.stuck++
		g.Wait()
		return false
	}
	step := path[0].Sub(pos)
	step = geometry.Point{X: clamp1(step.X), Y: clamp1(step.Y)}
	g.ManualMovePlayer(step.ToDirection())
	if g.Player.Position() == pos {
		a.stuck++
	} else {
		a.stuck = 0
	}
	return false
}

func clamp1(v int) int { return max(-1, min(1, v)) }

func (a *Autoplay) bump(target geometry.Point) {
	step := target.Sub(a.g.Player.Position())
	a.g.ManualMovePlayer(geometry.Point{X: clamp1(step.X), Y: clamp1(step.Y)}.ToDirection())
}

// queue adds an intent; stuck walking for too long fails the run.
// defend: a visible actor attacking the player gets a Kill intent put in front of everything else, like a player would react.
func (a *Autoplay) defend() bool {
	if len(a.intents) > 0 && strings.HasPrefix(a.intents[0].desc, "Kill ") {
		return false
	}
	attacker := a.g.playerAttacker()
	if attacker == nil {
		return false
	}
	rest := a.intents
	a.intents = nil
	a.verbs()["Kill"](attacker.GetInternalName())
	a.intents = append(a.intents, rest...)
	a.logf("defending against %s", attacker.GetInternalName())
	return true
}

// waitUntil: rest until a game time, resuming after interruptions (a fight in between is handled by defend).
func (a *Autoplay) waitUntil(desc string, at func() time.Time) (interface{}, error) {
	var target time.Time
	return a.queue(desc, func() bool {
		if target.IsZero() {
			target = at()
		}
		if left := target.Sub(a.g.gameTime.Time); left > 0 {
			a.g.PlayerRest(false, left)
			return false
		}
		return true
	})
}

// trace: AUTOPLAY_TRACE=<internal name> logs that actor's map, position and state on every step and world tick.
func (a *Autoplay) trace() {
	name := os.Getenv("AUTOPLAY_TRACE")
	if name == "" {
		return
	}
	for mapName, m := range a.g.activeMaps {
		for _, actor := range m.Actors() {
			if actor.GetInternalName() == name {
				line := fmt.Sprintf("TRACE %s on %s at %v zone %q state %s hp=%d", name, mapName, actor.Position(), m.FirstZoneAt(actor.Position()), actor.FSM.State().ToString(), actor.GetHitPoints())
				if line != a.lastTrace {
					a.lastTrace = line
					a.logf("%s %s", a.g.gameTime.Time.Format("Mon 15:04:05"), line)
				}
			}
		}
	}
}

func (a *Autoplay) queue(desc string, step func() bool) (interface{}, error) {
	a.intents = append(a.intents, intent{desc: desc, step: func() bool {
		if a.stuck > 60 {
			return a.fail("stuck: %s on %s", desc, a.g.currentMapName)
		}
		return step()
	}})
	return true, nil
}

func (a *Autoplay) reply(kind, label string, args []interface{}) (interface{}, error) {
	ans := answer{kind: kind, label: label}
	if len(args) > 1 {
		ans.count = int(args[1].(float64))
	}
	a.answers = append(a.answers, ans)
	return true, nil
}

// approach walks next to a named actor or object on the current map, then runs onArrival once.
func (a *Autoplay) approach(desc, name string, onArrival func() bool) (interface{}, error) {
	return a.queue(desc, func() bool {
		pos, found := a.locate(name)
		if !found {
			return a.fail("%s: %q is not on %s", desc, name, a.g.currentMapName)
		}
		if a.walkStep(pos, true) {
			return onArrival()
		}
		return false
	})
}

func (a *Autoplay) verbs() map[string]govaluate.ExpressionFunction {
	g := a.g
	str := func(args []interface{}, i int) string { return args[i].(string) }
	num := func(args []interface{}, i int) float64 { return args[i].(float64) }
	return map[string]govaluate.ExpressionFunction{
		// GoTo('named_location'): walk there. A transition as target takes you to the next map.
		"GoTo": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			startMap := ""
			return a.queue("GoTo "+name, func() bool {
				if startMap == "" {
					startMap = g.currentMapName
				} else if startMap != g.currentMapName {
					return true // went through a transition
				}
				pos, found := g.currentMap().TryGetNamedLocation(name)
				if !found {
					return a.fail("GoTo: no location %q on %s (known: %v)", name, g.currentMapName, g.currentMap().NamedLocationNames())
				}
				if a.walkStep(pos, false) {
					if g.currentMap().IsTransitionAt(pos) { // already standing on it: nothing moved us onto it, so ask for it
						g.CheckTransition()
						g.Wait()
						return false
					}
					return true
				}
				return false
			})
		},
		// Approach('actor_or_object')
		"Approach": func(args ...interface{}) (interface{}, error) {
			return a.approach("Approach "+str(args, 0), str(args, 0), func() bool { return true })
		},
		// Talk('actor'): walk up and start the conversation. Answer it with Choose().
		"Talk": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("Talk "+name, func() bool {
				actor := g.actorWithName(name)
				if actor == nil {
					return a.fail("Talk: %q is not on %s", name, g.currentMapName)
				}
				if g.canPlayerTalkToActor(actor) {
					a.logf("talk to %s", name)
					g.PlayerStartDialogue(actor.GetDialogueFile(), actor)
					return true
				}
				if a.walkStep(actor.Position(), true) { // next to them and still no talk: asleep, no dialogue, ...
					if a.stuck == 0 {
						a.logf("cannot talk to %s yet (me %v them %v sees=%v dialogue=%v asleep=%v)", name, g.Player.Position(), actor.Position(), g.Player.CanSee(actor.Position()), actor.HasDialogue(), actor.IsSleeping())
					}
					a.stuck++
					g.Wait()
				}
				return false
			})
		},
		// Use('object'): walk up and bump it: opens containers (picking locks first), terminals, doors.
		"Use": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.approach("Use "+name, name, func() bool {
				obj := a.object(name)
				if lock, isLockable := obj.(interface {
					IsLocked() bool
					GetLockFlag() string
				}); isLockable && lock.IsLocked() { // first bump unlocks (key) or picks, the next one uses
					if a.tries++; a.tries > 30 {
						return a.fail("Use: could not unlock %s", name)
					}
					a.bump(obj.Position())
					return false
				}
				a.logf("use %s", name)
				a.bump(obj.Position())
				return true
			})
		},
		// Kick('object'): walk up and kick it until it breaks (doors have hit points; a locked one opens for good).
		"Kick": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.approach("Kick "+name, name, func() bool {
				obj := a.object(name)
				if door, isDoor := obj.(*Door); isDoor && (door.IsBroken() || !door.IsLocked()) {
					return true
				}
				if a.tries++; a.tries > 30 {
					return a.fail("Kick: %s does not give", name)
				}
				a.logf("kick %s", name)
				g.playerMeleeAttackLocation(obj.Position())
				return false
			})
		},
		// Interact('actor_or_object', 'context menu entry')
		"Interact": func(args ...interface{}) (interface{}, error) {
			name, label := str(args, 0), str(args, 1)
			return a.approach("Interact "+name, name, func() bool {
				pos, _ := a.locate(name)
				a.withAnswer(answer{kind: "choose", label: label}, func() {
					if !g.OpenContextMenuFor(pos) {
						a.fail("Interact: no context menu on %s", name)
					}
				})
				return true
			})
		},
		// Kill('actor'): shoot while there is ammo, then melee. Done when the actor is dead or gone.
		"Kill": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			var victim *Actor // picked once, so one Kill('rat') is one dead rat
			return a.queue("Kill "+name, func() bool {
				if victim == nil {
					victim = a.nearestActorWithName(name)
				}
				if victim == nil || !victim.IsAlive() {
					return true
				}
				weapon, armed := g.Player.GetInventory().GetEquippedWeapon()
				if armed && weapon.IsRangedWeapon() && !weapon.HasAmmo() {
					g.PlayerReloadWeapon()
				}
				if armed && weapon.IsRangedWeapon() && weapon.HasAmmo() {
					if g.IsInShootingRange(g.Player, victim) && g.Player.CanSee(victim.Position()) {
						a.withAnswer(answer{kind: "target", label: name, target: victim}, g.PlayerRangedAttack)
						return false
					}
					if a.walkStep(victim.Position(), true) {
						a.stuck++ // as close as we get, still no shot: wait for the victim to move
						g.Wait()
					}
					return false
				}
				if !a.walkStep(victim.Position(), true) {
					return false
				}
				if g.currentMap().MoveDistance(g.Player.Position(), victim.Position()) > 1 {
					a.stuck++ // as close as we get, still no shot: wait for the victim to move
					g.Wait()
					return false
				}
				if victim.IsHostileTowards(g.Player) {
					a.bump(victim.Position())
				} else {
					a.withAnswer(answer{kind: "choose", label: "Melee Attack"}, func() { g.OpenContextMenuFor(victim.Position()) })
				}
				return false
			})
		},
		// FireMode('Burst'): switch the equipped weapon to the attack mode of that name.
		"FireMode": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("FireMode "+name, func() bool {
				weapon, armed := g.Player.GetInventory().GetEquippedWeapon()
				if !armed {
					return a.fail("FireMode: no weapon equipped")
				}
				for range weapon.AttackModes {
					if strings.Contains(strings.ToLower(weapon.GetCurrentAttackMode().Mode.ToString()), strings.ToLower(name)) {
						return true
					}
					g.CycleTargetMode()
				}
				return a.fail("FireMode: %q has no mode %q", weapon.GetInternalName(), name)
			})
		},
		// Sneak(): start sneaking (no-op when already sneaking).
		"Sneak": func(args ...interface{}) (interface{}, error) {
			return a.queue("Sneak", func() bool {
				if !g.Player.HasFlag(foundation.FlagSneaking) {
					g.PlayerToggleSneak()
				}
				return true
			})
		},
		// Choose('o_id' or 'menu entry'): answers the next dialogue or menu.
		"Choose": func(args ...interface{}) (interface{}, error) { return a.reply("choose", str(args, 0), args) },
		// Leave(): close the open dialogue, like pressing escape.
		"Leave": func(args ...interface{}) (interface{}, error) { return a.reply("leave", "", args) },
		// Take('item'[, n]) / Put('item'[, n]): answers the next opened container.
		"Take": func(args ...interface{}) (interface{}, error) { return a.reply("take", str(args, 0), args) },
		"Put":  func(args ...interface{}) (interface{}, error) { return a.reply("put", str(args, 0), args) },
		// Buy('item'[, n]): answers the next vendor menu.
		"Buy": func(args ...interface{}) (interface{}, error) { return a.reply("buy", str(args, 0), args) },
		// UseItem('dynamite'[, 5]): apply an inventory item; the optional value answers a prompt (e.g. the timer).
		"UseItem": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("UseItem "+name, func() bool {
				item := g.Player.GetInventory().GetItemByName(name)
				if item == nil {
					return a.fail("UseItem: no %q in inventory", name)
				}
				a.logf("use item %s", name)
				if len(args) > 1 {
					a.withAnswer(answer{kind: "text", label: fmt.Sprint(args[1])}, func() { g.PlayerApplyItem(item) })
				} else {
					g.PlayerApplyItem(item)
				}
				return true
			})
		},
		// PickUp('item'): walk to an item lying on this map (or on a corpse) and pick it up.
		"PickUp": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("PickUp "+name, func() bool {
				if g.Player.GetInventory().GetItemByName(name) != nil {
					return true // auto-pickup got it on the way
				}
				for _, item := range g.currentMap().Items() {
					if item.GetInternalName() == name {
						if !a.walkStep(item.Position(), false) {
							return false
						}
						a.logf("pick up %s", name)
						g.PlayerPickupItemAt(item.Position())
						return true
					}
				}
				for _, corpse := range g.currentMap().DownedActors() { // or loot it off a corpse: step onto it with the take queued
					if corpse.GetInventory().GetItemByName(name) != nil {
						a.withAnswer(answer{kind: "take", label: name}, func() { a.walkStep(corpse.Position(), false) })
						return false
					}
				}
				return a.fail("PickUp: no %q lying on %s (items: %s)", name, g.currentMapName, itemNames(g.currentMap().Items()))
			})
		},
		// Drop('item'): drop an inventory item where the player stands.
		"Drop": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("Drop "+name, func() bool {
				item := g.Player.GetInventory().GetItemByName(name)
				if item == nil {
					return a.fail("Drop: no %q in inventory", name)
				}
				g.PlayerDropItem(item)
				return true
			})
		},
		// Equip('item'): equip (wear/wield) an item from the inventory.
		"Equip": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("Equip "+name, func() bool {
				item := g.Player.GetInventory().GetItemByName(name)
				if item == nil {
					return a.fail("Equip: no %q in inventory", name)
				}
				if !g.IsEquipped(item) {
					g.EquipToggle(item)
				}
				if !g.IsEquipped(item) {
					return a.fail("Equip: could not equip %q", name)
				}
				return true
			})
		},
		"Unequip": func(args ...interface{}) (interface{}, error) {
			name := str(args, 0)
			return a.queue("Unequip "+name, func() bool {
				if item := g.Player.GetInventory().GetItemByName(name); item != nil && g.IsEquipped(item) {
					g.EquipToggle(item)
				}
				return true
			})
		},
		// SetStat('Perception', 7): character setup, like picking a build.
		"SetStat": func(args ...interface{}) (interface{}, error) {
			g.Player.GetCharSheet().SetStat(d100.StatFromString(str(args, 0)), int(num(args, 1)))
			return true, nil
		},
		// SetSkill('Mechanics', 110): character setup, like picking a build.
		"SetSkill": func(args ...interface{}) (interface{}, error) {
			g.Player.GetCharSheet().SetSkillAbsoluteValue(d100.SkillFromString(str(args, 0)), int(num(args, 1)))
			return true, nil
		},
		// GameOver(): true once the game showed its end screen (e.g. EndDemo).
		"GameOver": func(args ...interface{}) (interface{}, error) { return a.gameOver, nil },
		"WaitTurns": func(args ...interface{}) (interface{}, error) {
			left := int(num(args, 0))
			return a.queue("WaitTurns", func() bool {
				g.Wait()
				left--
				return left <= 0
			})
		},
		"WaitMinutes": func(args ...interface{}) (interface{}, error) {
			d := time.Duration(num(args, 0)) * time.Minute
			return a.waitUntil("Wait", func() time.Time { return g.gameTime.Time.Add(d) })
		},
		// WaitUntil('22:40'): the wait menu's "until <time>" action; rests until that time of day (tomorrow if it already passed), resuming after a fight.
		"WaitUntil": func(args ...interface{}) (interface{}, error) {
			clock, err := time.Parse("15:04", str(args, 0))
			if err != nil {
				return nil, err
			}
			var target time.Time
			return a.queue("WaitUntil "+str(args, 0), func() bool {
				if target.IsZero() {
					target = g.NextClockTime(clock.Hour(), clock.Minute())
				}
				if g.gameTime.Time.Before(target) {
					g.PlayerRestUntil(false, clock.Hour(), clock.Minute())
				}
				return !g.gameTime.Time.Before(target)
			})
		},
		"WaitHours": func(args ...interface{}) (interface{}, error) {
			d := time.Duration(num(args, 0)) * time.Hour
			return a.waitUntil("Wait", func() time.Time { return g.gameTime.Time.Add(d) })
		},
		// ForceChecks('success' | 'fail' | 'random'): player dialogue and skill checks.
		"ForceChecks": func(args ...interface{}) (interface{}, error) {
			g.ForcedChecks = str(args, 0)
			return true, nil
		},
		// Fail('why'): in an outcome, marks it as a failing ending.
		"Fail": func(args ...interface{}) (interface{}, error) {
			a.finish(false, str(args, 0))
			return true, nil
		},
		"IsOnMap": func(args ...interface{}) (interface{}, error) { return g.currentMapName == str(args, 0), nil },
	}
}

// autoplayUI forwards drawing to the real (or null) UI and turns every modal into
// a pending decision for the autoplayer. Nothing waits for a key press.
type autoplayUI struct {
	foundation.GameUI
	a *Autoplay
}

func (u autoplayUI) SkipAnimations() { u.a.trace(); u.GameUI.SkipAnimations() }

// a scripted player waits in one-minute steps a lot; a second of fading per step would be most of a visual run
func (u autoplayUI) FadeToBlack()   {}
func (u autoplayUI) FadeFromBlack() {}

func (u autoplayUI) open(m *modal) {
	if u.a.modal != nil && u.a.modal.kind != "conversation" {
		u.a.fail("%s %q opened while %s %q is pending", m.kind, m.title, u.a.modal.kind, u.a.modal.title)
		return
	}
	u.a.modal = m
}
func (u autoplayUI) unexpected(what string) { u.a.fail("unsupported modal: %s", what) }

func (u autoplayUI) SetConversationState(text string, options []foundation.MenuItem, partner foundation.ChatterSource, isTerminal bool) {
	u.a.logf("%s: %s", partner.Name(), strings.ReplaceAll(text, "\n", " "))
	u.a.modal = &modal{kind: "conversation", title: partner.Name(), items: options}
	u.GameUI.SetConversationState(text, options, partner, isTerminal)
}
func (u autoplayUI) CloseConversation() {
	if u.a.modal != nil && u.a.modal.kind == "conversation" {
		u.a.modal = nil
	}
	u.GameUI.CloseConversation()
}
func (u autoplayUI) OpenMenu(items []foundation.MenuItem) { u.open(&modal{kind: "menu", items: items}) }
func (u autoplayUI) OpenMenuWithTitle(title string, items []foundation.MenuItem) {
	u.open(&modal{kind: "menu", title: title, items: items})
}
func (u autoplayUI) OpenMenuWithTitleAndClose(title string, items []foundation.MenuItem, onClose func()) {
	u.open(&modal{kind: "menu", title: title, items: items, onClose: onClose})
}
func (u autoplayUI) AskForConfirmation(title string, message string, onConfirm func(bool)) {
	u.open(&modal{kind: "confirm", title: message, confirm: onConfirm})
}
func (u autoplayUI) ShowGiveAndTakeContainer(leftName string, leftItems []foundation.Item, rightName string, rightItems []foundation.Item, transferToLeft func(foundation.Item, int), transferToRight func(foundation.Item, int), takeAll func()) {
	u.open(&modal{kind: "container", title: rightName, mine: leftItems, theirs: rightItems, take: transferToLeft, put: transferToRight})
}
func (u autoplayUI) ShowTakeOnlyContainer(name string, items []foundation.Item, transfer func(foundation.Item)) {
	u.open(&modal{kind: "container", title: name, theirs: items, take: func(item foundation.Item, _ int) { transfer(item) }})
}
func (u autoplayUI) SelectTarget(_ func(foundation.ActorForUI) foundation.AttackInfo, onSelected func(geometry.Point)) {
	u.open(&modal{kind: "target", target: onSelected})
}
func (u autoplayUI) IndicateConversationStartByNPC(partner foundation.ChatterSource, done func()) {
	u.a.logf("%s is addressing you", partner.Name())
	u.open(&modal{kind: "wait", title: partner.Name(), done: done})
}
func (u autoplayUI) ForceListening(partner foundation.ChatterSource, monologue []string, done func()) {
	for _, line := range monologue {
		u.TryAddChatter(partner, line)
	}
	u.open(&modal{kind: "wait", title: partner.Name(), done: done})
}
func (u autoplayUI) TryAddChatter(src foundation.ChatterSource, text string) bool {
	u.a.logf("%s says: %s", src.Name(), text)
	return u.GameUI.TryAddChatter(src, text)
}
func (u autoplayUI) OpenTextWindow(description string) { u.a.logf("text: %s", description) }
func (u autoplayUI) ShowTextFileFullscreen(filename string, onClose func()) {
	u.open(&modal{kind: "wait", title: filename, done: onClose})
}
func (u autoplayUI) ShowGameOver(score foundation.ScoreInfo, _ []foundation.ScoreInfo) {
	u.a.logf("game over: %s", score.DescriptiveMessage)
	u.a.gameOver = true
}
func (u autoplayUI) QuitGame() { u.a.gameOver = true }

func (u autoplayUI) SelectDirection(func(geometry.CompassDirection)) { u.unexpected("SelectDirection") }
func (u autoplayUI) SelectBodyPart(d100.BodyPart, func(foundation.ActorForUI, d100.BodyPart)) {
	u.unexpected("SelectBodyPart (aimed shot)")
}
func (u autoplayUI) AskForString(prompt string, _ string, onDone func(string)) {
	u.open(&modal{kind: "string", title: prompt, text: onDone})
}
func (u autoplayUI) OpenInventoryForManagement() { u.unexpected("OpenInventoryForManagement") }
func (u autoplayUI) OpenInventoryForSelection(_ []foundation.Item, prompt string, _ func(foundation.Item)) {
	u.unexpected("OpenInventoryForSelection " + prompt)
}
func (u autoplayUI) OpenInventoryForSelectionWithClose(_ []foundation.Item, prompt string, _ func(foundation.Item), _ func()) {
	u.unexpected("OpenInventoryForSelection " + prompt)
}

// a keypad is hacked with e-picks, never answered with the code: the script has no way to know it
func (u autoplayUI) OpenKeypad(_ string, _ []rune, hack func() bool, done func(bool)) { done(hack()) }
func (u autoplayUI) OpenVendorMenu(title string, items []foundation.Item, buy func(foundation.Item, int, int), _ func(foundation.Item), onClose func()) {
	u.open(&modal{kind: "vendor", title: title, theirs: items, buy: buy, onClose: onClose})
}
func (u autoplayUI) OpenAimedShotPicker(foundation.ActorForUI, d100.BodyPart, func(foundation.ActorForUI, d100.BodyPart)) {
	u.unexpected("OpenAimedShotPicker")
}
func (u autoplayUI) SelectSaveName() { u.unexpected("SelectSaveName") }
func (u autoplayUI) SelectLoadName() { u.unexpected("SelectLoadName") }

var _ foundation.GameUI = nullUI{}
