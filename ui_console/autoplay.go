package ui_console

import (
	"github.com/gdamore/tcell/v2"
	"time"
)

// Autoplay hands the game to a script: every key is swallowed, and step runs
// once per delay on the UI event loop (like auto-run, via a synthetic key).
// Dialogues linger (readDelay) so they can be read. Quits 2s after the run ends.
func (u *UI) Autoplay(step func() (over bool), reading func() bool, delay, readDelay time.Duration) {
	const stepKey = tcell.KeyF39
	over := false
	u.application.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyCtrlC {
			return ev
		}
		if ev.Key() == stepKey && !over {
			over = step()
			go func() {
				switch {
				case over:
					time.Sleep(2 * time.Second)
					u.QuitGame()
					return
				case reading():
					time.Sleep(readDelay)
				default:
					time.Sleep(delay)
				}
				u.application.QueueEvent(tcell.NewEventKey(stepKey, 0, tcell.ModNone))
			}()
		}
		return nil
	})
	go func() {
		time.Sleep(time.Second) // let the screen come up
		u.application.QueueEvent(tcell.NewEventKey(stepKey, 0, tcell.ModNone))
	}()
}
