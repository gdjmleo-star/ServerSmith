package cron

import (
	"log"
	"serversmith/db"
	"time"
)

// StartOfflineCheck runs every 60 seconds, marks servers as offline if no report for 5 min
func StartOfflineCheck(stop chan struct{}) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Println("[cron] Offline checker started (every 60s)")

	for {
		select {
		case <-stop:
			log.Println("[cron] Offline checker stopped")
			return
		case <-ticker.C:
			checkOffline()
		}
	}
}

func checkOffline() {
	cutoff := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)

	rows, err := db.DB.Query(`
		SELECT id, name, host FROM servers
		WHERE status = 'online'
		  AND (last_report_at IS NULL OR last_report_at < ?)
	`, cutoff)
	if err != nil {
		log.Printf("[cron] offline check query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, host string
		if err := rows.Scan(&id, &name, &host); err != nil {
			continue
		}

		log.Printf("[cron] Server OFFLINE: %s (%s)", name, host)
		_, _ = db.DB.Exec("UPDATE servers SET status='offline', updated_at=? WHERE id=?", db.NowUTC(), id)

		now := db.NowUTC()
		_, _ = db.DB.Exec(
			"INSERT INTO alerts (server_id, alert_type, message, level, created_at) VALUES (?, 'offline', ?, 'critical', ?)",
			id, "Server offline: "+name, now,
		)
	}
}
