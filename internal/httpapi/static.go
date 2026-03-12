package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed all:webdist
var embeddedWebDist embed.FS

func (s *Server) mountStaticUI(router chi.Router) {
	staticFS, source, ok := resolveStaticUIFS()
	if !ok {
		s.logger.Warn("static ui not found", "embedded", "internal/httpapi/webdist", "disk_candidates", []string{
			filepath.Join(".", "web-next", "out"),
			filepath.Join(executableDir(), "web-next", "out"),
		})
		return
	}

	s.logger.Info("serving static ui", "source", source)
	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}
		serveStaticUI(w, r, staticFS)
	}))
}

func resolveStaticUIFS() (fs.FS, string, bool) {
	if embeddedFS, err := fs.Sub(embeddedWebDist, "webdist"); err == nil && fsFileExists(embeddedFS, "index.html") {
		return embeddedFS, "embedded", true
	}

	candidates := []string{
		filepath.Join(".", "web-next", "out"),
		filepath.Join(executableDir(), "web-next", "out"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			diskFS := os.DirFS(candidate)
			if fsFileExists(diskFS, "index.html") {
				absPath, absErr := filepath.Abs(candidate)
				if absErr == nil {
					return diskFS, absPath, true
				}
				return diskFS, candidate, true
			}
		}
	}

	return nil, "", false
}

func executableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

func serveStaticUI(w http.ResponseWriter, r *http.Request, staticFS fs.FS) {
	requestPath := path.Clean("/" + strings.TrimSpace(r.URL.Path))
	if requestPath == "/" {
		serveFSFile(w, r, staticFS, "index.html")
		return
	}

	trimmed := strings.TrimPrefix(requestPath, "/")
	candidates := []string{
		trimmed,
		trimmed + ".html",
		path.Join(trimmed, "index.html"),
	}

	if requestPath == "/favicon.ico" {
		candidates = append([]string{"icon.svg"}, candidates...)
	}

	for _, candidate := range candidates {
		if fsFileExists(staticFS, candidate) {
			serveFSFile(w, r, staticFS, candidate)
			return
		}
	}

	serveFSFile(w, r, staticFS, "index.html")
}

func fsFileExists(staticFS fs.FS, filename string) bool {
	info, err := fs.Stat(staticFS, filename)
	return err == nil && !info.IsDir()
}

func serveFSFile(w http.ResponseWriter, r *http.Request, staticFS fs.FS, filename string) {
	fileServer := http.FileServer(http.FS(staticFS))
	request := r.Clone(r.Context())
	request.URL = cloneURL(r.URL)
	request.URL.Path = "/" + strings.TrimPrefix(filename, "/")
	fileServer.ServeHTTP(w, request)
}

func cloneURL(source *url.URL) *url.URL {
	if source == nil {
		return &url.URL{}
	}
	cloned := *source
	return &cloned
}
