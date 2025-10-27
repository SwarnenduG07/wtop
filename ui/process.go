package ui

import (
	"fmt"
	"time"

	"github.com/yourusername/wtop/metrics"
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

func RenderProcessTable(row *int) {
	processInfos := metrics.GetTopProcesses(20)
	
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s  PID USER      PRI  NI    VIRT    RES    SHR S  %%CPU %%MEM     TIME+ SERVICE/PROCESS%s", Bold+White, Reset)
	*row++
	
	for _, info := range processInfos {
		MoveCursor(*row, 1)
		ClearLine()
		
		user := info.User
		if len(user) > 9 {
			user = user[:9]
		}
		
		serviceName := info.Name
		if len(serviceName) > 20 {
			serviceName = serviceName[:17] + "..."
		}
		
		runtime := time.Now().Unix() - info.CreateTime/1000
		timeStr := formatTime(runtime)
		
		virtStr := fmt.Sprintf("%.0fM", float64(info.VirtMem)/1024/1024)
		resStr := fmt.Sprintf("%.0fM", float64(info.ResMem)/1024/1024)
		shrStr := fmt.Sprintf("%.0fM", float64(info.ShrMem)/1024/1024)
		
		fmt.Printf("%s%5d %-9s %3d %3d %7s %7s %7s %s %5.1f %4.1f %9s %-20s%s",
			Green,
			info.PID,
			user,
			info.Priority,
			info.Nice,
			virtStr,
			resStr,
			shrStr,
			info.Status,
			info.CPUPercent,
			info.MemPercent,
			timeStr,
			serviceName,
			Reset)
		
		*row++
	}
}
