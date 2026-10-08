"""Content packs for contractor. See CONTENT.md.

A pack is one Python file (content/<pack>.py) that describes NPCs, objects, items, dialogues,
scripts, journal entries and playtests. Running it regenerates every file the pack owns, so
it is safe to edit and rerun:

  maps/<map>/actors_<pack>.rec, objects_<pack>.rec      placements (positions found automatically)
  definitions/{journal,items,actors,spawned_teams}_<pack>.rec
  dialogues/<name>.rec, scripts/<name>.rec               whole files the pack created
  playtests/<pack>_*.rec
  marked blocks inside other dialogues                   options added to existing NPCs

The list of owned files is kept in data_atom/packs/<pack>.manifest; files from the previous run
that the pack no longer writes are deleted.

CLI:
  python3 tools/content.py ref                 script functions, playtest verbs, maps and their spots
  python3 tools/content.py lint [pack]         unknown items, scripts, dialogues, spots, teams, flags, timers, options
  python3 tools/content.py remove <pack>       delete every file and block the pack generated
  python3 tools/content.py spot <map> X,Y      nearest free reachable tile to X,Y
"""
import collections, os, re, subprocess, sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DATA = os.path.join(ROOT, 'data_atom')
MAPS = os.path.join(DATA, 'maps')

# How a playtest gets from one map to another. Everything goes through the south hub by taxi;
# corporate and industry are on foot from commerce, the wall on foot from the south.
ROUTE = {
    'zone_residential_south': ([], []),
    'zone_commerce': ([["Talk('taxi_driver')", "Choose('go_eastern_rim')"]], [["Talk('taxi_driver')", "Choose('go_home')"]]),
    'zone_residential_east': ([["Talk('taxi_driver')", "Choose('go_residential_east')"]], [["Talk('taxi_driver')", "Choose('go_home')"]]),
    'zone_residential_west': ([["Talk('taxi_driver')", "Choose('go_residential_west')"]], [["Talk('taxi_driver')", "Choose('go_home')"]]),
    'hq_terra_vitae': ([["Talk('taxi_driver')", "Choose('go_terra_vitae')"]], [["Talk('taxi_driver')", "Choose('go_home')"]]),
    'zone_corporate': ([["Talk('taxi_driver')", "Choose('go_eastern_rim')"], ["GoTo('street_to_corporate')"]], [["GoTo('street_to_commerce')"], ["Talk('taxi_driver')", "Choose('go_home')"]]),
    'zone_industry': ([["Talk('taxi_driver')", "Choose('go_eastern_rim')"], ["GoTo('street_to_industry')"]], [["GoTo('street_to_commerce')"], ["Talk('taxi_driver')", "Choose('go_home')"]]),
    'zone_wall_south': ([["GoTo('street_to_wall')"]], [["GoTo('street_to_south')"]]),
}

# A fighter loadout for playtests.
GUN = ["Give('10mm_smg(100)')", "Give('10mm_jhp(48)')", "Give('10mm_ap(48)')", "Equip('10mm_smg')", "SetSkill('RangedCombat', 100)"]


class Trip:
    """Playtest frames that follow the player across maps: Trip().go('zone_commerce').do("Talk('lou')", "Choose('x')")."""

    def __init__(self):
        self.at, self.frames = 'zone_residential_south', []

    def go(self, m):
        if m not in ROUTE:
            raise ValueError(f'no route to {m}; add it to ROUTE in tools/content.py')
        if m != self.at:
            self.frames += ROUTE[self.at][1] + ROUTE[m][0]
            self.at = m
        return self

    def do(self, *steps):
        self.frames.append(list(steps))
        return self


# ---------- map reading ----------

_grids = {}


def grid(m):
    """The map as rows of characters ('.' is walkable floor), via cmd/mapdump."""
    if m not in _grids:
        out = subprocess.run(['go', 'run', './cmd/mapdump', os.path.join(MAPS, m)], cwd=ROOT, capture_output=True, text=True, check=True)
        _grids[m] = [line[3:] for line in out.stdout.splitlines()[1:]]
    return _grids[m]


def _map_files(m, kinds, skip_pack=None):
    for name in sorted(os.listdir(os.path.join(MAPS, m))):
        for kind in kinds:
            if name == kind + '.rec' or (name.startswith(kind + '_') and name.endswith('.rec') and name != f'{kind}_{skip_pack}.rec'):
                yield os.path.join(MAPS, m, name)


