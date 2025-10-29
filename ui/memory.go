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
	lines := []string{
		fmt.Sprintf("Mem  %s%s  %s/%s",
			renderUsageBar(mem.UsedPercent, barWidth, d.theme),
			memSpark,
			formatBytes(float64(mem.Used)),
			formatBytes(float64(mem.Total))),
	}

	if snap.Swap != nil && snap.Swap.Total > 0 {
		swapSpark := ""
		if d.swapHistory != nil && sparkWidth >= 8 {
			swapSpark = "  " + renderSparkline(d.swapHistory.Series(), sparkWidth, d.theme)
		}
		lines = append(lines, fmt.Sprintf("Swap %s  %s/%s",
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
		lines = append(lines, fmt.Sprintf("Disk %s  %s/%s (%s)",
			renderUsageBar(diskPercent, barWidth, d.theme)+diskSpark,
			formatBytes(float64(snap.Disk.Used)),
			formatBytes(float64(snap.Disk.Total)),
			snap.DiskPath))
	}

	if snap.ProcessSummary.Total > 0 {
		lines = append(lines, fmt.Sprintf("Tasks %d  Threads %d  Running %d",
			snap.ProcessSummary.Total,
			snap.ProcessSummary.Threads,
			snap.ProcessSummary.Running))
	}

	if width >= 48 {
		extra := []string{}
		extra = append(extra, fmt.Sprintf("Avail %s", formatBytes(float64(mem.Available))))
		if mem.Cached > 0 {
			extra = append(extra, fmt.Sprintf("Cached %s", formatBytes(float64(mem.Cached))))
		}
		if mem.Buffers > 0 {
			extra = append(extra, fmt.Sprintf("Buffers %s", formatBytes(float64(mem.Buffers))))
		}
		if len(extra) > 0 {
			lines = append(lines, strings.Join(extra, "  "))
		}
	}

	d.memoryView.SetText(strings.Join(lines, "\n"))
}
