package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed web-out/*
var webFS embed.FS

func webHandler() http.Handler {
	subFS, err := fs.Sub(webFS, "web-out")
	if err != nil {
		log.Fatalf("Failed to create embedded sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try the exact path first
		path := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(subFS, path); err != nil {
			// Not found: serve index.html as SPA fallback (client-side router handles 404s)
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
