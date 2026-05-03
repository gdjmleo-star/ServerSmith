package api

import (
	"encoding/json"
	"net/http"
	"serversmith/db"
	"sync"
)

type BatchRebootRequest struct {
	ServerIDs []int64 `json:"server_ids"`
}

type BatchRebootResult struct {
	ServerID int64  `json:"server_id"`
	Result   string `json:"result"`
	Error    string `json:"error,omitempty"`
}

func HandleBatchReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	var req BatchRebootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body: "+err.Error())
		return
	}
	if len(req.ServerIDs) == 0 {
		writeError(w, 400, "server_ids required")
		return
	}
	if len(req.ServerIDs) > 50 {
		writeError(w, 400, "max 50 servers per batch")
		return
	}

	type serverBasic struct {
		host    string
		sshPort int
		sshUser string
	}

	serverMap := make(map[int64]serverBasic)
	for _, id := range req.ServerIDs {
		row := db.DB.QueryRow("SELECT host, ssh_port, COALESCE(ssh_user,'root') FROM servers WHERE id=?", id)
		var s serverBasic
		if err := row.Scan(&s.host, &s.sshPort, &s.sshUser); err != nil {
			continue
		}
		serverMap[id] = s
	}

	results := make([]BatchRebootResult, len(req.ServerIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, id := range req.ServerIDs {
		wg.Add(1)
		go func(idx int, serverID int64) {
			defer wg.Done()
			res := BatchRebootResult{ServerID: serverID}

			s, ok := serverMap[serverID]
			if !ok {
				res.Result = "failed"
				res.Error = "server not found"
			} else {
				_, errMsg := sshReboot(s.host, s.sshPort, s.sshUser)
				if errMsg == "" {
					res.Result = "success"
					now := db.NowUTC()
					_, _ = db.DB.Exec(
						"INSERT INTO reboot_tasks (server_id, action, triggered_at, executed_at, result) VALUES (?,?,?,?,?)",
						serverID, "manual_reboot", now, now, "success",
					)
				} else {
					res.Result = "failed"
					res.Error = errMsg
					now := db.NowUTC()
					_, _ = db.DB.Exec(
						"INSERT INTO reboot_tasks (server_id, action, triggered_at, executed_at, result, error_msg) VALUES (?,?,?,?,?,?)",
						serverID, "manual_reboot", now, now, "failed", errMsg,
					)
				}
			}

			mu.Lock()
			results[idx] = res
			mu.Unlock()
		}(i, id)
	}

	wg.Wait()
	writeOK(w, results)
}
