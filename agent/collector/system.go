package collector

import (
	"fmt"
	"math"
	"os"
	"runtime"
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
	prevNetIn   int64
	prevNetOut  int64
	hostname    string
}

func New() *Collector {
	hostname, _ := os.Hostname()
	return &Collector{hostname: hostname}
}

func (c *Collector) Hostname() string {
	return c.hostname
}

func (c *Collector) Collect() (*SystemStats, error) {
	stats := &SystemStats{Hostname: c.hostname}

	if cpu, err := getCPUPercent(); err == nil {
		stats.CPUPercent = math.Round(cpu*10) / 10
	}
	if mem, err := getMemPercent(); err == nil {
		stats.MemPercent = math.Round(mem*10) / 10
	}
	if disk, err := getDiskPercent(); err == nil {
		stats.DiskPercent = math.Round(disk*10) / 10
	}
	if netIn, netOut, err := getNetStats(); err == nil {
		stats.NetIn = netIn
		stats.NetOut = netOut
	}
	stats.UptimeSec = getUptime()

	return stats, nil
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
	if len(sub) == 0 || len(s) < len(sub) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func init() {
	if runtime.GOOS != "linux" {
		fmt.Printf("[agent] Running on %s — some metrics use platform-specific implementation\n", runtime.GOOS)
	}
}
