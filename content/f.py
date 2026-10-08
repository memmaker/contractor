"""Package F: ten side contracts after Night City Stories, Greenwar and Northwest Passage.

Each contract has its own shape: ambush, betrayed client, part collection, night stakeout, timed
forgery, info-only job, evidence, deadline, betrayal insurance, a rescue the rescued may refuse.
Rerun after editing: python3 content/f.py
"""
import os, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'tools'))
from content import GUN, Pack, Trip

p = Pack('f', 'Package F: ten more contracts')

# ===== F1 The Slow Boat: investigation, then an ambush; three buyers =====
p.actor('zone_commerce', 'fixer_kade', 'Kade', 'A fixer in a silk shirt two sizes too clean for this street. He checks the door every time it opens.', 'K', near=(30,12), equipment=['gold(450)'])
p.corpse('zone_residential_east', 'courier_hiro', 'a dead courier in a Chiba freight jacket', near=(30,18), equipment=["key('luggage_7', 'claim ticket #7')"])
p.object('zone_commerce', '''Category: UnknownContainer
Name: left_luggage
Description: a left-luggage locker, #7
lockdifficulty: hard
lockflag: luggage_7
item: chiba_chip''', near=(26,20))
p.item('chiba_chip', 'a data chip', 'A sealed data chip off the Chiba freighter, stamped with an EBI cargo seal. Berth lists, if Kade is to be believed.', 400, pickup_flag='slow_boat_chip_taken')
p.spawnable('chiba_gunman', 'a gunman in a long coat', 'One of two men who followed the Chiba courier off the boat. They want the chip and they do not ask twice.', 'g', equipment=['10mm_pistol', '10mm_jhp(24)'], faction='chiba', aggressive='true')
p.team('chiba_gunmen', 'chiba_gunman', ['chiba_gunman'])
p.dialogue('fixer_kade', '''# F1 The Slow Boat. Flags: JobAccepted(slow_boat), slow_boat_kade. Script f_slow_boat: gunmen hunt whoever holds the chip in commerce.
%rec: OpeningBranch

cond: HasFlag('slow_boat_kade') || HasFlag('slow_boat_shaw') || HasFlag('slow_boat_olaf')
goto: Done

cond: HasFlag('JobAccepted(slow_boat)') && HasItem('chiba_chip')
goto: HasChip

cond: HasFlag('JobAccepted(slow_boat)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Times are slow, so here's easy money. A courier came in on the Chiba freighter carrying something for me. Meet him at the east side taxi stand, take the package, bring it here. Four hundred.
#
o_text: What does the package look like?
o_id: slow_boat_accept
o_goto: Accepted
#
o_text: Not my kind of job.
o_id: slow_boat_decline
o_goto: Bye

name: Accepted
npc: No idea. That's why I need someone who can think. His name is Hiro. Freight jacket. Go.
effect: SetFlag('JobAccepted(slow_boat)')
effect: RunScript('f_slow_boat')
effect: EndConversation

name: Waiting
npc: Hiro, east side. Don't come back empty-handed.
effect: EndConversation

name: HasChip
npc: (He sees the chip and goes very still.) You found it. Did anybody follow you?
#
o_text: (Hand it over.) Four hundred.
o_id: slow_boat_give
o_goto: Paid
#
o_text: Not yet.
o_id: slow_boat_keep
o_goto: Bye

name: Paid
npc: (He pockets it.) Berth lists. Who goes up to the stars and who stays down here. Now you know why the courier died. Forget it.
effect: StackTransferTo(NPC, 'chiba_chip', 1)
effect: StackTransferFrom(NPC, 'gold', 400)
effect: SetFlag('slow_boat_kade')
effect: EndConversation

name: Done
npc: We're done. Leave the door how you found it.
effect: EndConversation

name: Bye
npc: The door's behind you.
effect: EndConversation
''')
p.script('f_slow_boat', '''# F1: twenty minutes after the chip leaves the locker, the courier's killers catch up in the commerce district.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: HasFlag('slow_boat_chip_taken')
do: SaveTimeNow('slow_boat_t')

if: IsMinutesAfter('slow_boat_t', 20) && IsMap('zone_commerce')
do: SpawnTeamHunting('chiba_gunmen', 'zone_commerce', 'taxi_stand', 'player')
do: SetFlag('slow_boat_ambush')
''')
p.extend_dialogue('richard_shaw', 'Default', '''o_text: (Show him the Chiba chip.) Berth lists. Interested?
o_id: shaw_buy_chip
o_cond: HasItem('chiba_chip')
o_goto: ShawChip''', '''name: ShawChip
npc: (He reads the seal and stops breathing for a second.) The real thing. Six hundred, and every screen in the dome shows it by morning.
effect: StackTransferTo(NPC, 'chiba_chip', 1)
effect: PlayerAddGold(600)
effect: SetFlag('slow_boat_shaw')
effect: EndConversation''')
p.extend_dialogue('town_father_olaf', 'Default', '''o_text: (Give him the Chiba chip.) These are the names of who gets to leave.
o_id: olaf_take_chip
o_cond: HasItem('chiba_chip')
o_goto: OlafChip''', '''name: OlafChip
npc: (He reads for a long time.) None of ours. Not one. (He closes his eyes.) Then we stop waiting for the ships. Thank you. Terra Vitae will remember.
effect: StackTransferTo(NPC, 'chiba_chip', 1)
effect: SetFlag('slow_boat_olaf')
effect: SetFlag('terra_vitae_friend')
effect: EndConversation''')
p.journal_entry('''id: slow_boat
name: The Slow Boat
skill_points: 5
#
start_cond: HasFlag('JobAccepted(slow_boat)')
start_text: Kade, a fixer in the commerce district, wants a package from Hiro, a courier off the Chiba freighter. Meet him at the east side taxi stand.
#
prog_cond: HasItem('chiba_chip')
prog_text: Hiro was dead. His claim ticket opened locker #7 and a chip with an EBI seal. Somebody else wants it too. Kade pays 400, Shaw would pay more, and Father Olaf would want to read it.
#
end_cond: HasFlag('slow_boat_kade')
end_text: Kade has the berth lists. You never learned who he sells them to.
end_id: kade
#
end_cond: HasFlag('slow_boat_shaw')
end_text: Shaw paid 600 and the berth lists are public. Not one name from the poor districts.
end_id: shaw
#
end_cond: HasFlag('slow_boat_olaf')
end_text: Father Olaf read the berth lists. Terra Vitae has stopped waiting for the ships.
end_id: olaf
''')
t=Trip().go('zone_commerce').do("Talk('fixer_kade')","Choose('slow_boat_accept')").go('zone_residential_east').do("PickUp('luggage_7')").go('zone_commerce').do("Use('left_luggage')","Take('chiba_chip')").do(*GUN,"WaitMinutes(30)").do("Kill('chiba_gunman')").do("Kill('chiba_gunman')").do("Unequip('10mm_smg')","Talk('fixer_kade')","Choose('slow_boat_give')")
p.playtest('f_slow_boat_kade', 'The Slow Boat (fighter): dead courier, claim ticket, locker, the gunmen catch up, deliver to Kade.', "HasFlag('QuestCompleted(slow_boat, kade)') && HasFlag('slow_boat_ambush')", t)
t=Trip().go('zone_commerce').do("Talk('fixer_kade')","Choose('slow_boat_accept')").go('zone_residential_east').do("PickUp('luggage_7')").go('zone_commerce').do("Use('left_luggage')","Take('chiba_chip')").do("Talk('richard_shaw')","Choose('shaw_buy_chip')")
p.playtest('f_slow_boat_shaw', 'The Slow Boat (fast): sell the chip to Shaw before the gunmen arrive.', "HasFlag('QuestCompleted(slow_boat, shaw)') && !HasFlag('slow_boat_ambush')", t)
t=Trip().go('zone_commerce').do("Talk('fixer_kade')","Choose('slow_boat_accept')").go('zone_residential_east').do("PickUp('luggage_7')").go('zone_commerce').do("Use('left_luggage')","Take('chiba_chip')").go('hq_terra_vitae').do("Talk('town_father_olaf')","Choose('olaf_take_chip')")
p.playtest('f_slow_boat_olaf', 'The Slow Boat (grey): give the berth lists to Terra Vitae.', "HasFlag('QuestCompleted(slow_boat, olaf)')", t)

