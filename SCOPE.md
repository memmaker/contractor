# Scope: agent-tries-to-finish

Decided 2026-10-04 (grill session). Anything not listed is out of scope.

## Done means
1. **The Debt** (`_notes/starter_quest_outline.txt`) is playable end-to-end, with every listed branch:
   steal stash, force (lethal/non-lethal), persuade, negotiate double, get hired against the client,
   free/kill/interrogate Drake, pay yourself, ignore (Quinn Rix attacks the doc), the client branches
   (kill, extort, intimidate, steal from, goons, better job), steal-from-client-and-give-back,
   and the clinic takeover (Drake or EBI). Each branch has a journal entry and an XP reward.
2. **Chapter 1 stub**: after The Debt resolves, a lead to "find the mansion" is delivered. Following it
   ends the demo with an outcome summary and "to be continued".
3. **Every bug in the BUGS section of `_notes/todo_bugs.txt` is fixed.**

## Engine work
Only what The Debt needs (e.g. a locked stash container, skill checks applied to lockpicking/terminals, journal).
No sneak system, repair, criticals, or other general items from `fallout_scope.txt`.

## Out of scope
The cryo arc (Mansion, Cryo Lab), the `//TODO` wishlist in todo_bugs.txt, and new zones.

## Verification
Go tests drive each Debt branch through dialogue/script conditions and assert the resulting flags. The
existing validation must pass. The wasm build must boot in a browser.

## Delivery
Small commits on `agent-tries-to-finish`, no PR.

## Status (2026-10-04)
- The Debt: every branch is wired, and each ending is checked by `game/starter_quest_test.go`.
- Not done, and not a bug: "Actor AI in general / interaction markers / sandbox behaviours" is a feature wish.
- Not playtested by hand: Bob's melee fallback, the vendor buy fix, the goons fight, and the clinic takeover timing.
- Ignoring the job: declining it, or leaving the offer unanswered for 48 game hours, sends Quinn after Harker.
