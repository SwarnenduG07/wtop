package ui

import (
	"os"

	"golang.org/x/term"
)

func GetTerminalSize() (int, int) {
	fd := int(os.Stdout.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		return 120, 40
	}
	if width < 20 {
		width = 20
	}
	if height < 10 {
		height = 10
	}
	return width, height
}
