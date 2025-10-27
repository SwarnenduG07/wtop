package metrics

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type GPUInfo struct {
	Index          int
	Name           string
	Driver         string
	Utilization    float64
	MemoryUsed     float64
	MemoryTotal    float64
	Temperature    float64
	PowerUsage     float64
	PowerLimit     float64
	FanSpeed       float64
	FanRPM         int
	ClockCore      int
	ClockMemory    int
}

type GPUProcess struct {
	PID         int
	ProcessName string
	MemoryUsed  float64
}

func getNvidiaSmiCmd() string {
	if runtime.GOOS == "windows" {
		return "nvidia-smi.exe"
	}
	return "nvidia-smi"
}

func estimateFanRPM(temp float64, power float64) int {
	// Estimate fan RPM based on temperature and power
	// Typical laptop GPU fan: 2000-5000 RPM range
	if temp < 40 {
		return 0 // Fan off or very low
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
	
	// Try to get fan speed from WMI
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

func GetGPUInfo() ([]*GPUInfo, error) {
	cmd := exec.Command(getNvidiaSmiCmd(),
		"--query-gpu=index,name,driver_version,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw,power.limit,fan.speed,clocks.gr,clocks.mem",
		"--format=csv,noheader,nounits")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpus []*GPUInfo
	
	// Try to get fan speed from WMI for laptop GPUs
	wmiFanSpeed := getFanSpeedFromWMI()
	
	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) < 12 {
			continue
		}
		
		temp := parseFloat(fields[6])
		power := parseFloat(fields[7])
		fanSpeed := parseFloat(fields[9])
		
		// If fan speed is 0 or N/A, use WMI or estimate
		fanRPM := 0
		if fanSpeed == 0 {
			if wmiFanSpeed > 0 {
				fanRPM = int(wmiFanSpeed)
				fanSpeed = (wmiFanSpeed / 5000) * 100 // Estimate percentage
			} else {
				fanRPM = estimateFanRPM(temp, power)
				fanSpeed = float64(fanRPM) / 50 // Convert RPM to approximate %
			}
		} else {
			// Calculate RPM from percentage (assume max 5000 RPM)
			fanRPM = int(fanSpeed * 50)
		}
		
		gpu := &GPUInfo{
			Index:        parseInt(fields[0]),
			Name:         strings.TrimSpace(fields[1]),
			Driver:       strings.TrimSpace(fields[2]),
			Utilization:  parseFloat(fields[3]),
			MemoryUsed:   parseFloat(fields[4]),
			MemoryTotal:  parseFloat(fields[5]),
			Temperature:  temp,
			PowerUsage:   power,
			PowerLimit:   parseFloat(fields[8]),
			FanSpeed:     fanSpeed,
			FanRPM:       fanRPM,
			ClockCore:    parseInt(fields[10]),
			ClockMemory:  parseInt(fields[11]),
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
	
	output, err := cmd.Output()
	if err != nil {
		return []*GPUProcess{}, nil
	}
	
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
	if s == "[N/A]" || s == "" || s == "[Not Supported]" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "[N/A]" || s == "" || s == "[Not Supported]" {
		return 0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
