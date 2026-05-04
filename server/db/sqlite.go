package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	if dbPath == "" {
		dbPath = "/data/serversmith.db"
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	// WAL supports concurrent readers; allow multiple conns to prevent cron blocking HTTP handlers
	DB.SetMaxOpenConns(5)
	DB.SetMaxIdleConns(3)
	DB.SetConnMaxLifetime(0)

	// Enable WAL mode
	if _, err := DB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := DB.Exec("PRAGMA busy_timeout=10000"); err != nil {
		return fmt.Errorf("set busy timeout: %w", err)
	}
	if _, err := DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := runMigrations(); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	log.Printf("SQLite initialized: %s (WAL mode)", dbPath)
	return nil
}

func runMigrations() error {
	schema := `
	CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		ssh_port INTEGER DEFAULT 22,
		ssh_user TEXT DEFAULT 'root',
		server_type TEXT NOT NULL DEFAULT 'node',
		carrier_type TEXT,
		monthly_rent REAL,
		status TEXT DEFAULT 'offline',
		last_report_at DATETIME,
		agent_version TEXT,
		created_at DATETIME DEFAULT (datetime('now')),
		updated_at DATETIME DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS server_plans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
		plan_type TEXT NOT NULL,
		total_quota REAL,
		initial_used_gb REAL DEFAULT 0,
		cycle_day INTEGER,
		cycle_time TEXT DEFAULT '02:00',
		last_cycle_end DATETIME,
		next_cycle_at DATETIME,
		used_bytes INTEGER DEFAULT 0,
		expire_at DATETIME,
		expire_notify_days INTEGER DEFAULT 7
	);

	CREATE TABLE IF NOT EXISTS traffic_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
		report_at DATETIME NOT NULL,
		net_in INTEGER NOT NULL DEFAULT 0,
		net_out INTEGER NOT NULL DEFAULT 0,
		cpu_percent REAL NOT NULL DEFAULT 0,
		mem_percent REAL NOT NULL DEFAULT 0,
		disk_percent REAL NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS traffic_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
		date DATE NOT NULL,
		bytes_in INTEGER NOT NULL DEFAULT 0,
		bytes_out INTEGER NOT NULL DEFAULT 0,
		avg_cpu REAL NOT NULL DEFAULT 0,
		avg_mem REAL NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS reboot_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
		action TEXT NOT NULL,
		triggered_at DATETIME NOT NULL DEFAULT (datetime('now')),
		executed_at DATETIME,
		result TEXT DEFAULT 'pending',
		error_msg TEXT
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER REFERENCES servers(id) ON DELETE SET NULL,
		alert_type TEXT NOT NULL,
		message TEXT NOT NULL,
		level TEXT NOT NULL DEFAULT 'info',
		created_at DATETIME DEFAULT (datetime('now')),
		acknowledged INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_snapshots_server_report ON traffic_snapshots(server_id, report_at);
	CREATE INDEX IF NOT EXISTS idx_snapshots_report_at ON traffic_snapshots(report_at);
	CREATE INDEX IF NOT EXISTS idx_history_server_date ON traffic_history(server_id, date);
	CREATE INDEX IF NOT EXISTS idx_reboot_server ON reboot_tasks(server_id);
	CREATE INDEX IF NOT EXISTS idx_alerts_server ON alerts(server_id);
	CREATE INDEX IF NOT EXISTS idx_alerts_ack ON alerts(acknowledged);
	CREATE INDEX IF NOT EXISTS idx_plans_server ON server_plans(server_id);

	CREATE TABLE IF NOT EXISTS alert_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
		alert_type TEXT NOT NULL,
		enabled INTEGER DEFAULT 1,
		threshold TEXT NOT NULL,
		created_at DATETIME DEFAULT (datetime('now')),
		updated_at DATETIME DEFAULT (datetime('now')),
		UNIQUE(server_id, alert_type)
	);

	CREATE INDEX IF NOT EXISTS idx_alert_configs_server ON alert_configs(server_id);
	CREATE INDEX IF NOT EXISTS idx_plans_next_cycle ON server_plans(next_cycle_at);
	CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status);
	CREATE INDEX IF NOT EXISTS idx_servers_last_report ON servers(last_report_at);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	`

	if _, err := DB.Exec(schema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	// Migration: add ssh_user column for existing databases
	_, _ = DB.Exec("ALTER TABLE servers ADD COLUMN ssh_user TEXT DEFAULT 'root'")
	// Migration: add acknowledged fields to alerts
	_, _ = DB.Exec("ALTER TABLE alerts ADD COLUMN acknowledged INTEGER DEFAULT 0")
	_, _ = DB.Exec("ALTER TABLE alerts ADD COLUMN acknowledged_at DATETIME")
	// Migration: agent version tracking
	_, _ = DB.Exec("ALTER TABLE servers ADD COLUMN agent_version TEXT")	// Migration: expire_notify_days (added in P4, for existing databases without the column)
	_, _ = DB.Exec("ALTER TABLE server_plans ADD COLUMN expire_notify_days INTEGER DEFAULT 7")
	// Migration P9: split used_bytes into in/out
	_, _ = DB.Exec("ALTER TABLE server_plans ADD COLUMN used_bytes_in INTEGER DEFAULT 0")
	_, _ = DB.Exec("ALTER TABLE server_plans ADD COLUMN used_bytes_out INTEGER DEFAULT 0")
	_, _ = DB.Exec("UPDATE server_plans SET used_bytes_in = used_bytes / 2, used_bytes_out = used_bytes - used_bytes / 2 WHERE used_bytes > 0 AND used_bytes_in = 0 AND used_bytes_out = 0")
	// Migration P9: add uptime_sec to snapshots
	_, _ = DB.Exec("ALTER TABLE traffic_snapshots ADD COLUMN uptime_sec INTEGER DEFAULT 0")

	log.Println("Database schema initialized")
	return nil
}


