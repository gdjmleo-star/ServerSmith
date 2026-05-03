package cron

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"serversmith/db"
	"strings"
	"time"
)

// StartCycleReset runs every 5 minutes, checks for nodes that need cycle reset
func StartCycleReset(stop chan struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("[cron] Cycle reset checker started (every 5 min)")

	for {
		select {
		case <-stop:
			log.Println("[cron] Cycle reset checker stopped")
			return
		case <-ticker.C:
			checkAndReset()
		}
	}
}

func checkAndReset() {
	now := db.NowUTC()

	rows, err := db.DB.Query(`
		SELECT p.id, p.server_id, p.cycle_day, p.cycle_time, s.host, s.ssh_port, s.ssh_user
		FROM server_plans p
		JOIN servers s ON s.id = p.server_id
		WHERE p.plan_type = 'traffic'
		  AND p.next_cycle_at IS NOT NULL
		  AND p.next_cycle_at <= ?
	`, now)
	if err != nil {
		log.Printf("[cron] cycle reset query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var planID, serverID int64
		var cycleDay int
		var cycleTime, host, sshUser string
		var sshPort int

		if err := rows.Scan(&planID, &serverID, &cycleDay, &cycleTime, &host, &sshPort, &sshUser); err != nil {
			log.Printf("[cron] scan error: %v", err)
			continue
		}

		log.Printf("[cron] Resetting cycle for server %d (host: %s)", serverID, host)

		// Record last cycle end
		// The current used_bytes is the usage for the completed period
		// We can log it, then zero out

		// Calculate next cycle date
		nextCycle := calcNextCycle(cycleDay, cycleTime)

		_, err := db.DB.Exec(`
			UPDATE server_plans
			SET used_bytes = 0, last_cycle_end = ?, next_cycle_at = ?
			WHERE id = ?
		`, now, nextCycle, planID)
		if err != nil {
			log.Printf("[cron] reset error for server %d: %v", serverID, err)
			continue
		}

		// Record reboot task
		_, _ = db.DB.Exec(
			"INSERT INTO reboot_tasks (server_id, action, triggered_at, result) VALUES (?, 'cycle_reboot', ?, 'pending')",
			serverID, now,
		)

		// SSH reboot
		result := "success"
		errMsg := ""
		if err := sshReboot(host, sshPort, sshUser); err != nil {
			result = "failed"
			errMsg = err.Error()
			log.Printf("[cron] SSH reboot failed for %s: %v", host, err)
		}

		executedAt := db.NowUTC()
		_, _ = db.DB.Exec(
			"UPDATE reboot_tasks SET result=?, error_msg=?, executed_at=? WHERE server_id=? AND result='pending'",
			result, errMsg, executedAt, serverID,
		)

		_ = fmt.Sprintf("%d %s", planID, host) // suppress unused
	}
}

func calcNextCycle(day int, timeStr string) string {
	// Cycle day/time are interpreted as Asia/Shanghai
	now := time.Now().In(db.TZ)
	var hour, minute int
	fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)

	next := time.Date(now.Year(), now.Month(), day, hour, minute, 0, 0, db.TZ)
	if !next.After(now) {
		next = next.AddDate(0, 1, 0)
	}

	// Store as UTC
	return next.UTC().Format(time.RFC3339)
}

func sshReboot(host string, port int, user string) error {
	if user == "" {
		user = "root"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", user, host),
		"reboot",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SSH timeout after 10s: %s:%d", host, port)
		}
		return fmt.Errorf("SSH failed: %s - %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