def taken(m, skip_pack=None):
    s = set()
    for f in _map_files(m, ('actors', 'objects', 'items'), skip_pack):
        for x, y in re.findall(r'Position: \((\d+),\s*(\d+)\)', open(f).read()):
            s.add((int(x), int(y)))
    return s


def reachable(m):
    """Floor tiles connected to the map's taxi stand or street exits."""
    g = grid(m)
    src = ''.join(open(f).read() for f in _map_files(m, ('objects',)))
    seeds = [(int(x), int(y)) for x, y in re.findall(r'Identifier: taxi_stand\n(?:.*\n)*?Position: \((\d+),\s*(\d+)\)', src)]
    seeds += [(int(x), int(y)) for x, y in re.findall(r'Location: street_to\w*\n(?:.*\n)*?Position: \((\d+),\s*(\d+)\)', src)]
    floor = lambda p: 0 <= p[1] < len(g) and 0 <= p[0] < len(g[p[1]]) and g[p[1]][p[0]] == '.'
    seen = {(sx + dx, sy + dy) for sx, sy in seeds for dx in (-1, 0, 1) for dy in (-1, 0, 1) if floor((sx + dx, sy + dy))}
    q = collections.deque(seen)
    while q:
        x, y = q.popleft()
        for n in ((x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)):
            if n not in seen and floor(n):
                seen.add(n)
                q.append(n)
    return seen


def free_spot(m, near, avoid=frozenset(), skip_pack=None):
    """The reachable tile nearest to `near` whose 3x3 neighbourhood holds nothing else."""
    t = taken(m, skip_pack) | set(avoid)
    ok = lambda p: all((p[0] + dx, p[1] + dy) not in t for dx in (-1, 0, 1) for dy in (-1, 0, 1))
    candidates = [(abs(x - near[0]) + abs(y - near[1]), x, y) for x, y in reachable(m) if ok((x, y))]
    if not candidates:
        raise ValueError(f'no free reachable tile on {m}')
    _, x, y = min(candidates)
    return x, y


# ---------- the pack ----------

# Stats for a plain townsperson. Override any of them with keywords, e.g. p.actor(..., Awareness=8).
NPC_STATS = {'Age': 35, 'BodyType': 0, 'KillType': 0, 'DamageType': 0, 'TeamNum': 1, 'HitPoints': 28,
             'ActionPoints': 6, 'ArmorClass': 5, 'MeleeDamage': 2, 'CarryWeight': 100, 'Sequence': 8, 'HealingRate': 1,
             'CriticalChance': 2, 'Awareness': 4, 'Luck': 5, 'DREMP': 500}


def _fields(d):
    return ''.join(f'{k}: {v}\n' for k, v in d.items())


