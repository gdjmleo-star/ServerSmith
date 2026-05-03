package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"serversmith/db"
	"strconv"
	"strings"
	"time"
)

// HandleServerReboot handles POST /api/servers/:id/reboot
// Implements SSH reboot with 10-second timeout and failure recording.
func HandleServerReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/servers/"), "/reboot")
	idStr = strings.Split(idStr, "/")[0]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "invalid server id")
		return
	}

	var name, host, sshUser string
	var port int
	err = db.DB.QueryRow(
		"SELECT name, host, ssh_port, ssh_user FROM servers WHERE id = ?", id,
	).Scan(&name, &host, &port, &sshUser)
	if err != nil {
		writeError(w, 404, "server not found")
		return
	}

	if port <= 0 {
		port = 22
	}
	if sshUser == "" {
		sshUser = "root"
	}

	now := db.NowUTC()
	_, _ = db.DB.Exec(
		"INSERT INTO reboot_tasks (server_id, action, triggered_at, result) VALUES (?, 'manual_reboot', ?, 'pending')",
		id, now,
	)

	// Execute SSH reboot with 10-second timeout
	result, errMsg := sshReboot(host, port, sshUser)
	executedAt := db.NowUTC()

	_, _ = db.DB.Exec(
		"UPDATE reboot_tasks SET result=?, error_msg=?, executed_at=? WHERE server_id=? AND result='pending'",
		result, errMsg, executedAt, id,
	)

	if result == "failed" {
		log.Printf("[reboot] FAILED server=%s host=%s:%d err=%s", name, host, port, errMsg)
		writeOK(w, map[string]string{
			"status": "failed", "host": host, "error": errMsg,
		})
		return
	}

	log.Printf("[reboot] SUCCESS server=%s host=%s:%d", name, host, port)
	writeOK(w, map[string]string{
		"status": "success", "host": host,
	})
}

// sshReboot executes ssh reboot with 10-second timeout.
// Returns (result, errorMsg).
func sshReboot(host string, port int, user string) (string, string) {
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
			return "failed", "SSH timeout (10s)"
		}
		errMsg := strings.TrimSpace(string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "failed", errMsg
	}

	_ = output // reboot usually has no stdout
	return "success", ""
}

// SSHExec executes a command via SSH with timeout (for cron reboot usage).
func SSHExec(host string, port int, user string, cmdStr string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", user, host),
		cmdStr,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), err
	}
	return strings.TrimSpace(string(output)), nil
}
