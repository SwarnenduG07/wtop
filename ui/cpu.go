package ui

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
)

func RenderCPU(row *int) {
	cpuPercents, _ := cpu.Percent(0, true)
	cpuCount := len(cpuPercents)
	
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s%sCPU USAGE:%s", Bold+Cyan, " ", Reset)
	*row++
	
	maxCores := cpuCount
	if maxCores > 16 {
		maxCores = 16
	}
	
	for i := 0; i < maxCores; i += 4 {
		MoveCursor(*row, 1)
		ClearLine()
		
		for j := 0; j < 4 && (i+j) < maxCores; j++ {
			coreIdx := i + j
			if coreIdx < len(cpuPercents) {
				fmt.Printf("  %s%2d%s%s", Bold+Cyan, coreIdx+1, Reset, DrawColorBar(cpuPercents[coreIdx], 12))
				if j < 3 && (i+j+1) < maxCores {
					fmt.Print("  ")
				}
			}
		}
		*row++
	}
	
	*row++
	*row++
}