# ===== F2 Sweet Revenge: the star hires you; the exec offers to buy her location =====
p.actor('hq_terra_vitae', 'star_shani', 'Shani Kell', 'A woman with a famous face under a cheap hood. She flinches whenever a camera drone passes over the commune.', 's', near=(20,12), equipment=['gold(350)'])
p.actor('zone_corporate', 'exec_stromberg', 'Stromberg', 'A media executive with an earpiece that is never quiet. Behind him, a billboard of Shani Kell.', 'S', near=(40,18), equipment=['gold(900)'])
p.object('zone_corporate', '''Category: UnknownContainer
Name: stromberg_safe
Description: Stromberg's office safe
lockdifficulty: hard
item: stromberg_ledger
item: gold(150)''', near=(42,18))
p.item('stromberg_ledger', 'a black ledger', 'Stromberg’s private accounts. Braindance masters bought from EBI: "south wall, night, 14 angles". Buyers. Prices.', 300)
p.dialogue('star_shani', '''# F2 Sweet Revenge. Shani wants Stromberg ruined. Flags: JobAccepted(sweet_revenge), sweet_revenge_exposed/_killed/_released, sweet_revenge_betrayed.
%rec: OpeningBranch

cond: HasFlag('sweet_revenge_exposed') || HasFlag('sweet_revenge_avenged') || HasFlag('sweet_revenge_freed')
goto: Done

cond: HasFlag('sweet_revenge_betrayed')
goto: Suspicious

cond: HasFlag('JobAccepted(sweet_revenge)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Please don't say my name out loud. (She studies you.) You're a contractor. I can pay. Stromberg made me, and then he sold the night of the cleansing as a braindance. I recorded it for him. I didn't know.
#
o_text: What do you want done?
o_id: sweet_revenge_accept
o_goto: Accepted
#
o_text: Not my business.
o_id: sweet_revenge_decline
o_goto: Bye

name: Accepted
npc: Ruin him. Kill him if you must. Better, his black ledger is in his office safe; with it I can show the dome who bought those recordings. Or make him let me go. Three hundred, it's what I have left.
effect: SetFlag('JobAccepted(sweet_revenge)')
effect: EndConversation

name: Waiting
npc: Is it done?
#
o_text: (Give her the ledger.)
o_id: sweet_revenge_give_ledger
o_cond: HasItem('stromberg_ledger')
o_goto: Exposed
#
o_text: Stromberg is dead.
o_id: sweet_revenge_dead
o_cond: HasFlag('Killed(exec_stromberg)')
o_goto: Avenged
#
o_text: He tore up your contract. You're free.
o_id: sweet_revenge_free
o_cond: HasFlag('sweet_revenge_released')
o_goto: Freed
#
o_text: Not yet.
o_id: sweet_revenge_notyet
o_goto: Bye

name: Exposed
npc: (She turns the pages, white.) Fourteen angles. He sold fourteen angles of it. (Then, very calm.) Every screen. Tonight.
effect: StackTransferTo(NPC, 'stromberg_ledger', 1)
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('sweet_revenge_exposed')
effect: EndConversation

name: Avenged
npc: (She doesn't cry. She doesn't smile.) I thought it would feel like something. Take the money.
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('sweet_revenge_avenged')
effect: EndConversation

name: Freed
npc: Free. (She pulls the hood back for the first time.) I'm going to sing at Terra Vitae, for nobody's camera. Thank you.
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('sweet_revenge_freed')
effect: EndConversation

name: Suspicious
npc: (She looks at the drone in the sky, then at you.) Somebody told him where I am. Get away from me.
effect: EndConversation

name: Done
npc: (She hums something under her breath. It's good.)
effect: EndConversation

name: Bye
npc: Don't tell anyone you saw me.
effect: EndConversation
''')
p.dialogue('exec_stromberg', '''# F2 Sweet Revenge: the target. Release her (social Hard) or sell her out (800).
%rec: OpeningBranch

cond: true
goto: Default

%rec: Nodes

name: Default
npc: My time costs more than yours. Talk.
#
o_text: Shani Kell wants out of her contract. Let her go before this gets ugly.
o_id: stromberg_release
o_cond: HasFlag('JobAccepted(sweet_revenge)') && !HasFlag('sweet_revenge_released')
o_test: RollSkill('social', 'Hard')
o_succ: Released
o_fail: Refused
#
o_text: I know where Shani Kell is hiding.
o_id: stromberg_betray
o_cond: HasFlag('JobAccepted(sweet_revenge)') && !HasFlag('sweet_revenge_betrayed')
o_goto: Betray
#
o_text: What do you sell?
o_id: stromberg_lore
o_goto: Lore
#
o_text: Nothing.
o_id: stromberg_bye
o_goto: Bye

name: Lore
npc: Feelings. Other people's. A braindance of a first kiss sells. A braindance of a last breath sells better. The dome is bored, and I am its cure.
#
o_text: Back.
o_id: stromberg_lore_back
o_goto: Default

name: Released
npc: (A long look.) She's damaged goods anyway. Fine. (He taps his earpiece.) Legal, void the Kell contract. Now get out of my building.
effect: SetFlag('sweet_revenge_released')
effect: EndConversation

name: Refused
npc: She signed. Ink is forever. Security will see you out.
effect: EndConversation

name: Betray
npc: (He smiles for the first time.) Eight hundred. My people collect her tonight.
effect: PlayerAddGold(800)
effect: SetFlag('sweet_revenge_betrayed')
effect: RunScript('f_shani_taken')
effect: EndConversation

name: Bye
npc: James, my next appointment.
effect: EndConversation
''')
p.script('f_shani_taken', '''# F2: Stromberg's people collect Shani Kell six hours after you sold her out.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('shani_sold_t')

if: IsHoursAfter('shani_sold_t', 6)
do: LoadMap('hq_terra_vitae')
do: RemoveActor('star_shani')
do: SetFlag('shani_taken')
''')
p.journal_entry('''id: sweet_revenge
name: Sweet Revenge
skill_points: 5
#
start_cond: HasFlag('JobAccepted(sweet_revenge)')
start_text: Shani Kell, a braindance star hiding at Terra Vitae, wants revenge on Stromberg, the exec who sold her recording of the cleansing. Kill him, steal his black ledger from his office safe, or make him release her.
#
end_cond: HasFlag('sweet_revenge_betrayed')
end_text: You sold Shani Kell to Stromberg for 800. His people collected her from Terra Vitae that night.
end_id: betrayed
end_skill_points: 2
#
end_cond: HasFlag('sweet_revenge_exposed')
end_text: Stromberg's ledger is on every screen. Fourteen angles of the cleansing, and the names of everyone who bought them.
end_id: exposed
#
end_cond: HasFlag('sweet_revenge_avenged')
end_text: Stromberg is dead. Shani Kell felt nothing, and paid anyway.
end_id: killed
#
end_cond: HasFlag('sweet_revenge_freed')
end_text: Stromberg voided Shani's contract. She sings at Terra Vitae now, for no camera.
end_id: released
''')
t=Trip().go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_accept')").go('zone_corporate').do("SetSkill('Stealth', 200)","SetSkill('Mechanics', 200)","Give('mechanical_lockpick(5)')","Sneak()","Use('stromberg_safe')","Take('stromberg_ledger')").go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_give_ledger')")
p.playtest('f_sweet_revenge_exposed', 'Sweet Revenge (avoider): crack Stromberg\'s safe, the ledger goes public.', "HasFlag('QuestCompleted(sweet_revenge, exposed)')", t)
t=Trip().go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_accept')").go('zone_corporate').do(*GUN,"Kill('exec_stromberg')").do("Unequip('10mm_smg')").go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_dead')")
p.playtest('f_sweet_revenge_killed', 'Sweet Revenge (fighter): kill Stromberg.', "HasFlag('QuestCompleted(sweet_revenge, killed)')", t)
t=Trip().go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_accept')").go('zone_corporate').do("ForceChecks('success')","Talk('exec_stromberg')","Choose('stromberg_release')").go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_free')")
p.playtest('f_sweet_revenge_released', 'Sweet Revenge (talker): talk Stromberg into voiding her contract.', "HasFlag('QuestCompleted(sweet_revenge, released)')", t)
t=Trip().go('hq_terra_vitae').do("Talk('star_shani')","Choose('sweet_revenge_accept')").go('zone_corporate').do("Talk('exec_stromberg')","Choose('stromberg_betray')").do("WaitHours(7)")
p.playtest('f_sweet_revenge_betrayed', 'Sweet Revenge (grey): sell her location to Stromberg; she is gone by night.', "HasFlag('QuestCompleted(sweet_revenge, betrayed)') && HasFlag('shani_taken')", t)

# ===== F3 The '94 Harley: three parts on three maps, a rebuild that takes time, a rival buyer =====
p.actor('zone_residential_south', 'biker_dizzy', 'Dizzy', 'An old biker with a grey ponytail and an empty garage. He talks about engines the way other people talk about the dead.', 'D', near=(30,10), equipment=['gold(350)'])
p.actor('zone_residential_west', 'chopper_ratchet', 'Ratchet', 'A chop-shop mechanic in oil to the elbows, whistling.', 'R', near=(30,15), equipment=['gold(500)', 'harley_ignition'])
p.object('zone_residential_west', '''Category: UnknownContainer
Name: burned_car
Description: a burned-out car
lockdifficulty: easy
item: harley_fuel_cell
item: junk''', near=(36,15))
for n,d,l in (('harley_ignition','an ignition module','The ignition module off a 1994 Harley-Davidson. Chrome, brass, and somebody’s whole life.'),('harley_fuel_cell','a fuel cell','A dented fuel cell that fits a bike older than the dome.'),('harley_tank','a chrome tank','A chrome fuel tank, flames painted on, scratched by a hundred owners.')):
    p.item(n, d, l, 120)
