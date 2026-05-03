//go:build !windows

package collector

import (
	"fmt"
	"math"
	"os"
	"strings"
	"syscall"
	"time"
)

// readCPUStat reads the aggregate CPU jiffies from /proc/stat.
func readCPUStat() (total, idle int64, err error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	var user, nice, system, idleJ, iowait, irq, softirq, steal int64
	n, _ := fmt.Sscanf(string(data), "cpu %d %d %d %d %d %d %d %d",
		&user, &nice, &system, &idleJ, &iowait, &irq, &softirq, &steal)
	if n < 4 {
		return 0, 0, fmt.Errorf("cannot parse /proc/stat")
	}
	total = user + nice + system + idleJ + iowait + irq + softirq + steal
	idle = idleJ + iowait
	return total, idle, nil
}

// getCPUPercent samples /proc/stat twice with a 1-second gap and returns
// the real-time CPU usage percentage over that interval.
func getCPUPercent() (float64, error) {
	total1, idle1, err := readCPUStat()
	if err != nil {
		return 0, err
	}
	time.Sleep(1 * time.Second)
	total2, idle2, err := readCPUStat()
	if err != nil {
		return 0, err
	}
	dTotal := total2 - total1
	dIdle := idle2 - idle1
	if dTotal <= 0 {
		return 0, nil
	}
	return math.Round(float64(dTotal-dIdle)/float64(dTotal)*100*10) / 10, nil
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
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line) // strip leading spaces before interface name
		if line == "" || strings.HasPrefix(line, "Inter") || strings.HasPrefix(line, "face") {
			continue
		}
		if strings.HasPrefix(line, "lo:") {
			continue
		}
		var ifName string
		var in, out int64
		if n, _ := fmt.Sscanf(line, "%s %d %*d %*d %*d %*d %*d %*d %*d %d", &ifName, &in, &out); n >= 3 {
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
