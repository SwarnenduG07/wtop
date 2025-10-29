package ui

import (
	"github.com/gdamore/tcell/v2"
)

func (d *Dashboard) bindKeys() {
	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			d.stop()
			d.app.Stop()
			return nil
		case tcell.KeyF1:
			d.footer.SetText("Help is coming soon. Visit the README for now.")
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				d.stop()
				d.app.Stop()
				return nil
			case 's', 'S':
				d.cycleSortMode()
				return nil
			case 't', 'T':
				d.toggleTheme()
				return nil
			case '/':
				d.footer.SetText("[yellow]Process filtering not implemented yet[-]")
				return nil
			}
		}
		return event
	})

	d.processTable.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			d.stop()
			d.app.Stop()
		}
	})
}
