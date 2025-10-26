package main

import (
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	log.Println("Testing system metrics collection...")
	
	for i := 0; i < 3; i++ {
		// CPU
		percent, err := cpu.Percent(0, false)
		if err == nil && len(percent) > 0 {
			log.Printf("CPU: %.2f%%", percent[0])
		}

		// Memory
		v, err := mem.VirtualMemory()
		if err == nil {
			log.Printf("Memory: Used=%dMB, Total=%dMB (%.1f%%)", 
				v.Used/1024/1024, v.Total/1024/1024, v.UsedPercent)
		}

		time.Sleep(1 * time.Second)
	}
	
	log.Println("Test completed!")
}
