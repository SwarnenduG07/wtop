package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/SwarnenduG07/wtop/metrics"
	"github.com/gdamore/tcell/v2"
)

func (d *Dashboard) updateGPU(snap *snapshot) {
	if snap == nil || len(snap.GPUInfos) == 0 {
		d.gpuView.SetText("[gray]No discrete GPU detected[-]")
		return
	}

	_, _, width, _ := d.gpuView.GetInnerRect()
	if width <= 0 {
		width = 80
	}

	var lines []string

	for _, gpu := range snap.GPUInfos {
		// Header line: [index] GPU Name (Driver: version)
		driver := strings.TrimSpace(gpu.Driver)
		if driver == "" {
			driver = "Unknown"
		}
		headerLine := fmt.Sprintf("  [[%d]] %s (Driver: %s)", gpu.Index, gpu.Name, driver)
		lines = append(lines, headerLine)

		// P-State, Compute, Bus, PCIe line
		stateParts := []string{}

		// P-State
		pstate := strings.TrimSpace(gpu.PerformanceState)
		if pstate == "" {
			pstate = strings.TrimSpace(gpu.PowerState)
		}
		if pstate != "" && pstate != "[N/A]" && pstate != "[Unknown]" {
			stateParts = append(stateParts, fmt.Sprintf("P-State: %s", pstate))
		}

		// Compute Mode
		computeMode := strings.TrimSpace(gpu.ComputeMode)
		if computeMode != "" && computeMode != "[N/A]" {
			stateParts = append(stateParts, fmt.Sprintf("Compute: %s", computeMode))
		}

		// Memory Bus Width
		if gpu.MemoryBusWidth > 0 {
			stateParts = append(stateParts, fmt.Sprintf("Bus: %d-bit", gpu.MemoryBusWidth))
		}

		// PCIe
		if gpu.PCIeGen > 0 && gpu.PCIeWidth > 0 {
			stateParts = append(stateParts, fmt.Sprintf("PCIe: Gen%d x%d", gpu.PCIeGen, gpu.PCIeWidth))
		}

		if len(stateParts) > 0 {
			lines = append(lines, "  "+strings.Join(stateParts, "  "))
		}

		// GPU and Memory bars line
		gpuBar := renderBtopBar(gpu.Utilization, 15, d.theme)
		memPercent := 0.0
		if gpu.MemoryTotal > 0 {
			memPercent = (gpu.MemoryUsed / gpu.MemoryTotal) * 100
		}
		memBar := renderBtopBar(memPercent, 15, d.theme)
		usedGB := gpu.MemoryUsed / 1024
		totalGB := gpu.MemoryTotal / 1024
		memUsage := fmt.Sprintf("%.1fG/%.1fG", usedGB, totalGB)
		gpuMemLine := fmt.Sprintf("  GPU: %s %.1f%%  Mem: %s %.1f%% %s",
			gpuBar, gpu.Utilization, memBar, memPercent, memUsage)
		lines = append(lines, gpuMemLine)

		// Memory Controller and SM clock line
		memCtrlBar := renderBtopBar(gpu.MemoryUtilization, 15, d.theme)
		smClock := ""
		if gpu.ClockSM > 0 {
			smClock = fmt.Sprintf("SM: %dMHz", gpu.ClockSM)
		}
		memCtrlLine := fmt.Sprintf("  Mem Ctrl: %s %.1f%%  %s", memCtrlBar, gpu.MemoryUtilization, smClock)
		lines = append(lines, memCtrlLine)

		// Temperature bar with Power and Fan
		tempBar := renderBtopBar(gpu.Temperature, 15, d.theme)
		tempLine := fmt.Sprintf("  Temp: %s %.0f°C", tempBar, gpu.Temperature)

		if gpu.PowerUsage > 0 {
			tempLine += fmt.Sprintf("  Power: %.1fW", gpu.PowerUsage)
		}

		if gpu.FanRPM > 0 {
			tempLine += fmt.Sprintf("  Fan: %d RPM", gpu.FanRPM)
		} else if gpu.FanSpeed > 0 {
			tempLine += fmt.Sprintf("  Fan: %.0f%%", gpu.FanSpeed)
		}

		lines = append(lines, tempLine)

		// Clocks line
		if gpu.ClockCore > 0 || gpu.ClockMemory > 0 {
			clockLine := "  Clocks:"
			if gpu.ClockCore > 0 {
				clockLine += fmt.Sprintf(" Core: %dMHz", gpu.ClockCore)
			}
			if gpu.ClockMemory > 0 {
				clockLine += fmt.Sprintf("  Memory: %dMHz", gpu.ClockMemory)
			}
			lines = append(lines, clockLine)
		}

		// Throttle line
		throttleReasons := formatGPUThrottle(gpu.ThrottleReasons)
		if throttleReasons != "" {
			lines = append(lines, fmt.Sprintf("  Throttle: %s", throttleReasons))
		} else {
			lines = append(lines, "  Throttle: None")
		}

		// GPU Processes - render as a compact table with totals
		if snap.GPUProcesses != nil {
			if procs := snap.GPUProcesses[gpu.Index]; len(procs) > 0 {
				lines = append(lines, "  GPU Processes:")

				// Header
				lines = append(lines, fmt.Sprintf("    %-6s  %-25s  %-8s  %6s", "PID", "Process", "Type", "Mem"))

				totalProcMem := 0.0
				for _, proc := range procs {
					procType := strings.TrimSpace(proc.Type)
					if procType == "" {
						procType = "Compute"
					}
					name := truncateLabel(proc.ProcessName, 25)
					memMB := proc.MemoryUsed
					totalProcMem += memMB
					procLine := fmt.Sprintf("    %-6d  %-25s  %-8s  %6.0fMB",
						proc.PID,
						name,
						procType,
						memMB)
					lines = append(lines, procLine)
				}

				// Summary of process memory usage
				lines = append(lines, fmt.Sprintf("    Processes: %d  Total GPU Mem: %.0fMB",
					len(procs), totalProcMem))
			}
		}

		lines = append(lines, "")
	}

	d.gpuView.SetText(strings.TrimSpace(strings.Join(lines, "\n")))
}

