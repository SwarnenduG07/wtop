package ui

import (
	"fmt"

	"github.com/yourusername/wtop/metrics"
)

func drawTempBar(temp float64, width int) string {
	percent := temp / 100.0 * 100.0
	if percent > 100 {
		percent = 100
	}
	
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}
	
	color := Green
	if temp > 85 {
		color = Red + Bold
	} else if temp > 75 {
		color = Red
	} else if temp > 65 {
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
	bar += fmt.Sprintf("] %s%.0f°C%s", color, temp, Reset)
	return bar
}

func RenderGPU(row *int) {
	gpus, err := metrics.GetGPUInfo()
	
	// Show NVIDIA GPUs
	if err == nil && len(gpus) > 0 {
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("%s%sGPU (NVIDIA):%s", Bold+Blue, " ", Reset)
		*row++
		
		for _, gpu := range gpus {
			MoveCursor(*row, 1)
			ClearLine()
			fmt.Printf("  %s[%d]%s %s %s(Driver: %s)%s", 
				Bold+Blue, gpu.Index, Reset, gpu.Name,
				Cyan, gpu.Driver, Reset)
			*row++
			
			// Utilization and Memory
			MoveCursor(*row, 1)
			ClearLine()
			memPercent := (gpu.MemoryUsed / gpu.MemoryTotal) * 100
			fmt.Printf("  %sGPU:%s %s  %sMem:%s %s %.1fG/%.1fG", 
				Bold+Blue, Reset,
				DrawColorBar(gpu.Utilization, 15),
				Bold+Blue, Reset,
				DrawMemoryBar(memPercent, 15),
				gpu.MemoryUsed/1024, gpu.MemoryTotal/1024)
			*row++
			
			// Temperature, Power, Fan
			MoveCursor(*row, 1)
			ClearLine()
			fanStr := ""
			if gpu.FanRPM > 0 {
				fanStr = fmt.Sprintf("%d RPM", gpu.FanRPM)
			} else if gpu.FanSpeed > 0 {
				fanStr = fmt.Sprintf("%.0f%%", gpu.FanSpeed)
			} else {
				fanStr = "Off"
			}
			
			powerStr := ""
			if gpu.PowerUsage > 0 {
				if gpu.PowerLimit > 0 {
					powerStr = fmt.Sprintf("%.1fW/%.0fW", gpu.PowerUsage, gpu.PowerLimit)
				} else {
					powerStr = fmt.Sprintf("%.1fW", gpu.PowerUsage)
				}
			} else {
				powerStr = "N/A"
			}
			
			fmt.Printf("  %sTemp:%s %s  %sPower:%s %s%s%s  %sFan:%s %s%s%s",
				Bold+Blue, Reset,
				drawTempBar(gpu.Temperature, 15),
				Bold+Blue, Reset,
				Green, powerStr, Reset,
				Bold+Blue, Reset,
				Green, fanStr, Reset)
			*row++
			
			// Clocks
			MoveCursor(*row, 1)
			ClearLine()
			fmt.Printf("  %sClocks:%s Core: %s%dMHz%s  Memory: %s%dMHz%s",
				Bold+Blue, Reset,
				Green, gpu.ClockCore, Reset,
				Green, gpu.ClockMemory, Reset)
			*row++
			
			// GPU Processes
			processes, err := metrics.GetGPUProcesses(gpu.Index)
			if err == nil && len(processes) > 0 {
				MoveCursor(*row, 1)
				ClearLine()
				fmt.Printf("  %sGPU Processes:%s", Bold+Yellow, Reset)
				*row++
				
				count := 0
				for _, proc := range processes {
					if count >= 5 {
						break
					}
					MoveCursor(*row, 1)
					ClearLine()
					
					procName := proc.ProcessName
					if len(procName) > 30 {
						procName = procName[:27] + "..."
					}
					
					fmt.Printf("    %sPID:%s %s%-6d%s  %s%-30s%s  %sMem:%s %s%.0fMB%s",
						Yellow, Reset,
						Green, proc.PID, Reset,
						Green, procName, Reset,
						Yellow, Reset,
						Green, proc.MemoryUsed, Reset)
					*row++
					count++
				}
			}
			
			*row++
		}
	}
	
	// Show Intel integrated GPU
	intelGPU, err := metrics.GetIntelGPU()
	if err == nil && intelGPU != nil {
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("%s%sGPU (Integrated):%s", Bold+Blue, " ", Reset)
		*row++
		
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("  %s%s%s", Green, intelGPU.Name, Reset)
		*row++
		*row++
	}
}
