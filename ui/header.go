package ui

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/host"
)

func RenderHeader(row *int) {
	MoveCursor(*row, 1)
	ClearLine()
	
	hostInfo, _ := host.Info()
	uptime := time.Duration(hostInfo.Uptime) * time.Second
	
	fmt.Printf("%swtop - %s%s                                    %sUptime: %v%s", 
		Bold+Cyan, hostInfo.Hostname, Reset, 
		Green, uptime.Truncate(time.Second), Reset)
	*row++
	
	MoveCursor(*row, 1)
	ClearLine()
	fmt.Printf("%s%s%s", Yellow, Repeat("=", 80), Reset)
	*row++
}
