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

	mux := http.NewServeMux()

	// CORS
	handler := api.CORSMiddleware(mux)

	// API Routes
	mux.HandleFunc("/api/report", api.HandleReport)
	mux.HandleFunc("/api/servers", api.HandleServers)
	mux.HandleFunc("/api/servers/batch-reboot", api.HandleBatchReboot)
	mux.HandleFunc("/api/servers/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// /api/servers/:id/reboot
		if len(path) > len("/api/servers/")+1 && path[len(path)-7:] == "/reboot" {
			api.HandleServerReboot(w, r)
			return
		}
		// /api/servers/:id/history
		if len(path) > len("/api/servers/")+1 && path[len(path)-8:] == "/history" {
			api.HandleServerHistory(w, r)
			return
		}
		// /api/servers/:id/reboot-tasks
		if len(path) > len("/api/servers/")+1 && strings.HasSuffix(path, "/reboot-tasks") {
			api.HandleServerRebootTasks(w, r)
			return
		}
		// /api/servers/:id
		api.HandleServerByID(w, r)
	})
	mux.HandleFunc("/api/dashboard", api.HandleDashboard)
	mux.HandleFunc("/api/dashboard/by-carrier", api.HandleDashboardByCarrier)
	mux.HandleFunc("/api/alerts", api.HandleAlerts)
	mux.HandleFunc("/api/alerts/", func(w http.ResponseWriter, r *http.Request) {
		// /api/alerts/:id/acknowledge
		if strings.HasSuffix(r.URL.Path, "/acknowledge") {
			api.HandleAcknowledgeAlert(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/alert-configs", api.HandleAlertConfigs)
	mux.HandleFunc("/api/backup", api.HandleBackup)

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
