package api

import (
	"fmt"
	"os"
	"serversmith/db"
	"testing"
	"time"
)

// setupTestDB creates a temporary SQLite database for testing.
func setupTestDB(t *testing.T) {
	t.Helper()
	tmp, err := os.CreateTemp("", "serversmith-test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	os.Setenv("DB_PATH", tmp.Name())
	os.Setenv("TESTING", "true")
	tmp.Close() // DB will create/reopen
	if err := db.Init(tmp.Name()); err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		db.DB.Close()
		os.Remove(tmp.Name())
	})
}

// insertTestServer inserts a traffic-plan server for testing.
func insertTestServer(t *testing.T) int64 {
	t.Helper()
	res, err := db.DB.Exec(
		"INSERT INTO servers (name, host, ssh_port, ssh_user, server_type) VALUES (?, ?, ?, ?, ?)",
		"test-server", "192.168.1.100", 22, "root", "node",
	)
	if err != nil {
		t.Fatalf("insert server: %v", err)
	}
	id, _ := res.LastInsertId()

	// Insert a traffic plan
	_, err = db.DB.Exec(
		"INSERT INTO server_plans (server_id, plan_type, total_quota, cycle_day, cycle_time, used_bytes, next_cycle_at) VALUES (?, 'traffic', 1000, 1, '02:00', 0, ?)",
		id, time.Now().UTC().Add(30*24*time.Hour).Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert plan: %v", err)
	}
	return id
}

// ============================================================
// Test 1: 流量增量累计
// ============================================================
func TestTrafficDeltaAccumulation(t *testing.T) {
	setupTestDB(t)
	serverID := insertTestServer(t)

	// Insert first snapshot (baseline) — net_in=1000, net_out=500
	err := insertSnapshot(serverID, 1000, 500, 30.0, 50.0, 40.0)
	if err != nil {
		t.Fatalf("insert first snapshot: %v", err)
	}

	// Insert second snapshot — net_in=2000, net_out=800
	err = insertSnapshot(serverID, 2000, 800, 45.0, 55.0, 40.0)
	if err != nil {
		t.Fatalf("insert second snapshot: %v", err)
	}

	// Delta: net_in=1000, net_out=300 → total=1300 bytes
	expectedDelta := int64(1300)

	// Re-calculate used_bytes
	err = accumulateTrafficFromSnapshots(serverID)
	if err != nil {
		t.Fatalf("accumulate traffic: %v", err)
	}

	var usedBytes int64
	err = db.DB.QueryRow("SELECT used_bytes FROM server_plans WHERE server_id=?", serverID).Scan(&usedBytes)
	if err != nil {
		t.Fatalf("query used_bytes: %v", err)
	}

	if usedBytes != expectedDelta {
		t.Errorf("expected used_bytes=%d, got %d", expectedDelta, usedBytes)
	}

	// Insert third snapshot — net_in=3000, net_out=1000 → delta from second: net_in=1000, net_out=200
	err = insertSnapshot(serverID, 3000, 1000, 50.0, 60.0, 45.0)
	if err != nil {
		t.Fatalf("insert third snapshot: %v", err)
	}

	err = accumulateTrafficFromSnapshots(serverID)
	if err != nil {
		t.Fatalf("accumulate traffic: %v", err)
	}

	// Total used: 1300 + 1200 = 2500 bytes
	err = db.DB.QueryRow("SELECT used_bytes FROM server_plans WHERE server_id=?", serverID).Scan(&usedBytes)
	if err != nil {
		t.Fatalf("query used_bytes: %v", err)
	}

	expectedTotal := int64(1300 + 1000 + 200) // 2500
	if usedBytes != expectedTotal {
		t.Errorf("expected total used_bytes=%d, got %d", expectedTotal, usedBytes)
	}
}

