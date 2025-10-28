package ui

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/process"
)

func RenderSystemInfo(row *int, width int, limit int) {
	if *row > limit {
		return
	}
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s SYSTEM INFO:%s", Bold+Green, Reset)
	*row++
	if *row > limit {
		return
	}

	processes, _ := process.Processes()
	totalTasks := len(processes)
	runningTasks := 0
	totalThreads := 0

	for _, p := range processes {
		status, _ := p.Status()
		if len(status) > 0 && status[0] == "R" {
			runningTasks++
		}
		threads, _ := p.NumThreads()
		totalThreads += int(threads)
	}

	MoveCursor(*row, 1)
	ClearLine()
	infoText := fmt.Sprintf("Tasks:%d, %d thr; %d running", totalTasks, totalThreads, runningTasks)
	if width < 40 {
		infoText = fmt.Sprintf("Tasks:%d Run:%d", totalTasks, runningTasks)
	}
	available := width - 3
	if available < 1 {
		available = width
	}
	infoText = FitPlainString(infoText, available)
	line := fmt.Sprintf("  %s%s%s", Bold+White, infoText, Reset)
	fmt.Print(line)
	*row++
	if *row > limit {
		return
	}

	if runtime.GOOS != "windows" {
		loadAvg, err := load.Avg()
		if err == nil {
			MoveCursor(*row, 1)
			ClearLine()
			loadText := fmt.Sprintf("Load avg: %.2f %.2f %.2f", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
			if width < 40 {
				loadText = fmt.Sprintf("Load: %.2f %.2f", loadAvg.Load1, loadAvg.Load5)
			}
			available = width - 3
			if available < 1 {
				available = width
			}
			loadText = FitPlainString(loadText, available)
			line = fmt.Sprintf("  %s%s%s", Bold+Green, loadText, Reset)
			fmt.Print(line)
			*row++
			if *row > limit {
				return
			}
		}
	}

	if *row <= limit {
		*row++
	}
}
