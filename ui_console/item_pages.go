package ui_console

import (
	"contractor/foundation"
	"github.com/gdamore/tcell/v2"
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

func pageTitle(title string, ammo bool) string {
	if ammo {
		return title + " - Ammo (Tab: Items)"
	}
	return title + " - Items (Tab: Ammo)"
}

func isPageToggle(event *tcell.EventKey) bool {
	return event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab
}
