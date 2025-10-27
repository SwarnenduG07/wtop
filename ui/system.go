package ui

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/process"
)

func RenderSystemInfo(row *int) {
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s%sSYSTEM INFO:%s", Bold+Green, " ", Reset)
	*row++
	
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
	fmt.Printf("  %sTasks:%s %s%d%s, %s%d%s thr; %s%d%s running", 
		Bold+White, Reset,
		Green, totalTasks, Reset,
		Green, totalThreads, Reset,
		Green, runningTasks, Reset)
	*row++
	
	if runtime.GOOS != "windows" {
		loadAvg, err := load.Avg()
		if err == nil {
			MoveCursor(*row, 1)
			ClearLine()
			fmt.Printf("  %sLoad avg:%s %s%.2f %.2f %.2f%s", 
				Bold+Green, Reset,
				Green, loadAvg.Load1, loadAvg.Load5, loadAvg.Load15, Reset)
			*row++
		}
	}
	
	*row++
	*row++
}
