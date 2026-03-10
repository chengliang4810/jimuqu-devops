package httpapi

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed web
//go:embed web/assets/*
var webFiles embed.FS

func (s *Server) mountUI(router chi.Router) {
	indexFile, err := webFiles.ReadFile("web/index.html")
	if err == nil {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexFile)
		})
	}

	assetsFS, err := fs.Sub(webFiles, "web/assets")
	if err == nil {
		fileServer := http.FileServer(http.FS(assetsFS))
		router.Handle("/assets/*", http.StripPrefix("/assets/", fileServer))
	}
}