p.dialogue('biker_dizzy', '''# F3 The '94 Harley. Three parts: ignition (Ratchet, west), fuel cell (a burned car, west), chrome tank (Mott, at the wall). Six hours to rebuild.
%rec: OpeningBranch

cond: HasFlag('harley_rebuilt') || HasFlag('harley_sold_ratchet')
goto: Done

cond: HasFlag('harley_parts_given') && IsHoursAfter('harley_t', 6)
goto: Ride

cond: HasFlag('harley_parts_given')
goto: Working

cond: HasFlag('JobAccepted(old_harley)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: My '94 Harley. Real gasoline engine, the last one in the dome. They stripped her while I was in the clinic. Ignition, fuel cell, the chrome tank. Without them she's a sculpture.
#
o_text: I'll find your parts.
o_id: harley_accept
o_goto: Accepted
#
o_text: Good luck with that.
o_id: harley_decline
o_goto: Bye

name: Accepted
npc: Ratchet's chop shop in the west took the ignition, I'd bet my teeth. Fuel cells end up in burned-out cars over there. The tank? Scavengers at the wall buy chrome. Three hundred when she runs.
effect: SetFlag('JobAccepted(old_harley)')
effect: EndConversation

name: Waiting
npc: Ignition, fuel cell, tank. All three, or she doesn't breathe.
#
o_text: (Hand over all three parts.)
o_id: harley_give_parts
o_cond: HasItem('harley_ignition') && HasItem('harley_fuel_cell') && HasItem('harley_tank')
o_goto: Rebuilding
#
o_text: Still looking.
o_id: harley_looking
o_goto: Bye

name: Rebuilding
npc: (He lays the parts out on a rag like surgical tools.) Come back in six hours. Don't talk to me before then.
effect: StackTransferTo(NPC, 'harley_ignition', 1)
effect: StackTransferTo(NPC, 'harley_fuel_cell', 1)
effect: StackTransferTo(NPC, 'harley_tank', 1)
effect: SaveTimeNow('harley_t')
effect: SetFlag('harley_parts_given')
effect: EndConversation

name: Working
npc: (Not looking up.) Six hours, I said.
effect: EndConversation

name: Ride
npc: (He kicks her and she ROARS. Every window on the street opens.) Hear that? Three hundred. And any time you need a ride, you ask old Dizzy.
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('harley_rebuilt')
effect: EndConversation

name: Done
npc: Keep the rubber side down.
effect: EndConversation

name: Bye
npc: She's waiting.
effect: EndConversation
''')
p.dialogue('chopper_ratchet', '''# F3: Ratchet has the ignition. Buy it (150), talk it free (social Easy), or sell him all three parts (450).
%rec: OpeningBranch

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Parts, labour, no questions. What do you need?
#
o_text: That Harley ignition. It's Dizzy's. He's old and it's all he has.
o_id: ratchet_ignition_talk
o_cond: HasFlag('JobAccepted(old_harley)') && !HasItem('harley_ignition') && !HasFlag('ratchet_ignition_gone')
o_test: RollSkill('social', 'Easy')
o_succ: GiveIgnition
o_fail: Refuse
#
o_text: (Buy the ignition.) 150.
o_id: ratchet_ignition_buy
o_cond: HasFlag('JobAccepted(old_harley)') && HasItem('gold', 150) && !HasFlag('ratchet_ignition_gone')
o_goto: SellIgnition
#
o_text: (Show him all three parts.) What would a corporate collector pay?
o_id: ratchet_buy_parts
o_cond: HasItem('harley_ignition') && HasItem('harley_fuel_cell') && HasItem('harley_tank') && !HasFlag('harley_parts_given')
o_goto: BuyParts
#
o_text: Business good?
o_id: ratchet_lore
o_goto: Lore
#
o_text: Bye.
o_id: ratchet_bye
o_goto: Bye

name: GiveIgnition
npc: (He sighs.) Dizzy's? I didn't know. Here. Tell him no hard feelings.
effect: StackTransferFrom(NPC, 'harley_ignition', 1)
effect: SetFlag('ratchet_ignition_gone')
effect: EndConversation

name: SellIgnition
npc: Pleasure.
effect: StackTransferTo(NPC, 'gold', 150)
effect: StackTransferFrom(NPC, 'harley_ignition', 1)
effect: SetFlag('ratchet_ignition_gone')
effect: EndConversation

name: Refuse
npc: It's mine now. That's how parts work.
effect: EndConversation

name: BuyParts
npc: (He whistles.) A complete set. There's a corporate kid who wants a real racer. Four-fifty for the lot, cash.
#
o_text: Sold.
o_id: ratchet_parts_sold
o_goto: PartsSold
#
o_text: No. They're Dizzy's.
o_id: ratchet_parts_no
o_goto: Bye

name: PartsSold
npc: Smart. Old men and old bikes, both on borrowed time.
effect: StackTransferTo(NPC, 'harley_ignition', 1)
effect: StackTransferTo(NPC, 'harley_fuel_cell', 1)
effect: StackTransferTo(NPC, 'harley_tank', 1)
effect: StackTransferFrom(NPC, 'gold', 450)
effect: SetFlag('harley_sold_ratchet')
effect: EndConversation

name: Lore
npc: Since the dome sealed nothing new comes in. Old parts on older machines. I'm not a thief, I'm an archaeologist.
#
o_text: Back.
o_id: ratchet_lore_back
o_goto: Default

name: Bye
npc: Mind the oil.
effect: EndConversation
''')
p.extend_dialogue('scavenger_mott', 'Default', '''o_text: You have a chrome bike tank. I'll buy it. (100)
o_id: mott_buy_tank
o_cond: HasFlag('JobAccepted(old_harley)') && HasItem('gold', 100) && !HasItem('harley_tank') && !HasFlag('mott_tank_sold')
o_goto: MottTank''', '''name: MottTank
npc: Flames and all. Some fool carried it to the wall to sell for water.
effect: StackTransferTo(NPC, 'gold', 100)
effect: PlayerAddItem('harley_tank')
effect: SetFlag('mott_tank_sold')
effect: EndConversation''')
p.journal_entry('''id: old_harley
name: The '94 Harley
skill_points: 4
#
start_cond: HasFlag('JobAccepted(old_harley)')
start_text: Dizzy's 1994 Harley was stripped: ignition, fuel cell, chrome tank. Ratchet's chop shop in the west, a burned-out car over there, and the scavengers at the wall.
#
prog_cond: HasItem('harley_ignition') && HasItem('harley_fuel_cell') && HasItem('harley_tank')
prog_text: All three parts. Dizzy is in the south. Ratchet would pay more.
#
end_cond: HasFlag('harley_rebuilt')
end_text: Dizzy's Harley roars. The only real engine in the dome, and you can hear it from the wall.
end_id: rebuilt
#
end_cond: HasFlag('harley_sold_ratchet')
end_text: Ratchet paid 450 for the parts. A corporate kid has a racer. Dizzy still polishes a sculpture.
end_id: sold
end_skill_points: 2
''')
for name,talk in (('f_old_harley_rebuilt',False),('f_old_harley_sold',True)):
    t=Trip().do("Give('gold(300)')","Talk('biker_dizzy')","Choose('harley_accept')").go('zone_residential_west').do("Talk('chopper_ratchet')","Choose('ratchet_ignition_buy')").do("Give('mechanical_lockpick(5)')","SetSkill('Mechanics', 200)","Use('burned_car')","Take('harley_fuel_cell')").go('zone_residential_south').go('zone_wall_south').do("Talk('scavenger_mott')","Choose('mott_buy_tank')")
    if talk:
        t.go('zone_residential_south').go('zone_residential_west').do("Talk('chopper_ratchet')","Choose('ratchet_buy_parts')","Choose('ratchet_parts_sold')")
        p.playtest(name, "The '94 Harley (grey): collect the parts and sell them to Ratchet.", "HasFlag('QuestCompleted(old_harley, sold)')", t)
    else:
        t.go('zone_residential_south').do("Talk('biker_dizzy')","Choose('harley_give_parts')").do("WaitHours(6)").do("Talk('biker_dizzy')")
        p.playtest(name, "The '94 Harley: three parts on three maps, six hours of work, one roar.", "HasFlag('QuestCompleted(old_harley, rebuilt)')", t)
# ===== F4 Ecto: a night stakeout; the ghost only appears after dark =====
p.actor('zone_commerce', 'club_lou', 'Lou', 'The manager of club Ecto, white-haired at thirty, with eyes that keep going to the corners of the room.', 'L', near=(50,10), equipment=['gold(400)'])
p.named_location('zone_commerce', 'ecto_floor', near=(54,10))
p.object('zone_commerce', '''Category: StateChanger
Name: ecto_junction_box
Description: a humming junction box
state_0_label: a humming junction box
state_0_change: ClearFlag('ecto_power_cut')
state_0_rune: J
state_1_label: a dead junction box
state_1_change: SetFlag('ecto_power_cut') && RemoveActor('ecto_ghost')
state_1_rune: j''', near=(47,8))
p.spawnable('ecto_ghost', 'a white face', 'A pale face hanging in the air, mouth open in a scream with no sound. When it turns, you see a cable glint behind it.', '@', hp=999, Dialogue='ecto_ghost', faction='ecto')
p.dialogue('club_lou', '''# F4 Ecto. The club is haunted at night (script f_ecto spawns the ghost). Cut the power, or talk to whoever is behind it.
%rec: OpeningBranch

cond: HasFlag('ecto_fixed') || HasFlag('ecto_lied')
goto: Done

cond: HasFlag('JobAccepted(ecto)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: My club's haunted. After dark there's a face on the dance floor. My last manager they found in a cupboard, skinny as a rake, hair white as mine. Nobody comes any more.
#
o_text: Ghosts aren't real. I'll stay the night.
o_id: ecto_accept
o_goto: Accepted
#
o_text: Call a priest.
o_id: ecto_decline
o_goto: Bye

name: Accepted
npc: It comes after ten. Stay on the floor and don't look it in the eyes. Make it stop and you get three-fifty.
effect: SetFlag('JobAccepted(ecto)')
effect: RunScript('f_ecto')
effect: EndConversation

name: Waiting
npc: Well?
#
o_text: It was a projector. I cut its power.
o_id: ecto_report_cut
o_cond: HasFlag('ecto_power_cut')
o_goto: Paid
#
o_text: It was a kid with a projector. She's stopped.
o_id: ecto_report_kid
o_cond: HasFlag('ecto_pixel_stopped')
o_goto: Paid
#
o_text: It's gone. You can open tonight.
o_id: ecto_report_lie
o_cond: HasFlag('ecto_haunt_kept')
o_goto: LiePaid
#
o_text: Still watching.
o_id: ecto_report_wait
o_goto: Bye

name: Paid
npc: A box? I went white over a box? (He laughs until he coughs.) Three-fifty. Drinks are on the ghost.
effect: StackTransferFrom(NPC, 'gold', 350)
effect: SetFlag('ecto_fixed')
effect: EndConversation

name: LiePaid
npc: Gone? (He believes you, because he wants to.) Three-fifty. Opening tonight.
effect: StackTransferFrom(NPC, 'gold', 350)
effect: SetFlag('ecto_lied')
effect: EndConversation

name: Done
npc: No ghosts tonight. Just paying customers. Mostly.
effect: EndConversation

name: Bye
npc: Stay out of the mirrors.
effect: EndConversation
''')
p.dialogue('ecto_ghost', '''# F4: the ghost. Trace the cable, or talk back through the projector to Pixel.
%rec: OpeningBranch

cond: true
goto: Default

%rec: Nodes

name: Default
npc: (A white face, mouth open, no sound. It flickers when you step closer.)
#
o_text: (Follow the flicker.) There's a cable behind it.
o_id: ghost_trace
o_goto: Traced
#
o_text: (Technology) Find the projector's mic and talk back through it.
o_id: ghost_talk
o_test: RollSkill('technology', 'Easy')
o_succ: Pixel
o_fail: Static
#
o_text: (Back away.)
o_id: ghost_leave
o_goto: Bye

name: Traced
npc: (The cable runs along the ceiling to a humming junction box by the back wall.)
effect: SetFlag('ecto_traced')
effect: EndConversation

name: Static
npc: (Static. The face screams louder, without a sound.)
effect: EndConversation

name: Pixel
npc: (The face stops. A girl's voice, young and scared.) Who is this? ... Okay. Okay. I'm Pixel. Lou cheated my mom out of this club. The ghost is the only thing that ever scared him.
#
o_text: Stop now, before EBI traces you instead of me.
o_id: ghost_pixel_stop
o_goto: PixelStops
#
o_text: Keep haunting. I'll tell Lou it's fixed, and you pay me.
o_id: ghost_pixel_deal
o_goto: PixelDeal

name: PixelStops
npc: EBI? ... Fine. Fine! It's off. Tell him... no. Don't tell him anything. (The face fades out.)
effect: SetFlag('ecto_pixel_stopped')
effect: RemoveActor('ecto_ghost')
effect: EndConversation

name: PixelDeal
npc: (A laugh.) Deal. Two hundred, dropped in your account. Boo.
effect: PlayerAddGold(200)
effect: SetFlag('ecto_haunt_kept')
effect: EndConversation

name: Bye
npc: (It watches you go.)
effect: EndConversation
''')
p.script('f_ecto', '''# F4: at night the ghost appears on Ecto's floor.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: IsNight()
do: SpawnActor('ecto_ghost', 'zone_commerce', 'ecto_floor')
do: SetFlag('ecto_ghost_seen')
''')
p.journal_entry('''id: ecto
name: Ecto
skill_points: 4
#
start_cond: HasFlag('JobAccepted(ecto)')
start_text: Lou's club Ecto in the commerce district is haunted after dark. Stay the night and make it stop. 350 sat.
#
prog_cond: HasFlag('ecto_ghost_seen')
prog_text: The ghost is real enough to see. It flickers like a picture.
#
end_cond: HasFlag('ecto_lied')
end_text: Pixel paid you and Lou paid you. The ghost still dances at Ecto. Somebody had it coming.
end_id: haunted
#
end_cond: HasFlag('ecto_fixed') && HasFlag('ecto_power_cut')
end_text: The ghost was a projector on a junction box. You pulled the plug. Ecto is open again.
end_id: power_cut
#
end_cond: HasFlag('ecto_fixed')
end_text: The ghost was a girl called Pixel with a grudge. She switched it off herself.
end_id: talked
''')
base=lambda: Trip().go('zone_commerce').do("Talk('club_lou')","Choose('ecto_accept')").do("WaitUntil('23:00')")
p.playtest('f_ecto_cut', 'Ecto: wait for the dark, follow the cable, kill the power.', "HasFlag('QuestCompleted(ecto, power_cut)')", base().do("Talk('ecto_ghost')","Choose('ghost_trace')").do("Interact('ecto_junction_box', 'a dead junction box')").do("Talk('club_lou')","Choose('ecto_report_cut')"))
p.playtest('f_ecto_talked', 'Ecto (talker): talk back through the projector, Pixel stops.', "HasFlag('QuestCompleted(ecto, talked)')", base().do("ForceChecks('success')","Talk('ecto_ghost')","Choose('ghost_talk')","Choose('ghost_pixel_stop')").do("Talk('club_lou')","Choose('ecto_report_kid')"))
p.playtest('f_ecto_haunted', 'Ecto (grey): take Pixel\'s money, lie to Lou.', "HasFlag('QuestCompleted(ecto, haunted)')", base().do("ForceChecks('success')","Talk('ecto_ghost')","Choose('ghost_talk')","Choose('ghost_pixel_deal')").do("Talk('club_lou')","Choose('ecto_report_lie')"))

