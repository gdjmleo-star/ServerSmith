package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"serversmith/db"
)

type ReportRequest struct {
	ServerID    string  `json:"server_id"`
	NetIn       int64   `json:"net_in"`
	NetOut      int64   `json:"net_out"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float64 `json:"mem_percent"`
	DiskPercent float64 `json:"disk_percent"`
	UptimeSec   int64   `json:"uptime_sec"`
}

type ReportResponse struct {
	OK          bool   `json:"ok"`
	NextCycleAt string `json:"next_cycle_at,omitempty"`
}

func HandleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid request body: "+err.Error())
		return
	}

	if req.ServerID == "" {
		writeError(w, 400, "server_id is required")
		return
	}

	now := db.NowUTC()

	agentVersion := r.Header.Get("X-Agent-Version")

	// Find server by host (server_id = IP/hostname)
	var serverID int64
	err := db.DB.QueryRow("SELECT id FROM servers WHERE host = ?", req.ServerID).Scan(&serverID)
	if err != nil {
		log.Printf("Server %s not found, auto-registering", req.ServerID)
		res, err := db.DB.Exec(
			"INSERT INTO servers (name, host, status, last_report_at, created_at, updated_at) VALUES (?, ?, 'online', ?, ?, ?)",
			req.ServerID, req.ServerID, now, now, now,
		)
		if err != nil {
			writeError(w, 500, "failed to register server: "+err.Error())
			return
		}
		serverID, _ = res.LastInsertId()
		_, _ = db.DB.Exec(
			"INSERT INTO server_plans (server_id, plan_type) VALUES (?, 'no_limit')",
			serverID,
		)
	} else {
		updateSQL := "UPDATE servers SET status = 'online', last_report_at = ?, updated_at = ? WHERE id = ?"
		updateArgs := []interface{}{now, now, serverID}
		if agentVersion != "" {
			updateSQL = "UPDATE servers SET status = 'online', last_report_at = ?, updated_at = ?, agent_version = ? WHERE id = ?"
			updateArgs = []interface{}{now, now, agentVersion, serverID}
		}
		_, _ = db.DB.Exec(updateSQL, updateArgs...)
	}

	// Get last snapshot for delta
	var lastNetIn, lastNetOut int64
	row := db.DB.QueryRow(
		"SELECT net_in, net_out FROM traffic_snapshots WHERE server_id = ? ORDER BY report_at DESC LIMIT 1",
		serverID,
	)
	_ = row.Scan(&lastNetIn, &lastNetOut)

	// Insert new snapshot
	_, err = db.DB.Exec(
		`INSERT INTO traffic_snapshots (server_id, report_at, net_in, net_out, cpu_percent, mem_percent, disk_percent)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		serverID, now, req.NetIn, req.NetOut, req.CPUPercent, req.MemPercent, req.DiskPercent,
	)
	if err != nil {
		log.Printf("insert snapshot error: %v", err)
	}

	// Calculate delta and accumulate
	if lastNetIn > 0 || lastNetOut > 0 {
		deltaIn := req.NetIn - lastNetIn
		deltaOut := req.NetOut - lastNetOut
		if deltaIn < 0 {
			deltaIn = req.NetIn
		}
		if deltaOut < 0 {
			deltaOut = req.NetOut
		}
		deltaBytes := deltaIn + deltaOut

		if deltaBytes > 0 {
			_, _ = db.DB.Exec(
				"UPDATE server_plans SET used_bytes = used_bytes + ? WHERE server_id = ? AND plan_type = 'traffic'",
				deltaBytes, serverID,
			)

			var usedBytes int64
			var totalQuota float64
			row := db.DB.QueryRow(
				"SELECT used_bytes, total_quota FROM server_plans WHERE server_id = ? AND plan_type = 'traffic'",
				serverID,
			)
			if err := row.Scan(&usedBytes, &totalQuota); err == nil && totalQuota > 0 {
				usedGB := float64(usedBytes) / 1_000_000_000
				usagePercent := (usedGB / totalQuota) * 100
				if usagePercent >= 100 {
					writeAlert(serverID, "traffic_exceeded",
						fmt.Sprintf("Traffic quota exceeded: %.1f/%.0f GB", usedGB, totalQuota), "critical")
				} else if usagePercent >= 90 {
					writeAlert(serverID, "traffic_exceeded",
						fmt.Sprintf("Traffic at %.0f%%: %.1f/%.0f GB", usagePercent, usedGB, totalQuota), "warning")
				}
			}
		}
	}

	// Check system thresholds
	if req.CPUPercent > 85 {
		writeAlert(serverID, "high_cpu", fmt.Sprintf("CPU at %.0f%%", req.CPUPercent), "warning")
	}
	if req.MemPercent > 90 {
		writeAlert(serverID, "high_mem", fmt.Sprintf("Memory at %.0f%%", req.MemPercent), "warning")
	}
	if req.DiskPercent > 90 {
		writeAlert(serverID, "high_disk", fmt.Sprintf("Disk at %.0f%%", req.DiskPercent), "warning")
	}

	// Get next_cycle_at
	var nextCycleAt string
	row2 := db.DB.QueryRow(
		"SELECT COALESCE(next_cycle_at, '') FROM server_plans WHERE server_id = ? AND plan_type = 'traffic'",
		serverID,
	)
	_ = row2.Scan(&nextCycleAt)

	writeOK(w, ReportResponse{
		OK:          true,
		NextCycleAt: nextCycleAt,
	})
}

func writeAlert(serverID int64, alertType, message, level string) {
	now := db.NowUTC()
	_, err := db.DB.Exec(
		"INSERT INTO alerts (server_id, alert_type, message, level, created_at) VALUES (?, ?, ?, ?, ?)",
		serverID, alertType, message, level, now,
	)
	if err != nil {
		log.Printf("write alert error: %v", err)
	}
}