class Pack:
    def __init__(self, name, title):
        if not re.fullmatch(r'[a-z0-9]+', name):
            raise ValueError('pack names are lowercase letters and digits')
        self.name, self.title = name, title
        self.files = {}  # path relative to data_atom -> content
        self.placements = collections.defaultdict(lambda: {'actors': [], 'objects': []})
        self.used = collections.defaultdict(set)
        self.journal, self.items, self.spawnables, self.teams = [], [], [], []
        self.extensions = []  # (dialogue, node, option, nodes)

    def _place(self, m, near=None, at=None):
        if at:
            pos = at
        else:
            pos = free_spot(m, near or (40, 12), self.used[m], skip_pack=self.name)
        self.used[m].add(pos)
        return pos

    # placements
    def actor(self, m, name, desc, long, icon, near=None, at=None, equipment=(), dialogue=True, hp=None, faction=None, **extra):
        """An NPC standing on map m. dialogue=True expects dialogues/<name>.rec. Extra keywords become record fields."""
        x, y = self._place(m, near, at)
        rec = {'Name': name, 'Description': desc, 'LongDescription': long}
        if dialogue:
            rec['Dialogue'] = name if dialogue is True else dialogue
        rec.update({'faction': faction or name, 'Icon': icon, 'Foreground': 'light_gray_2'})
        rec.update(NPC_STATS)
        if hp:
            rec['HitPoints'] = hp
        rec.update(extra)
        body = _fields(rec) + ''.join(f'equipment: {e}\n' for e in equipment)
        self.placements[m]['actors'].append(body + f'Position: ({x},{y})\n')
        return x, y

    def corpse(self, m, name, desc, near=None, at=None, equipment=()):
        x, y = self._place(m, near, at)
        body = f'Name: {name}\nIcon: M\nForeground: light_gray_2\nFlags: Spawn_Dead\nDescription: {desc}\n' + ''.join(f'equipment: {e}\n' for e in equipment)
        self.placements[m]['actors'].append(body + f'HitPoints: 12\nPosition: ({x},{y})\n')
        return x, y

    def object(self, m, body, near=None, at=None):
        """Any object record (container, readable, state changer...); Position is added."""
        x, y = self._place(m, near, at)
        self.placements[m]['objects'].append(body.strip() + f'\nPosition: ({x},{y})\n')
        return x, y

    def container(self, m, name, desc, items, near=None, at=None, lock=None, lockflag=None):
        body = f'Category: UnknownContainer\nName: {name}\nDescription: {desc}\n'
        body += f'lockdifficulty: {lock}\n' if lock else ''
        body += f'lockflag: {lockflag}\n' if lockflag else ''
        body += ''.join(f'item: {i}\n' for i in items)
        return self.object(m, body, near, at)

    def named_location(self, m, ident, near=None, at=None):
        return self.object(m, f'Category: NamedLocation\nIdentifier: {ident}', near, at)

    # definitions
    def item(self, name, desc, long, cost=50, pickup_flag=None):
        self.items.append(f'Name: {name}\nDescription: {desc}\nLong_Description: {long}\nCategory: Other\nWeight: 1\nCost: {cost}\n' + (f'PickupFlag: {pickup_flag}\n' if pickup_flag else ''))

    def spawnable(self, name, desc, long, icon, equipment=(), hp=30, **extra):
        """An actor template for SpawnActor / teams. Extra keywords become fields (faction=..., aggressive='true')."""
        self.spawnables.append(_fields({'Name': name, 'Description': desc, 'Long_Description': long, 'Icon': icon, 'Foreground': 'light_gray_2', **NPC_STATS, 'HitPoints': hp, **extra}) + ''.join(f'equipment: {e}\n' for e in equipment))

    def team(self, name, leader, members):
        self.teams.append(f'name: {name}\nleader: {leader}\n' + ''.join(f'member: {mbr}\n' for mbr in members))

    def journal_entry(self, text):
        self.journal.append(text.strip() + '\n')

    # whole files
    def dialogue(self, name, text):
        self.files[f'dialogues/{name}.rec'] = text.strip() + '\n'

    def script(self, name, frames, comment=''):
        """A script. Pass just the frames ('if: ...\\ndo: ...' blocks) or a whole script file."""
        if '%rec: frames' in frames:
            self.files[f'scripts/{name}.rec'] = frames.strip() + '\n'
            return
        head = (f'# {comment}\n' if comment else '') + '%rec: definitions\n\n%rec: outcomes\n\n%rec: cancel\n\n%rec: frames\n\n'
        self.files[f'scripts/{name}.rec'] = head + frames.strip() + '\n'

    def extend_dialogue(self, dialogue, node, option, nodes=''):
        """Add an option to an existing NPC's node (inserted after the npc text) plus extra nodes at the end."""
        self.extensions.append((dialogue, node, option.strip(), nodes.strip()))

    def terminal_lead(self, job, hide_when, menu, lines, lead):
        """A vague job offer on the apartment terminal that names who to ask. Hidden once hide_when is true."""
        self.extend_dialogue('home_terminal', 'DefaultIntroduction',
                             f"o_text: (Job offer) {menu}\no_id: lead_{job}\no_cond: !({hide_when})\no_goto: lead_{job}",
                             f"name: lead_{job}\nnpc: ** Contractor Client Interface **\n" + ''.join(f'+ > {x}\n' for x in lines) +
                             f"+ > {lead}\neffect: SetFlag('JobRead({job})')\n#\no_text: Go back.\no_goto: DefaultIntroduction\n#\no_text: Leave.\no_id: lead_leave\no_goto: End")

    def playtest(self, name, comment, outcome, trip):
        if not name.startswith(self.name + '_'):
            raise ValueError(f'playtest names start with "{self.name}_" so content_check.sh finds them')
        body = '\n\n'.join('if: true\n' + '\n'.join('do: ' + s for s in f) for f in trip.frames if f)
        self.files[f'playtests/{name}.rec'] = f'# {comment}\n\n%rec: outcomes\n\nif: {outcome}\n\n%rec: frames\n\n{body}\n'

    # output
    def write(self):
        n, files = self.name, dict(self.files)
        header = f'# Generated by content/{n}.py ({self.title}). Edit that file and rerun it, not this one.\n'
        for m, kinds in self.placements.items():
            for kind, recs in kinds.items():
                if recs:
                    files[f'maps/{m}/{kind}_{n}.rec'] = header + '%rec: default\n\n' + '\n'.join(recs)
        if self.journal:
            files[f'definitions/journal_{n}.rec'] = header + '%rec: journalEntry\n%key: id\n\n' + '\n'.join(self.journal)
        if self.items:
            files[f'definitions/items_{n}.rec'] = header + '%rec: Misc\n\n' + '\n'.join(self.items)
        if self.spawnables:
            files[f'definitions/actors_{n}.rec'] = header + '%rec: Actor\n\n' + '\n'.join(self.spawnables)
        if self.teams:
            files[f'definitions/spawned_teams_{n}.rec'] = header + '%rec: team\n\n' + '\n'.join(self.teams)

        manifest = os.path.join(DATA, 'packs', f'{n}.manifest')
        old = open(manifest).read().split() if os.path.exists(manifest) else []
        for rel in old:
            if rel not in files and os.path.exists(os.path.join(DATA, rel)):
                os.remove(os.path.join(DATA, rel))

        begin, end = f'# >>> pack {n}', f'# <<< pack {n}'
        strip = re.compile(rf'(?ms)(?:(?<=\n)\n)?^{re.escape(begin)}\n.*?^{re.escape(end)}\n')  # plus the blank line before an appended block
        touched = set()
        for dlg, node, option, nodes in self.extensions:
            touched.add(dlg)
        ddir = os.path.join(DATA, 'dialogues')
        for fname in os.listdir(ddir):  # remove last run's blocks everywhere, including dialogues no longer extended
            p = os.path.join(ddir, fname)
            s = open(p).read()
            if begin in s:
                open(p, 'w').write(strip.sub('', s))
        for dlg, node, option, nodes in self.extensions:
            p = os.path.join(ddir, dlg + '.rec')
            s = files.get(f'dialogues/{dlg}.rec') or open(p).read()
            mt = re.search(rf'(?m)^name: {re.escape(node)}\n(?:(?!^name:).*\n)*?npc:.*\n(?:\+.*\n)*', s)
            if not mt:
                raise ValueError(f'node {node} not found in dialogues/{dlg}.rec')
            s = s[:mt.end()] + f'{begin}\n#\n{option}\n{end}\n' + s[mt.end():]
            if nodes:
                s = s.rstrip('\n') + f'\n\n{begin}\n{nodes}\n{end}\n'
            if f'dialogues/{dlg}.rec' in files:
                files[f'dialogues/{dlg}.rec'] = s
            else:
                open(p, 'w').write(s)

        for rel, content in files.items():
            p = os.path.join(DATA, rel)
            os.makedirs(os.path.dirname(p), exist_ok=True)
            open(p, 'w').write(content)
        os.makedirs(os.path.dirname(manifest), exist_ok=True)
        open(manifest, 'w').write('\n'.join(sorted(files)) + '\n')
        print(f'pack {n}: wrote {len(files)} files, extended {len(touched)} dialogues')



