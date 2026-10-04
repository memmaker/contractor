// mapoverview renders every map reachable from the player start into one HTML page,
// with every transition drawn as a numbered, colored connector to its target tile.
//
//	go run ./cmd/mapoverview > overview.html
package main

import (
	"fmt"
	"html"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
)

type edge struct {
	From, To string
	FromPos  geometry.Point
	ToLoc    string
	Kind     string // street, taxi
}

type mapInfo struct {
	Name, Display string
	Size          geometry.Point
	Tiles         []textiles.TextIcon
	Locations     map[string]geometry.Point
	Edges         []edge
}

var taxiRe = regexp.MustCompile(`Transition(?:WithDriver)?\((?:NPC, )?'([^']+)', '([^']+)'\)`)

func main() {
	root := "data_atom"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(filepath.Join(root, "definitions", "palette.rec")))
	startRecs, _ := recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(root, "definitions", "player_start.rec")))
	start := startRecs[0].FindValueForKeyIgnoreCase("mapName")

	// taxi destinations come from the dialogue, not the map
	var taxiDests [][2]string
	for _, f := range []string{"taxi_driver.rec", "taxi.rec"} {
		data, _ := os.ReadFile(filepath.Join(root, "dialogues", f))
		for _, m := range taxiRe.FindAllStringSubmatch(string(data), -1) {
			taxiDests = append(taxiDests, [2]string{m[1], m[2]})
		}
		if len(taxiDests) > 0 {
			break
		}
	}

	maps := map[string]*mapInfo{}
	queue := []string{start}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, done := maps[name]; done {
			continue
		}
		m := loadMap(filepath.Join(root, "maps", name), name, palette, taxiDests)
		if m == nil {
			fmt.Fprintln(os.Stderr, "missing map:", name)
			continue
		}
		maps[name] = m
		for _, e := range m.Edges {
			queue = append(queue, e.To)
		}
	}

	names := make([]string, 0, len(maps))
	for n := range maps {
		names = append(names, n)
	}
	sort.Strings(names)
	// start map first
	sort.SliceStable(names, func(i, j int) bool { return names[i] == start && names[j] != start })
	writeHTML(os.Stdout, names, maps, start)
}

func loadMap(dir, name string, palette textiles.ColorPalette, taxiDests [][2]string) *mapInfo {
	if !fxtools.DirExists(dir) {
		return nil
	}
	tileSet := textiles.ReadTilesFileAndClose(fxtools.MustOpen(filepath.Join(dir, "tileSet.rec")), palette)
	size, tileMap := textiles.ReadTileMap16(filepath.Join(dir, "tiles.bin"))
	m := &mapInfo{Name: name, Display: name, Size: size, Locations: map[string]geometry.Point{}}
	m.Tiles = make([]textiles.TextIcon, len(tileMap))
	for i, t := range tileMap {
		if int(t) < len(tileSet) {
			m.Tiles[i] = tileSet[t].Icon
		}
	}
	if fxtools.FileExists(filepath.Join(dir, "meta.rec")) {
		recs, _ := recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(dir, "meta.rec")))
		if n := recs[0].FindValueForKeyIgnoreCase("Name"); n != "" {
			m.Display = n
		}
	}
	icons := map[string]textiles.TextIcon{}
	if fxtools.FileExists(filepath.Join(dir, "iconsForObjects.rec")) {
		for n, r := range textiles.ReadIconRecordsIntoMap(fxtools.MustOpen(filepath.Join(dir, "iconsForObjects.rec"))) {
			icons[strings.ToLower(n)] = textiles.NewTextIconFromNamedColorChar(r.Icon, palette)
		}
	}
	if !fxtools.FileExists(filepath.Join(dir, "objects.rec")) {
		return m
	}
	objs, _ := recfile.ReadAndClose(fxtools.MustOpen(filepath.Join(dir, "objects.rec")))
	for _, rec := range objs {
		cat := strings.ToLower(rec.FindValueForKeyIgnoreCase("Category"))
		pos, _ := geometry.NewPointFromEncodedString(rec.FindValueForKeyIgnoreCase("Position"))
		idx := pos.Y*size.X + pos.X
		switch cat {
		case "transition":
			m.Locations[rec.FindValueForKeyIgnoreCase("Location")] = pos
			m.Edges = append(m.Edges, edge{From: name, To: rec.FindValueForKeyIgnoreCase("TargetMap"),
				FromPos: pos, ToLoc: rec.FindValueForKeyIgnoreCase("TargetLocation"), Kind: "street"})
		case "namedlocation":
			m.Locations[rec.FindValueForKeyIgnoreCase("identifier")] = pos
		case "bakedlight":
		default:
			iconName := cat
			if o := rec.FindValueForKeyIgnoreCase("IconOverride"); o != "" {
				iconName = strings.ToLower(o)
			}
			if ic, ok := icons[iconName]; ok && idx >= 0 && idx < len(m.Tiles) {
				if ic.Bg.A == 0 {
					ic.Bg = m.Tiles[idx].Bg
				}
				m.Tiles[idx] = ic
			}
			if cat == "terminal" && strings.EqualFold(rec.FindValueForKeyIgnoreCase("Dialogue"), "taxi") {
				for _, d := range taxiDests {
					if d[0] != name {
						m.Edges = append(m.Edges, edge{From: name, To: d[0], FromPos: pos, ToLoc: d[1], Kind: "taxi"})
					}
				}
			}
		}
	}
	return m
}

