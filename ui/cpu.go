package ui

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
)

func RenderCPU(row *int, width int, limit int) {
	if *row > limit {
		return
	}
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil || len(cpuPercents) == 0 {
		return
	}

	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s CPU USAGE:%s", Bold+Cyan, Reset)
	*row++
	if *row > limit {
		return
	}

	coresPerRow := 1
	switch {
	case width >= 140:
		coresPerRow = 4
	case width >= 110:
		coresPerRow = 3
	case width >= 80:
		coresPerRow = 2
	}
	if coresPerRow > len(cpuPercents) {
		coresPerRow = len(cpuPercents)
	}
	if coresPerRow == 0 {
		coresPerRow = 1
	}

	barWidth := width/coresPerRow - 18
	if barWidth < 3 {
		barWidth = 3
	}
	if barWidth > width-12 {
		barWidth = width - 12
	}

	availableRows := limit - *row + 1
	if availableRows <= 0 {
		return
	}

	maxLines := (len(cpuPercents) + coresPerRow - 1) / coresPerRow
	if maxLines > availableRows {
		maxLines = availableRows
	}
	maxCores := maxLines * coresPerRow
	if maxCores > len(cpuPercents) {
		maxCores = len(cpuPercents)
	}

	for i := 0; i < maxCores; i += coresPerRow {
		if *row > limit {
			break
		}
		MoveCursor(*row, 1)
		ClearLine()

		var line strings.Builder
		for j := 0; j < coresPerRow; j++ {
			coreIdx := i + j
			if coreIdx >= maxCores {
				break
			}
			if line.Len() > 0 {
				line.WriteString("  ")
			}
			label := fmt.Sprintf("%s%2d%s", Bold+Cyan, coreIdx+1, Reset)
			line.WriteString(label)
			line.WriteByte(' ')
			line.WriteString(DrawColorBar(cpuPercents[coreIdx], barWidth))
		}

		output := line.String()
		if VisibleLength(output) > width {
			// Reduce bar width and rebuild if we exceeded the width budget.
			adjustedWidth := width/coresPerRow - 18
			if adjustedWidth < 3 {
				adjustedWidth = 3
			}
			for {
				line.Reset()
				for j := 0; j < coresPerRow; j++ {
					coreIdx := i + j
					if coreIdx >= maxCores {
						break
					}
					if line.Len() > 0 {
						line.WriteString("  ")
					}
					label := fmt.Sprintf("%s%2d%s", Bold+Cyan, coreIdx+1, Reset)
					line.WriteString(label)
					line.WriteByte(' ')
					line.WriteString(DrawColorBar(cpuPercents[coreIdx], adjustedWidth))
				}
				output = line.String()
				if VisibleLength(output) <= width || adjustedWidth <= 3 {
					break
				}
				adjustedWidth--
			}
		}

		fmt.Print(output)
		*row++
	}

	if *row <= limit {
		*row++
	}
}