# ---------- lint: cross-file references ----------

def _read_all(pattern_dir, suffix='.rec'):
    for base, _, names in os.walk(pattern_dir):
        for name in names:
            if name.endswith(suffix):
                path = os.path.join(base, name)
                yield os.path.relpath(path, DATA), open(path).read()


def lint(only=None):
    """Check names used in data_atom against what is defined. only: set of data_atom-relative paths to report on."""
    files = {rel: txt for rel, txt in _read_all(DATA) if rel != 'definitions/objectTemplates.rec'}  # editor templates, not content
    go = ''.join(open(os.path.join(ROOT, 'game', f)).read() for f in os.listdir(os.path.join(ROOT, 'game')) if f.endswith('.go') and not f.endswith('_test.go'))
    q = r"""['"]([^'"]+)['"]"""
    every = lambda rx, where=files: {(rel, m) for rel, txt in where.items() for m in re.findall(rx, txt)}
    names = lambda rx, where=files: {m for _, m in every(rx, where)}
    defs = {rel: txt for rel, txt in files.items() if rel.startswith('definitions/')}
    maps = {rel: txt for rel, txt in files.items() if rel.startswith('maps/')}
    tests = {rel: txt for rel, txt in files.items() if rel.startswith('playtests/')}
    content = {rel: txt for rel, txt in files.items() if not rel.startswith('playtests/')}

    items = names(r'(?m)^Name:\s*(\S+)', defs) | {'gold'} | set(re.findall(r'InternalName:\s*"(\w+)"', go)) | names(r'key\(' + q)  # a key's item name is its lock flag
    actors = names(r'(?m)^Name:\s*(\S+)', maps) | names(r'(?m)^Name:\s*(\S+)', defs)
    spots = names(r'(?m)^(?:Identifier|Location):\s*(\S+)', maps)
    teams = names(r'(?m)^name:\s*(\S+)', {r: t for r, t in defs.items() if 'spawned_teams' in r})
    scripts = {os.path.basename(r)[:-4] for r in files if r.startswith('scripts/')}
    dialogues = {os.path.basename(r)[:-4] for r in files if r.startswith('dialogues/')}
    option_ids = names(r'(?m)^o_id:\s*(\S+)')
    option_texts = names(r'(?m)^o_text:\s*(.+?)\s*$')

    families = set(re.findall(r'Sprintf\("(\w+)\(%', go)) | {'JobAccepted', 'JobDeclined', 'JobRead'}
    set_flags = names(r'(?:SetFlag|IncrementFlag|SetFlagTo)\(' + q) | names(r'(?m)^(?:PickupFlag|DropFlag|lockflag|LockFlag):\s*(\S+)')
    set_flags |= names(r"key\(" + q) | set(re.findall(r'gameFlags\.(?:Set|SetFlag|Increment)\("(\w+)"', go))
    set_flags |= {line.split(':')[0] for r, t in files.items() if r.endswith('initFlags.rec') for line in t.splitlines() if ':' in line and not line.startswith(('%', '#'))}
    timers = names(r'SaveTimeNow\(' + q)

    errors = []
    def err(rel, msg):
        if only is None or rel in only:
            errors.append(f'{rel}: {msg}')
    base = lambda n: re.sub(r'\(.*\)$', '', n.strip())

    for rel, n in every(r'(?:HasItem|PlayerAddItem|RemoveItem|GiveItem)\(' + q) | every(r'StackTransfer(?:To|From)\(\w+,\s*' + q):
        if base(n) not in items:
            err(rel, f'unknown item {n!r}')
    for rel, n in every(r'(?m)^(?:equipment|item):\s*(.+?)\s*$', maps) | every(r'(?m)^equipment:\s*(.+?)\s*$', defs):
        if not n.startswith('key(') and base(n) not in items:
            err(rel, f'unknown item {n!r}')
    # every key opens a lock, and every lock has a way to open it: key, script flag, flag control or number code (lockpicking doesn't count)
    key_flags = names(r'key\(' + q) | names(r'(?mi)^lockflag:\s*(\S+)', defs)  # key() items and key item definitions
    script_opened = names(r'(?:SetFlag|ClearFlag)\(' + q) | set(re.findall(r'gameFlags\.(?:Set|SetFlag|Unset|Clear)\w*\("(\w+)"', go))
    lock_flags = set()
    for rel, txt in maps.items():
        for rec in txt.split('\n\n'):
            m = re.search(r'(?mi)^lockflag:\s*(\S+)', rec)
            if not m:
                continue
            lock_flags.add(m.group(1))
            if not re.search(r'(?mi)^(?:IsFlagControlled:\s*true|NumberLock:)', rec) and m.group(1) not in key_flags | script_opened:
                err(rel, f'lock {m.group(1)!r} cannot be opened (no key, script or number code)')
    for rel, n in every(r'key\(' + q) | every(r'(?mi)^lockflag:\s*(\S+)', defs):
        if n not in lock_flags and n != 'CHANGEME':  # CHANGEME: the bare 'key' template, never placed (checked below)
            err(rel, f'key {n!r} opens no lock')
    for rel, _ in every(r"(?m)^(?:item|equipment):\s*key\s*$") | every(r"(?:PlayerAddItem|StackTransferFrom|GiveItem)\([^)]*['\"]key['\"]"):
        err(rel, "bare 'key' template opens no lock: use key('flag', 'name')")
    for rel, n in every(r'RunScript\(' + q):
        if n not in scripts:
            err(rel, f'no script {n!r} (scripts/{n}.rec)')
    for rel, n in every(r'(?m)^Dialogue:\s*(\S+)', maps) | every(r'(?m)^Dialogue:\s*(\S+)', defs):
        if n not in dialogues:
            err(rel, f'no dialogue {n!r} (dialogues/{n}.rec)')
    for rel, txt in files.items():
        for d, _, loc in re.findall(r'SpawnActor\(' + q + r',\s*' + q + r',\s*' + q, txt):
            if d not in actors:
                err(rel, f'SpawnActor of unknown actor {d!r}')
            if loc not in spots:
                err(rel, f'SpawnActor at unknown spot {loc!r}')
        for team, _, loc in re.findall(r'SpawnTeam\w*\(' + q + r',\s*' + q + r',\s*' + q, txt):
            if team not in teams:
                err(rel, f'unknown team {team!r}')
            if loc not in spots:
                err(rel, f'team spawns at unknown spot {loc!r}')
        for t in re.findall(r'Is(?:Turns|Minutes|Hours|Days)After\(' + q, txt):
            if t not in timers:
                err(rel, f'timer {t!r} is never started with SaveTimeNow')
    for rel, f in every(r'(?:HasFlag|GetFlag)\(' + q, content):
        if f not in set_flags and f.split('(')[0] not in families:
            err(rel, f'flag {f!r} is queried but never set')
    for rel, n in every(r'(?:Talk|Kill)\(' + q, tests):
        if n not in actors:
            err(rel, f'unknown actor {n!r}')
    for rel, n in every(r'Choose\(' + q, tests):
        if n not in option_ids and n not in option_texts:
            err(rel, f'no dialogue option with o_id {n!r}')
    for rel, n in every(r'GoTo\(' + q, tests):
        if n not in spots and n not in actors:
            err(rel, f'GoTo unknown spot {n!r}')
    return sorted(set(errors))

