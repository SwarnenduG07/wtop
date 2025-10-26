package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

type ProcessInfo struct {
	PID        int32
	PPID       int32
	Name       string
	User       string
	Priority   int32
	Nice       int32
	CPUPercent float64
	Memory     uint64
	MemPercent float32
	VirtMem    uint64
	ResMem     uint64
	ShrMem     uint64
	Status     string
	Command    string
	Threads    int32
	CreateTime int64
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func clearLine() {
	fmt.Print("\033[2K")
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func drawColorBar(percent float64, width int) string {
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}
	
	color := Green
	if percent > 80 {
		color = Red
	} else if percent > 60 {
		color = Yellow
	}
	
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += color + "█" + Reset
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %s%.1f%%%s", color, percent, Reset)
	return bar
}

func drawMemoryBar(percent float64, width int) string {
	filled := int(percent * float64(width) / 100)
	if filled > width {
		filled = width
	}
	
	// Memory-specific colors (more aggressive)
	color := Green
	if percent > 90 {
		color = Red + Bold
	} else if percent > 75 {
		color = Red
	} else if percent > 50 {
		color = Yellow
	}
	
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += color + "█" + Reset
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %s%.1f%%%s", color, percent, Reset)
	return bar
}

func updateHeader(row *int) {
	moveCursor(*row, 1)
	clearLine()
	
	hostInfo, _ := host.Info()
	uptime := time.Duration(hostInfo.Uptime) * time.Second
	
	fmt.Printf("%swtop - %s%s                                    %sUptime: %v%s", 
		Bold+Cyan, hostInfo.Hostname, Reset, 
		Green, uptime.Truncate(time.Second), Reset)
	*row++
	
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%s%s%s", Yellow, repeat("=", 80), Reset)
	*row++
}

func updateCPUBars(row *int) {
	cpuPercents, _ := cpu.Percent(0, true)
	cpuCount := len(cpuPercents)
	
	// CPU header with spacing
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%s%sCPU USAGE:%s", Bold+Cyan, " ", Reset)
	*row++
	
	// CPU cores in compact htop style with better spacing
	maxCores := cpuCount
	if maxCores > 16 { maxCores = 16 }
	
	for i := 0; i < maxCores; i += 4 {
		moveCursor(*row, 1)
		clearLine()
		
		for j := 0; j < 4 && (i+j) < maxCores; j++ {
			coreIdx := i + j
			if coreIdx < len(cpuPercents) {
				fmt.Printf("  %s%2d%s%s", Bold+Cyan, coreIdx+1, Reset, drawColorBar(cpuPercents[coreIdx], 12))
				if j < 3 && (i+j+1) < maxCores {
					fmt.Print("  ")
				}
			}
		}
		*row++
	}
	
	// Add spacing after CPU section
	*row++
	*row++
}

func updateMemoryBar(row *int) {
	v, err := mem.VirtualMemory()
	if err == nil {
		// Memory header
		moveCursor(*row, 1)
		clearLine()
		fmt.Printf("%s%sMEMORY & SWAP:%s", Bold+Purple, " ", Reset)
		*row++
		
		moveCursor(*row, 1)
		clearLine()
		fmt.Printf("%s  Mem:%s %s  %.1fG/%.1fG", 
			Bold+Purple, Reset,
			drawMemoryBar(v.UsedPercent, 35),
			float64(v.Used)/1024/1024/1024, 
			float64(v.Total)/1024/1024/1024)
		*row++
		
		// Swap
		swap, _ := mem.SwapMemory()
		if swap.Total > 0 {
			moveCursor(*row, 1)
			clearLine()
			fmt.Printf("%s  Swp:%s %s  %.1fG/%.1fG", 
				Bold+Purple, Reset,
				drawMemoryBar(swap.UsedPercent, 35),
				float64(swap.Used)/1024/1024/1024, 
				float64(swap.Total)/1024/1024/1024)
			*row++
		}
		
		// Add spacing after memory section
		*row++
	}
}

func updateSystemInfo(row *int) {
	// System info header
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%s%sSYSTEM INFO:%s", Bold+Green, " ", Reset)
	*row++
	
	// Task count
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
	
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("  %sTasks:%s %s%d%s, %s%d%s thr; %s%d%s running", 
		Bold+White, Reset,
		Green, totalTasks, Reset,
		Green, totalThreads, Reset,
		Green, runningTasks, Reset)
	*row++
	
	// Load average (Linux only)
	if runtime.GOOS != "windows" {
		loadAvg, err := load.Avg()
		if err == nil {
			moveCursor(*row, 1)
			clearLine()
			fmt.Printf("  %sLoad avg:%s %s%.2f %.2f %.2f%s", 
				Bold+Green, Reset,
				Green, loadAvg.Load1, loadAvg.Load5, loadAvg.Load15, Reset)
			*row++
		}
	}
	
	// Add spacing before process table
	*row++
	*row++
}

func getCleanProcessName(name, cmdline string) string {
	// If no command line, use process name
	if cmdline == "" {
		return name
	}
	
	// Extract service names from common Windows paths
	if len(cmdline) > 30 {
		// Chrome processes
		if strings.Contains(cmdline, "chrome.exe") {
			return "Google Chrome"
		}
		// VS Code
		if strings.Contains(cmdline, "Code.exe") {
			return "Visual Studio Code"
		}
		// Slack
		if strings.Contains(cmdline, "slack.exe") {
			return "Slack"
		}
		// Discord
		if strings.Contains(cmdline, "discord.exe") || strings.Contains(cmdline, "Discord.exe") {
			return "Discord"
		}
		// NVIDIA services
		if strings.Contains(cmdline, "NVIDIA") || strings.Contains(cmdline, "nvidia") {
			if strings.Contains(cmdline, "nvcontainer") {
				return "NVIDIA Container"
			}
			return "NVIDIA Service"
		}
		// Windows system processes
		if strings.Contains(cmdline, "sihost.exe") {
			return "Shell Infrastructure Host"
		}
		if strings.Contains(cmdline, "svchost.exe") {
			return "Service Host Process"
		}
		if strings.Contains(cmdline, "RuntimeBroker.exe") {
			return "Runtime Broker"
		}
		if strings.Contains(cmdline, "conhost.exe") {
			return "Console Window Host"
		}
		if strings.Contains(cmdline, "SystemApps") {
			return "Windows System App"
		}
		// Extract filename from path
		parts := strings.Split(cmdline, "\\")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			// Remove quotes and .exe extension
			filename = strings.Trim(filename, "\"")
			filename = strings.TrimSuffix(filename, ".exe")
			if filename != "" {
				return filename
			}
		}
	}
	
	// Fallback to process name
	return name
}

