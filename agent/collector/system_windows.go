//go:build windows

package collector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func getCPUPercent() (float64, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-WmiObject Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average").Output()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return val, err
}

func getMemPercent() (float64, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"$os=Get-WmiObject Win32_OperatingSystem; [math]::Round(($os.TotalVisibleMemorySize-$os.FreePhysicalMemory)/$os.TotalVisibleMemorySize*100,1)").Output()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return val, err
}

func getDiskPercent() (float64, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"$d=Get-PSDrive C; [math]::Round($d.Used/($d.Used+$d.Free)*100,1)").Output()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return val, err
}

func getNetStats() (int64, int64, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"$a=Get-NetAdapterStatistics | Where-Object {$_.Name -notlike '*Loopback*'}; ($a | Measure-Object -Property ReceivedBytes -Sum).Sum, ($a | Measure-Object -Property SentBytes -Sum).Sum").Output()
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected output: %s", string(out))
	}
	in, _ := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
	out2, _ := strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64)
	return in, out2, nil
}

func getUptime() int64 {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-Date) - (gcim Win32_OperatingSystem).LastBootUpTime | Select-Object -ExpandProperty TotalSeconds").Output()
	if err != nil {
		return 0
	}
	val, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return int64(val)
}
