package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"serversmith/api"
	"serversmith/cron"
	"serversmith/db"
)

// Build version, set via -ldflags at compile time
var Version = "dev"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("ServerSmith starting...")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/serversmith.db"
	}

	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create default admin on first run
	api.InitDefaultAdmin()

	mux := http.NewServeMux()

	// ── Public routes (no auth) ──
	mux.HandleFunc("/api/login", api.HandleLogin)
	mux.HandleFunc("/api/verify", api.HandleVerify)
	mux.HandleFunc("/api/report", api.HandleReport)
	mux.HandleFunc("/public/status", api.HandlePublicStatus)

	// ── Protected routes (require JWT) ──
	mux.HandleFunc("/api/servers", api.RequireAuth(api.HandleServers))
	mux.HandleFunc("/api/servers/batch-reboot", api.RequireAuth(api.HandleBatchReboot))
	mux.HandleFunc("/api/servers/", api.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/servers/")+1 && path[len(path)-7:] == "/reboot" {
			api.HandleServerReboot(w, r)
			return
		}
		if len(path) > len("/api/servers/")+1 && path[len(path)-8:] == "/history" {
			api.HandleServerHistory(w, r)
			return
		}
		if len(path) > len("/api/servers/")+1 && strings.HasSuffix(path, "/reboot-tasks") {
			api.HandleServerRebootTasks(w, r)
			return
		}
		api.HandleServerByID(w, r)
	}))
	mux.HandleFunc("/api/dashboard", api.RequireAuth(api.HandleDashboard))
	mux.HandleFunc("/api/dashboard/by-carrier", api.RequireAuth(api.HandleDashboardByCarrier))
	mux.HandleFunc("/api/alerts", api.RequireAuth(api.HandleAlerts))
	// User management
	mux.HandleFunc("/api/users/change-password", api.RequireAuth(api.HandleChangePassword))
	mux.HandleFunc("/api/users/", api.RequireAuth(api.HandleUserByID))
	mux.HandleFunc("/api/users", api.RequireAuth(api.HandleUsers))
	mux.HandleFunc("/api/alerts/", api.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/acknowledge") {
			api.HandleAcknowledgeAlert(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	mux.HandleFunc("/api/alert-configs", api.RequireAuth(api.HandleAlertConfigs))
	mux.HandleFunc("/api/backup", api.RequireAuth(api.HandleBackup))
	mux.HandleFunc("/api/ssh-key", api.RequireAuth(api.HandleSSHKey))
	// Agent 下载接口 (public): 二进制 + 安装脚本
	mux.HandleFunc("/agent/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/agent/"):]
		if name == "install.sh" || name == "install.ps1" {
			api.HandleAgentInstallScript(w, r)
		} else {
			api.HandleAgentBinary(w, r)
		}
	})

	// CORS
	handler := api.CORSMiddleware(mux)

	// Web static files (Go embed — serves the Next.js static export)
	// SPA fallback: non-matching paths serve index.html for client-side routing
	mux.Handle("/", webHandler())

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, Version)
	})

	// Start cron tasks
	stopCron := make(chan struct{})
	go cron.StartCycleReset(stopCron)
	go cron.StartOfflineCheck(stopCron)
	go cron.StartExpiryNotify(stopCron)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		close(stopCron)
		os.Exit(0)
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("ServerSmith API listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
