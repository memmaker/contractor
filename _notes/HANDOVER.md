# Contractor: handover

State as of 2026-10-07, branch `agent-tries-to-finish` (HEAD `f565d80`). Web build published at
https://ruzzoli.de/games/contractor/play/ (`./publish_web.sh`).

## 1. What the game is right now

A complete, winnable arc: tutorial (Grim's keys), the starter contract (Medical Extortion), then
chapters 1-3 (find the mansion, the recluse Logan Faust, the cryo lab) with four endings
(`cryo_sleep`, `against_all_odds`, `sold_lab`, `gave_up`). Every main quest has a fighter, avoider
and manipulator route. The "golden ticket" main quest resolves as a scam.

Design pillars: `readme.md`. Chapter spec: `_notes/chapters_outline.txt` (binding). Done list:
`_notes/chapters_checklist.md`. Player-facing route guide: `walkthrough.html`.

## 2. Work in flight: the world pass

Spec: `_notes/world_outline.txt`. Checklist: `_notes/world_checklist.md` (only the orchestrator ticks it).
Five builder agents were dispatched in parallel, one per package, each with strict file ownership:

| Pkg | Scope | Owns |
|---|---|---|
| A | side contracts south: The Duck, Tags, Rats; reactive R1-R4 | maps zone_residential_south, _south_underground, zone_wall_south, hq_terra_vitae; dialogues anna_whitaker, big_bob, cleaner, gate_captain, synth_duck, town_lucy, town_father_olaf, richard_shaw; `journal_a/actors_a/items_a.rec`; `a_*` scripts/playtests |
| B | Bar Tab, Dead Freight, Shift Change; R5-R7 | maps zone_commerce, zone_spaceport, zone_industry; dialogues bartender, dec_vega, solara_carter, foreman_okafor, gun_thug_leader, ebi_security_officer, manifest_terminal, spaceport_console; `*_b.rec`; `b_*` |
| C | density for zone_corporate, campus_ebi, zone_residential_west, hq_project_21, hq_bat; Lost Badge, Silence, Rent | those maps; dialogues adrian_voss, jana_miller, haruki_tanaka, bat_receptionist + new NPCs; `*_c.rec`; `c_*` |
| D | mansion/lab gaps D1-D6 (failed persuasion recovers, warn Faust, thief cost, dead guards acknowledged, broadcast costs sleepers, move_sleeper) | mansion + cryo_lab maps, their dialogues, `journal.rec` recluse/cryo only, `ebi_strike.rec`, `d_*` |
| E | living world E1-E6 (news terminal, squatter, beggar informant/shrine, taxi memory, wanted at wall, Grim leaves) | dialogues public_terminal, beggar, taxi_driver_chatter, jacob_thorne, grim_beard, squatter_lenny; zone_residential_east map; `spawned_teams.rec` append; `e_*`; may append `RunScript:` lines to zone_residential_south/meta.rec |

When a builder reports: verify its playtests yourself (`runall.sh <glob>`), add any requested one-liners
it was not allowed to make (e.g. `Dialogue: richard_shaw` in zone_commerce/actors.rec, `RunScript:` lines
in zone_residential_south/meta.rec), paste its walkthrough snippet into `walkthrough.html`, tick the
checklist, then the wrap-up section: all playtests headless, `go test ./game`, ONE visual playtest at
the very end, commit, push, `./publish_web.sh`.

Builders were told: no engine (`game/*.go`) changes, no commits, headless only, report deviations.
Expect merge chores, not conflicts: ownership is disjoint except the files named above.

## 3. How things work (engine facts you will need)

- **Data** lives in `data_atom/` as recfiles. Maps are `data_atom/maps/<name>/` with `tiles.bin`,
  `zones.bin`, `tileSet.rec`, `meta.rec` (name, indoor/outdoor, `RunScript:` autostart lines),
  `objects.rec`, `actors.rec`, `items.rec`. New maps are written as `layout.txt` (23 rows x 80 cols,
  blank line, legend `<char> <tile>`, optional `zone <name>` blocks) and compiled with `cmd/mapgen`.
- **Definitions**: `definitions/actors.rec` + `actors_*.rec`, item files + `items_*.rec`,
  `journal.rec` + `journal_*.rec`, `spawned_teams.rec`. The `_*.rec` globs exist so content packages
  own their own files.
- **Dialogues** (`dialogues/<npc>.rec`): `%rec: OpeningBranch` (cond/goto) and `%rec: Nodes`
  (name, npc text, effect, options o_text/o_id/o_cond/o_test `RollSkill(skill,'Easy|Hard')`/o_goto/
  o_effect). Function-shaped effects `X(...)` are the script functions in `game/script_funcs.go`
  (`GetScriptFuncs`), including `SetFlag/ClearFlag/GetFlag`, `PlayerAddItem/Gold`, `StackTransferTo`,
  `RunScript`, `ProvokeFaction`, `SpawnTeamHunting`, `SpawnActor`, `RemoveActor`, `EndGame`.
  Terminals are objects with `Name:` (internal) and `Dialogue:`; NPC_NAME lets one file serve several.
- **Scripts** (`scripts/*.rec`): `definitions` vars, `outcomes`, `cancel`, `frames` (if:/do:), frames
  run strictly in order; time via `SaveTimeNow` + `IsHoursAfter/IsDaysAfter`. Started by `RunScript`
  effects or a map's `meta.rec`; tick every turn regardless of current map.
- **Journal**: quest = start_cond(s), prog_cond(s), end_cond(s) with end_id/end_text/end_xp. First
  matching end in file order wins. A completed quest switches to a better match later (XP paid once),
  so order the outcomes by priority. Flags: `QuestStarted(id)`, `QuestCompleted(id)`,
  `QuestCompleted(id, outcome)`.
- **Free flags**: `TalkedTo(npc)`, `Killed(npc)` (counter per internal name), `KilledByPlayer(npc)`,
  `ContainerRemoved(container, item)`, item `PickupFlag`, `Ending(id)`.
- **Doors**: a `LockedDoor` opens with the key whose `LockFlag` matches, by lockpicking, or when that
  flag is cleared (`ClearFlag('cryo_keycard')`). NPC pathing passes locked doors only with the key.
- **AI/crime**: factions are strings; actors neutral unless `aggressive: true`. Crimes are sight-based.
  Gunfire is heard: idle actors of another faction within 25 tiles investigate the shooter
  (`gunshotHeard` in `game/crime.go`, through walls, no suppressors). Investigators without a path calm
  down. Schedules: `schedules/<npc>.rec`, days incl. `every_day` and `every_hour` (minutes only).
- **Endings**: `EndGame('<id>')` sets `Ending(id)` and shows the `cryo` quest's outcome text.

## 4. Testing

- Playtests: `data_atom/playtests/*.rec` (`%rec: outcomes` if:, `%rec: frames` if:/do:). Verbs in
  `game/autoplay.go`: Talk, Choose(o_id), Use, GoTo(named location), Take, PickUp (also corpses),
  Kill, Kick(object), Sneak, Equip, FireMode, SetSkill, SetStat (six d100 stats), ForceChecks
  ('success'|'fail'|'random', dialogue checks only, not combat), WaitMinutes/Hours/Until;
  outcomes GameOver(), IsOnMap, HasFlag, HasItem. RNG is seeded: runs are deterministic.
- Run headless from the project dir: `go build -o <scratch>/contractor . && <scratch>/contractor
  autoplay -headless <name>`; last line PASS/FAIL. `AUTOPLAY_TRACE=<internal name>` logs that actor's
  position/state per step. Visual: build with tags `ebitensinglethread,ebiten,terminal`, run
  `autoplay <name>` without `-headless`.
- Scratchpad helper: `<scratch>/runall.sh [glob]` (120 s timeout can mask an idling run: an
  out-of-frames script idles to the 5000-turn cap, read the last line).
- Go tests: `go test ./game -run 'ChapterMaps|Chapters|StarterQuest'` (don't run the full suite
  unless asked). `go vet` reports pre-existing unreachable code in `state_core.go`.
- Rules from the owner: only the relevant tests; isolate with traces, don't brute-force reruns;
  at most ONE visual playtest, at the very end, after all headless pass.

## 5. Playtest inventory (all PASS at HEAD)

Chapter 1: `chapter1_a_taxi_driver`, `chapter1_b_gate_captain`, `chapter1_c_invoice`.
Chapter 2: `chapter2_a_fighter_kill_faust`, `chapter2_b_thief`, `chapter2_c_diplomat_allied`,
`chapter2_d_extort`, `chapter2_e_sold_out`. Chapter 3: `chapter3_a_pull_plug`, `chapter3_b_fuse`,
`chapter3_c_ark_reroute`, `chapter3_d_faust_pod`, `chapter3_e_broadcast`, `chapter3_f_sell_lab`,
`chapter3_g_retire`. Starter: `diplomat_*`, `fighter_*`, `thief_*` as present in the folder.

## 6. Known open points

- No suppressed weapons; gunfire radius is a flat constant. Add a weapon loudness field if needed.
- Lockpicking in an NPC's view is a minor crime (two warnings, then combat); the thief must sneak.
- `go vet` unreachable-code warnings in `game/state_core.go` predate this work.
- GitHub reports two dependabot vulnerabilities on the default branch (not addressed).
- Branch is `agent-tries-to-finish`; nothing has been merged to `main`.
