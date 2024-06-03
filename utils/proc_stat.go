package utils

import (
	"fmt"
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v3/process"
)

func ProcCpuUsage() (float64, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, err
	}

	return proc.CPUPercent()
}

func ProcMemUsage() (int, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, err
	}
	info, err := proc.MemoryInfo()
	if err != nil {
		return 0, err
	}

	fmt.Println("meminfo: ", info)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tSys = %v MiB \n", bToMb(m.Sys))

	return 0, nil
}

func bToMb(b uint64) uint64 {
	return b / (1024 * 1024)
}
