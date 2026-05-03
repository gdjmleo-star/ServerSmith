package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed all:web-out
var webFS embed.FS

func webHandler() http.Handler {
	subFS, err := fs.Sub(webFS, "web-out")
	if err != nil {
		log.Fatalf("Failed to create embedded sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		if path == "" {
			// Root → serve index.html via FileServer (handles ETag, Range, etc.)
			r.URL.Path = "/"
		} else if _, err := fs.Stat(subFS, path); err != nil {
			// File not found — try directory index (Next.js trailingSlash: true generates
			// /login/index.html, /servers/index.html, etc. but embed.FS can't Stat directories).
			// IMPORTANT: http.FileServer redirects any path ending in /index.html to the parent
			// directory (Go's documented behavior), so we MUST read the file directly instead of
			// passing it through fileServer.ServeHTTP.
			indexPath := strings.TrimSuffix(path, "/") + "/index.html"
			if data, err := fs.ReadFile(subFS, indexPath); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.WriteHeader(http.StatusOK)
				w.Write(data)
				return
			}
			// Truly not found → SPA fallback to root index.html
			// Client-side Next.js router handles 404 routes
			r.URL.Path = "/"
		}
		// path exists as a file → let FileServer serve it directly

		fileServer.ServeHTTP(w, r)
	})
}
