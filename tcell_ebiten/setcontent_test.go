package tcell_ebiten

import (
    "image"
    "testing"

    "github.com/gdamore/tcell/v2"
)

// Unchanged writes must keep a cell synced, or Show redraws the whole grid every frame.
func TestSetContentKeepsUnchangedCellsSynced(t *testing.T) {
    et := &etcell{grid: make([]cell, 1), grid_size: image.Pt(1, 1)}
    st := tcell.StyleDefault.Foreground(tcell.ColorRed)
    et.SetContent(0, 0, 'a', nil, st)
    et.grid[0].synced = true

    et.SetContent(0, 0, 'a', nil, st)
    if !et.grid[0].synced {
        t.Fatal("identical write marked cell dirty")
    }
    et.SetContent(0, 0, 'b', nil, st)
    if et.grid[0].synced {
        t.Fatal("changed write left cell synced")
    }
}
