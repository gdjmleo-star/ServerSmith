package collector

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"syscall"
	"time"
)

type SystemStats struct {
	NetIn       int64   `json:"net_in"`
	NetOut      int64   `json:"net_out"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float64 `json:"mem_percent"`
	DiskPercent float64 `json:"disk_percent"`
	UptimeSec   int64   `json:"uptime_sec"`
	Hostname    string  `json:"-"`
}

type Collector struct {
	prevNetIn  int64
	prevNetOut int64
	prevCPUTime int64
	prevCPUIdle int64
	hostname   string
}

func New() *Collector {
	hostname, _ := os.Hostname()
	return &Collector{
		hostname: hostname,
	}
}

func (c *Collector) Hostname() string {
	return c.hostname
}

func (c *Collector) Collect() (*SystemStats, error) {
	stats := &SystemStats{
		Hostname: c.hostname,
	}

	// CPU
	if cpu, err := getCPUPercent(); err == nil {
		stats.CPUPercent = math.Round(cpu*10) / 10
	} else {
		stats.CPUPercent = 0
	}

	// Memory
	if mem, err := getMemPercent(); err == nil {
		stats.MemPercent = math.Round(mem*10) / 10
	} else {
		stats.MemPercent = 0
	}

	// Disk
	if disk, err := getDiskPercent(); err == nil {
		stats.DiskPercent = math.Round(disk*10) / 10
	} else {
		stats.DiskPercent = 0
	}

	// Network
	netIn, netOut, err := getNetStats()
	if err == nil {
		stats.NetIn = netIn
		stats.NetOut = netOut
	}

	// Uptime
	stats.UptimeSec = getUptime()

	return stats, nil
}

func getCPUPercent() (float64, error) {
	// Simple approach: read /proc/stat on Linux
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
	var total, free, buffers, cached int64
	fmt.Sscanf(string(data), "MemTotal: %d kB\nMemFree: %d kB\nMemAvailable: %d kB",
		&total, &free, &buffers)
	if total == 0 {
		return 0, fmt.Errorf("cannot parse MemTotal")
	}
	// Use MemAvailable if possible (third field)
	if cached > 0 {
		available := cached
		return float64(total-available) / float64(total) * 100, nil
	}
	return float64(total-free) / float64(total) * 100, nil
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
	return float64(total-free) / float64(total) * 100, nil
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

		// Skip headers and loopback
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

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func contains(s string, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0)
}

// Placeholder for non-Linux systems
func init() {
	if runtime.GOOS != "linux" {
		fmt.Println("Warning: ServerSmith agent designed for Linux. Some metrics may be unavailable.")
	}
}