func truncateLabel(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func renderBtopBar(value float64, width int, theme Theme) string {
	width = clampInt(width, 6, 60)

	// For temperature, scale to 0-100 range (assuming 100°C max)
	if value > 100 {
		value = 100
	}

	filled := int(math.Round(value / 100 * float64(width)))
	if filled > width {
		filled = width
	}

	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < width; i++ {
		if i < filled {
			b.WriteRune('█')
		} else {
			b.WriteRune(' ')
		}
	}
	b.WriteString("]")
	return b.String()
}

func getTempColor(temp float64, theme Theme) tcell.Color {
	if temp < 50 {
		return tcell.ColorGreen
	} else if temp < 70 {
		return theme.Warning
	} else if temp < 85 {
		return tcell.ColorOrange
	}
	return theme.Critical
}

func formatGPUFan(g *metrics.GPUInfo) string {
	if g.FanRPM > 0 {
		return fmt.Sprintf("%d RPM", g.FanRPM)
	}
	if g.FanSpeed > 0 {
		return fmt.Sprintf("%.0f%%", g.FanSpeed)
	}
	return "Off"
}

func formatGPUThrottle(reasons []string) string {
	filtered := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		reason = strings.TrimSpace(reason)
		if reason == "" || reason == "None" || reason == "GPU Idle" {
			continue
		}
		filtered = append(filtered, reason)
	}
	return strings.Join(filtered, ", ")
}

func formatGPUPower(g *metrics.GPUInfo) string {
	if g.PowerLimit > 0 {
		return fmt.Sprintf("Power %.0f/%.0fW", g.PowerUsage, g.PowerLimit)
	}
	if g.PowerUsage > 0 {
		return fmt.Sprintf("Power %.0fW", g.PowerUsage)
	}
	return "Power N/A"
}

func formatGPUClocks(g *metrics.GPUInfo) string {
	if g.ClockCore > 0 || g.ClockMemory > 0 {
		return fmt.Sprintf("Clock %d/%d MHz", g.ClockCore, g.ClockMemory)
	}
	return "Clock N/A"
}