# ===== F5 False Papers: courier, a twelve-hour wait, or sell her face to EBI =====
p.actor('zone_residential_south', 'refugee_amira', 'Amira', 'A refugee woman with two children asleep against her legs. She got inside the dome the night before it sealed and has had no ID since.', 'A', near=(55,15), equipment=['amira_ring'])
p.actor('zone_residential_west', 'forger_nils', 'Nils', 'A soft-spoken forger with magnifying lenses and very clean hands. Papers take time, he says. Art takes time.', 'N', near=(55,12), equipment=['gold(400)', 'forged_sin'])
p.item('amira_photo', 'a passport photo', 'A photo of Amira, taken against a bedsheet. She tried to smile.', 5)
p.item('forged_sin', 'a forged SIN card', 'A forged identity card. Ration-ready. The name on it belongs to nobody, which is the point.', 250)
p.item('amira_ring', 'a wedding ring', 'A thin gold ring. Her husband did not make it through the gate.', 200)
p.dialogue('refugee_amira', '''# F5 False Papers. Carry her photo to Nils, wait twelve hours, carry the card back. Or sell her face to EBI.
%rec: OpeningBranch

cond: HasFlag('papers_delivered')
goto: Done

cond: HasFlag('JobAccepted(false_papers)') && HasItem('forged_sin')
goto: Return

cond: HasFlag('JobAccepted(false_papers)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: No SIN, no ration card. No ration card, my children don't eat. There is a forger in the west, Nils. They say his papers take forever to make. Forever is fine. Hunger isn't.
#
o_text: I'll take your photo to him.
o_id: papers_accept
o_goto: Accepted
#
o_text: I can't help.
o_id: papers_decline
o_goto: Bye

name: Accepted
npc: (She gives you a photo and a fold of notes.) A hundred. It's all of it. He asks two-fifty. Please.
effect: SetFlag('JobAccepted(false_papers)')
effect: PlayerAddItem('amira_photo')
effect: PlayerAddGold(100)
effect: EndConversation

name: Waiting
npc: Did he say how long?
effect: EndConversation

name: Return
npc: (She reads the name on the card. It isn't hers. She practises saying it.)
#
o_text: (Give her the card.)
o_id: papers_give
o_goto: Paid

name: Paid
npc: My children have a name now. (She presses her ring into your hand and won't take it back.)
effect: StackTransferTo(NPC, 'forged_sin', 1)
effect: StackTransferFrom(NPC, 'amira_ring', 1)
effect: SetFlag('papers_delivered')
effect: EndConversation

name: Done
npc: (One of the children waves at you.)
effect: EndConversation

name: Bye
npc: Go well.
effect: EndConversation
''')
p.dialogue('forger_nils', '''# F5: Nils forges Amira's papers in twelve hours (pay or intimidate), or buys her photo for EBI.
%rec: OpeningBranch

cond: HasFlag('nils_working') && IsHoursAfter('nils_t', 12) && !HasFlag('papers_picked')
goto: Ready

cond: HasFlag('nils_working') && !HasFlag('papers_picked')
goto: NotReady

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Papers are an art. Art has a price. What do you need?
#
o_text: (Show him Amira's photo.) Papers for this woman.
o_id: nils_order
o_cond: HasItem('amira_photo')
o_goto: Price
#
o_text: Who buys papers?
o_id: nils_lore
o_goto: Lore
#
o_text: Nothing.
o_id: nils_bye
o_goto: Bye

name: Price
npc: Two-fifty. Twelve hours. Good papers take time.
#
o_text: (Pay.) Two-fifty.
o_id: nils_pay
o_cond: HasItem('gold', 250)
o_goto: Started
#
o_text: You'll do it for nothing, or EBI hears your name.
o_id: nils_threaten
o_test: RollSkill('intimidate', 'Easy')
o_succ: StartedFree
o_fail: Laughs
#
o_text: EBI pays for faces like hers, doesn't it?
o_id: nils_sell_photo
o_goto: SellOut

name: Started
npc: Twelve hours.
effect: StackTransferTo(NPC, 'amira_photo', 1)
effect: StackTransferTo(NPC, 'gold', 250)
effect: SaveTimeNow('nils_t')
effect: SetFlag('nils_working')
effect: EndConversation

name: StartedFree
npc: (A thin smile.) For the cause, then. Twelve hours.
effect: StackTransferTo(NPC, 'amira_photo', 1)
effect: SaveTimeNow('nils_t')
effect: SetFlag('nils_working')
effect: EndConversation

name: Laughs
npc: EBI buys my papers too. Threaten someone else.
effect: EndConversation

name: SellOut
npc: (He takes the photo with two fingers.) Three hundred. An EBI clerk collects faces of undocumented people. What happens next is not my department.
effect: StackTransferTo(NPC, 'amira_photo', 1)
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('papers_sold')
effect: RunScript('f_amira_taken')
effect: EndConversation

name: NotReady
npc: Forever takes a while.
effect: EndConversation

name: Ready
npc: (He blows on the card.) Nobody's name, nobody's face. Perfect.
effect: StackTransferFrom(NPC, 'forged_sin', 1)
effect: SetFlag('papers_picked')
effect: EndConversation

name: Lore
npc: Refugees. Runaways. EBI, when they want someone to disappear on paper first. A Saudi passport with a level-three chip, four thousand. These things take forever to make.
#
o_text: Back.
o_id: nils_lore_back
o_goto: Default

name: Bye
npc: Stay legal.
effect: EndConversation
''')
p.script('f_amira_taken', '''# F5: Amira's photo reached EBI. A day later she and her children are gone.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('amira_sold_t')

if: IsHoursAfter('amira_sold_t', 8)
do: LoadMap('zone_residential_south')
do: RemoveActor('refugee_amira')
do: SetFlag('amira_taken')
''')
p.journal_entry('''id: false_papers
name: False Papers
skill_points: 4
#
start_cond: HasFlag('JobAccepted(false_papers)')
start_text: Amira, a refugee in the south district, needs papers for her family. Take her photo to Nils the forger in the west. He asks 250; she has 100.
#
prog_cond: HasFlag('nils_working')
prog_text: Nils is making the papers. Twelve hours.
#
end_cond: HasFlag('papers_delivered')
end_text: Amira's children have a name and a ration card. She gave you her wedding ring.
end_id: delivered
#
end_cond: HasFlag('papers_sold')
end_text: You sold Amira's face to an EBI clerk for 300. The next day her corner was empty.
end_id: sold
end_skill_points: 1
''')
t=Trip().do("Talk('refugee_amira')","Choose('papers_accept')").go('zone_residential_west').do("Talk('forger_nils')","Choose('nils_order')","Choose('nils_pay')").do("WaitHours(12)").do("Talk('forger_nils')").go('zone_residential_south').do("Talk('refugee_amira')","Choose('papers_give')")
p.playtest('f_false_papers_paid', 'False Papers: carry the photo, pay, wait twelve hours, carry the card back.', "HasFlag('QuestCompleted(false_papers, delivered)') && HasItem('amira_ring')", t)
t=Trip().do("Talk('refugee_amira')","Choose('papers_accept')").go('zone_residential_west').do("ForceChecks('success')","Talk('forger_nils')","Choose('nils_order')","Choose('nils_threaten')").do("WaitHours(12)").do("Talk('forger_nils')").go('zone_residential_south').do("Talk('refugee_amira')","Choose('papers_give')")
p.playtest('f_false_papers_threat', 'False Papers (talker): lean on Nils for free papers.', "HasFlag('QuestCompleted(false_papers, delivered)') && HasItem('gold', 250)", t)
t=Trip().do("Talk('refugee_amira')","Choose('papers_accept')").go('zone_residential_west').do("Talk('forger_nils')","Choose('nils_order')","Choose('nils_sell_photo')").do("WaitHours(9)")
p.playtest('f_false_papers_sold', 'False Papers (grey): sell her face to EBI.', "HasFlag('QuestCompleted(false_papers, sold)') && HasFlag('amira_taken')", t)

