package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"serversmith/db"
)

func HandleBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/serversmith.db"
	}

	backupDir := filepath.Dir(dbPath)
	timestamp := time.Now().UTC().Format("20060102_150405")
	backupName := fmt.Sprintf("serversmith_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Use SQLite VACUUM INTO for a safe, consistent backup while the server is running
	_, err := db.DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
	if err != nil {
		log.Printf("[backup] VACUUM INTO failed: %v", err)
		writeError(w, 500, "backup failed: "+err.Error())
		return
	}

	writeOK(w, map[string]string{
		"backup_path": backupPath,
		"backup_name": backupName,
		"timestamp":   timestamp,
	})
	log.Printf("[backup] Created: %s", backupPath)
}
