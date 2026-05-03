package api

import (
	"database/sql"
	"net/http"
	"serversmith/db"
)

type RebootTaskResponse struct {
	ID          int64   `json:"id"`
	Action      string  `json:"action"`
	TriggeredAt string  `json:"triggered_at"`
	ExecutedAt  *string `json:"executed_at"`
	Result      string  `json:"result"`
	ErrorMsg    *string `json:"error_msg"`
}

func HandleServerRebootTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	idStr := GetID(r.URL.Path, "/api/servers")
	if idStr == "" {
		writeError(w, 400, "invalid server id")
		return
	}

	limit := "5"
	if l := r.URL.Query().Get("limit"); l != "" {
		limit = l
	}

	rows, err := db.DB.Query(`
		SELECT id, action, triggered_at, executed_at, result, error_msg
		FROM reboot_tasks
		WHERE server_id = ?
		ORDER BY triggered_at DESC
		LIMIT ?
	`, idStr, limit)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	var tasks []RebootTaskResponse
	for rows.Next() {
		var t RebootTaskResponse
		var execAt, errMsg sql.NullString
		if err := rows.Scan(&t.ID, &t.Action, &t.TriggeredAt, &execAt, &t.Result, &errMsg); err != nil {
			continue
		}
		if execAt.Valid {
			t.ExecutedAt = &execAt.String
		}
		if errMsg.Valid {
			t.ErrorMsg = &errMsg.String
		}
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []RebootTaskResponse{}
	}
	writeOK(w, tasks)
}
