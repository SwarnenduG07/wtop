package ui

import (
	"fmt"
	"time"

	"github.com/SwarnenduG07/wtop/metrics"
)

func formatTime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("0:%02d.00", seconds)
	}
	minutes := seconds / 60
	secs := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%d:%02d.00", minutes, secs)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, mins, secs)
}

func formatBytesMB(value uint64) string {
	return fmt.Sprintf("%.0fM", float64(value)/1024/1024)
}

func RenderProcessTable(row *int, width int, limit int) {
	if *row > limit {
		return
	}
	availableRows := limit - *row + 1
	if availableRows <= 3 {
		return
	}

	layout := "wide"
	nameWidth := width - 72
	userWidth := 9

	switch {
	case width >= 120 && nameWidth >= 14:
		layout = "wide"
	case width >= 100:
		layout = "medium"
		nameWidth = width - 40
		userWidth = 9
	case width >= 70:
		layout = "compact"
		nameWidth = width - 30
		userWidth = 8
	case width >= 46:
		layout = "minimal"
		nameWidth = width - 18
	default:
		MoveCursor(*row, 1)
		ClearLine()
		fmt.Printf("%sResize window to view process table%s", Yellow, Reset)
		*row++
		return
	}

	if nameWidth < 8 {
		nameWidth = 8
	}

	MoveCursor(*row, 1)
	ClearLine()
	var header string
	switch layout {
	case "wide":
		header = "  PID USER      PRI  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND"
	case "medium":
		header = "  PID USER      S  %CPU %MEM     TIME+ COMMAND"
	case "compact":
		header = "  PID USER   S %CPU %MEM COMMAND"
	case "minimal":
		header = "  PID %CPU %MEM COMMAND"
	}
	if VisibleLength(header) > width {
		header = FitString(header, width)
	}
	fmt.Printf("%s%s%s", Bold+White, header, Reset)
	*row++
	if *row > limit {
		return
	}

	maxRows := limit - *row + 1
	processInfos := metrics.GetTopProcesses(maxRows)
	currentTime := time.Now().Unix()

	for _, info := range processInfos {
		if *row > limit {
			break
		}
		MoveCursor(*row, 1)
		ClearLine()

		user := FitString(info.User, userWidth)
		name := FitString(info.Name, nameWidth)
		runtime := currentTime - info.CreateTime/1000
		if runtime < 0 {
			runtime = 0
		}
		timeStr := formatTime(runtime)

		line := ""
		switch layout {
		case "wide":
			virt := formatBytesMB(info.VirtMem)
			res := formatBytesMB(info.ResMem)
			shr := formatBytesMB(info.ShrMem)
			line = fmt.Sprintf("%s%5d%s %-*s %3d %3d %7s %7s %7s %-1s %6.1f %6.1f %9s %-*s",
				Green, info.PID, Reset,
				userWidth, user,
				info.Priority,
				info.Nice,
				virt,
				res,
				shr,
				info.Status,
				info.CPUPercent,
				info.MemPercent,
				timeStr,
				nameWidth, name)
		case "medium":
			line = fmt.Sprintf("%s%5d%s %-*s %-1s %6.1f %6.1f %9s %-*s",
				Green, info.PID, Reset,
				userWidth, user,
				info.Status,
				info.CPUPercent,
				info.MemPercent,
				timeStr,
				nameWidth, name)
		case "compact":
			line = fmt.Sprintf("%s%5d%s %-*s %-1s %5.1f %5.1f %-*s",
				Green, info.PID, Reset,
				userWidth, user,
				info.Status,
				info.CPUPercent,
				info.MemPercent,
				nameWidth, name)
		case "minimal":
			line = fmt.Sprintf("%s%5d%s %5.1f %5.1f %-*s",
				Green, info.PID, Reset,
				info.CPUPercent,
				info.MemPercent,
				nameWidth, name)
		}

		if VisibleLength(line) > width {
			trimWidth := nameWidth
			for VisibleLength(line) > width && trimWidth > 4 {
				trimWidth--
				name = FitString(info.Name, trimWidth)
				switch layout {
				case "wide":
					line = fmt.Sprintf("%s%5d%s %-*s %3d %3d %7s %7s %7s %-1s %6.1f %6.1f %9s %-*s",
						Green, info.PID, Reset,
						userWidth, user,
						info.Priority,
						info.Nice,
						formatBytesMB(info.VirtMem),
						formatBytesMB(info.ResMem),
						formatBytesMB(info.ShrMem),
						info.Status,
						info.CPUPercent,
						info.MemPercent,
						timeStr,
						trimWidth, name)
				case "medium":
					line = fmt.Sprintf("%s%5d%s %-*s %-1s %6.1f %6.1f %9s %-*s",
						Green, info.PID, Reset,
						userWidth, user,
						info.Status,
						info.CPUPercent,
						info.MemPercent,
						timeStr,
						trimWidth, name)
				case "compact":
					line = fmt.Sprintf("%s%5d%s %-*s %-1s %5.1f %5.1f %-*s",
						Green, info.PID, Reset,
						userWidth, user,
						info.Status,
						info.CPUPercent,
						info.MemPercent,
						trimWidth, name)
				case "minimal":
					line = fmt.Sprintf("%s%5d%s %5.1f %5.1f %-*s",
						Green, info.PID, Reset,
						info.CPUPercent,
						info.MemPercent,
						trimWidth, name)
				}
			}
		}

		fmt.Print(line)
		*row++
	}
}
