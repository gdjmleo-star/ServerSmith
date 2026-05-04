package api

import (
	"net/http"
	"serversmith/db"
)

type LiveStatsItem struct {
	ServerID     int64    `json:"server_id"`
	ServerName   string   `json:"server_name"`
	Status       string   `json:"status"`
	CPUPercent   float64  `json:"cpu_percent"`
	MemPercent   float64  `json:"mem_percent"`
	DiskPercent  float64  `json:"disk_percent"`
	NetInRate    float64  `json:"net_in_rate"`
	NetOutRate   float64  `json:"net_out_rate"`
	NetInTotal   int64    `json:"net_in_total"`
	NetOutTotal  int64    `json:"net_out_total"`
	UsedBytes    int64    `json:"used_bytes"`
	UsedBytesIn  int64    `json:"used_bytes_in"`
	UsedBytesOut int64    `json:"used_bytes_out"`
	TotalQuotaGB float64  `json:"total_quota_gb"`
	UptimeSec    int64    `json:"uptime_sec"`
	ExpireAt     *string  `json:"expire_at"`
	LastReportAt *string  `json:"last_report_at"`
}

func HandleLiveStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	// Fetch latest snapshot per server using a subquery
	rows, err := db.DB.Query(`
		SELECT
			s.id,
			s.name,
			s.status,
			s.last_report_at,
			COALESCE(t.cpu_percent, 0),
			COALESCE(t.mem_percent, 0),
			COALESCE(t.disk_percent, 0),
			COALESCE(t.net_in, 0),
			COALESCE(t.net_out, 0),
			COALESCE(t.uptime_sec, 0),
			COALESCE(t2.net_in, 0),
			COALESCE(t2.net_out, 0),
			COALESCE(CAST((julianday(t.report_at) - julianday(t2.report_at)) * 86400 AS INTEGER), 0),
			COALESCE(sp.used_bytes, 0),
			COALESCE(sp.used_bytes_in, 0),
			COALESCE(sp.used_bytes_out, 0),
			COALESCE(sp.total_quota, 0),
			sp.expire_at
		FROM servers s
		LEFT JOIN traffic_snapshots t ON t.id = (
			SELECT id FROM traffic_snapshots WHERE server_id = s.id ORDER BY report_at DESC LIMIT 1
		)
		LEFT JOIN traffic_snapshots t2 ON t2.id = (
			SELECT id FROM traffic_snapshots WHERE server_id = s.id ORDER BY report_at DESC LIMIT 1 OFFSET 1
		)
		LEFT JOIN server_plans sp ON sp.server_id = s.id
		ORDER BY s.name
	`)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	result := make([]LiveStatsItem, 0)
	for rows.Next() {
		var item LiveStatsItem
		var prevNetIn, prevNetOut int64
		var diffSec int64

		if err := rows.Scan(
			&item.ServerID, &item.ServerName, &item.Status, &item.LastReportAt,
			&item.CPUPercent, &item.MemPercent, &item.DiskPercent,
			&item.NetInTotal, &item.NetOutTotal, &item.UptimeSec,
			&prevNetIn, &prevNetOut, &diffSec,
			&item.UsedBytes, &item.UsedBytesIn, &item.UsedBytesOut,
			&item.TotalQuotaGB, &item.ExpireAt,
		); err != nil {
			continue
		}

		// Calculate real-time rates (bytes/s)
		if diffSec > 0 {
			dIn := item.NetInTotal - prevNetIn
			dOut := item.NetOutTotal - prevNetOut
			if dIn < 0 { dIn = 0 }
			if dOut < 0 { dOut = 0 }
			item.NetInRate = float64(dIn) / float64(diffSec)
			item.NetOutRate = float64(dOut) / float64(diffSec)
		}

		result = append(result, item)
	}

	writeOK(w, result)
}
