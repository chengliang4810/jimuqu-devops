package httpapi

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) mountStaticUI(router chi.Router) {
	staticDir, ok := resolveStaticUIDir()
	if !ok {
		s.logger.Warn("static ui directory not found", "candidates", []string{
			filepath.Join(".", "web-next", "out"),
			filepath.Join(executableDir(), "web-next", "out"),
		})
		return
	}

	s.logger.Info("serving static ui", "dir", staticDir)
	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}
		serveStaticUI(w, r, staticDir)
	}))
}

func resolveStaticUIDir() (string, bool) {
	candidates := []string{
		filepath.Join(".", "web-next", "out"),
		filepath.Join(executableDir(), "web-next", "out"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			absPath, absErr := filepath.Abs(candidate)
			if absErr == nil {
				return absPath, true
			}
			return candidate, true
		}
	}

	return "", false
}

func executableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

func serveStaticUI(w http.ResponseWriter, r *http.Request, staticDir string) {
	requestPath := path.Clean("/" + strings.TrimSpace(r.URL.Path))
	if requestPath == "/" {
		serveExistingFile(w, r, filepath.Join(staticDir, "index.html"))
		return
	}

	trimmed := strings.TrimPrefix(requestPath, "/")
	candidates := []string{
		filepath.Join(staticDir, filepath.FromSlash(trimmed)),
		filepath.Join(staticDir, filepath.FromSlash(trimmed)+".html"),
		filepath.Join(staticDir, filepath.FromSlash(trimmed), "index.html"),
	}

	if requestPath == "/favicon.ico" {
		candidates = append([]string{filepath.Join(staticDir, "icon.svg")}, candidates...)
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			serveExistingFile(w, r, candidate)
			return
		}
	}

	serveExistingFile(w, r, filepath.Join(staticDir, "index.html"))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func serveExistingFile(w http.ResponseWriter, r *http.Request, filename string) {
	http.ServeFile(w, r, filename)
}