# ===== F6 Payback: an information job; you never carry anything =====
p.actor('zone_residential_east', 'junkie_stripe', 'Stripe', 'A twitchy man with a braindance jack burnt black. Somebody sold him out, and he cannot stop talking about it.', 'S', near=(25,18), equipment=['gold(500)'])
p.actor('zone_wall_south', 'informant_pinky', 'Pinky', 'A nervous man with pink hair and colour-shift contacts, hiding among the scavengers at the wall. He is always facing the street.', 'P', near=(56,6), equipment=['gold(180)'])
p.dialogue('junkie_stripe', '''# F6 Payback. Stripe pays for knowledge: where is Pinky? What happens next happens off-screen (script f_pinky_dies).
%rec: OpeningBranch

cond: HasFlag('pinky_sold') || HasFlag('pinky_killed_paid') || HasFlag('stripe_lied') || HasFlag('stripe_doubts')
goto: Done

cond: HasFlag('JobAccepted(payback)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: A thousand for knowledge, that's what I'd pay if I had it. I've got three hundred. Pink hair, shift tacts, calls himself Pinky. Sold me out to EBI. I want to know where he sleeps.
#
o_text: I'll find him.
o_id: payback_accept
o_goto: Accepted
#
o_text: Find him yourself.
o_id: payback_decline
o_goto: Bye

name: Accepted
npc: Just find him. Tell me where. You don't have to do anything else. (He smiles. It's horrible.)
effect: SetFlag('JobAccepted(payback)')
effect: EndConversation

name: Waiting
npc: Well? Where?
#
o_text: He hides with the scavengers at the south wall.
o_id: payback_tell
o_cond: HasFlag('pinky_found') && !HasFlag('Killed(informant_pinky)') && !HasFlag('pinky_warned')
o_goto: Told
#
o_text: He's dead. I did it.
o_id: payback_dead
o_cond: HasFlag('Killed(informant_pinky)')
o_goto: DeadPaid
#
o_text: He went over the wall. He's gone.
o_id: payback_lie
o_cond: HasFlag('pinky_warned')
o_test: RollSkill('social', 'Easy')
o_succ: Believed
o_fail: Doubts
#
o_text: Still looking.
o_id: payback_looking
o_goto: Bye

name: Told
npc: The wall. (He's already standing.) Three hundred. Go home, contractor.
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('pinky_sold')
effect: RunScript('f_pinky_dies')
effect: EndConversation

name: DeadPaid
npc: (He laughs and cries at once.) You did it for me? Four hundred. Everything.
effect: StackTransferFrom(NPC, 'gold', 400)
effect: SetFlag('pinky_killed_paid')
effect: EndConversation

name: Believed
npc: Over the wall. (He sags.) Then the dead have him. Good. Take a hundred for your trouble.
effect: StackTransferFrom(NPC, 'gold', 100)
effect: SetFlag('stripe_lied')
effect: EndConversation

name: Doubts
npc: Over the wall. Sure. (He looks at you for a long time.) Get out.
effect: SetFlag('stripe_doubts')
effect: EndConversation

name: Done
npc: Yeah, yeah. We're done.
effect: EndConversation

name: Bye
npc: Pink hair. Don't forget.
effect: EndConversation
''')
p.dialogue('informant_pinky', '''# F6: Pinky. Finding him sets pinky_found. Warn him (he pays and runs) or kill him.
%rec: OpeningBranch

cond: HasFlag('pinky_warned')
goto: Leaving

cond: true
goto: Default

%rec: Nodes

name: Default
npc: I don't know you. I don't know anyone.
effect: SetFlag('pinky_found')
#
o_text: Stripe is looking for you.
o_id: pinky_warn
o_cond: HasFlag('JobAccepted(payback)')
o_goto: Plead
#
o_text: Why are you hiding at the wall?
o_id: pinky_lore
o_goto: Lore
#
o_text: Nothing.
o_id: pinky_bye
o_goto: Bye

name: Plead
npc: (He goes grey.) EBI made me. They asked nicely and then not nicely. Please.
#
o_text: Run. Tonight.
o_id: pinky_run
o_goto: Warned
#
o_text: Just checking it was you.
o_id: pinky_checked
o_goto: Bye

name: Warned
npc: (He empties his pockets into your hands.) One-eighty, all of it. Tell him I'm dead. Tell him anything.
effect: StackTransferFrom(NPC, 'gold', 180)
effect: SetFlag('pinky_warned')
effect: RunScript('f_pinky_leaves')
effect: EndConversation

name: Leaving
npc: I'm going, I'm going.
effect: EndConversation

name: Lore
npc: Because nobody looks at the wall. Nobody wants to see what's out there. Best hiding place in the dome.
#
o_text: Back.
o_id: pinky_lore_back
o_goto: Default

name: Bye
npc: You didn't see me.
effect: EndConversation
''')
p.extend_dialogue('noodle_vendor_suki', 'Default', '''o_text: Seen a man with pink hair and shift tacts?
o_id: suki_pinky
o_cond: HasFlag('JobAccepted(payback)') && !HasFlag('pinky_found')
o_goto: SukiPinky''', '''name: SukiPinky
npc: Pinky? Bought noodles with shaking hands, then went south to the wall to hide with the scavengers. Everybody knows. He's not good at hiding.
effect: SetFlag('pinky_found')
effect: EndConversation''')
p.script('f_pinky_dies', '''# F6: you told Stripe. Twelve hours later Pinky is dead at the wall.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('pinky_sold_t')

if: IsHoursAfter('pinky_sold_t', 12)
do: LoadMap('zone_wall_south')
do: ActorsDie('informant_pinky')
do: SetFlag('pinky_dead')
''')
p.script('f_pinky_leaves', '''# F6: warned, Pinky leaves the dome by morning.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('pinky_warned_t')

if: IsHoursAfter('pinky_warned_t', 2)
do: LoadMap('zone_wall_south')
do: RemoveActor('informant_pinky')
''')
p.journal_entry('''id: payback
name: Payback
skill_points: 4
#
start_cond: HasFlag('JobAccepted(payback)')
start_text: Stripe, a lace junkie in the eastern district, pays 300 to know where Pinky hides, the informant who sold him to EBI. Just the where.
#
prog_cond: HasFlag('pinky_found')
prog_text: Pinky hides with the scavengers at the south wall. Stripe only wants to know where.
#
end_cond: HasFlag('pinky_sold')
end_text: You told Stripe where Pinky sleeps. You never asked what happened next. The scavengers found him in the morning.
end_id: sold
#
end_cond: HasFlag('pinky_killed_paid')
end_text: You killed Pinky yourself and Stripe paid everything he had.
end_id: killed
#
end_cond: HasFlag('stripe_lied')
end_text: Pinky ran and Stripe believes he went over the wall.
end_id: lied
#
end_cond: HasFlag('stripe_doubts')
end_text: Pinky ran. Stripe didn't believe you, and now he wonders about you.
end_id: doubted
end_skill_points: 2
''')
t=Trip().go('zone_residential_east').do("Talk('junkie_stripe')","Choose('payback_accept')").do("Talk('noodle_vendor_suki')","Choose('suki_pinky')").do("Talk('junkie_stripe')","Choose('payback_tell')").do("WaitHours(13)")
p.playtest('f_payback_sold', 'Payback: ask around, sell the location, never meet Pinky. He is dead by morning.', "HasFlag('QuestCompleted(payback, sold)') && HasFlag('pinky_dead') && !HasFlag('TalkedTo(informant_pinky)')", t)
t=Trip().go('zone_residential_east').do("Talk('junkie_stripe')","Choose('payback_accept')").go('zone_wall_south').do(*GUN,"Kill('informant_pinky')").do("Unequip('10mm_smg')").go('zone_residential_east').do("Talk('junkie_stripe')","Choose('payback_dead')")
p.playtest('f_payback_killed', 'Payback (fighter): do it yourself.', "HasFlag('QuestCompleted(payback, killed)')", t)
t=Trip().go('zone_residential_east').do("Talk('junkie_stripe')","Choose('payback_accept')").go('zone_wall_south').do("Talk('informant_pinky')","Choose('pinky_warn')","Choose('pinky_run')").go('zone_residential_east').do("ForceChecks('success')","Talk('junkie_stripe')","Choose('payback_lie')")
p.playtest('f_payback_lied', 'Payback (grey): warn Pinky, take his money, lie to Stripe.', "HasFlag('QuestCompleted(payback, lied)')", t)

