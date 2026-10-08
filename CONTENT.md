# Adding content

New NPCs, contracts, items and world reactions go into **content packs**. A pack is one Python file,
`content/<pack>.py`, that describes everything; running it writes the game's data files. Packs never
edit hand-made files except for clearly marked option blocks, so a pack can be edited and rerun as
often as you like.

```bash
python3 tools/content.py ref        # what exists: script functions, playtest verbs, maps, actors, spots
./content_check.sh <pack>           # regenerate + lint + validate + go tests + the pack's playtests
./content_check.sh <pack> ecto      # same, only playtests named <pack>_ecto*
```

`content/f.py` (package F, ten contracts) is the full worked example.

## Workflow

1. `python3 tools/content.py ref` and read the maps you want to use (actors and spots per map).
2. Write `content/<pack>.py`. Pack names are lowercase letters and digits.
3. `./content_check.sh <pack>` until it is green. Fix the first problem it reports and rerun.
4. Add the contract to `walkthrough.html` (side contract tab), commit.

## What a pack writes

| Call | Ends up in |
|---|---|
| `p.actor`, `p.corpse` | `maps/<map>/actors_<pack>.rec` |
| `p.object`, `p.container`, `p.named_location` | `maps/<map>/objects_<pack>.rec` |
| `p.item` | `definitions/items_<pack>.rec` |
| `p.spawnable`, `p.team` | `definitions/actors_<pack>.rec`, `definitions/spawned_teams_<pack>.rec` |
| `p.journal_entry` | `definitions/journal_<pack>.rec` |
| `p.dialogue`, `p.script` | `dialogues/<name>.rec`, `scripts/<name>.rec` |
| `p.extend_dialogue`, `p.terminal_lead` | a `# >>> pack <pack>` block inside an existing dialogue |
| `p.playtest` | `playtests/<pack>_*.rec` |

The engine loads `actors_*.rec`, `objects_*.rec` and `items_*.rec` next to each map's own files, and
every `journal_*`, `items_*`, `actors_*` and `spawned_teams_*` file in `definitions/`.
`data_atom/packs/<pack>.manifest` lists what the pack owns; files it stops writing are deleted on the
next run. Don't hand-edit generated files: change the pack and rerun.

**Placement.** `near=(x, y)` puts the thing on the closest free floor tile that is reachable from the
map's taxi stand or street exits and has nothing else within one tile. `at=(x, y)` pins it exactly.
`python3 tools/content.py spot <map> X,Y` shows where `near` would land. Positions are stable across
reruns as long as the hand-made content around them doesn't change. The map editor only edits a map's
own files, so move pack content by changing `near`/`at`.

## A complete small contract

```python
import os, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'tools'))
from content import GUN, Pack, Trip

p = Pack('g', 'Package G: example')

# People and things. Extra keyword arguments become record fields (Awareness=8, aggressive='true').
p.actor('zone_residential_south', 'widow_ruth', 'Ruth', 'An old woman who keeps looking at her empty wrist.', 'R',
        near=(30, 15), equipment=['gold(200)'])
p.container('zone_residential_west', 'ruth_pawn_box', 'a pawnshop lockbox', ['ruth_watch'], near=(50, 10), lock='easy')
p.item('ruth_watch', 'a silver watch', 'Engraved: For R., forty years.', cost=150, pickup_flag='ruth_watch_found')

p.dialogue('widow_ruth', '''
# G1 Ruth's watch. Return it or sell it.
%rec: OpeningBranch

cond: HasFlag('ruth_done')
goto: Done

cond: HasFlag('JobAccepted(ruth)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: My grandson pawned my watch. The pawnshop in the west keeps it in a lockbox.
#
o_text: I'll get it back.
o_id: ruth_accept
o_goto: Accepted
#
o_text: Not today.
o_id: ruth_decline
o_goto: Bye

name: Accepted
npc: Bless you.
effect: SetFlag('JobAccepted(ruth)')
effect: EndConversation

name: Waiting
npc: Any luck?
#
o_text: (Give her the watch.)
o_id: ruth_give
o_cond: HasItem('ruth_watch')
o_goto: Paid
#
o_text: Not yet.
o_id: ruth_wait
o_goto: Bye

name: Paid
npc: Forty years. Here, take this.
effect: StackTransferTo(NPC, 'ruth_watch', 1)
effect: StackTransferFrom(NPC, 'gold', 100)
effect: SetFlag('ruth_done')
effect: EndConversation

name: Done
npc: It still ticks.
effect: EndConversation

name: Bye
npc: Mind the step.
effect: EndConversation
''')

# A second way out, added to an NPC that already exists.
p.extend_dialogue('pawn_rook', 'Default', '''o_text: Want a silver watch? (sell it)
o_id: rook_buy_watch
o_cond: HasItem('ruth_watch')
o_goto: RookWatch''', '''name: RookWatch
npc: Engraved. Worth less. Two hundred.
effect: StackTransferTo(NPC, 'ruth_watch', 1)
effect: PlayerAddGold(200)
effect: SetFlag('ruth_watch_sold')
effect: EndConversation''')

p.journal_entry('''id: ruth
name: Ruth's Watch
skill_points: 2
#
start_cond: HasFlag('JobAccepted(ruth)')
start_text: Ruth in the south wants her watch back from the pawnshop lockbox in the west.
#
prog_cond: HasFlag('ruth_watch_found')
prog_text: You have the watch.
#
end_cond: HasFlag('ruth_done')
end_text: Ruth has her watch back.
end_id: returned
#
end_cond: HasFlag('ruth_watch_sold')
end_text: You sold Ruth's watch back to Rook.
end_id: sold
end_skill_points: 1
''')

p.terminal_lead('ruth', "HasFlag('JobAccepted(ruth)')", 'Recover a keepsake, southside.',
                ['An old woman lost something to a pawnshop.'], 'Ask for Ruth, southside residential.')

# One playtest per route. Trip() starts in the south district and inserts the taxi rides.
take = lambda: (Trip().do("Talk('widow_ruth')", "Choose('ruth_accept')")
                .go('zone_residential_west').do("Give('mechanical_lockpick(5)')", "SetSkill('Mechanics', 200)",
                                               "Use('ruth_pawn_box')", "Take('ruth_watch')"))
p.playtest('g_ruth_returned', "Ruth's Watch: crack the lockbox, bring it home.", "HasFlag('QuestCompleted(ruth, returned)')",
           take().go('zone_residential_south').do("Talk('widow_ruth')", "Choose('ruth_give')"))
p.playtest('g_ruth_sold', "Ruth's Watch (grey): sell it straight back to Rook.", "HasFlag('QuestCompleted(ruth, sold)')",
           take().do("Talk('pawn_rook')", "Choose('rook_buy_watch')"))

p.write()
```

