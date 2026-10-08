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

// pageTitle draws a tab strip into the list's top border: the active tab inverted, the other dimmed.
func pageTitle(title string, ammo bool) string {
	on, off := "[::r] %s [::-]", "[::d] %s [::-]"
	items, ammoTab := on, off
	if ammo {
		items, ammoTab = off, on
	}
	return cview.Escape(title) + " " + fmt.Sprintf(items, "Items") + fmt.Sprintf(ammoTab, "Ammo") + "[::d] Tab[::-]"
}

func isPageToggle(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab
}
