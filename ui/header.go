package ui

import (
	"fmt"
	"strings"
)

func (d *Dashboard) updateHeader(snap *snapshot, rates netRates) {
	if snap == nil {
		d.header.SetText("[yellow]collecting metrics...[-]")
		return
	}

	if !rates.Valid && d.lastRates.Valid {
		rates = d.lastRates
	}

	_, _, width, _ := d.header.GetInnerRect()
	if width <= 0 {
		width = 80
	}

	reset := resetTag()
	accent := colorTag(d.theme.Accent)

	hostname := strings.TrimSpace(snap.Hostname)
	if hostname == "" {
		hostname = "unknown"
	}

	lineOne := joinWithSpacing([]string{
		fmt.Sprintf("%s%s%s", accent, hostname, reset),
		fmt.Sprintf("up %s", formatUptime(snap.Uptime)),
		fmt.Sprintf("tasks %d/%d", snap.ProcessSummary.Running, snap.ProcessSummary.Total),
	})

	if snap.LoadReported {
		loadStr := fmt.Sprintf("load %.2f %.2f %.2f", snap.Load1, snap.Load5, snap.Load15)
		lineOne = joinWithSpacing([]string{lineOne, loadStr})
	}

	cpuBarWidth := clampInt(width/3, 12, 40)
	cpuBar := renderUsageBar(snap.TotalCPU, cpuBarWidth, d.theme)
	cpuSpark := ""
	sparkWidth := clampInt(width/3, 8, 40)
	if d.cpuHistory != nil && sparkWidth >= 8 {
		cpuSpark = "  " + renderSparkline(d.cpuHistory.Series(), sparkWidth, d.theme)
	}

	partsLineTwo := []string{
		fmt.Sprintf("⚙ CPU %s%s", cpuBar, cpuSpark),
	}

	if snap.Memory != nil {
		memPercent := snap.Memory.UsedPercent
		memBar := renderUsageBar(memPercent, cpuBarWidth, d.theme)
		memSpark := ""
		if d.memHistory != nil && sparkWidth >= 8 {
			memSpark = "  " + renderSparkline(d.memHistory.Series(), sparkWidth, d.theme)
		}
		partsLineTwo = append(partsLineTwo,
			fmt.Sprintf("💾 MEM %s%s %s/%s",
				memBar, memSpark,
				formatBytes(float64(snap.Memory.Used)),
				formatBytes(float64(snap.Memory.Total))))
	}

	if snap.Swap != nil && snap.Swap.Total > 0 {
		swapBar := renderUsageBar(snap.Swap.UsedPercent, clampInt(cpuBarWidth, 10, 30), d.theme)
		partsLineTwo = append(partsLineTwo,
			fmt.Sprintf("� SWP %s %s/%s",
				swapBar,
				formatBytes(float64(snap.Swap.Used)),
				formatBytes(float64(snap.Swap.Total))))
	} else if snap.Disk != nil && snap.Disk.Total > 0 {
		diskPercent := (float64(snap.Disk.Used) / float64(snap.Disk.Total)) * 100
		diskBar := renderUsageBar(diskPercent, clampInt(cpuBarWidth, 10, 30), d.theme)
		partsLineTwo = append(partsLineTwo,
			fmt.Sprintf("📀 DISK %s %s/%s",
				diskBar,
				formatBytes(float64(snap.Disk.Used)),
				formatBytes(float64(snap.Disk.Total))))
	}

	lineTwo := joinWithSpacing(partsLineTwo)

	netLine := ""
	if rates.Valid {
		up := formatBytesPerSec(rates.Up)
		down := formatBytesPerSec(rates.Down)
		upSpark, downSpark := "", ""
		netSparkWidth := clampInt(width/4, 8, 32)
		if d.netUpHistory != nil && netSparkWidth >= 8 {
			upSpark = "  " + renderSparkline(d.netUpHistory.Series(), netSparkWidth, d.theme)
		}
		if d.netDnHistory != nil && netSparkWidth >= 8 {
			downSpark = "  " + renderSparkline(d.netDnHistory.Series(), netSparkWidth, d.theme)
		}
		netLine = joinWithSpacing([]string{
			fmt.Sprintf("↑ %s%s", up, upSpark),
			fmt.Sprintf("↓ %s%s", down, downSpark),
		})
		netLine = fmt.Sprintf("🌐 %s", netLine)
	}

	if netLine == "" {
		d.header.SetText(lineOne + "\n" + lineTwo)
	} else {
		d.header.SetText(lineOne + "\n" + lineTwo + "\n" + netLine)
	}
}