# ---------- CLI ----------

def _ref():
    src = open(os.path.join(ROOT, 'game', 'script_funcs.go')).read()
    print('# Script functions (dialogue conditions/effects, script if:/do:, journal conditions)')
    sigs = {}
    parts = re.split(r'^\s*"(\w+)":\s*func', src, flags=re.M)
    for name, body in zip(parts[1::2], parts[2::2]):
        args = dict(re.findall(r'args\[(\d+)\]\.\((\w+)\)', body))
        kind = {'string': "'text'", 'float64': 'n', 'bool': 'true'}
        sigs[name] = ', '.join(kind.get(args.get(str(i)), '?') for i in range(max(map(int, args), default=-1) + 1))
    for name in sorted(sigs):
        print(f'  {name}({sigs[name]})')
    auto = open(os.path.join(ROOT, 'game', 'autoplay.go')).read()
    print('\n# Playtest verbs (do: lines)')
    for name in sorted(set(re.findall(r'^\s*"([A-Z]\w+)":\s*func', auto, re.M))):
        print(' ', name)
    print('\n# Maps: actors, named spots and transitions')
    for m in sorted(os.listdir(MAPS)):
        if not os.path.isdir(os.path.join(MAPS, m)):
            continue
        actors = [re.search(r'Name: (\S+)', b).group(1) for f in _map_files(m, ('actors',)) for b in open(f).read().split('\n\n') if 'Name:' in b and 'Spawn_Dead' not in b]
        objs = ''.join(open(f).read() for f in _map_files(m, ('objects',)))
        spots = re.findall(r'(?:Identifier|Location): (\S+)', objs)
        print(f'  {m}{"  (routed)" if m in ROUTE else ""}')
        print(f'    actors: {" ".join(sorted(set(actors)))}')
        print(f'    spots:  {" ".join(sorted(set(spots)))}')


if __name__ == '__main__':
    if sys.argv[1:2] == ['ref']:
        _ref()
    elif sys.argv[1:2] == ['remove'] and len(sys.argv) == 3:  # delete everything a pack generated
        Pack(sys.argv[2], 'removed').write()
        os.remove(os.path.join(DATA, 'packs', sys.argv[2] + '.manifest'))
    elif sys.argv[1:2] == ['lint']:
        only = None
        if len(sys.argv) > 2:  # just the files a pack owns
            only = set(open(os.path.join(DATA, 'packs', sys.argv[2] + '.manifest')).read().split())
        problems = lint(only)
        print('\n'.join(problems) or 'lint: ok')
        sys.exit(1 if problems else 0)
    elif sys.argv[1:2] == ['spot'] and len(sys.argv) == 4:
        x, y = map(int, sys.argv[3].split(','))
        print(free_spot(sys.argv[2], (x, y)))
    else:
        print(__doc__)
