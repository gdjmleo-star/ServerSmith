package api

import (
	"encoding/json"
	"net/http"
	"serversmith/db"
	"strconv"
	"time"
)

type ServerRequest struct {
	Name        string   `json:"name"`
	Host        string   `json:"host"`
	SSHPort     int      `json:"ssh_port"`
	SSHUser     string   `json:"ssh_user"`
	ServerType  string   `json:"server_type"`
	CarrierType *string  `json:"carrier_type"`
	MonthlyRent *float64 `json:"monthly_rent"`
	PlanType    string   `json:"plan_type"`
	TotalQuota  *float64 `json:"total_quota"`
	InitialUsed *float64 `json:"initial_used_gb"`
	CycleDay    *int     `json:"cycle_day"`
	CycleTime   string   `json:"cycle_time"`
	ExpireAt    *string  `json:"expire_at"`
	ExpireNotifyDays *int `json:"expire_notify_days"`
}

type ServerResponse struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Host         string   `json:"host"`
	SSHPort      int      `json:"ssh_port"`
	SSHUser      string   `json:"ssh_user"`
	ServerType   string   `json:"server_type"`
	CarrierType  *string  `json:"carrier_type"`
	MonthlyRent  float64  `json:"monthly_rent"`
	Status       string   `json:"status"`
	LastReport   *string  `json:"last_report_at"`
	AgentVersion *string  `json:"agent_version"`
	CreatedAt    string   `json:"created_at"`
	PlanType     string   `json:"plan_type"`
	TotalQuota   *float64 `json:"total_quota"`
	UsedBytes    int64    `json:"used_bytes"`
	UsedGB       float64  `json:"used_gb"`
	RemainingGB  float64  `json:"remaining_gb"`
	UsagePct     float64  `json:"usage_percent"`
	NextCycle    *string  `json:"next_cycle_at"`
	ExpireAt     *string  `json:"expire_at"`
	CycleDay     *int     `json:"cycle_day"`
	CycleTime    *string  `json:"cycle_time"`
}

func HandleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		listServers(w, r)
	case "POST":
		createServer(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func HandleServerByID(w http.ResponseWriter, r *http.Request) {
	idStr := GetID(r.URL.Path, "/api/servers")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "invalid server id")
		return
	}
	switch r.Method {
	case "GET":
		getServer(w, id)
	case "PUT":
		updateServer(w, r, id)
	case "DELETE":
		deleteServer(w, id)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func listServers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(listServersSQL)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	servers := make([]ServerResponse, 0)
	for rows.Next() {
		sr := scanServerRow(rows)
		if sr == nil {
			continue
		}
		enrichServerUsage(sr)
		servers = append(servers, *sr)
	}
	writeOK(w, servers)
}