## Data formats in one page

**Dialogue** (`%rec: OpeningBranch` then `%rec: Nodes`). Opening branches are tried top to bottom; the
first `cond` that is true picks the starting node. A node has `name`, `npc` (continue lines with `+ `),
any number of `effect:` lines, and options separated by `#`. An option has `o_text`, `o_id` (playtests
choose by it), optional `o_cond`, and either `o_goto` or a skill test `o_test: RollSkill('social', 'Easy')`
with `o_succ`/`o_fail`. End a conversation with `effect: EndConversation`. Skills: social, intimidate,
technology, mechanics, stealth... (see `ref`). `NPC` in an effect is the person you talk to.

**Script** (`p.script(name, frames)`): frames run strictly in order. Each frame waits until its `if:` is
true, then runs its `do:` lines once. Start a script from a dialogue with `effect: RunScript('name')`.
Use `SaveTimeNow('t')` then `IsMinutesAfter('t', 20)` / `IsHoursAfter` / `IsDaysAfter` for delays,
`IsMap('zone_x')` to wait for the player, `LoadMap('zone_x')` before changing actors on another map.

**Journal entry**: `start_cond`/`start_text`, optional `prog_cond`/`prog_text`, then any number of
`end_cond`/`end_text`/`end_id` (+ optional `end_skill_points`). The first end condition that is true wins, so put
the specific ones first. Quests are the only progression: `skill_points:` (about 1 per 50 old XP, a
side contract gives 3–6) and `perk_points: 1` for the big ones. There is no XP and no level. Completion sets `QuestCompleted(<id>, <end_id>)`, which playtests check.

**Flags the game sets on its own**: `TalkedTo(npc)`, `Killed(npc)`, `QuestCompleted(id, end_id)`,
`WasAttacked(npc)`, `PlayerVisited(map)` and an item's `PickupFlag` when it is picked up, looted or taken
from a container. A key item `key('flag', 'description')` is called `flag` and opens containers or
doors with `lockflag: flag`.

**Playtest**: `%rec: outcomes` holds one `if:` that must become true; frames are `do:` lines run in
order. Verbs are listed by `ref`: Talk, Choose (by o_id), GoTo (spot or actor), Use (object), Take
(from the open container), PickUp (from floor or corpse), Kill, Give, SetSkill, SetDerivedStat('HitPoints', 45), ForceChecks('success' |
'fail' | 'none'), Sneak, WaitHours, WaitMinutes, WaitUntil('23:00'), Equip, Unequip, Interact(object,
menu label). `GUN` is a fighter loadout.

**Character model**: skills only, no attributes, no XP. Every skill starts at 20, the three tagged
skills at 40; quests hand out `skill_points` and `perk_points`. Derived stats are fixed (HitPoints 30,
Speed 10, Awareness 5 …, see `rules.rec`); gear, drugs and perks change them. NPCs set their own
`HitPoints`, `Speed`, `Awareness` (1–10, how well they notice sneaking) and `ActionPoints`.

## Conventions

- A contract is accepted with `SetFlag('JobAccepted(<id>)')`, the same id as its journal entry.
- Give each contract at least three routes (fight, talk, sneak or grey) and a playtest for each.
  Vary the shape: not every job is "get item X from Y".
- Every world reaction (someone leaves, dies, moves in) is a script or a changed opening branch, and
  someone or something comments on it.
- Write in-world text with typographic apostrophes (’) inside Python strings to avoid escape trouble.

## Traps

- `WaitMinutes` keeps waiting while you are attacked. Before an ambush, wait until just past the spawn
  time, then `Kill`.
- Several enemies at once kill a playtest character with default stats. Use `GUN` and fewer or weaker
  enemies (`hp=`).
- `Kill('name')` picks the nearest actor with that name; one call kills one.
- An NPC with `faction` other than its own name fights alongside that faction (`ProvokeFaction`).
- Extend hand-made dialogues or your own pack's. Another pack rewrites its own dialogue files on each
  run, which drops blocks other packs put there.
- Placing things next to visible proximity mines makes the playtest walk into them.
- Lint (`python3 tools/content.py lint <pack>`) catches unknown items, scripts, dialogues, spots, teams,
  options, timers that were never started, and flags that are queried but never set.
