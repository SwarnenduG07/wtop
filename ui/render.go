package ui

import "fmt"

func RenderFrame(width, height int) {
	if width < 20 {
		width = 20
	}
	if height < 10 {
		height = 10
	}

	footerHeight := 3
	contentLimit := height - footerHeight
	if contentLimit < 3 {
		contentLimit = height
	}

	row := 1
	RenderHeader(&row, width)

	RenderCPU(&row, width, contentLimit)
	RenderMemory(&row, width, contentLimit)
	RenderGPU(&row, width, contentLimit)
	RenderSystemInfo(&row, width, contentLimit)
	RenderProcessTable(&row, width, contentLimit)

	if row <= contentLimit {
		MoveCursor(row, 1)
		fmt.Print("\033[J")
	}

	RenderFooter(width, height)
}
