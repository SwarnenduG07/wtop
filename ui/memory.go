package ui

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
)

func RenderMemory(row *int) {
	v, err := mem.VirtualMemory()
	if err == nil {
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("%s%sMEMORY & SWAP:%s", Bold+Purple, " ", Reset)
		*row++
		
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("%s  Mem:%s %s  %.1fG/%.1fG", 
			Bold+Purple, Reset,
			DrawMemoryBar(v.UsedPercent, 35),
			float64(v.Used)/1024/1024/1024, 
			float64(v.Total)/1024/1024/1024)
		*row++
		
		swap, _ := mem.SwapMemory()
		if swap.Total > 0 {
			MoveCursor(*row, 1)
			ClearLine()
			fmt.Printf("%s  Swp:%s %s  %.1fG/%.1fG", 
				Bold+Purple, Reset,
				DrawMemoryBar(swap.UsedPercent, 35),
				float64(swap.Used)/1024/1024/1024, 
				float64(swap.Total)/1024/1024/1024)
			*row++
		}
		
		*row++
	}
}
