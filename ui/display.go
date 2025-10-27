package ui

import "fmt"

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

func MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func ClearLine() {
	fmt.Print("\033[2K")
}

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

func Repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func DrawColorBar(percent float64, width int) string {
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}
	
	color := Green
	if percent > 80 {
		color = Red
	} else if percent > 60 {
		color = Yellow
	}
	
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += color + "█" + Reset
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %s%.1f%%%s", color, percent, Reset)
	return bar
}

func DrawMemoryBar(percent float64, width int) string {
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}
	
	color := Green
	if percent > 90 {
		color = Red + Bold
	} else if percent > 75 {
		color = Red
	} else if percent > 50 {
		color = Yellow
	}
	
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += color + "█" + Reset
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %s%.1f%%%s", color, percent, Reset)
	return bar
}
