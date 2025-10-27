package ui

import "fmt"

func RenderFooter(row *int) {
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s%s%s", Yellow, Repeat("=", 80), Reset)
	*row++
	
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%sF1Help F2Setup F3Search F4Filter F5Tree F6SortBy F7Nice F8Nice+ F9Kill F10Quit%s", Green, Reset)
	*row++
	
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%sPress Ctrl+C to quit • Refreshing every 2 seconds%s", Cyan, Reset)
}
