package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"serversmith/db"

	"github.com/golang-jwt/jwt/v5"
)

// --- JWT ---

var jwtSecret []byte

func getJWTSecret() []byte {
	if jwtSecret != nil {
		return jwtSecret
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-not-for-production"
		log.Println("[auth] WARNING: JWT_SECRET not set, using development secret")
	}
	jwtSecret = []byte(secret)
	return jwtSecret
}

type contextKey string

const userContextKey contextKey = "user"

// Claims carries the JWT payload
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func generateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func validateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// --- Password hashing ---

func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

// --- Default admin config ---

func defaultAdminCreds() (string, string) {
	user := os.Getenv("ADMIN_USERNAME")
	if user == "" {
		user = "admin"
	}
	pass := os.Getenv("ADMIN_PASSWORD")
	if pass == "" {
		pass = "serversmith123"
	}
	return user, pass
}

// --- Auth middleware ---

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, 401, "unauthorized")
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := validateToken(tokenStr)
		if err != nil {
			writeError(w, 401, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, claims.Username)
		next(w, r.WithContext(ctx))
	}
}

// --- Handlers ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid body: "+err.Error())
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, 400, "username and password required")
		return
	}

	// Check DB
	var storedHash string
	err := db.DB.QueryRow(
		"SELECT password_hash FROM users WHERE username = ?", req.Username,
	).Scan(&storedHash)
	if err != nil {
		writeError(w, 401, "invalid credentials")
		return
	}

	if hashPassword(req.Password) != storedHash {
		writeError(w, 401, "invalid credentials")
		return
	}

	token, err := generateToken(req.Username)
	if err != nil {
		writeError(w, 500, "token generation failed")
		return
	}

	writeOK(w, LoginResponse{Token: token, Username: req.Username})
}

func HandleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeError(w, 401, "unauthorized")
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := validateToken(tokenStr)
	if err != nil {
		writeError(w, 401, "invalid or expired token")
		return
	}

	writeOK(w, map[string]string{"username": claims.Username})
}

// --- Public status API (no auth required) ---

type PublicServerStatus struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	IsOnline    bool    `json:"is_online"`
	LastReportAt *string `json:"last_report_at,omitempty"`
}

func HandlePublicStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, name, status, last_report_at
		FROM servers
		ORDER BY name
	`)
	if err != nil {
		writeError(w, 500, "query error: "+err.Error())
		return
	}
	defer rows.Close()

	servers := make([]PublicServerStatus, 0)
	for rows.Next() {
		var s PublicServerStatus
		var status string
		var lastReport *string
		if err := rows.Scan(&s.ID, &s.Name, &status, &lastReport); err != nil {
			continue
		}
		s.IsOnline = status == "online"
		s.LastReportAt = lastReport
		servers = append(servers, s)
	}

	writeOK(w, servers)
}

// --- Init default admin ---

func InitDefaultAdmin() {
	user, pass := defaultAdminCreds()

	var count int
	_ = db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return
	}

	hash := hashPassword(pass)
	_, err := db.DB.Exec(
		"INSERT INTO users (username, password_hash, created_at) VALUES (?, ?, ?)",
		user, hash, db.NowUTC(),
	)
	if err != nil {
		log.Printf("[auth] Failed to create default admin: %v", err)
		return
	}
	log.Printf("[auth] Default admin created: %s", user)
}
