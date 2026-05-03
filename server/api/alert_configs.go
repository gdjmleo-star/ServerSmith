package api

import (
	"encoding/json"
	"net/http"
	"serversmith/db"
	"strconv"
)

type AlertConfigRequest struct {
	ServerID  int64  `json:"server_id"`
	AlertType string `json:"alert_type"`
	Enabled   bool   `json:"enabled"`
	Threshold string `json:"threshold"`
}

type AlertConfigResponse struct {
	ID        int64  `json:"id"`
	ServerID  int64  `json:"server_id"`
	AlertType string `json:"alert_type"`
	Enabled   bool   `json:"enabled"`
	Threshold string `json:"threshold"`
	UpdatedAt string `json:"updated_at"`
}

func HandleAlertConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getAlertConfigs(w, r)
	case "PUT":
		upsertAlertConfig(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func getAlertConfigs(w http.ResponseWriter, r *http.Request) {
	serverIDStr := r.URL.Query().Get("server_id")
	if serverIDStr == "" {
		writeError(w, 400, "server_id required")
		return
	}
	serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "invalid server_id")
		return
	}

	rows, err := db.DB.Query(
		"SELECT id, server_id, alert_type, enabled, threshold, updated_at FROM alert_configs WHERE server_id=?",
		serverID,
	)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	var configs []AlertConfigResponse
	for rows.Next() {
		var cfg AlertConfigResponse
		if err := rows.Scan(&cfg.ID, &cfg.ServerID, &cfg.AlertType, &cfg.Enabled, &cfg.Threshold, &cfg.UpdatedAt); err != nil {
			continue
		}
		configs = append(configs, cfg)
	}

	if configs == nil {
		configs = []AlertConfigResponse{}
	}
	writeOK(w, configs)
}

func upsertAlertConfig(w http.ResponseWriter, r *http.Request) {
	var req AlertConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body: "+err.Error())
		return
	}
	if req.ServerID == 0 {
		writeError(w, 400, "server_id required")
		return
	}
	if req.AlertType != "offline" && req.AlertType != "traffic_usage" {
		writeError(w, 400, "alert_type must be 'offline' or 'traffic_usage'")
		return
	}
	if req.Threshold == "" {
		writeError(w, 400, "threshold required")
		return
	}

	now := db.NowUTC()
	enabled := 0
	if req.Enabled {
		enabled = 1
	}

	_, err := db.DB.Exec(`
		INSERT INTO alert_configs (server_id, alert_type, enabled, threshold, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(server_id, alert_type) DO UPDATE SET
			enabled=excluded.enabled,
			threshold=excluded.threshold,
			updated_at=excluded.updated_at
	`, req.ServerID, req.AlertType, enabled, req.Threshold, now, now)
	if err != nil {
		writeError(w, 500, "upsert error: "+err.Error())
		return
	}

	row := db.DB.QueryRow(
		"SELECT id, server_id, alert_type, enabled, threshold, updated_at FROM alert_configs WHERE server_id=? AND alert_type=?",
		req.ServerID, req.AlertType,
	)
	var cfg AlertConfigResponse
	if err := row.Scan(&cfg.ID, &cfg.ServerID, &cfg.AlertType, &cfg.Enabled, &cfg.Threshold, &cfg.UpdatedAt); err != nil {
		writeError(w, 500, "fetch after upsert: "+err.Error())
		return
	}
	writeOK(w, cfg)
}
