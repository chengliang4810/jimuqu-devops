package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWriteJSONDisablesCaching(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeJSON(recorder, http.StatusOK, map[string]string{"status": "ok"})

	cacheControl := recorder.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "no-store") {
		t.Fatalf("expected Cache-Control to include no-store, got %q", cacheControl)
	}
	if got := recorder.Header().Get("CDN-Cache-Control"); got != "no-store" {
		t.Fatalf("expected CDN-Cache-Control no-store, got %q", got)
	}
	if got := recorder.Header().Get("Surrogate-Control"); got != "no-store" {
		t.Fatalf("expected Surrogate-Control no-store, got %q", got)
	}
}

func TestStaticHTMLDisablesCaching(t *testing.T) {
	staticFS := fstest.MapFS{
		"index.html": {Data: []byte("<!doctype html><html></html>")},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	serveFSFile(recorder, request, staticFS, "index.html")

	cacheControl := recorder.Header().Get("Cache-Control")
	if !strings.Contains(cacheControl, "no-store") {
		t.Fatalf("expected html Cache-Control to include no-store, got %q", cacheControl)
	}
	if got := recorder.Header().Get("CDN-Cache-Control"); got != "no-store" {
		t.Fatalf("expected html CDN-Cache-Control no-store, got %q", got)
	}
}

func TestNextStaticAssetsUseImmutableCaching(t *testing.T) {
	staticFS := fstest.MapFS{
		"_next/static/chunks/app.js": {Data: []byte("console.log('ok')")},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/_next/static/chunks/app.js", nil)

	serveFSFile(recorder, request, staticFS, "_next/static/chunks/app.js")

	want := "public, max-age=31536000, immutable"
	if got := recorder.Header().Get("Cache-Control"); got != want {
		t.Fatalf("expected immutable cache header %q, got %q", want, got)
	}
}
