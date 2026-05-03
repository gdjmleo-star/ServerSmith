package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"serversmith-agent/collector"
)

// AgentVersion is the current version of the ServerSmith agent
const AgentVersion = "1.0.0"

type Report struct {
	ServerID    string  `json:"server_id"`
	NetIn       int64   `json:"net_in"`
	NetOut      int64   `json:"net_out"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float64 `json:"mem_percent"`
	DiskPercent float64 `json:"disk_percent"`
	UptimeSec   int64   `json:"uptime_sec"`
}

func main() {
	server := flag.String("server", "http://localhost:18080", "ServerSmith management API URL")
	serverID := flag.String("server-id", "", "Server identifier (default: hostname)")
	interval := flag.Int("interval", 30, "Report interval in seconds")
	flag.Parse()

	if *serverID == "" {
		hostname, _ := os.Hostname()
		serverID = &hostname
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("ServerSmith Agent v%s starting...", AgentVersion)
	log.Printf("  Server: %s", *server)
	log.Printf("  ID:     %s", *serverID)
	log.Printf("  Interval: %ds", *interval)

	collector := collector.New()
	client := &http.Client{Timeout: 10 * time.Second}

	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	// Do a first report immediately
	doReport(client, *server, *serverID, collector)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			doReport(client, *server, *serverID, collector)
		case <-sigCh:
			log.Println("Agent shutting down")
			return
		}
	}
}

func doReport(client *http.Client, serverURL, serverID string, col *collector.Collector) {
	stats, err := col.Collect()
	if err != nil {
		log.Printf("Collect error: %v", err)
		return
	}

	report := Report{
		ServerID:    serverID,
		NetIn:       stats.NetIn,
		NetOut:      stats.NetOut,
		CPUPercent:  stats.CPUPercent,
		MemPercent:  stats.MemPercent,
		DiskPercent: stats.DiskPercent,
		UptimeSec:   0,
	}

	body, err := json.Marshal(report)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return
	}

	url := fmt.Sprintf("%s/api/report", serverURL)
	var lastErr error
	for retry := 0; retry < 3; retry++ {
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Agent-Version", AgentVersion)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Reported OK (%d bytes in, %d bytes out, CPU %.1f%%, Mem %.1f%%, Disk %.1f%%)",
				report.NetIn, report.NetOut, report.CPUPercent, report.MemPercent, report.DiskPercent)
			return
		}
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		break
	}

	log.Printf("Report failed after retries: %v", lastErr)
}
