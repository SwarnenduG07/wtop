package ui

import (
	"fmt"
	"strings"
)

func (d *Dashboard) updateMemory(snap *snapshot) {
	if snap == nil || snap.Memory == nil {
		d.memoryView.SetText("[yellow]Memory metrics unavailable[-]")
		return
	}

	_, _, width, _ := d.memoryView.GetInnerRect()
	if width <= 0 {
		width = 60
	}
	barWidth := clampInt(width-20, 10, 60)
	sparkWidth := clampInt(width-barWidth-12, 8, 40)

	mem := snap.Memory
	memSpark := ""
	if d.memHistory != nil && sparkWidth >= 8 {
		memSpark = "  " + renderSparkline(d.memHistory.Series(), sparkWidth, d.theme)
	}
	// Split output into memory (left) and disk (right) panes.
	memLines := []string{}
	diskLines := []string{}

	// Memory: Total
	memLines = append(memLines, fmt.Sprintf("Total: %s", formatBytes(float64(mem.Total))))

	// Memory: Used (bar + spark + value + percent)
	memLines = append(memLines, fmt.Sprintf("Used:  %s  %s (%.0f%%)",
		renderUsageBar(mem.UsedPercent, barWidth, d.theme)+memSpark,
		formatBytes(float64(mem.Used)),
		mem.UsedPercent))

	// Memory: Available
	if mem.Available > 0 {
		availPercent := (float64(mem.Available) / float64(mem.Total)) * 100
		memLines = append(memLines, fmt.Sprintf("Available: %s  %s (%.0f%%)",
			formatBytes(float64(mem.Available)),
			renderUsageBar(availPercent, clampInt(barWidth, 10, 60), d.theme),
			availPercent))
	}

	// Memory: Cached and Buffers
	if mem.Cached > 0 {
		cachedPercent := (float64(mem.Cached) / float64(mem.Total)) * 100
		memLines = append(memLines, fmt.Sprintf("Cached:    %s  (%.0f%%)",
			formatBytes(float64(mem.Cached)),
			cachedPercent))
	}
	if mem.Buffers > 0 {
		bufPercent := (float64(mem.Buffers) / float64(mem.Total)) * 100
		memLines = append(memLines, fmt.Sprintf("Buffers:   %s  (%.0f%%)",
			formatBytes(float64(mem.Buffers)),
			bufPercent))
	}

	// Tasks (keep with memory pane)
	if snap.ProcessSummary.Total > 0 {
		memLines = append(memLines, fmt.Sprintf("Tasks %d  Threads %d  Running %d",
			snap.ProcessSummary.Total,
			snap.ProcessSummary.Threads,
			snap.ProcessSummary.Running))
	}

	// Disk/Swap: put on the disk pane
	if snap.Swap != nil && snap.Swap.Total > 0 {
		swapSpark := ""
		if d.swapHistory != nil && sparkWidth >= 8 {
			swapSpark = "  " + renderSparkline(d.swapHistory.Series(), sparkWidth, d.theme)
		}
		diskLines = append(diskLines, fmt.Sprintf("Swap %s  %s/%s",
			renderUsageBar(snap.Swap.UsedPercent, barWidth, d.theme)+swapSpark,
			formatBytes(float64(snap.Swap.Used)),
			formatBytes(float64(snap.Swap.Total))))
	}

	if snap.Disk != nil && snap.Disk.Total > 0 {
		diskPercent := (float64(snap.Disk.Used) / float64(snap.Disk.Total)) * 100
		diskSpark := ""
		if d.diskHistory != nil && sparkWidth >= 8 {
			diskSpark = "  " + renderSparkline(d.diskHistory.Series(), sparkWidth, d.theme)
		}
		diskLines = append(diskLines, fmt.Sprintf("Disk %s  %s/%s (%s)",
			renderUsageBar(diskPercent, barWidth, d.theme)+diskSpark,
			formatBytes(float64(snap.Disk.Used)),
			formatBytes(float64(snap.Disk.Total)),
			snap.DiskPath))
	}

	// Write into respective views
	d.memoryView.SetText(strings.Join(memLines, "\n"))
	if d.diskView != nil {
		d.diskView.SetText(strings.Join(diskLines, "\n"))
	}
}
