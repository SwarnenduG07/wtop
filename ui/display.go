package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

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
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

func DrawColorBar(percent float64, width int) string {
	if width < 1 {
		width = 1
	}
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

	var builder strings.Builder
	builder.Grow(width + 16)
	builder.WriteByte('[')
	for i := 0; i < width; i++ {
		if i < filled {
			builder.WriteString(color)
			builder.WriteRune('█')
			builder.WriteString(Reset)
		} else {
			builder.WriteByte(' ')
		}
	}
	builder.WriteString("] ")
	builder.WriteString(color)
	builder.WriteString(fmt.Sprintf("%.1f%%", percent))
	builder.WriteString(Reset)
	return builder.String()
}

func DrawMemoryBar(percent float64, width int) string {
	if width < 1 {
		width = 1
	}
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

	var builder strings.Builder
	builder.Grow(width + 16)
	builder.WriteByte('[')
	for i := 0; i < width; i++ {
		if i < filled {
			builder.WriteString(color)
			builder.WriteRune('█')
			builder.WriteString(Reset)
		} else {
			builder.WriteByte(' ')
		}
	}
	builder.WriteString("] ")
	builder.WriteString(color)
	builder.WriteString(fmt.Sprintf("%.1f%%", percent))
	builder.WriteString(Reset)
	return builder.String()
}

func FitString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func VisibleLength(s string) int {
	count := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) {
				c := s[i]
				if (c >= '0' && c <= '9') || c == ';' {
					i++
					continue
				}
				// Consume final byte of control sequence.
				i++
				break
			}
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		if size == 0 {
			break
		}
		i += size
		count++
	}
	return count
}

func FitPlainString(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}