# ===== F7 Carmen: gather evidence, then accuse, confront or sell out =====
p.actor('zone_residential_west', 'detective_rourke', 'Detective Rourke', 'An old city detective EBI forgot to fire. Five bodies in the west and nobody but him counting.', 'R', near=(40,10), equipment=['gold(500)'])
p.corpse('zone_residential_west', 'carmen_victim', 'the latest victim, a velvet ribbon round his wrist', near=(44,14), equipment=['velvet_ribbon'])
p.actor('zone_commerce', 'singer_carmen', 'Carmen', 'A singer in a long coat, beautiful and very calm. A velvet ribbon in her hair.', 'C', near=(60,18), equipment=['gold(600)', 'scalpel'], hp=40)
p.actor('zone_commerce', 'stagehand_tito', 'Tito', 'A stagehand coiling cable by the bar’s back door. He sees everyone leave and never says so.', 'T', near=(64,16), equipment=['gold(20)'])
p.object('zone_commerce', '''Category: UnknownContainer
Name: carmen_trunk
Description: a dressing-room trunk
lockdifficulty: hard
item: carmen_diary''', near=(62,20))
p.item('velvet_ribbon', 'a velvet ribbon', 'A dark red velvet ribbon, tied in a neat bow. It came off a dead man’s wrist.', 5)
p.item('carmen_diary', 'Carmen’s diary', 'A diary in tidy handwriting. Five names. Five dates. Five reasons, each one worse.', 50)
EVID="((HasItem('velvet_ribbon') && HasFlag('carmen_witness')) || (HasItem('velvet_ribbon') && HasItem('carmen_diary')) || (HasFlag('carmen_witness') && HasItem('carmen_diary')))"
p.dialogue('detective_rourke', f'''# F7 Carmen. Two of three pieces of evidence (ribbon on the victim, Tito's testimony, the diary) and Rourke makes the arrest.
%rec: OpeningBranch

cond: HasFlag('carmen_arrested') || HasFlag('carmen_dead_paid')
goto: Done

cond: HasFlag('JobAccepted(carmen)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Five bodies since the dome sealed, men and women, same knife. My bosses call it gang trouble. Word on the street says Carmen ain't dead. The singer at the commerce bar.
#
o_text: I'll find you proof.
o_id: carmen_accept
o_goto: Accepted
#
o_text: Not my case.
o_id: carmen_decline
o_goto: Bye

name: Accepted
npc: Two pieces. One's not enough for upstairs. The latest victim is still lying behind the pawnshop, nobody's moved him. Somebody at the bar must have seen her leave with him. And if she keeps notes... Four hundred, my own money.
effect: SetFlag('JobAccepted(carmen)')
effect: EndConversation

name: Waiting
npc: What have you got?
#
o_text: It's enough. Arrest her.
o_id: carmen_accuse
o_cond: {EVID}
o_goto: Arrest
#
o_text: She's dead. She came at me.
o_id: carmen_dead
o_cond: HasFlag('Killed(singer_carmen)')
o_goto: DeadPaid
#
o_text: Not enough yet.
o_id: carmen_notyet
o_goto: Bye

name: Arrest
npc: (He reads, listens, nods.) That'll hold. I'll bring her in at dawn. Four hundred. And thank you. Really.
effect: StackTransferFrom(NPC, 'gold', 400)
effect: SetFlag('carmen_arrested')
effect: RunScript('f_carmen_arrest')
effect: EndConversation

name: DeadPaid
npc: (A long silence.) No trial, then. No answers for the families. (He pays half.) Two hundred.
effect: StackTransferFrom(NPC, 'gold', 200)
effect: SetFlag('carmen_dead_paid')
effect: EndConversation

name: Done
npc: No new bodies this week. First week since the dome sealed.
effect: EndConversation

name: Bye
npc: Lock your door tonight.
effect: EndConversation
''')
p.dialogue('singer_carmen', '''# F7: Carmen. Confront her (she attacks) or let her buy you.
%rec: OpeningBranch

cond: HasFlag('carmen_bought')
goto: Bought

cond: true
goto: Default

%rec: Nodes

name: Default
npc: You came to hear me sing? Everybody comes to hear me sing.
#
o_text: I know about the five.
o_id: carmen_confront
o_cond: HasFlag('JobAccepted(carmen)')
o_goto: Confront
#
o_text: Rourke is on to you. What's it worth to make that go away?
o_id: carmen_sellout
o_cond: HasFlag('JobAccepted(carmen)') && !HasFlag('carmen_arrested')
o_goto: Offer
#
o_text: Why does everyone think you're dead?
o_id: carmen_lore
o_goto: Lore
#
o_text: Another time.
o_id: carmen_bye
o_goto: Bye

name: Confront
npc: (Her smile doesn't move.) Then you know how this ends, darling.
effect: ProvokeFaction('singer_carmen')
effect: EndConversation

name: Offer
npc: (She counts out five hundred without looking.) Tell the old man the trail went cold. Sit in the front row sometime.
effect: StackTransferFrom(NPC, 'gold', 500)
effect: SetFlag('carmen_bought')
effect: EndConversation

name: Bought
npc: Front row, darling.
effect: EndConversation

name: Lore
npc: Because I wanted them to. A woman disappears, a man appears. The dome is small, darling. People are so easy to lose.
#
o_text: Back.
o_id: carmen_lore_back
o_goto: Default

name: Bye
npc: Sweet dreams.
effect: EndConversation
''')
p.dialogue('stagehand_tito', '''# F7: Tito saw Carmen leave with the latest victim. Ask nicely (social Easy) or pay 50.
%rec: OpeningBranch

cond: HasFlag('carmen_witness')
goto: Said

cond: true
goto: Default

%rec: Nodes

name: Default
npc: I just coil cable.
#
o_text: The night the last man died. Did Carmen leave with anyone?
o_id: tito_ask
o_cond: HasFlag('JobAccepted(carmen)')
o_test: RollSkill('social', 'Easy')
o_succ: Witness
o_fail: Silent
#
o_text: Bye.
o_id: tito_bye
o_goto: Bye

name: Silent
npc: I coil cable. Cable doesn't talk, I don't talk.
#
o_text: (Pay him 50.)
o_id: tito_pay
o_cond: HasItem('gold', 50)
o_goto: PaidWitness
#
o_text: Fine.
o_id: tito_silent_bye
o_goto: Bye

name: PaidWitness
npc: (The money goes.) She left with him. Back door. Red ribbon in her hair. I'll say it to the detective, once.
effect: StackTransferTo(NPC, 'gold', 50)
effect: SetFlag('carmen_witness')
effect: EndConversation

name: Witness
npc: (He looks at the stage door.) She left with him. Back door. Red ribbon in her hair. I'll say it to the detective, once.
effect: SetFlag('carmen_witness')
effect: EndConversation

name: Said
npc: I said it already.
effect: EndConversation

name: Bye
npc: Mind the cable.
effect: EndConversation
''')
p.script('f_carmen_arrest', '''# F7: Rourke brings Carmen in at dawn.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('carmen_arrest_t')

if: IsHoursAfter('carmen_arrest_t', 6)
do: LoadMap('zone_commerce')
do: RemoveActor('singer_carmen')
do: SetFlag('carmen_taken')
''')
p.journal_entry('''id: carmen
name: Carmen
skill_points: 6
perk_points: 1
#
start_cond: HasFlag('JobAccepted(carmen)')
start_text: Detective Rourke in the west thinks Carmen, a singer in the commerce district, killed five people. He needs two pieces of proof: the latest victim behind the pawnshop, a witness at the bar, her notes.
#
end_cond: HasFlag('carmen_arrested')
end_text: Rourke arrested Carmen at dawn. The killings in the west have stopped.
end_id: arrested
#
end_cond: HasFlag('carmen_dead_paid')
end_text: Carmen is dead. No trial, no answers for the families.
end_id: killed
#
end_cond: HasFlag('carmen_bought')
end_text: Carmen paid you 500. The bodies keep turning up in the west.
end_id: bought
end_skill_points: 1
''')
start=lambda: Trip().go('zone_residential_west').do("Talk('detective_rourke')","Choose('carmen_accept')")
p.playtest('f_carmen_ribbon_witness', 'Carmen: the ribbon off the victim and Tito\'s word, then the arrest.', "HasFlag('QuestCompleted(carmen, arrested)')", start().do("PickUp('velvet_ribbon')").go('zone_commerce').do("ForceChecks('success')","Talk('stagehand_tito')","Choose('tito_ask')").go('zone_residential_west').do("Talk('detective_rourke')","Choose('carmen_accuse')"))
p.playtest('f_carmen_diary', 'Carmen (avoider): bribe Tito, crack the trunk for her diary.', "HasFlag('QuestCompleted(carmen, arrested)') && HasItem('carmen_diary')", start().go('zone_commerce').do("ForceChecks('fail')","Talk('stagehand_tito')","Choose('tito_ask')","Choose('tito_pay')").do("ForceChecks('none')","SetSkill('Stealth', 200)","SetSkill('Mechanics', 200)","Give('mechanical_lockpick(5)')","Sneak()","Use('carmen_trunk')","Take('carmen_diary')").go('zone_residential_west').do("Talk('detective_rourke')","Choose('carmen_accuse')"))
p.playtest('f_carmen_confront', 'Carmen (fighter): tell her you know.', "HasFlag('QuestCompleted(carmen, killed)')", start().go('zone_commerce').do(*GUN,"Talk('singer_carmen')","Choose('carmen_confront')").do("Kill('singer_carmen')").do("Unequip('10mm_smg')").go('zone_residential_west').do("Talk('detective_rourke')","Choose('carmen_dead')"))
p.playtest('f_carmen_bought', 'Carmen (grey): she pays you to forget.', "HasFlag('QuestCompleted(carmen, bought)')", start().go('zone_commerce').do("Talk('singer_carmen')","Choose('carmen_sellout')"))
# ===== F8 Proxy Vote: a two-day deadline; earn, forge, or give the vote away =====
p.actor('zone_corporate', 'exec_carleton', 'Carleton', 'A board-room climber in a suit worth more than your apartment. He needs one more vote and has two days to get it.', 'C', near=(30,12), equipment=['gold(800)'])
p.actor('zone_residential_south', 'shareholder_hal', 'Hal', 'An old dockworker who got one share of Terra Vitae stock as a retirement gift. A broken robot dog sits at his feet.', 'H', near=(40,18), equipment=['proxy_card'])
p.actor('zone_residential_south', 'coop_agent_rhee', 'Rhee', 'An organiser for the food co-op, collecting shareholder proxies one old pensioner at a time.', 'R', near=(46,18), equipment=['gold(300)'])
p.item('proxy_card', 'a signed proxy card', 'Hal’s share vote, signed over in shaky capitals.', 50)
p.item('forged_proxy', 'a forged proxy card', 'Hal’s signature, better than Hal does it.', 50)
p.dialogue('exec_carleton', '''# F8 Proxy Vote. Two days (script f_proxy_deadline). Hal's real vote, a forged one, or nothing.
%rec: OpeningBranch

cond: HasFlag('proxy_done')
goto: Done

cond: HasFlag('proxy_too_late')
goto: TooLate

cond: HasFlag('JobAccepted(proxy)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: The board votes in two days. I'm one share short. An old dockworker in the south, Hal, holds it and doesn't answer his messages. Get me his proxy.
#
o_text: Two days. Deal.
o_id: proxy_accept
o_goto: Accepted
#
o_text: Not interested.
o_id: proxy_decline
o_goto: Bye

name: Accepted
npc: Five hundred if it's in my hand before the vote. Nothing after.
effect: SetFlag('JobAccepted(proxy)')
effect: RunScript('f_proxy_deadline')
effect: EndConversation

name: Waiting
npc: Clock's running.
#
o_text: (Hand over Hal's proxy.)
o_id: proxy_give_real
o_cond: HasItem('proxy_card')
o_goto: PaidReal
#
o_text: (Hand over the forged proxy.)
o_id: proxy_give_forged
o_cond: HasItem('forged_proxy')
o_goto: PaidForged
#
o_text: Working on it.
o_id: proxy_wait
o_goto: Bye

name: PaidReal
npc: Signed and all. Five hundred.
effect: StackTransferTo(NPC, 'proxy_card', 1)
effect: StackTransferFrom(NPC, 'gold', 500)
effect: SetFlag('proxy_done')
effect: SetFlag('proxy_real')
effect: EndConversation

name: PaidForged
npc: (He doesn't look closely. He doesn't want to.) Five hundred.
effect: StackTransferTo(NPC, 'forged_proxy', 1)
effect: StackTransferFrom(NPC, 'gold', 500)
effect: SetFlag('proxy_done')
effect: SetFlag('proxy_forged')
effect: EndConversation

name: TooLate
npc: The vote was yesterday. We lost by one. Get out of my building.
effect: EndConversation

name: Done
npc: Motion carried.
effect: EndConversation

name: Bye
npc: Two days.
effect: EndConversation
''')
p.dialogue('shareholder_hal', '''# F8: Hal signs for a favour: parts for his robot dog. Or he gives his vote to the co-op.
%rec: OpeningBranch

cond: HasFlag('hal_signed') || HasFlag('proxy_coop')
goto: Done

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Rex here was a good dog. Then a part went, and now he just sits. (He pats the dented metal head.)
#
o_text: I need your share vote for Carleton.
o_id: hal_ask
o_cond: HasFlag('JobAccepted(proxy)')
o_goto: Favour
#
o_text: (Give him the robot repair parts.)
o_id: hal_parts
o_cond: HasFlag('JobAccepted(proxy)') && HasItem('robot_repair_parts')
o_goto: Signed
#
o_text: Rhee says the co-op could use your vote.
o_id: hal_coop
o_cond: HasFlag('JobAccepted(proxy)') && HasFlag('rhee_met')
o_goto: Coop
#
o_text: Good dog.
o_id: hal_bye
o_goto: Bye

name: Favour
npc: Votes, votes. Nobody visits Hal unless they want his vote. Fix Rex. There's spare parts in a crate somewhere in the industry district. Then we talk.
effect: EndConversation

name: Signed
npc: (Rex's tail twitches, then wags.) Ha! Good boy! Here, son. Sign it however you want.
effect: StackTransferTo(NPC, 'robot_repair_parts', 1)
effect: StackTransferFrom(NPC, 'proxy_card', 1)
effect: SetFlag('hal_signed')
effect: EndConversation

name: Coop
npc: The co-op? They feed my neighbours. Carleton never fed anybody. (He signs it over to Rhee.)
effect: SetFlag('proxy_coop')
effect: EndConversation

name: Done
npc: Rex says hello.
effect: EndConversation

name: Bye
npc: Mind the dog.
effect: EndConversation
''')
p.dialogue('coop_agent_rhee', '''# F8: the co-op wants Hal's vote too.
%rec: OpeningBranch

cond: HasFlag('proxy_coop') && !HasFlag('rhee_paid')
goto: Thanks

cond: true
goto: Default

%rec: Nodes

name: Default
npc: One share, one vote. Terra Vitae's board wants to sell the hydroponics to Carleton's people. Every pensioner's proxy counts.
effect: SetFlag('rhee_met')
effect: EndConversation

name: Thanks
npc: Hal signed to us? (She hugs you, briefly, hard.) We can't pay what Carleton pays. One-fifty.
effect: StackTransferFrom(NPC, 'gold', 150)
effect: SetFlag('rhee_paid')
effect: EndConversation
''')
p.extend_dialogue('forger_nils', 'Default', '''o_text: Can you do a shareholder's signature? Hal, from the south.
o_id: nils_proxy
o_cond: HasFlag('JobAccepted(proxy)') && HasItem('gold', 200) && !HasFlag('proxy_done')
o_goto: NilsProxy''', '''name: NilsProxy
npc: Signatures are quick. Two hundred, and ten minutes. (He scribbles. It is better than Hal.)
effect: StackTransferTo(NPC, 'gold', 200)
effect: PlayerAddItem('forged_proxy')
effect: EndConversation''')
p.script('f_proxy_deadline', '''# F8: the vote is held two days after Carleton hires you.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('proxy_t')

if: IsDaysAfter('proxy_t', 2) && !HasFlag('proxy_done')
do: SetFlag('proxy_too_late')
''')
p.journal_entry('''id: proxy
name: Proxy Vote
skill_points: 4
#
start_cond: HasFlag('JobAccepted(proxy)')
start_text: Carleton in the corporate district needs Hal’s share vote, from the south district, within two days. 500 sat.
#
end_cond: HasFlag('proxy_real')
end_text: You fixed Hal’s robot dog and he signed. Carleton won his vote.
end_id: favour
#
end_cond: HasFlag('proxy_forged')
end_text: Nils forged Hal’s signature. Carleton won his vote, and Hal never knew he voted.
end_id: forged
#
end_cond: HasFlag('proxy_coop')
end_text: Hal gave his vote to the food co-op. Carleton lost by one.
end_id: coop
#
end_cond: HasFlag('proxy_too_late')
end_text: The vote came and went. Carleton lost by one.
end_id: too_late
end_skill_points: 0
''')
start=lambda: Trip().go('zone_corporate').do("Talk('exec_carleton')","Choose('proxy_accept')")
p.playtest('f_proxy_favour', 'Proxy Vote: fetch parts for Hal’s dog, he signs.', "HasFlag('QuestCompleted(proxy, favour)')", start().go('zone_residential_south').do("Talk('shareholder_hal')","Choose('hal_ask')").go('zone_industry').do("Use('tool_crate')","Take('robot_repair_parts')").go('zone_residential_south').do("Talk('shareholder_hal')","Choose('hal_parts')").go('zone_corporate').do("Talk('exec_carleton')","Choose('proxy_give_real')"))
p.playtest('f_proxy_forged', 'Proxy Vote (avoider): Nils forges the signature, Hal is never asked.', "HasFlag('QuestCompleted(proxy, forged)') && !HasFlag('TalkedTo(shareholder_hal)')", start().do("Give('gold(200)')").go('zone_residential_west').do("Talk('forger_nils')","Choose('nils_proxy')").go('zone_corporate').do("Talk('exec_carleton')","Choose('proxy_give_forged')"))
p.playtest('f_proxy_coop', 'Proxy Vote (grey): talk Hal into giving his vote to the co-op.', "HasFlag('QuestCompleted(proxy, coop)') && HasFlag('rhee_paid')", start().go('zone_residential_south').do("Talk('coop_agent_rhee')").do("Talk('shareholder_hal')","Choose('hal_coop')").do("Talk('coop_agent_rhee')"))
p.playtest('f_proxy_too_late', 'Proxy Vote: miss the deadline.', "HasFlag('QuestCompleted(proxy, too_late)')", start().do("WaitHours(49)"))

