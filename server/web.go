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
		// embed.FS doesn't support directory Stat, so check for index.html instead
		path := strings.TrimPrefix(r.URL.Path, "/")
		checkPath := strings.TrimSuffix(path, "/") + "/index.html"
		if checkPath == "/index.html" {
			checkPath = "index.html"
		}
		if _, err := fs.Stat(subFS, checkPath); err != nil {
			// Really 404 → SPA fallback to index.html (client-side router handles it)
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
