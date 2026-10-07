// mapdump prints a map's tiles as text: '#' blocks movement, '.' is walkable. Rows and columns are numbered.
//
//	go run ./cmd/mapdump data_atom/maps/<name>
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/textiles"
)

func main() {
	dir := os.Args[1]
	palette := textiles.ReadPaletteFileOrDefault(fxtools.MustOpen(filepath.Join(dir, "..", "..", "definitions", "palette.rec")))
	tileSet := textiles.ReadTilesFileAndClose(fxtools.MustOpen(filepath.Join(dir, "tileSet.rec")), palette)
	size, tiles := textiles.ReadTileMap16(filepath.Join(dir, "tiles.bin"))
	fmt.Print("   ")
	for x := 0; x < size.X; x++ {
		fmt.Print(x % 10)
	}
	fmt.Println()
	for y := 0; y < size.Y; y++ {
		fmt.Printf("%2d ", y)
		for x := 0; x < size.X; x++ {
			if t := tiles[y*size.X+x]; int(t) < len(tileSet) && tileSet[t].IsWalkable {
				fmt.Print(".")
			} else {
				fmt.Print("#")
			}
		}
		fmt.Println()
	}
}
