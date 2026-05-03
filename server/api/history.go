// Register the HandleServerHistory function for dashboard
// This is placed here to avoid circular imports while keeping the dashboard file clean

package api

import (
	"net/http"
	"serversmith/db"
	"strconv"
	"strings"
	"time"
)

type HistoryPoint struct {
	Date    string  `json:"date"`
	BytesIn int64   `json:"bytes_in"`
	BytesOut int64  `json:"bytes_out"`
	AvgCPU  float64 `json:"avg_cpu"`
	AvgMem  float64 `json:"avg_mem"`
}

func HandleServerHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/history")
	parts := strings.Split(path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "invalid server id")
		return
	}

	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)

	rows, err := db.DB.Query(`
		SELECT report_at, net_in, net_out, cpu_percent, mem_percent
		FROM traffic_snapshots
		WHERE server_id = ? AND report_at >= ?
		ORDER BY report_at ASC
	`, id, cutoff)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	var history []HistoryPoint
	for rows.Next() {
		var hp HistoryPoint
		if err := rows.Scan(&hp.Date, &hp.BytesIn, &hp.BytesOut, &hp.AvgCPU, &hp.AvgMem); err != nil {
			continue
		}
		history = append(history, hp)
	}

	if history == nil {
		history = []HistoryPoint{}
	}
	writeOK(w, history)
}