# ===== F9 A Holiday on Ice: the client means to burn you =====
p.actor('zone_commerce', 'agent_nueman', 'Nueman', 'A smiling man from an orbital consultancy. His teeth are perfect. His shoes are not from this dome.', 'N', near=(36,10), equipment=['gold(900)'])
p.actor('zone_corporate', 'engineer_morrow', 'Morrow', 'An orbital engineer stuck in the dome since the seal, with a briefcase he never puts down.', 'M', near=(48,14), equipment=['orbital_specs', 'gold(400)'])
p.object('zone_corporate', '''Category: UnknownContainer
Name: morrow_briefcase
Description: a steel briefcase
lockdifficulty: hard
item: orbital_specs''', near=(50,14))
p.item('orbital_specs', 'orbital mass-driver specs', 'A data chip of mass-driver specifications. Morrow’s life’s work. Orbital Air would kill for it, literally.', 600)
p.spawnable('nueman_cleaner', 'a man in a grey suit', 'A cleaner from an orbital consultancy. He has no face you will remember.', 'c', equipment=['10mm_pistol', '10mm_jhp(24)'], faction='nueman', aggressive='true')
p.team('nueman_cleaners', 'nueman_cleaner', ['nueman_cleaner'])
p.dialogue('agent_nueman', '''# F9 A Holiday on Ice. Nueman pays for Morrow's specs, then sends cleaners (script f_nueman). Unless you suspected and took insurance.
%rec: OpeningBranch

cond: HasFlag('nueman_paid') || HasFlag('nueman_insured')
goto: Done

cond: HasFlag('JobAccepted(holiday)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: An engineer called Morrow carries some specifications that belong to my employer. Recover them. Six hundred, and a nice holiday afterwards. Somewhere cold.
#
o_text: Sounds easy.
o_id: holiday_accept
o_goto: Accepted
#
o_text: No.
o_id: holiday_decline
o_goto: Bye

name: Accepted
npc: Corporate district. He never puts the briefcase down. (He smiles.)
effect: SetFlag('JobAccepted(holiday)')
effect: EndConversation

name: Waiting
npc: Well?
#
o_text: (Hand over the specs.)
o_id: holiday_give
o_cond: HasItem('orbital_specs')
o_goto: Paid
#
o_text: I know what you did to Orbital Air's last contractor. I left a copy with a lawyer.
o_id: holiday_insure
o_cond: HasItem('orbital_specs') && HasFlag('nueman_suspected')
o_test: RollSkill('intimidate', 'Easy')
o_succ: Insured
o_fail: Paid
#
o_text: Not yet.
o_id: holiday_wait
o_goto: Bye

name: Paid
npc: Lovely. (Six hundred, counted twice.) Enjoy your holiday.
effect: StackTransferTo(NPC, 'orbital_specs', 1)
effect: StackTransferFrom(NPC, 'gold', 600)
effect: SetFlag('nueman_paid')
effect: RunScript('f_nueman')
effect: EndConversation

name: Insured
npc: (The smile stays. The eyes stop.) How prudent. Eight hundred, then, and we never meet again.
effect: StackTransferTo(NPC, 'orbital_specs', 1)
effect: StackTransferFrom(NPC, 'gold', 800)
effect: SetFlag('nueman_insured')
effect: EndConversation

name: Done
npc: We're finished, you and I.
effect: EndConversation

name: Bye
npc: Pity.
effect: EndConversation
''')
p.dialogue('engineer_morrow', '''# F9: Morrow. Bluff him (technology Hard), learn about Nueman, or sell the specs back.
%rec: OpeningBranch

cond: HasFlag('specs_returned')
goto: Done

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Don't touch the case.
#
o_text: (Technology) Safety inspection. I need to verify the data on that chip.
o_id: morrow_bluff
o_cond: HasFlag('JobAccepted(holiday)') && !HasItem('orbital_specs')
o_test: RollSkill('technology', 'Hard')
o_succ: Bluffed
o_fail: Suspicious
#
o_text: A man called Nueman wants your specs.
o_id: morrow_nueman
o_cond: HasFlag('JobAccepted(holiday)')
o_goto: Warning
#
o_text: (Give the specs back.) These are yours.
o_id: morrow_return
o_cond: HasItem('orbital_specs')
o_goto: Returned
#
o_text: Never mind.
o_id: morrow_bye
o_goto: Bye

name: Bluffed
npc: An inspector? Finally, someone official. (He hands you the chip.) Bring it right back.
effect: StackTransferFrom(NPC, 'orbital_specs', 1)
effect: EndConversation

name: Suspicious
npc: You're no inspector.
effect: EndConversation

name: Warning
npc: Nueman. (He grips the case.) He hired a team for Orbital Air last year. They got the data. Then they got shot. He doesn't leave people who know.
effect: SetFlag('nueman_suspected')
effect: EndConversation

name: Returned
npc: (He holds it like a child.) Three hundred is all I have. Thank you.
effect: StackTransferTo(NPC, 'orbital_specs', 1)
effect: StackTransferFrom(NPC, 'gold', 300)
effect: SetFlag('specs_returned')
effect: EndConversation

name: Done
npc: I owe you.
effect: EndConversation

name: Bye
npc: Hm.
effect: EndConversation
''')
p.script('f_nueman', '''# F9: thirty minutes after the deal, Nueman's cleaners come for the contractor.
%rec: definitions

%rec: outcomes

%rec: cancel

%rec: frames

if: true
do: SaveTimeNow('nueman_t')

if: IsMinutesAfter('nueman_t', 30) && IsMap('zone_commerce')
do: SpawnTeamHunting('nueman_cleaners', 'zone_commerce', 'taxi_stand', 'player')
do: SetFlag('nueman_cleaners_sent')
''')
p.journal_entry('''id: holiday
name: A Holiday on Ice
skill_points: 5
#
start_cond: HasFlag('JobAccepted(holiday)')
start_text: Nueman in the commerce district pays 600 for an engineer’s specs. Morrow, corporate district, never puts his briefcase down.
#
end_cond: HasFlag('specs_returned')
end_text: You gave Morrow his specs back. Nueman will find someone else.
end_id: returned
#
end_cond: HasFlag('nueman_insured')
end_text: You knew what Nueman does to his contractors and made sure he couldn’t. He paid 800 instead.
end_id: insured
#
end_cond: HasFlag('nueman_paid') && HasFlag('Killed(nueman_cleaner)')
end_text: Nueman paid, then sent cleaners. They are the ones on holiday now.
end_id: survived
''')
start=lambda: Trip().go('zone_commerce').do("Talk('agent_nueman')","Choose('holiday_accept')")
p.playtest('f_holiday_survived', 'Holiday on Ice (fighter): bluff Morrow, deliver, then shoot the cleaners.', "HasFlag('QuestCompleted(holiday, survived)')", start().go('zone_corporate').do("ForceChecks('success')","Talk('engineer_morrow')","Choose('morrow_bluff')").go('zone_commerce').do("Talk('agent_nueman')","Choose('holiday_give')").do(*GUN,"WaitMinutes(31)").do("Kill('nueman_cleaner')").do("Kill('nueman_cleaner')"))
p.playtest('f_holiday_insured', 'Holiday on Ice (talker): learn who Nueman is first, then take insurance.', "HasFlag('QuestCompleted(holiday, insured)') && !HasFlag('nueman_cleaners_sent')", start().go('zone_corporate').do("ForceChecks('success')","Talk('engineer_morrow')","Choose('morrow_nueman')").do("Talk('engineer_morrow')","Choose('morrow_bluff')").go('zone_commerce').do("Talk('agent_nueman')","Choose('holiday_insure')").do("WaitMinutes(35)"))
p.playtest('f_holiday_returned', 'Holiday on Ice (avoider): crack the briefcase, then give it back to Morrow.', "HasFlag('QuestCompleted(holiday, returned)')", start().go('zone_corporate').do("SetSkill('Stealth', 200)","SetSkill('Mechanics', 200)","Give('mechanical_lockpick(5)')","Sneak()","Use('morrow_briefcase')","Take('orbital_specs')").do("Sneak()","Talk('engineer_morrow')","Choose('morrow_return')"))

