// mapgen writes tiles.bin and zones.bin for a map from its layout.txt.
//
//	go run ./cmd/mapgen data_atom/maps/<name>
//
// layout.txt: the first h lines are the map (w columns), then a legend of lines
// "<char> <tile name>" naming tiles in that map's tileSet.rec (first match by name),
// then optional "zone <name>" blocks of h lines where every non-space char belongs to the zone.
// ponytail: dimensions are those of the first line and the number of lines before the first blank one.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/textiles"
)

func main() {
	dir := os.Args[1]
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(filepath.Join(dir, "..", "..", "definitions", "palette.rec")))
	tileSet := textiles.ReadTilesFileAndClose(fxtools.MustOpen(filepath.Join(dir, "tileSet.rec")), palette)

	var lines []string
	scanner := bufio.NewScanner(fxtools.MustOpen(filepath.Join(dir, "layout.txt")))
	for scanner.Scan() {
		lines = append(lines, strings.TrimRight(scanner.Text(), "\r"))
	}
	var rows []string
	for _, line := range lines {
		if line == "" {
			break
		}
		rows = append(rows, line)
	}
	w, h := len([]rune(rows[0])), len(rows)

	legend := map[rune]int16{}
	zones := map[string]map[geometry.Point]bool{}
	for i := h; i < len(lines); i++ {
		line := lines[i]
		switch {
		case strings.HasPrefix(line, "zone "):
			name := strings.TrimSpace(line[5:])
			zones[name] = map[geometry.Point]bool{}
			for y := 0; y < h; y++ {
				for x, c := range []rune(lines[i+1+y]) {
					if c != ' ' {
						zones[name][geometry.Point{X: x, Y: y}] = true
					}
				}
			}
			i += h
		case len([]rune(line)) > 2:
			r := []rune(line)
			tileName := strings.TrimSpace(string(r[1:]))
			idx := -1
			for j, tile := range tileSet {
				if tile.Name == tileName {
					idx = j
					break
				}
			}
			if idx < 0 {
				fail("legend: no tile named %q in tileSet.rec", tileName)
			}
			legend[r[0]] = int16(idx)
		}
	}

	tiles := make([]int16, w*h)
	for y, row := range rows {
		r := []rune(row)
		if len(r) != w {
			fail("row %d has %d columns, want %d", y, len(r), w)
		}
		for x, c := range r {
			idx, ok := legend[c]
			if !ok {
				fail("row %d col %d: char %q not in legend", y, x, c)
			}
			tiles[y*w+x] = idx
		}
	}
	if err := textiles.SaveTileMap16(tiles, geometry.Point{X: w, Y: h}, filepath.Join(dir, "tiles.bin")); err != nil {
		fail("%v", err)
	}
	if err := textiles.SaveZones(filepath.Join(dir, "zones.bin"), zones); err != nil {
		fail("%v", err)
	}
	fmt.Printf("%s: %dx%d, %d zones\n", dir, w, h, len(zones))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
