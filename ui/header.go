package ui

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/host"
)

func formatUptime(d time.Duration) string {
	if d < time.Minute {
		return d.Truncate(time.Second).String()
	}
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func RenderHeader(row *int, width int) {
	if width < 20 {
		width = 20
	}
	MoveCursor(*row, 1)
	ClearLine()

	hostInfo, _ := host.Info()
	if hostInfo == nil {
		hostInfo = &host.InfoStat{Hostname: "unknown"}
	}
	uptime := time.Duration(hostInfo.Uptime) * time.Second

	title := fmt.Sprintf("%swtop%s %s%s%s", Bold+Cyan, Reset, Bold, hostInfo.Hostname, Reset)
	uptimeStr := fmt.Sprintf("%sUptime:%s %s", Green, Reset, formatUptime(uptime))
	padding := width - VisibleLength(title) - VisibleLength(uptimeStr)
	if padding < 1 {
		padding = 1
	}

	fmt.Printf("%s%s%s", title, Repeat(" ", padding), uptimeStr)
	*row++

	MoveCursor(*row, 1)
	ClearLine()
	sepLen := width
	if sepLen < 1 {
		sepLen = 1
	}
	fmt.Printf("%s%s%s", Yellow, Repeat("=", sepLen), Reset)
	*row++
}
