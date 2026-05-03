package cron

import (
	"log"
	"serversmith/db"
	"time"
)

// StartExpiryNotify runs daily at 08:00, checks for expiring servers
func StartExpiryNotify(stop chan struct{}) {
	log.Println("[cron] Expiry notifier started (runs daily at 08:00)")

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	checkExpiry()

	for {
		select {
		case <-stop:
			log.Println("[cron] Expiry notifier stopped")
			return
		case <-ticker.C:
			bjNow := time.Now().In(db.TZ)
			if bjNow.Hour() == 8 && bjNow.Minute() < 5 {
				checkExpiry()
			}
		}
	}
}

func checkExpiry() {
	rows, err := db.DB.Query(`
		SELECT s.id, s.name, p.expire_at, COALESCE(p.expire_notify_days, 7)
		FROM servers s
		JOIN server_plans p ON p.server_id = s.id
		WHERE p.expire_at IS NOT NULL
	`)
	if err != nil {
		log.Printf("[cron] expiry query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, expireAt string
		var notifyDays int64
		if err := rows.Scan(&id, &name, &expireAt, &notifyDays); err != nil {
			continue
		}

		expireTime, err := time.Parse(time.RFC3339, expireAt)
		if err != nil {
			continue
		}

		daysUntilExpiry := time.Until(expireTime).Hours() / 24

		// Skip servers that are not within the notification window
		if daysUntilExpiry > float64(notifyDays) || daysUntilExpiry < 0 {
			continue
		}

		level := "warning"
		if daysUntilExpiry <= 3 {
			level = "critical"
		}

		_, _ = db.DB.Exec(
			"INSERT INTO alerts (server_id, alert_type, message, level, created_at) VALUES (?, 'expiring', ?, ?, ?)",
			id, "Server expiring: "+name+" on "+expireAt, level, db.NowUTC(),
		)
		log.Printf("[cron] Expiry alert: %s expires %s (notify_days=%d)", name, expireAt, notifyDays)
	}
}
