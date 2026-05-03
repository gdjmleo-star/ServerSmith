//go:build !windows

package collector

import (
	"fmt"
	"math"
	"os"
	"syscall"
)

func getCPUPercent() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}
	var user, nice, system, idle, iowait, irq, softirq, steal int64
	n, _ := fmt.Sscanf(string(data), "cpu %d %d %d %d %d %d %d %d",
		&user, &nice, &system, &idle, &iowait, &irq, &softirq, &steal)
	if n < 4 {
		return 0, fmt.Errorf("cannot parse /proc/stat")
	}
	total := user + nice + system + idle + iowait + irq + softirq + steal
	idleTotal := idle + iowait
	if total == 0 {
		return 0, nil
	}
	return float64(total-idleTotal) / float64(total) * 100, nil
}

func getMemPercent() (float64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	var total, free, available int64
	fmt.Sscanf(string(data), "MemTotal: %d kB\nMemFree: %d kB\nMemAvailable: %d kB",
		&total, &free, &available)
	if total == 0 {
		return 0, fmt.Errorf("cannot parse MemTotal")
	}
	if available > 0 {
		return math.Round(float64(total-available)/float64(total)*100*10) / 10, nil
	}
	return math.Round(float64(total-free)/float64(total)*100*10) / 10, nil
}

func getDiskPercent() (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	if total == 0 {
		return 0, nil
	}
	return math.Round(float64(total-free)/float64(total)*100*10) / 10, nil
}

func getNetStats() (int64, int64, error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	var totalIn, totalOut int64
	lines := string(data)
	for len(lines) > 0 {
		var line string
		idx := indexByte(lines, '\n')
		if idx < 0 {
			line = lines
			lines = ""
		} else {
			line = lines[:idx]
			lines = lines[idx+1:]
		}
		if line == "" || line[0] == ' ' || contains(line, "lo:") {
			continue
		}
		var ifName string
		var in, out int64
		if n, _ := fmt.Sscanf(line, "%s %d %*d %*d %*d %*d %*d %*d %d", &ifName, &in, &out); n >= 3 {
			totalIn += in
			totalOut += out
		}
	}
	return totalIn, totalOut, nil
}

func getUptime() int64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	var uptime float64
	fmt.Sscanf(string(data), "%f", &uptime)
	return int64(uptime)
}