# ===== F10 Deadline: a rescue where the rescued may not want rescuing =====
p.actor('zone_corporate', 'assistant_caitlin', 'Caitlin', 'A producer’s assistant with three phones, chewing her nails. Her presenter has not come home.', 'C', near=(22,16), equipment=['gold(700)'])
p.actor('zone_commerce', 'loanshark_fargo', 'Fargo', 'A loan shark with a gold tooth and a ledger. Holly Jones owes him, and he keeps her where he can see her.', 'F', near=(70,10), equipment=['gold(500)', 'holly_marker', '10mm_pistol', '10mm_jhp(24)'], hp=40)
for i in range(2): p.actor('zone_commerce', 'fargo_goon', 'a leg-breaker', 'One of Fargo’s collectors. Big hands, small patience.', 'g', near=(72,12), equipment=['club'], dialogue=False, hp=18, faction='loanshark_fargo')
p.actor('zone_commerce', 'holly_jones', 'Holly Jones', 'A news presenter, face known from every screen in the dome, now waiting tables in Fargo’s back room.', 'H', near=(68,8))
p.item('holly_marker', 'Holly’s marker', 'A promissory note: Holly Jones owes Fargo four hundred, plus whatever he feels like.', 5)
p.dialogue('assistant_caitlin', '''# F10 Deadline. Free Holly from Fargo. Holly decides whether to come back.
%rec: OpeningBranch

cond: HasFlag('deadline_paid')
goto: Done

cond: HasFlag('JobAccepted(deadline)')
goto: Waiting

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Holly Jones. The news. She hasn't been on air in a week. She gambled, she borrowed from a man called Fargo in the commerce district, and now she works off the debt in his back room. The show goes on without her in three days.
#
o_text: I'll bring her back.
o_id: deadline_accept
o_goto: Accepted
#
o_text: Not my problem.
o_id: deadline_decline
o_goto: Bye

name: Accepted
npc: Here's three hundred up front. Pay him, scare him, I don't care. Three-fifty more when she's on air.
effect: SetFlag('JobAccepted(deadline)')
effect: StackTransferFrom(NPC, 'gold', 300)
effect: EndConversation

name: Waiting
npc: Well?
#
o_text: She's on her way home.
o_id: deadline_home
o_cond: HasFlag('holly_home')
o_goto: Paid
#
o_text: She's free. She chose not to come back.
o_id: deadline_stays
o_cond: HasFlag('holly_stays')
o_goto: Stays
#
o_text: Working on it.
o_id: deadline_wait
o_goto: Bye

name: Paid
npc: (All three phones ring at once.) She's back! Three-fifty. You saved a show, you know that?
effect: StackTransferFrom(NPC, 'gold', 350)
effect: SetFlag('deadline_paid')
effect: EndConversation

name: Stays
npc: Chose? (She sits down.) ... Then it's her life. One-fifty. I hope she's happy.
effect: StackTransferFrom(NPC, 'gold', 150)
effect: SetFlag('deadline_paid')
effect: EndConversation

name: Done
npc: The show goes on.
effect: EndConversation

name: Bye
npc: Three days.
effect: EndConversation
''')
p.dialogue('loanshark_fargo', '''# F10: Fargo. Pay the marker, scare him off it, or fight him and his goons.
%rec: OpeningBranch

cond: HasFlag('holly_free')
goto: Done

cond: true
goto: Default

%rec: Nodes

name: Default
npc: Everyone pays Fargo. Some pay with money, some with time.
#
o_text: I'm here about Holly Jones.
o_id: fargo_holly
o_cond: HasFlag('JobAccepted(deadline)')
o_goto: Holly
#
o_text: Bye.
o_id: fargo_bye
o_goto: Bye

name: Holly
npc: Four hundred and she walks. Cash.
#
o_text: (Pay 400.)
o_id: fargo_pay
o_cond: HasItem('gold', 400)
o_goto: Paid
#
o_text: Tear up the marker, or I tear up you.
o_id: fargo_threaten
o_test: RollSkill('intimidate', 'Hard')
o_succ: Scared
o_fail: Fight
#
o_text: No.
o_id: fargo_no
o_goto: Bye

name: Paid
npc: Pleasure. (He tears the marker.) She's yours.
effect: StackTransferTo(NPC, 'gold', 400)
effect: SetFlag('holly_free')
effect: EndConversation

name: Scared
npc: (He weighs it. He tears the marker.) She was bad for business anyway.
effect: SetFlag('holly_free')
effect: EndConversation

name: Fight
npc: Boys.
effect: ProvokeFaction('loanshark_fargo')
effect: EndConversation

name: Done
npc: We're square.
effect: EndConversation

name: Bye
npc: Interest is daily.
effect: EndConversation
''')
p.dialogue('holly_jones', '''# F10: Holly is free once the marker is gone. She decides whether to go home.
%rec: OpeningBranch

cond: HasFlag('holly_home') || HasFlag('holly_stays')
goto: Done

cond: HasFlag('holly_free') || HasFlag('Killed(loanshark_fargo)')
goto: Free

cond: true
goto: Default

%rec: Nodes

name: Default
npc: More coffee? (She doesn't look up.)
effect: EndConversation

name: Free
npc: (She stares at you.) It's over? ... Caitlin sent you. The show.
#
o_text: Your show needs you. Go home.
o_id: holly_go_home
o_goto: Home
#
o_text: It's your choice. You don't owe anybody now.
o_id: holly_choice
o_goto: Choice

name: Home
npc: Yeah. Yeah, okay. I know how to read a teleprompter. (She takes off the apron.)
effect: SetFlag('holly_home')
effect: RemoveActor('holly_jones')
effect: EndConversation

name: Choice
npc: Choice. (She laughs.) Nobody asked me that in years. I read other people's lies for eight years. Here, at least, the coffee's honest. I'm staying.
effect: SetFlag('holly_stays')
effect: EndConversation

name: Done
npc: Coffee?
effect: EndConversation
''')
p.journal_entry('''id: deadline
name: Deadline
skill_points: 5
#
start_cond: HasFlag('JobAccepted(deadline)')
start_text: Holly Jones, the news presenter, works off a debt to Fargo, a loan shark in the commerce district. Her assistant Caitlin in the corporate district wants her back on air.
#
prog_cond: HasFlag('holly_free') || HasFlag('Killed(loanshark_fargo)')
prog_text: Holly is free. Whether she goes back is up to her.
#
end_cond: HasFlag('deadline_paid') && HasFlag('holly_home')
end_text: Holly Jones is back on air. Fargo’s back room needs a new waitress.
end_id: home
#
end_cond: HasFlag('deadline_paid') && HasFlag('holly_stays')
end_text: Holly is free and chose to stay. Caitlin paid less and understood more.
end_id: stayed
''')
start=lambda: Trip().go('zone_corporate').do("Talk('assistant_caitlin')","Choose('deadline_accept')").go('zone_commerce')
back=lambda t,o: t.go('zone_corporate').do("Talk('assistant_caitlin')",f"Choose('{o}')")
p.playtest('f_deadline_paid', 'Deadline: pay off the marker, send Holly home.', "HasFlag('QuestCompleted(deadline, home)')", back(start().do("Give('gold(100)')","Talk('loanshark_fargo')","Choose('fargo_holly')","Choose('fargo_pay')").do("Talk('holly_jones')","Choose('holly_go_home')"),'deadline_home'))
p.playtest('f_deadline_stayed', 'Deadline (talker): scare Fargo, give Holly the choice. She stays.', "HasFlag('QuestCompleted(deadline, stayed)')", back(start().do("ForceChecks('success')","Talk('loanshark_fargo')","Choose('fargo_holly')","Choose('fargo_threaten')").do("Talk('holly_jones')","Choose('holly_choice')"),'deadline_stays'))
p.playtest('f_deadline_fight', 'Deadline (fighter): the threat fails, so fight Fargo and his goons.', "HasFlag('QuestCompleted(deadline, home)') && HasFlag('Killed(loanshark_fargo)')", back(start().do(*GUN,"ForceChecks('fail')","Talk('loanshark_fargo')","Choose('fargo_holly')","Choose('fargo_threaten')").do("Kill('loanshark_fargo')").do("Kill('fargo_goon')").do("Kill('fargo_goon')").do("Unequip('10mm_smg')","Talk('holly_jones')","Choose('holly_go_home')"),'deadline_home'))


p.write()
