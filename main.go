package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SwarnenduG07/wtop/ui"
)

func main() {
	ui.HideCursor()
	defer ui.ShowCursor()

	fmt.Printf("%sStarting wtop... Press Ctrl+C to exit%s\n", ui.Green, ui.Reset)
	time.Sleep(1 * time.Second)

	ui.ClearScreen()

	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	render := func() {
		width, height := ui.GetTerminalSize()
		ui.RenderFrame(width, height)
	}

	render()

	for {
		select {
		case <-ticker.C:
			render()
		case <-resizeCh:
			ui.ClearScreen()
			render()
		}
	}
}
