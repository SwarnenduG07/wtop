package ui

import (
	"fmt"
	"strings"

	"github.com/SwarnenduG07/wtop/metrics"
)

func formatFan(g *metrics.GPUInfo) string {
	if g.FanRPM > 0 {
		return fmt.Sprintf("%d RPM", g.FanRPM)
	}
	if g.FanSpeed > 0 {
		return fmt.Sprintf("%.0f%%", g.FanSpeed)
	}
	return "Off"
}

func formatPower(g *metrics.GPUInfo) string {
	if g.PowerUsage <= 0 {
		return "N/A"
	}
	if g.PowerLimit > 0 {
		return fmt.Sprintf("%.0f/%.0fW", g.PowerUsage, g.PowerLimit)
	}
	return fmt.Sprintf("%.0fW", g.PowerUsage)
}

func formatThrottle(reasons []string) string {
	filtered := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		reason = strings.TrimSpace(reason)
		if reason == "" || reason == "None" || reason == "GPU Idle" {
			continue
		}
		filtered = append(filtered, reason)
	}
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, ", ")
}

func RenderGPU(row *int, width int, limit int) {
	if *row > limit {
		return
	}
	gpus, err := metrics.GetGPUInfo()
	if err != nil || len(gpus) == 0 {
		return
	}

	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s GPU:%s", Bold+Blue, Reset)
	*row++
	if *row > limit {
		return
	}

	nameWidth := width - 46
	if nameWidth < 12 {
		nameWidth = 12
	}

	for _, gpu := range gpus {
		if *row > limit {
			break
		}

		MoveCursor(*row, 1)
		ClearLine()

		name := FitString(gpu.Name, nameWidth)
		memPercent := 0.0
		if gpu.MemoryTotal > 0 {
			memPercent = (gpu.MemoryUsed / gpu.MemoryTotal) * 100
		}

		line := fmt.Sprintf("  %s[%d]%s %-*s",
			Bold+Blue, gpu.Index, Reset, nameWidth, name)
		segments := []string{
			fmt.Sprintf(" %sUtil:%s %5.1f%%", Bold+Blue, Reset, gpu.Utilization),
			fmt.Sprintf(" %sMem:%s %5.1f%%", Bold+Blue, Reset, memPercent),
		}
		for _, seg := range segments {
			if VisibleLength(line+seg) > width {
				break
			}
			line += seg
		}
		fmt.Print(line)
		*row++
		if *row > limit {
			break
		}

		MoveCursor(*row, 1)
		ClearLine()
		usedGB := gpu.MemoryUsed / 1024
		totalGB := gpu.MemoryTotal / 1024
		line = fmt.Sprintf("    %sTemp:%s %5.1f°C", Bold+Blue, Reset, gpu.Temperature)
		details := []string{
			fmt.Sprintf(" %sFan:%s %s", Bold+Blue, Reset, formatFan(gpu)),
			fmt.Sprintf(" %sPower:%s %s", Bold+Blue, Reset, formatPower(gpu)),
			fmt.Sprintf(" %sClock:%s %d/%dMHz", Bold+Blue, Reset, gpu.ClockCore, gpu.ClockMemory),
			fmt.Sprintf(" %sMem:%s %.1f/%.1fG", Bold+Blue, Reset, usedGB, totalGB),
		}
		for _, seg := range details {
			if VisibleLength(line+seg) > width {
				break
			}
			line += seg
		}
		fmt.Print(line)
		*row++
		if *row > limit {
			break
		}

		throttle := formatThrottle(gpu.ThrottleReasons)
		if throttle != "" {
			MoveCursor(*row, 1)
			ClearLine()
			trimmed := FitString(throttle, width-20)
			throttleLine := fmt.Sprintf("    %sThrottle:%s %s%s%s",
				Bold+Blue, Reset, Yellow, trimmed, Reset)
			fmt.Print(throttleLine)
			*row++
			if *row > limit {
				break
			}
		}

		if width >= 100 && *row <= limit {
			processes, err := metrics.GetGPUProcesses(gpu.Index)
			if err == nil && len(processes) > 0 {
				MoveCursor(*row, 1)
				ClearLine()
				fmt.Printf("    %sProcesses:%s", Bold+Yellow, Reset)
				*row++
				if *row > limit {
					break
				}
				maxRows := limit - *row + 1
				if maxRows > 3 {
					maxRows = 3
				}
				for i := 0; i < len(processes) && i < maxRows; i++ {
					proc := processes[i]
					MoveCursor(*row, 1)
					ClearLine()
					nameWidth := width - 40
					if nameWidth < 10 {
						nameWidth = 10
					}
					procName := FitString(proc.ProcessName, nameWidth)
					fmt.Printf("      %sPID:%s %-6d %sMem:%s %5.0fMB %-*s",
						Yellow, Reset, proc.PID,
						Yellow, Reset, proc.MemoryUsed,
						nameWidth, procName)
					*row++
					if *row > limit {
						break
					}
				}
			}
		}

		if *row <= limit {
			*row++
		}
	}

	if *row > limit {
		return
	}

	intelGPU, err := metrics.GetIntelGPU()
	if err == nil && intelGPU != nil {
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("%s Integrated GPU:%s", Bold+Blue, Reset)
		*row++
		if *row > limit {
			return
		}
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("  %s%s%s", Green, FitString(intelGPU.Name, width-6), Reset)
		*row++
	}
}