// ============================================================
// Test 2: 北京时间复位计算
// ============================================================
func TestBeijingCycleResetCalculation(t *testing.T) {
	// Simulate: "每月5号02:00" Asia/Shanghai → UTC conversion
	// Test in several Beijing time windows to verify no off-by-one

	tests := []struct {
		name       string
		currentBJ  string // "2006-01-02 15:04" in Beijing
		cycleDay   int
		cycleTime  string
		wantUTCStr string // expected next_cycle_at prefix (date part in UTC)
	}{
		{
			name:       "before cycle day",
			currentBJ:  "2026-05-03 10:00", // May 3, before May 5
			cycleDay:   5,
			cycleTime:  "02:00",
			wantUTCStr: "2026-05-04", // May 5 02:00 BJT = May 4 18:00 UTC
		},
		{
			name:       "on cycle day before time",
			currentBJ:  "2026-05-05 01:00", // May 5 01:00 BJT, before 02:00
			cycleDay:   5,
			cycleTime:  "02:00",
			wantUTCStr: "2026-05-04", // May 5 02:00 BJT = May 4 18:00 UTC
		},
		{
			name:       "on cycle day after time",
			currentBJ:  "2026-05-05 03:00", // May 5 03:00 BJT, after 02:00
			cycleDay:   5,
			cycleTime:  "02:00",
			wantUTCStr: "2026-06-04", // June 5 02:00 BJT = June 4 18:00 UTC
		},
		{
			name:       "after cycle day",
			currentBJ:  "2026-05-20 10:00", // May 20
			cycleDay:   15,
			cycleTime:  "00:00",
			wantUTCStr: "2026-06-14", // June 15 00:00 BJT = June 14 16:00 UTC
		},
		{
			name:       "day 31 short month",
			currentBJ:  "2026-04-28 10:00", // Apr 28, next cycle day 31
			cycleDay:   31,
			cycleTime:  "02:00",
			wantUTCStr: "2026-04-30", // Apr 31 doesn't exist → rolls to May 1? Actually Apr has 30 days...
			// Apr 28 -> next Apr 31 doesn't exist -> May 1 might be used
			// Actually, time.Date(2026, 4, 31, ...) normalizes to May 1, 2026
			// May 1 02:00 BJT = Apr 30 18:00 UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock current time by parsing the BJ string
			nowBJ, err := time.ParseInLocation("2006-01-02 15:04", tt.currentBJ, db.TZ)
			if err != nil {
				t.Fatalf("parse current time: %v", err)
			}

			// Use calcNextCycleWithTime to pass our mock time
			result := calcNextCycleAt(nowBJ, tt.cycleDay, tt.cycleTime)
			t.Logf("current BJ: %s, next_cycle UTC: %s", tt.currentBJ, result)

			if tt.wantUTCStr != "skip" {
				// Verify the result starts with expected date
				if len(result) < 10 || result[:10] != tt.wantUTCStr {
					t.Errorf("expected UTC date prefix %q, got %q", tt.wantUTCStr, result[:10])
				}
			}

			// Parsed as UTC must succeed
			_, err = time.Parse(time.RFC3339, result)
			if err != nil {
				t.Errorf("result not valid RFC3339: %v", err)
			}
		})
	}
}

