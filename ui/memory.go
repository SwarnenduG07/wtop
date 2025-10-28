package ui

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
)

func RenderMemory(row *int, width int, limit int) {
	if *row > limit {
		return
	}
	v, err := mem.VirtualMemory()
	if err != nil {
		return
	}

	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s MEMORY & SWAP:%s", Bold+Purple, Reset)
	*row++
	if *row > limit {
		return
	}

	barWidth := width - 32
	if barWidth < 10 {
		barWidth = width / 2
	}
	if barWidth < 8 {
		barWidth = 8
	}

	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s  Mem:%s %s  %.1fG/%.1fG",
		Bold+Purple, Reset,
		DrawMemoryBar(v.UsedPercent, barWidth),
		float64(v.Used)/1024/1024/1024,
		float64(v.Total)/1024/1024/1024)
	*row++
	if *row > limit {
		return
	}

	swap, _ := mem.SwapMemory()
	if swap.Total > 0 {
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf(
			"%s  Swp:%s %s  %.1fG/%.1fG",
			Bold+Purple, Reset,
			DrawMemoryBar(swap.UsedPercent, barWidth),
			float64(swap.Used)/1024/1024/1024,
			float64(swap.Total)/1024/1024/1024,
		)
		*row++
		if *row > limit {
			return
		}
	}

	*row++
}