func createServer(w http.ResponseWriter, r *http.Request) {
	var req ServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body: "+err.Error())
		return
	}
	if req.Name == "" || req.Host == "" {
		writeError(w, 400, "name and host are required")
		return
	}
	if req.ServerType == "" {
		req.ServerType = "node"
	}
	planType := req.PlanType
	if planType == "" {
		planType = "no_limit"
	}
	now := db.NowUTC()
	sshPort := req.SSHPort
	if sshPort <= 0 {
		sshPort = 22
	}
	sshUser := req.SSHUser
	if sshUser == "" {
		sshUser = "root"
	}
	monthlyRent := float64(0)
	if req.MonthlyRent != nil {
		monthlyRent = *req.MonthlyRent
	}

	tx, err := db.DB.Begin()
	if err != nil {
		writeError(w, 500, "tx error: "+err.Error())
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(insertServerSQL,
		req.Name, req.Host, sshPort, sshUser,
		req.ServerType, req.CarrierType, monthlyRent, now, now)
	if err != nil {
		writeError(w, 500, "insert server error: "+err.Error())
		return
	}
	serverID, _ := res.LastInsertId()

	if planType == "traffic" {
		if err := createTrafficPlan(tx, serverID, &req); err != nil {
			writeError(w, 500, "insert traffic plan error: "+err.Error())
			return
		}
	} else {
		if err := createNoLimitPlan(tx, serverID, &req); err != nil {
			writeError(w, 500, "insert no_limit plan error: "+err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, 500, "commit error: "+err.Error())
		return
	}
	getServer(w, serverID)
}

func getServer(w http.ResponseWriter, id int64) {
	row := db.DB.QueryRow(getServerByIDSQL, id)
	sr := scanServerRow(row)
	if sr == nil {
		writeError(w, 404, "server not found")
		return
	}
	enrichServerUsage(sr)
	writeOK(w, *sr)
}

func updateServer(w http.ResponseWriter, r *http.Request, id int64) {
	var req ServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body: "+err.Error())
		return
	}

	now := db.NowUTC()
	// Read current server first
	var curName, curHost, curSrvType, curSSHUser string
	var curSSHPort int
	var curCarrier *string
	var curRent float64
	err := db.DB.QueryRow(
		"SELECT name, host, ssh_port, ssh_user, server_type, carrier_type, COALESCE(monthly_rent,0) FROM servers WHERE id=?", id,
	).Scan(&curName, &curHost, &curSSHPort, &curSSHUser, &curSrvType, &curCarrier, &curRent)
	if err != nil {
		writeError(w, 404, "server not found")
		return
	}

	// Only overwrite fields that were explicitly sent (non-zero for their zero value type)
	name := curName
	if req.Name != "" {
		name = req.Name
	}
	host := curHost
	if req.Host != "" {
		host = req.Host
	}
	sshPort := curSSHPort
	if req.SSHPort > 0 {
		sshPort = req.SSHPort
	}
	sshUser := curSSHUser
	if req.SSHUser != "" {
		sshUser = req.SSHUser
	}
	srvType := curSrvType
	if req.ServerType != "" {
		srvType = req.ServerType
	}
	carrierType := curCarrier
	if req.CarrierType != nil {
		if *req.CarrierType == "" {
			carrierType = nil
		} else {
			carrierType = req.CarrierType
		}
	}
	monthlyRent := curRent
	if req.MonthlyRent != nil {
		monthlyRent = *req.MonthlyRent
	}

	_, err = db.DB.Exec(
		"UPDATE servers SET name=?, host=?, ssh_port=?, ssh_user=?, server_type=?, carrier_type=?, monthly_rent=?, updated_at=? WHERE id=?",
		name, host, sshPort, sshUser, srvType, carrierType, monthlyRent, now, id,
	)
	if err != nil {
		writeError(w, 500, "update error: "+err.Error())
		return
	}

	// Handle plan_type change (e.g. traffic ↔ no_limit)
	if req.PlanType != "" {
		var curPlanType string
		_ = db.DB.QueryRow("SELECT plan_type FROM server_plans WHERE server_id=?", id).
			Scan(&curPlanType)
		if req.PlanType != curPlanType {
			tx, txErr := db.DB.Begin()
			if txErr != nil {
				writeError(w, 500, "tx error: "+txErr.Error())
				return
			}
			defer tx.Rollback()
			if _, err := tx.Exec("DELETE FROM server_plans WHERE server_id=?", id); err != nil {
				writeError(w, 500, "plan delete error: "+err.Error())
				return
			}
			planReq := req
			if req.PlanType == "traffic" {
				if err := createTrafficPlan(tx, id, &planReq); err != nil {
					writeError(w, 500, "create traffic plan: "+err.Error())
					return
				}
			} else {
				if err := createNoLimitPlan(tx, id, &planReq); err != nil {
					writeError(w, 500, "create no_limit plan: "+err.Error())
					return
				}
			}
			if err := tx.Commit(); err != nil {
				writeError(w, 500, "commit error: "+err.Error())
				return
			}
		} else if req.ExpireAt != nil {
			// Same plan type — update expire_at and expire_notify_days in-place
			var expireAt *string
			if *req.ExpireAt != "" {
				parsed, parseErr := time.ParseInLocation("2006-01-02", *req.ExpireAt, db.TZ)
				if parseErr == nil {
					bjEnd := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, db.TZ)
					utcStr := bjEnd.UTC().Format(time.RFC3339)
					expireAt = &utcStr
				} else {
					expireAt = req.ExpireAt
				}
			}
			notifyDays := 7
			if req.ExpireNotifyDays != nil {
				notifyDays = *req.ExpireNotifyDays
			}
			_, _ = db.DB.Exec(
				"UPDATE server_plans SET expire_at=?, expire_notify_days=? WHERE server_id=?",
				expireAt, notifyDays, id,
			)
		}
	}

	getServer(w, id)
}

func deleteServer(w http.ResponseWriter, id int64) {
	_, err := db.DB.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		writeError(w, 500, "delete error: "+err.Error())
		return
	}
	writeOK(w, map[string]bool{"deleted": true})
}
