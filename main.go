package main

import (
	"fmt"
	"time"

	"github.com/yourusername/wtop/ui"
)

func main() {
	ui.HideCursor()
	defer ui.ShowCursor()
	
	fmt.Printf("%sStarting wtop... Press Ctrl+C to exit%s\n", ui.Green, ui.Reset)
	time.Sleep(1 * time.Second)
	
	ui.ClearScreen()
	
	for {
		row := 1
		
		ui.RenderHeader(&row)
		ui.RenderCPU(&row)
		ui.RenderMemory(&row)
		ui.RenderGPU(&row)
		ui.RenderSystemInfo(&row)
		ui.RenderProcessTable(&row)
		ui.RenderFooter(&row)
		
		time.Sleep(2 * time.Second)
	}
}
