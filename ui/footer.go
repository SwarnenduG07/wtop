package ui

import "fmt"

func RenderFooter(width, height int) {
	if width < 1 || height < 1 {
		return
	}
	lines := []struct {
		color string
		text  string
	}{
		{Yellow, Repeat("=", width)},
		{Green, FitPlainString("F1 Help  F2 Setup  F3 Search  F4 Filter  F5 Tree  F6 Sort  F7 Nice-  F8 Nice+  F9 Kill  F10 Quit", width)},
		{Cyan, FitPlainString("Press Ctrl+C to quit • Refreshing every 2 seconds", width)},
	}
	start := height - len(lines) + 1
	if start < 1 {
		start = 1
	}
	for i, line := range lines {
		row := start + i
		if row > height {
			break
		}
		MoveCursor(row, 1)
		ClearLine()
		fmt.Printf("%s%s%s", line.color, line.text, Reset)
	}
}
