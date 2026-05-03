package api

import (
	"database/sql"
	"fmt"
	"time"

	"serversmith/db"
)

// scanServerRow scans a row (or single-row query result) into a ServerResponse.
// Returns nil on scan error.
func scanServerRow(scanner interface {
	Scan(dest ...any) error
}) *ServerResponse {
	var sr ServerResponse
	var monthlyRent float64
	var nextCycle, expireAt, lastReport, cycleTime, agentVersion sql.NullString
	var cycleDay sql.NullInt64

	err := scanner.Scan(
		&sr.ID, &sr.Name, &sr.Host, &sr.SSHPort, &sr.SSHUser, &sr.ServerType,
		&sr.CarrierType, &monthlyRent, &sr.Status, &lastReport, &agentVersion, &sr.CreatedAt,
		&sr.PlanType, &sr.TotalQuota, &sr.UsedBytes, &nextCycle, &expireAt,
		&cycleDay, &cycleTime,
	)
	if err != nil {
		return nil
	}

	sr.MonthlyRent = monthlyRent
	if nextCycle.Valid {
		sr.NextCycle = &nextCycle.String
	}
	if expireAt.Valid {
		sr.ExpireAt = &expireAt.String
	}
	if lastReport.Valid {
		sr.LastReport = &lastReport.String
	}
	if agentVersion.Valid {
		sr.AgentVersion = &agentVersion.String
	}
	if cycleDay.Valid {
		v := int(cycleDay.Int64)
		sr.CycleDay = &v
	}
	if cycleTime.Valid {
		sr.CycleTime = &cycleTime.String
	}
	return &sr
}

// enrichServerUsage computes display fields (used_gb, remaining_gb, usage_percent).
func enrichServerUsage(sr *ServerResponse) {
	sr.UsedGB = float64(sr.UsedBytes) / 1_000_000_000
	if sr.TotalQuota != nil && *sr.TotalQuota > 0 {
		sr.RemainingGB = *sr.TotalQuota - sr.UsedGB
		if sr.RemainingGB < 0 {
			sr.RemainingGB = 0
		}
		sr.UsagePct = (sr.UsedGB / *sr.TotalQuota) * 100
	}
}

// calcNextCycle interprets cycle_day+cycle_time as Asia/Shanghai, then converts to UTC.
func calcNextCycle(day int, timeStr string) string {
	now := time.Now().In(db.TZ)
	var hour, minute int
	if n, _ := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute); n < 2 {
		hour, minute = 2, 0
	}

	next := time.Date(now.Year(), now.Month(), day, hour, minute, 0, 0, db.TZ)
	if !next.After(now) {
		next = next.AddDate(0, 1, 0)
	}
	return next.UTC().Format(time.RFC3339)
}

// createTrafficPlan inserts a traffic plan record for a node server.
func createTrafficPlan(tx *sql.Tx, serverID int64, req *ServerRequest) error {
	initialUsed := float64(0)
	if req.InitialUsed != nil {
		initialUsed = *req.InitialUsed
	}
	usedBytes := int64(initialUsed * 1_000_000_000)
	totalQuota := float64(0)
	if req.TotalQuota != nil {
		totalQuota = *req.TotalQuota
	}
	cycleDay := 1
	if req.CycleDay != nil {
		cycleDay = *req.CycleDay
	}
	cycleTime := req.CycleTime
	if cycleTime == "" {
		cycleTime = "02:00"
	}

	nextCycle := calcNextCycle(cycleDay, cycleTime)

	// expire_at applies to all plan types
	var expireAt *string
	if req.ExpireAt != nil && *req.ExpireAt != "" {
		parsed, parseErr := time.ParseInLocation("2006-01-02", *req.ExpireAt, db.TZ)
		if parseErr == nil {
			bjEnd := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, db.TZ)
			utcStr := bjEnd.UTC().Format(time.RFC3339)
			expireAt = &utcStr
		} else {
			expireAt = req.ExpireAt
		}
	}
	expireNotifyDays := 7
	if req.ExpireNotifyDays != nil {
		expireNotifyDays = *req.ExpireNotifyDays
	}

	_, err := tx.Exec(
		`INSERT INTO server_plans (server_id, plan_type, total_quota, initial_used_gb, cycle_day, cycle_time, used_bytes, next_cycle_at, expire_at, expire_notify_days)
		 VALUES (?, 'traffic', ?, ?, ?, ?, ?, ?, ?, ?)`,
		serverID, totalQuota, initialUsed, cycleDay, cycleTime, usedBytes, nextCycle, expireAt, expireNotifyDays,
	)
	return err
}

// createNoLimitPlan inserts a no_limit plan record for a project server.
func createNoLimitPlan(tx *sql.Tx, serverID int64, req *ServerRequest) error {
	var expireAt *string
	if req.ExpireAt != nil && *req.ExpireAt != "" {
		// expire_at input is Beijing date, convert to UTC end-of-day
		parsed, parseErr := time.ParseInLocation("2006-01-02", *req.ExpireAt, db.TZ)
		if parseErr == nil {
			bjEnd := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, db.TZ)
			utcStr := bjEnd.UTC().Format(time.RFC3339)
			expireAt = &utcStr
		} else {
			expireAt = req.ExpireAt
		}
	}
	_, err := tx.Exec(
		`INSERT INTO server_plans (server_id, plan_type, expire_at)
		 VALUES (?, 'no_limit', ?)`,
		serverID, expireAt,
	)
	return err
}
