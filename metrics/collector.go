package metrics

import (
	"context"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/process"
)

type SystemMetrics struct {
	CPUPercent    float64
	MemoryUsed    uint64
	MemoryTotal   uint64
	NetworkSent   uint64
	NetworkRecv   uint64
	DiskRead      uint64
	DiskWrite     uint64
}

func CollectSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// CPU
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	}

	// Memory
	vmem, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemoryUsed = vmem.Used
		metrics.MemoryTotal = vmem.Total
	}

	// Network
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		metrics.NetworkSent = netIO[0].BytesSent
		metrics.NetworkRecv = netIO[0].BytesRecv
	}

	// Disk
	diskIO, err := disk.IOCounters()
	if err == nil {
		for _, d := range diskIO {
			metrics.DiskRead += d.ReadBytes
			metrics.DiskWrite += d.WriteBytes
		}
	}

	return metrics, nil
}

func GetProcessList() ([]*process.Process, error) {
	return process.Processes()
}
