package ui

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/SwarnenduG07/wtop/metrics"
	"github.com/SwarnenduG07/wtop/types"
)

type processSummary struct {
	Total   int
	Running int
	Threads int
}

type snapshot struct {
	Timestamp time.Time

	Hostname string
	Uptime   time.Duration

	Load1        float64
	Load5        float64
	Load15       float64
	LoadReported bool

	CPUPerCore []float64
	TotalCPU   float64

	Memory *mem.VirtualMemoryStat
	Swap   *mem.SwapMemoryStat

	DiskPath string
	Disk     *disk.UsageStat

	GPUInfos     []*metrics.GPUInfo
	GPUProcesses map[int][]*metrics.GPUProcess

	ProcessSummary processSummary
	Processes      []*types.ProcessInfo

	NetBytesSent uint64
	NetBytesRecv uint64
}

func collectSnapshot(limit int) (*snapshot, error) {
	snap := &snapshot{Timestamp: time.Now()}

	if hostInfo, err := host.Info(); err == nil && hostInfo != nil {
		snap.Hostname = hostInfo.Hostname
		snap.Uptime = time.Duration(hostInfo.Uptime) * time.Second
	}

	if avg, err := load.Avg(); err == nil {
		snap.Load1, snap.Load5, snap.Load15 = avg.Load1, avg.Load5, avg.Load15
		snap.LoadReported = true
	}

	if totals, err := cpu.Percent(0, false); err == nil && len(totals) > 0 {
		snap.TotalCPU = totals[0]
	}
	if perCore, err := cpu.Percent(0, true); err == nil {
		snap.CPUPerCore = perCore
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		snap.Memory = vm
	}
	if sw, err := mem.SwapMemory(); err == nil {
		snap.Swap = sw
	}

	path := primaryDiskPath()
	snap.DiskPath = path
	if usage, err := disk.Usage(path); err == nil {
		snap.Disk = usage
	}

	if counters, err := gnet.IOCounters(false); err == nil && len(counters) > 0 {
		snap.NetBytesSent = counters[0].BytesSent
		snap.NetBytesRecv = counters[0].BytesRecv
	}

	if processes := metrics.GetTopProcesses(limit); len(processes) > 0 {
		snap.Processes = processes
	}

	snap.ProcessSummary = collectProcessSummary()

	if gpus, err := metrics.GetGPUInfo(); err == nil {
		snap.GPUInfos = gpus
		if len(gpus) > 0 {
			snap.GPUProcesses = make(map[int][]*metrics.GPUProcess)
			for _, gpu := range gpus {
				procs, err := metrics.GetGPUProcesses(gpu.Index)
				if err != nil || len(procs) == 0 {
					continue
				}
				if len(procs) > 5 {
					procs = procs[:5]
				}
				snap.GPUProcesses[gpu.Index] = procs
			}
		}
	}

	return snap, nil
}

func collectProcessSummary() processSummary {
	processes, err := process.Processes()
	if err != nil {
		return processSummary{}
	}
	var summary processSummary
	for _, p := range processes {
		summary.Total++
		if status, err := p.Status(); err == nil && len(status) > 0 {
			if status[0] == "R" {
				summary.Running++
			}
		}
		if threads, err := p.NumThreads(); err == nil {
			summary.Threads += int(threads)
		}
	}
	return summary
}

func primaryDiskPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}
