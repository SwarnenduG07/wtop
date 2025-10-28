package metrics

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type GPUInfo struct {
	Index              int
	Name               string
	Driver             string
	Utilization        float64
	MemoryUsed         float64
	MemoryTotal        float64
	Temperature        float64
	PowerUsage         float64
	PowerLimit         float64
	FanSpeed           float64
	FanRPM             int
	ClockCore          int
	ClockMemory        int
	ClockSM            int
	PerformanceState   string
	ThrottleReasons    []string
	MemoryUtilization  float64
	PCIeGen            int
	PCIeWidth          int
	ComputeMode        string
	MemoryBusWidth     int
	PowerState         string
	TempSlowdown       float64
}

type GPUProcess struct {
	PID         int
	ProcessName string
	MemoryUsed  float64
	Type        string
}

func getNvidiaSmiCmd() string {
	if runtime.GOOS == "windows" {
		return "nvidia-smi.exe"
	}
	return "nvidia-smi"
}

func estimateFanRPM(temp float64, power float64) int {
	if temp < 40 {
		return 0
	} else if temp < 50 {
		return 1500 + int(power*20)
	} else if temp < 60 {
		return 2000 + int(power*30)
	} else if temp < 70 {
		return 2500 + int(power*40)
	} else if temp < 80 {
		return 3500 + int(power*50)
	} else {
		return 4500 + int(power*60)
	}
}

func getFanSpeedFromWMI() float64 {
	if runtime.GOOS != "windows" {
		return 0
	}
	
	cmd := exec.Command("wmic", "path", "Win32_Fan", "get", "DesiredSpeed", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") {
			fields := strings.Split(line, ",")
			if len(fields) >= 2 {
				speed := parseFloat(fields[1])
				if speed > 0 {
					return speed
				}
			}
		}
	}
	
	return 0
}

func parseThrottleReasons(value string) []string {
	reasons := []string{}
	val := parseInt(value)
	
	if val == 0 {
		return []string{"None"}
	}
	
	if val&0x01 != 0 {
		reasons = append(reasons, "GPU Idle")
	}
	if val&0x02 != 0 {
		reasons = append(reasons, "App Clocks")
	}
	if val&0x04 != 0 {
		reasons = append(reasons, "SW Power Cap")
	}
	if val&0x08 != 0 {
		reasons = append(reasons, "HW Slowdown")
	}
	if val&0x10 != 0 {
		reasons = append(reasons, "Sync Boost")
	}
	if val&0x20 != 0 {
		reasons = append(reasons, "SW Thermal")
	}
	if val&0x40 != 0 {
		reasons = append(reasons, "HW Thermal")
	}
	if val&0x80 != 0 {
		reasons = append(reasons, "HW Power Brake")
	}
	if val&0x100 != 0 {
		reasons = append(reasons, "Display Clock")
	}
	
	if len(reasons) == 0 {
		reasons = append(reasons, "Unknown")
	}
	
	return reasons
}

func GetGPUInfo() ([]*GPUInfo, error) {
	cmd := exec.Command(getNvidiaSmiCmd(),
		"--query-gpu=index,name,driver_version,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw,power.limit,fan.speed,clocks.gr,clocks.mem,clocks.sm,pstate,clocks_throttle_reasons.active,utilization.memory,pcie.link.gen.current,pcie.link.width.current,compute_mode,temperature.gpu.tlimit",
		"--format=csv,noheader,nounits")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpus []*GPUInfo
	
	wmiFanSpeed := getFanSpeedFromWMI()
	
	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) < 20 {
			continue
		}
		
		temp := parseFloat(fields[6])
		power := parseFloat(fields[7])
		fanSpeed := parseFloat(fields[9])
		
		fanRPM := 0
		if fanSpeed == 0 {
			if wmiFanSpeed > 0 {
				fanRPM = int(wmiFanSpeed)
				fanSpeed = (wmiFanSpeed / 5000) * 100
			} else {
				fanRPM = estimateFanRPM(temp, power)
				fanSpeed = float64(fanRPM) / 50
			}
		} else {
			fanRPM = int(fanSpeed * 50)
		}
		
		pstate := strings.TrimSpace(fields[13])
		throttleReasons := parseThrottleReasons(fields[14])
		
		pcieGen := parseInt(fields[16])
		pcieWidth := parseInt(fields[17])
		memBusWidth := pcieWidth * 32
		
		gpu := &GPUInfo{
			Index:             parseInt(fields[0]),
			Name:              strings.TrimSpace(fields[1]),
			Driver:            strings.TrimSpace(fields[2]),
			Utilization:       parseFloat(fields[3]),
			MemoryUsed:        parseFloat(fields[4]),
			MemoryTotal:       parseFloat(fields[5]),
			Temperature:       temp,
			PowerUsage:        power,
			PowerLimit:        parseFloat(fields[8]),
			FanSpeed:          fanSpeed,
			FanRPM:            fanRPM,
			ClockCore:         parseInt(fields[10]),
			ClockMemory:       parseInt(fields[11]),
			ClockSM:           parseInt(fields[12]),
			PerformanceState:  pstate,
			ThrottleReasons:   throttleReasons,
			MemoryUtilization: parseFloat(fields[15]),
			PCIeGen:           pcieGen,
			PCIeWidth:         pcieWidth,
			PowerState:        pstate,
			MemoryBusWidth:    memBusWidth,
			ComputeMode:       strings.TrimSpace(fields[18]),
			TempSlowdown:      parseFloat(fields[19]),
		}
		
		gpus = append(gpus, gpu)
	}
	
	return gpus, nil
}

func GetGPUProcesses(gpuIndex int) ([]*GPUProcess, error) {
	cmd := exec.Command(getNvidiaSmiCmd(),
		"--query-compute-apps=pid,process_name,used_memory",
		"--format=csv,noheader,nounits",
		"-i", strconv.Itoa(gpuIndex))
	
	output, _ := cmd.Output()
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var processes []*GPUProcess
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) < 3 {
			continue
		}
		
		proc := &GPUProcess{
			PID:         parseInt(fields[0]),
			ProcessName: strings.TrimSpace(fields[1]),
			MemoryUsed:  parseFloat(fields[2]),
			Type:        "Compute",
		}
		processes = append(processes, proc)
	}
	
	return processes, nil
}

func GetIntelGPU() (*GPUInfo, error) {
	if runtime.GOOS != "windows" {
		return nil, nil
	}
	
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name,AdapterRAM", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "intel") {
			fields := strings.Split(line, ",")
			if len(fields) >= 2 {
				return &GPUInfo{
					Index: 99,
					Name:  strings.TrimSpace(fields[len(fields)-1]),
				}, nil
			}
		}
	}
	
	return nil, nil
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "[N/A]" || s == "" || s == "[Not Supported]" || s == "[Unknown Error]" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "[N/A]" || s == "" || s == "[Not Supported]" || s == "[Unknown Error]" {
		return 0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