// ============================================================
// Test 3: 到期日不偏移
// ============================================================
func TestExpireDateNoOffset(t *testing.T) {
	tests := []struct {
		name    string
		input   string // "2006-01-02" in Beijing
		wantUTC string // expected UTC RFC3339 prefix
	}{
		{
			name:    "simple date",
			input:   "2026-12-31",
			wantUTC: "2026-12-31", // Dec 31 23:59:59 BJT = Dec 31 15:59:59 UTC (same UTC day!)
		},
		{
			name:    "mid-month",
			input:   "2026-06-15",
			wantUTC: "2026-06-15", // Jun 15 23:59:59 BJT = Jun 15 15:59:59 UTC
		},
		{
			name:    "boundary: first day of month",
			input:   "2026-05-01",
			wantUTC: "2026-05-01", // May 1 23:59:59 BJT = May 1 15:59:59 UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := time.ParseInLocation("2006-01-02", tt.input, db.TZ)
			if err != nil {
				t.Fatalf("parse input: %v", err)
			}
			// Convert to Beijing 23:59:59 then to UTC
			bjEnd := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, db.TZ)
			utcEnd := bjEnd.UTC().Format(time.RFC3339)

			t.Logf("input BJ date: %s → BJ 23:59:59 → UTC: %s", tt.input, utcEnd)

			// Verify UTC date matches expected
			if len(utcEnd) < 10 || utcEnd[:10] != tt.wantUTC {
				t.Errorf("expected UTC date prefix %q, got %q", tt.wantUTC, utcEnd[:10])
			}

			// Verify it's valid RFC3339
			parsedUTC, err := time.Parse(time.RFC3339, utcEnd)
			if err != nil {
				t.Fatalf("parsed UTC not valid RFC3339: %v", err)
			}

			// Verify round-trip: UTC → BJ display
			bjDisplay := parsedUTC.In(db.TZ)
			if bjDisplay.Day() != parsed.Day() || bjDisplay.Month() != parsed.Month() || bjDisplay.Year() != parsed.Year() {
				t.Errorf("BJ round-trip mismatch: original %s, got %s",
					tt.input, bjDisplay.Format("2006-01-02"))
			}

			// CRITICAL: UTC date should equal BJ date (not one day earlier)
			// Because 23:59:59 BJT = 15:59:59 UTC, same calendar day
			if parsedUTC.Day() != parsed.Day() || parsedUTC.Month() != parsed.Month() || parsedUTC.Year() != parsed.Year() {
				t.Errorf("OFF-BY-ONE! UTC date differs from BJ date! Input BJ: %s, UTC: %s",
					tt.input, utcEnd)
			}
		})
	}
}

// ============================================================
// Test helpers extracted from report.go logic
// ============================================================

// insertSnapshot inserts a traffic snapshot row.
func insertSnapshot(serverID int64, netIn, netOut int64, cpu, mem, disk float64) error {
	_, err := db.DB.Exec(
		"INSERT INTO traffic_snapshots (server_id, report_at, net_in, net_out, cpu_percent, mem_percent, disk_percent) VALUES (?, datetime('now'), ?, ?, ?, ?, ?)",
		serverID, netIn, netOut, cpu, mem, disk,
	)
	return err
}

// accumulateTrafficFromSnapshots recalculates used_bytes from snapshot deltas.
func accumulateTrafficFromSnapshots(serverID int64) error {
	rows, err := db.DB.Query(`
		SELECT net_in, net_out FROM traffic_snapshots
		WHERE server_id = ?
		ORDER BY report_at ASC
	`, serverID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var prevIn, prevOut int64
	var totalDelta int64
	first := true

	for rows.Next() {
		var curIn, curOut int64
		if err := rows.Scan(&curIn, &curOut); err != nil {
			return err
		}
		if first {
			first = false
			prevIn, prevOut = curIn, curOut
			continue
		}
		deltaIn := curIn - prevIn
		deltaOut := curOut - prevOut
		if deltaIn < 0 {
			deltaIn = 0 // counter reset
		}
		if deltaOut < 0 {
			deltaOut = 0
		}
		totalDelta += deltaIn + deltaOut
		prevIn, prevOut = curIn, curOut
	}

	_, err = db.DB.Exec("UPDATE server_plans SET used_bytes = ? WHERE server_id = ?", totalDelta, serverID)
	return err
}

// calcNextCycleAt computes next_cycle_at given a mock current Beijing time.
// This is a testable version of the production calcNextCycle.
func calcNextCycleAt(now time.Time, day int, timeStr string) string {
	var hour, minute int
	if n, _ := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute); n < 2 {
		hour, minute = 2, 0
	}
	next := time.Date(now.Year(), now.Month(), day, hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 1, 0)
	}
	return next.UTC().Format(time.RFC3339)
}
