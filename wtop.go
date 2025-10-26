package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID        int32
	Name       string
	CPUPercent float64
	Memory     uint64
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func displayHeader() {
	fmt.Printf("wtop - System Monitor                                           %s\n", time.Now().Format("15:04:05"))
	fmt.Println("================================================================================")
}

func displayCPU() {
	cpuPercent, err := cpu.Percent(0, false)
	cpuCount, _ := cpu.Counts(true)
	
	if err == nil && len(cpuPercent) > 0 {
		fmt.Printf("CPU: %.1f%% (%d cores)\n", cpuPercent[0], cpuCount)
	} else {
		fmt.Printf("CPU: N/A (%d cores)\n", cpuCount)
	}
}

func displayMemory() {
	v, err := mem.VirtualMemory()
	if err == nil {
		fmt.Printf("Memory: %.1f GB / %.1f GB (%.1f%%)\n", 
			float64(v.Used)/1024/1024/1024, 
			float64(v.Total)/1024/1024/1024, 
			v.UsedPercent)
	} else {
		fmt.Println("Memory: N/A")
	}
}

func displayDisk() {
	var diskPath string
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	} else {
		diskPath = "/"
	}
	
	usage, err := disk.Usage(diskPath)
	if err == nil {
		fmt.Printf("Disk (%s): %.1f GB / %.1f GB (%.1f%%)\n",
			diskPath,
			float64(usage.Used)/1024/1024/1024,
			float64(usage.Total)/1024/1024/1024,
			usage.UsedPercent)
	} else {
		fmt.Printf("Disk (%s): N/A\n", diskPath)
	}
}

func displayNetwork() {
	stats, err := net.IOCounters(false)
	if err == nil && len(stats) > 0 {
		fmt.Printf("Network: ↑ %.1f MB ↓ %.1f MB\n",
			float64(stats[0].BytesSent)/1024/1024,
			float64(stats[0].BytesRecv)/1024/1024)
	} else {
		fmt.Println("Network: N/A")
	}
}

func getProcessInfo(p *process.Process) *ProcessInfo {
	name, nameErr := p.Name()
	if nameErr != nil {
		name = "Unknown"
	}
	
	cpuPercent, cpuErr := p.CPUPercent()
	if cpuErr != nil {
		cpuPercent = 0
	}
	
	memInfo, memErr := p.MemoryInfo()
	var memory uint64 = 0
	if memErr == nil && memInfo != nil {
		memory = memInfo.RSS
	}
	
	return &ProcessInfo{
		PID:        p.Pid,
		Name:       name,
		CPUPercent: cpuPercent,
		Memory:     memory,
	}
}

func displayProcesses() {
	processes, err := process.Processes()
	if err != nil {
		fmt.Println("\nProcesses: N/A")
		return
	}
	
	var processInfos []*ProcessInfo
	
	// Collect process info
	for _, p := range processes {
		if len(processInfos) >= 50 { // Limit to avoid too much processing
			break
		}
		
		info := getProcessInfo(p)
		if info.CPUPercent > 0 || info.Memory > 0 {
			processInfos = append(processInfos, info)
		}
	}
	
	// Sort by CPU usage
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].CPUPercent > processInfos[j].CPUPercent
	})
	
	fmt.Printf("\nTop Processes by CPU:\n")
	fmt.Printf("%-8s %-25s %-8s %-10s\n", "PID", "Name", "CPU%", "Memory")
	fmt.Println("------------------------------------------------------------------------")
	
	count := 0
	for _, info := range processInfos {
		if count >= 15 { // Show top 15
			break
		}
		
		// Truncate long process names
		name := info.Name
		if len(name) > 25 {
			name = name[:22] + "..."
		}
		
		fmt.Printf("%-8d %-25s %-8.1f %-10.1f MB\n", 
			info.PID, 
			name, 
			info.CPUPercent, 
			float64(info.Memory)/1024/1024)
		count++
	}
}

func main() {
	fmt.Println("Starting wtop... Press Ctrl+C to exit")
	time.Sleep(1 * time.Second)
	
	for {
		clearScreen()
		displayHeader()
		displayCPU()
		displayMemory()
		displayDisk()
		displayNetwork()
		displayProcesses()
		
		fmt.Println("\nRefreshing every 3 seconds... Press Ctrl+C to exit")
		time.Sleep(3 * time.Second)
	}
}
