# Chapters 1-3 checklist (spec: chapters_outline.txt)

## Engine
- [x] cmd/mapgen: layout.txt -> tiles.bin + zones.bin
- [x] EndGame('<id>') dialogue effect, journal-driven game over screen
- [x] endDemo removed if unused

## Maps
- [x] mansion: locations, doors, faust_safe, faust_terminal, cellar transition
- [x] cryo_lab: pods 1-4, ark_terminal, generator room, spare_parts
- [x] actors: logan_faust, mira_sykes, mansion_guard x3, ebi_trooper
- [x] items: cryo_keycard, mansion_key, faust_safe_key, cryo_fuse
- [x] game/chapter_maps_test.go passes

## Chapter 1 - find the mansion
- [x] taxi_driver: social Easy / pay 200
- [x] gate_captain: intimidate Hard / gate_log loot
- [x] invoice at hq_terra_vitae (exists)
- [x] home_terminal: chapter1_accept sets JobAccepted(mansion); report_faust / sell_lab / retire nodes
- [x] taxi: mansion destination
- [x] journal find_mansion
- [x] playtests chapter1_a/b/c pass headless

## Chapter 2 - the recluse
- [x] mansion_guard, mira_sykes, logan_faust, faust_terminal dialogues
- [x] scripts/ebi_strike
- [x] journal recluse (faust_dead, sold_out, allied, extorted); to_the_stars scam end
- [x] playtests: fighter (kill), thief (service door + safe + hack), diplomat (allied), sold_out strike

## Chapter 3 - the cryo lab
- [x] ark_terminal, cryo_pod dialogues
- [x] journal cryo (cryo_sleep, against_all_odds, sold_lab, gave_up)
- [x] playtests: pull plug, fuse, ARK reroute, Faust's pod 4; each ending reaches GameOver
- [x] game/chapters_test.go passes

## Wrap-up
- [x] walkthrough.html extended with chapters 1-3
- [x] relevant tests + new playtests pass headless and one visual
- [x] commit, push, web deploy

## Follow-up (2026-10-07)

- [x] Gunfire is heard: idle actors of other factions within 25 tiles investigate the shooter (`gunshotHeard`, through walls, no suppressors)
- [x] Investigating actors without a path calm down instead of standing in Investigate forever
- [x] Completed quests keep reacting: a better matching outcome replaces the old one, XP paid once (`recluse` allied → sold_out → faust_dead)
- [x] `every_hour` schedule day; Mira Sykes patrols hall, study, kitchen, gate
- [x] Playtests: chapter2_a rewritten for the converging household, chapter2_e asserts `QuestCompleted(recluse, sold_out)`, fighter_c_kill_jacob fixed (was dying deterministically)
