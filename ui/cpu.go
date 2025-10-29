package ui

import (
	"fmt"
	"math"
	"strings"
)

func (d *Dashboard) updateCPU(snap *snapshot) {
	if snap == nil || len(snap.CPUPerCore) == 0 {
		d.cpuView.SetText("[yellow]CPU metrics unavailable[-]")
		return
	}

	_, _, width, _ := d.cpuView.GetInnerRect()
	if width <= 0 {
		width = 80
	}
	totalBarWidth := clampInt(width-30, 12, 60)
	totalBar := renderUsageBar(snap.TotalCPU, totalBarWidth, d.theme)
	totalSpark := ""
	sparkWidth := clampInt(width-totalBarWidth-12, 8, 40)
	if d.cpuHistory != nil && sparkWidth >= 8 {
		totalSpark = "  " + renderSparkline(d.cpuHistory.Series(), sparkWidth, d.theme)
	}
	totalLine := fmt.Sprintf("Total %s%s", totalBar, totalSpark)

	cores := snap.CPUPerCore
	coresPerRow := determineCoresPerRow(width, len(cores))
	barWidth := computeBarWidth(width, coresPerRow)

	var lines []string
	lines = append(lines, totalLine)

	for i := 0; i < len(cores); i += coresPerRow {
		var builder strings.Builder
		for j := 0; j < coresPerRow; j++ {
			idx := i + j
			if idx >= len(cores) {
				break
			}
			if builder.Len() > 0 {
				builder.WriteString("  ")
			}
			label := fmt.Sprintf("%sC%02d%s", colorTag(d.theme.Accent), idx+1, resetTag())
			builder.WriteString(label)
			builder.WriteByte(' ')
			builder.WriteString(renderUsageBar(cores[idx], barWidth, d.theme))
		}
		lines = append(lines, builder.String())
	}

	d.cpuView.SetText(strings.Join(lines, "\n"))
}

func determineCoresPerRow(width int, total int) int {
	if total <= 0 {
		return 1
	}
	switch {
	case width >= 120 && total >= 4:
		return int(math.Min(4, float64(total)))
	case width >= 90 && total >= 3:
		return int(math.Min(3, float64(total)))
	case width >= 60 && total >= 2:
		return int(math.Min(2, float64(total)))
	default:
		return 1
	}
}

func computeBarWidth(width, coresPerRow int) int {
	if coresPerRow <= 0 {
		coresPerRow = 1
	}
	raw := (width / coresPerRow) - 10
	if raw < 6 {
		raw = 6
	}
	if raw > 60 {
		raw = 60
	}
	return raw
}
