package ui_console

import (
	"contractor/foundation"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/memmaker/go/cview"
)

// Ammo is a second inventory: lists show either items or ammo, never both, so each page
// needs at most 26 letters. Tab (gamepad RB/LB) flips the page; UI.ammoPage remembers it
// while a screen reopens after a transfer, closing the screen resets it.

func onPage(items []foundation.Item, ammo bool) []foundation.Item {
	var page []foundation.Item
	for _, item := range items {
		if item.IsAmmo() == ammo {
			page = append(page, item)
		}
	}
	return page
}

// withPageTabs draws an Items/Ammo tab strip on the row above the list's border: the active tab inverted, the other dimmed.
func withPageTabs(list *cview.List, ammo bool) {
	on, off := "[::r] %s [::-]", "[::d] %s [::-]"
	items, ammoTab := on, off
	if ammo {
		items, ammoTab = off, on
	}
	tabs := []byte(fmt.Sprintf(items, "Items") + fmt.Sprintf(ammoTab, "Ammo"))
	list.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if y > 0 {
			cview.Print(screen, tabs, x+1, y-1, width-2, cview.AlignLeft, tcell.ColorDefault)
		}
		return x + 1, y + 1, width - 2, height - 2 // the list's border, no padding
	})
}

func isPageToggle(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab
}
