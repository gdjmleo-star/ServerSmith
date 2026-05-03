package api

import (
	"database/sql"
	"net/http"
	"serversmith/db"
	"strconv"
	"strings"
)

func HandleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	limit := "50"
	if l := r.URL.Query().Get("limit"); l != "" {
		limit = l
	}

	whereClause := ""
	ackFilter := r.URL.Query().Get("acknowledged")
	if ackFilter == "false" {
		whereClause = "AND a.acknowledged = 0"
	} else if ackFilter == "true" {
		whereClause = "AND a.acknowledged = 1"
	}

	query := `
		SELECT a.id, a.server_id, COALESCE(s.name, ''), a.alert_type, a.message, a.level,
		       a.created_at, a.acknowledged, a.acknowledged_at
		FROM alerts a
		LEFT JOIN servers s ON s.id = a.server_id
		WHERE 1=1 ` + whereClause + `
		ORDER BY a.created_at DESC
		LIMIT ?`

	rows, err := db.DB.Query(query, limit)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	var alerts []AlertResponse
	for rows.Next() {
		var ar AlertResponse
		var serverID sql.NullInt64
		var ackAt sql.NullString
		if err := rows.Scan(&ar.ID, &serverID, &ar.ServerName, &ar.AlertType,
			&ar.Message, &ar.Level, &ar.CreatedAt, &ar.Acknowledged, &ackAt); err != nil {
			continue
		}
		if serverID.Valid {
			ar.ServerID = &serverID.Int64
		}
		if ackAt.Valid {
			ar.AcknowledgedAt = &ackAt.String
		}
		alerts = append(alerts, ar)
	}

	if alerts == nil {
		alerts = []AlertResponse{}
	}
	writeOK(w, alerts)
}

func HandleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		writeError(w, 405, "method not allowed")
		return
	}

	// path: /api/alerts/:id/acknowledge
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, 400, "invalid path")
		return
	}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		writeError(w, 400, "invalid id")
		return
	}

	now := db.NowUTC()
	res, err := db.DB.Exec(
		"UPDATE alerts SET acknowledged=1, acknowledged_at=? WHERE id=?",
		now, id,
	)
	if err != nil {
		writeError(w, 500, "update error: "+err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, 404, "alert not found")
		return
	}
	writeOK(w, map[string]bool{"acknowledged": true})
}
