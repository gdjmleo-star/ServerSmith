package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"serversmith/db"
)

// --- User management handlers ---

type UserInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleUsers: GET /api/users → list all users
// POST /api/users → create new user
func HandleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		listUsers(w, r)
	case "POST":
		createUser(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

// HandleUserByID: DELETE /api/users/{id} → delete user
func HandleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		writeError(w, 405, "method not allowed")
		return
	}

	// Extract ID from path: /api/users/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/users/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, 400, "missing user id")
		return
	}

	// Get caller username from context
	callerUsername, _ := r.Context().Value(userContextKey).(string)

	// Prevent self-deletion
	var targetUsername string
	err := db.DB.QueryRow("SELECT username FROM users WHERE id = ?", parts[0]).Scan(&targetUsername)
	if err != nil {
		writeError(w, 404, "user not found")
		return
	}
	if targetUsername == callerUsername {
		writeError(w, 400, "cannot delete your own account")
		return
	}

	// Prevent deleting the last user
	var count int
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count <= 1 {
		writeError(w, 400, "cannot delete the last admin account")
		return
	}

	_, err = db.DB.Exec("DELETE FROM users WHERE id = ?", parts[0])
	if err != nil {
		writeError(w, 500, "delete failed: "+err.Error())
		return
	}
	writeOK(w, map[string]bool{"ok": true})
}

// HandleChangePassword: POST /api/users/change-password
func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	callerUsername, _ := r.Context().Value(userContextKey).(string)

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, 400, "current_password and new_password required")
		return
	}
	if len(req.NewPassword) < 4 {
		writeError(w, 400, "new password must be at least 4 characters")
		return
	}

	// Verify current password
	var storedHash string
	err := db.DB.QueryRow("SELECT password_hash FROM users WHERE username = ?", callerUsername).Scan(&storedHash)
	if err != nil {
		writeError(w, 404, "user not found")
		return
	}
	if hashPassword(req.CurrentPassword) != storedHash {
		writeError(w, 401, "current password is incorrect")
		return
	}

	// Update password
	newHash := hashPassword(req.NewPassword)
	_, err = db.DB.Exec("UPDATE users SET password_hash = ? WHERE username = ?", newHash, callerUsername)
	if err != nil {
		writeError(w, 500, "update failed: "+err.Error())
		return
	}
	writeOK(w, map[string]bool{"ok": true})
}

func listUsers(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.DB.Query("SELECT id, username, created_at FROM users ORDER BY id")
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	users := make([]UserInfo, 0)
	for rows.Next() {
		var u UserInfo
		if err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}
	writeOK(w, users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeError(w, 400, "username and password required")
		return
	}
	if len(req.Password) < 4 {
		writeError(w, 400, "password must be at least 4 characters")
		return
	}

	// Check duplicate
	var count int
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
	if count > 0 {
		writeError(w, 409, "username already exists")
		return
	}

	hash := hashPassword(req.Password)
	res, err := db.DB.Exec(
		"INSERT INTO users (username, password_hash, created_at) VALUES (?, ?, ?)",
		req.Username, hash, time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		writeError(w, 500, "create failed: "+err.Error())
		return
	}
	id, _ := res.LastInsertId()
	writeOK(w, UserInfo{ID: id, Username: req.Username})
}