func css(c color.RGBA) string {
	if c.A == 0 {
		return "transparent"
	}
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func writeHTML(w *os.File, names []string, maps map[string]*mapInfo, start string) {
	fmt.Fprint(w, `<!doctype html><meta charset="utf-8"><title>Map overview</title>
<style>
body{background:#111;color:#ddd;font-family:system-ui;margin:0;padding:16px}
#maps{display:flex;flex-wrap:wrap;gap:48px;align-items:flex-start;position:relative}
.map{position:relative}
.map h2{font-size:14px;margin:0 0 4px;font-weight:600}
.map h2 small{color:#888;font-weight:400;margin-left:6px}
pre{margin:0;font:12px/14px "Menlo","DejaVu Sans Mono",monospace;background:#000;padding:4px;border:1px solid #333}
pre span{display:inline-block;width:8px;height:14px;text-align:center;overflow:hidden;vertical-align:top}
.T,.L{position:relative;outline:2px solid var(--c);outline-offset:-1px;cursor:pointer;z-index:2}
.T::after,.L::after{content:attr(data-n);position:absolute;left:8px;top:-9px;background:var(--c);color:#000;font:bold 10px/12px system-ui;padding:0 3px;border-radius:3px;white-space:nowrap;z-index:3}
.L::after{content:"→" attr(data-n)}
svg{position:absolute;inset:0;width:100%;height:100%;pointer-events:none;overflow:visible;z-index:1}
path{fill:none;stroke-width:2;opacity:.55}
path.taxi{stroke-dasharray:6 5}
path.hot{opacity:1;stroke-width:4;filter:drop-shadow(0 0 4px var(--c))}
#legend{font-size:12px;color:#aaa;margin-bottom:12px}
#legend b{color:#ddd}
.badge{display:inline-block;padding:0 4px;border-radius:3px;background:var(--c);color:#000;font-weight:bold}
</style>
<div id=legend><b>Transitions</b>: solid line = street/door transition, dashed = taxi. Numbered badge on the source tile, same number with → on the destination tile. Hover a badge or line to highlight it; click a badge to jump to the other end. Start map first.</div>
<div id=maps><svg id=lines></svg>
`)
	n := 0
	ids := map[string]int{} // edge id -> number
	for _, name := range names {
		m := maps[name]
		fmt.Fprintf(w, `<div class=map id="m-%s"><h2>%s<small>%s %dx%d</small></h2><pre data-w=%d>`, name, html.EscapeString(m.Display), name, m.Size.X, m.Size.Y, m.Size.X)
		// mark source tiles and target tiles
		marks := map[geometry.Point][]string{}
		for _, e := range m.Edges {
			n++
			id := fmt.Sprintf("%s|%d,%d|%s|%s", e.From, e.FromPos.X, e.FromPos.Y, e.To, e.ToLoc)
			ids[id] = n
			marks[e.FromPos] = append(marks[e.FromPos], fmt.Sprintf("T %d %s %s", n, e.To, e.ToLoc))
		}
		// incoming
		for _, other := range names {
			for _, e := range maps[other].Edges {
				if e.To != name {
					continue
				}
				pos, ok := m.Locations[e.ToLoc]
				if !ok {
					fmt.Fprintf(os.Stderr, "dangling: %s -> %s:%s\n", e.From, e.To, e.ToLoc)
					continue
				}
				marks[pos] = append(marks[pos], fmt.Sprintf("L %s|%d,%d|%s|%s", e.From, e.FromPos.X, e.FromPos.Y, e.To, e.ToLoc))
			}
		}
		for y := 0; y < m.Size.Y; y++ {
			for x := 0; x < m.Size.X; x++ {
				ic := m.Tiles[y*m.Size.X+x]
				ch := ic.Char
				if ch == 0 {
					ch = ' '
				}
				p := geometry.Point{X: x, Y: y}
				attrs := ""
				if ms, ok := marks[p]; ok {
					var cls, nums []string
					var eids []string
					for _, mk := range ms {
						f := strings.SplitN(mk, " ", 2)
						if f[0] == "T" {
							parts := strings.SplitN(f[1], " ", 3)
							cls = append(cls, "T")
							nums = append(nums, parts[0])
							eids = append(eids, fmt.Sprintf("%s|%d,%d|%s|%s", name, x, y, parts[1], parts[2]))
						} else {
							cls = append(cls, "L")
							nums = append(nums, fmt.Sprint(ids[f[1]]))
							eids = append(eids, f[1])
						}
					}
					attrs = fmt.Sprintf(` class="%s" data-n="%s" data-e="%s" title="%s"`, cls[0], strings.Join(nums, ","), html.EscapeString(strings.Join(eids, ";")), html.EscapeString(strings.Join(ms, "\n")))
				}
				fmt.Fprintf(w, `<span style="color:%s;background:%s"%s>%s</span>`, css(ic.Fg), css(ic.Bg), attrs, html.EscapeString(string(ch)))
			}
			fmt.Fprint(w, "\n")
		}
		fmt.Fprint(w, "</pre></div>\n")
	}
	// edge list for JS
	fmt.Fprint(w, "<script>const EDGES=[")
	for _, name := range names {
		for _, e := range maps[name].Edges {
			id := fmt.Sprintf("%s|%d,%d|%s|%s", e.From, e.FromPos.X, e.FromPos.Y, e.To, e.ToLoc)
			tgt, ok := maps[e.To]
			if !ok {
				continue
			}
			pos, ok := tgt.Locations[e.ToLoc]
			if !ok {
				continue
			}
			fmt.Fprintf(w, "[%q,%d,%q,%d,%d,%q,%d,%d,%q],", id, ids[id], e.From, e.FromPos.X, e.FromPos.Y, e.To, pos.X, pos.Y, e.Kind)
		}
	}
	fmt.Fprint(w, `];
const svg=document.getElementById('lines'),root=document.getElementById('maps');
const hue=n=>'hsl('+(n*47%360)+' 90% 60%)';
function cell(map,x,y){const pre=document.querySelector('#m-'+CSS.escape(map)+' pre');return pre&&pre.children[y*pre.dataset.w+x]}
function center(el){const r=el.getBoundingClientRect(),b=root.getBoundingClientRect();return [r.left-b.left+r.width/2,r.top-b.top+r.height/2]}
function draw(){svg.innerHTML='';for(const [id,n,fm,fx,fy,tm,tx,ty,kind] of EDGES){const a=cell(fm,fx,fy),b=cell(tm,tx,ty);if(!a||!b)continue;const [x1,y1]=center(a),[x2,y2]=center(b);const dx=(x2-x1)/2;const p=document.createElementNS('http://www.w3.org/2000/svg','path');p.setAttribute('d','M'+x1+' '+y1+' C'+(x1+dx)+' '+y1+' '+(x2-dx)+' '+y2+' '+x2+' '+y2);p.setAttribute('class',kind);p.style.stroke=hue(n);p.style.setProperty('--c',hue(n));p.dataset.e=id;svg.appendChild(p)}}
for(const el of document.querySelectorAll('.T,.L')){const ids=el.dataset.e.split(';');const n=parseInt(el.dataset.n);el.style.setProperty('--c',hue(n));
el.onmouseenter=()=>ids.forEach(i=>document.querySelectorAll('path').forEach(p=>{if(p.dataset.e===i)p.classList.add('hot')}));
el.onmouseleave=()=>document.querySelectorAll('path.hot').forEach(p=>p.classList.remove('hot'));
el.onclick=()=>{const id=ids[0];const other=[...document.querySelectorAll('.T,.L')].find(o=>o!==el&&o.dataset.e.split(';').includes(id));if(other){other.scrollIntoView({block:'center',inline:'center',behavior:'smooth'});other.animate([{outlineWidth:'6px'},{outlineWidth:'2px'}],{duration:900})}}}
draw();addEventListener('resize',draw);document.fonts&&document.fonts.ready.then(draw);
</script>`)
}