func getProcessInfo(p *process.Process) *ProcessInfo {
	name, _ := p.Name()
	if name == "" {
		name = "Unknown"
	}
	
	ppid, _ := p.Ppid()
	username, _ := p.Username()
	if username == "" {
		username = "N/A"
	}
	
	cpuPercent, _ := p.CPUPercent()
	
	memInfo, _ := p.MemoryInfo()
	var memory, virtMem, resMem, shrMem uint64 = 0, 0, 0, 0
	if memInfo != nil {
		memory = memInfo.RSS
		virtMem = memInfo.VMS
		resMem = memInfo.RSS
		shrMem = memInfo.RSS / 4 // Approximated shared memory
	}
	
	memPercent, _ := p.MemoryPercent()
	
	status, _ := p.Status()
	statusStr := "S"
	if len(status) > 0 {
		statusStr = status[0]
	}
	
	cmdline, _ := p.Cmdline()
	
	// Get clean service name
	cleanName := getCleanProcessName(name, cmdline)
	
	threads, _ := p.NumThreads()
	createTime, _ := p.CreateTime()
	
	// Priority and Nice (approximated for cross-platform)
	priority := int32(20)
	nice := int32(0)
	
	return &ProcessInfo{
		PID:        p.Pid,
		PPID:       ppid,
		Name:       cleanName, // Use clean name instead of raw name
		User:       username,
		Priority:   priority,
		Nice:       nice,
		CPUPercent: cpuPercent,
		Memory:     memory,
		MemPercent: memPercent,
		VirtMem:    virtMem,
		ResMem:     resMem,
		ShrMem:     shrMem,
		Status:     statusStr,
		Command:    cleanName, // Use clean name for command too
		Threads:    threads,
		CreateTime: createTime,
	}
}

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

func updateProcessTable(row *int) {
	processes, err := process.Processes()
	if err != nil {
		return
	}
	
	var processInfos []*ProcessInfo
	
	// Collect process info
	for _, p := range processes {
		if len(processInfos) >= 100 {
			break
		}
		
		info := getProcessInfo(p)
		processInfos = append(processInfos, info)
	}
	
	// Sort by CPU usage
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].CPUPercent > processInfos[j].CPUPercent
	})
	
	// Header
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%s  PID USER      PRI  NI    VIRT    RES    SHR S  %%CPU %%MEM     TIME+ SERVICE/PROCESS%s", Bold+White, Reset)
	*row++
	
	// Process list
	count := 0
	for _, info := range processInfos {
		if count >= 20 {
			break
		}
		
		moveCursor(*row, 1)
		clearLine()
		
		// Truncate fields
		user := info.User
		if len(user) > 9 {
			user = user[:9]
		}
		
		serviceName := info.Name
		if len(serviceName) > 20 {
			serviceName = serviceName[:17] + "..."
		}
		
		// Calculate runtime
		runtime := time.Now().Unix() - info.CreateTime/1000
		timeStr := formatTime(runtime)
		
		// Format memory in appropriate units
		virtStr := fmt.Sprintf("%.0fM", float64(info.VirtMem)/1024/1024)
		resStr := fmt.Sprintf("%.0fM", float64(info.ResMem)/1024/1024)
		shrStr := fmt.Sprintf("%.0fM", float64(info.ShrMem)/1024/1024)
		
		// All processes in green color
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
		count++
	}
}

func updateFooter(row *int) {
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%s%s%s", Yellow, repeat("=", 80), Reset)
	*row++
	
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%sF1Help F2Setup F3Search F4Filter F5Tree F6SortBy F7Nice F8Nice+ F9Kill F10Quit%s", Green, Reset)
	*row++
	
	moveCursor(*row, 1)
	clearLine()
	fmt.Printf("%sPress Ctrl+C to quit • Refreshing every 2 seconds%s", Cyan, Reset)
}

// String repeat helper
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func main() {
	hideCursor()
	defer showCursor()
	
	fmt.Printf("%sStarting wtop... Press Ctrl+C to exit%s\n", Green, Reset)
	time.Sleep(1 * time.Second)
	
	// Clear screen once
	fmt.Print("\033[2J\033[H")
	
	for {
		row := 1
		
		updateHeader(&row)
		updateCPUBars(&row)
		updateMemoryBar(&row)
		updateSystemInfo(&row)
		updateProcessTable(&row)
		updateFooter(&row)
		
		time.Sleep(2 * time.Second)
	}
}
