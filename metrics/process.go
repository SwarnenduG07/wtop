package metrics

import (
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/yourusername/wtop/types"
)

func GetCleanProcessName(name, cmdline string) string {
	if cmdline == "" {
		return name
	}
	
	if len(cmdline) > 30 {
		if strings.Contains(cmdline, "chrome.exe") {
			return "Google Chrome"
		}
		if strings.Contains(cmdline, "Code.exe") {
			return "Visual Studio Code"
		}
		if strings.Contains(cmdline, "slack.exe") {
			return "Slack"
		}
		if strings.Contains(cmdline, "discord.exe") || strings.Contains(cmdline, "Discord.exe") {
			return "Discord"
		}
		if strings.Contains(cmdline, "NVIDIA") || strings.Contains(cmdline, "nvidia") {
			if strings.Contains(cmdline, "nvcontainer") {
				return "NVIDIA Container"
			}
			return "NVIDIA Service"
		}
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
		
		parts := strings.Split(cmdline, "\\")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			filename = strings.Trim(filename, "\"")
			filename = strings.TrimSuffix(filename, ".exe")
			if filename != "" {
				return filename
			}
		}
	}
	
	return name
}

func GetProcessInfo(p *process.Process) *types.ProcessInfo {
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
		shrMem = memInfo.RSS / 4
	}
	
	memPercent, _ := p.MemoryPercent()
	
	status, _ := p.Status()
	statusStr := "S"
	if len(status) > 0 {
		statusStr = status[0]
	}
	
	cmdline, _ := p.Cmdline()
	cleanName := GetCleanProcessName(name, cmdline)
	
	threads, _ := p.NumThreads()
	createTime, _ := p.CreateTime()
	
	return &types.ProcessInfo{
		PID:        p.Pid,
		PPID:       ppid,
		Name:       cleanName,
		User:       username,
		Priority:   20,
		Nice:       0,
		CPUPercent: cpuPercent,
		Memory:     memory,
		MemPercent: memPercent,
		VirtMem:    virtMem,
		ResMem:     resMem,
		ShrMem:     shrMem,
		Status:     statusStr,
		Command:    cleanName,
		Threads:    threads,
		CreateTime: createTime,
	}
}

func GetTopProcesses(limit int) []*types.ProcessInfo {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}
	
	var processInfos []*types.ProcessInfo
	
	for _, p := range processes {
		if len(processInfos) >= 100 {
			break
		}
		info := GetProcessInfo(p)
		processInfos = append(processInfos, info)
	}
	
	sort.Slice(processInfos, func(i, j int) bool {
		return processInfos[i].CPUPercent > processInfos[j].CPUPercent
	})
	
	if len(processInfos) > limit {
		return processInfos[:limit]
	}
	return processInfos
}
