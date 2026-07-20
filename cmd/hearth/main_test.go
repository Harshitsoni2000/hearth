package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustWriteFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func TestRootHandlerListsPlayableURLs(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "Dune.mkv"), "video-a")
	mustWriteFile(t, filepath.Join(root, "Shows", "Ep1.mp4"), "video-b")
	mustWriteFile(t, filepath.Join(root, "notes.txt"), "not a video")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "192.168.1.5:8080"
	rec := httptest.NewRecorder()

	rootHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/plain")
	}

	wantLines := []string{
		"http://192.168.1.5:8080/media/Dune.mkv",
		"http://192.168.1.5:8080/media/Shows/Ep1.mp4",
	}
	body := rec.Body.String()
	for _, want := range wantLines {
		if !strings.Contains(body, want) {
			t.Errorf("body %q does not contain line %q", body, want)
		}
	}
	if strings.Contains(body, "notes.txt") {
		t.Errorf("body %q should not list non-video file notes.txt", body)
	}
}

func TestRootHandlerEmptyLibrary(t *testing.T) {
	root := t.TempDir()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()

	rootHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "" {
		t.Errorf("body = %q, want empty for a library with no video files", body)
	}
}

func TestRootHandlerScanErrorWritesErrorLine(t *testing.T) {
	missingRoot := filepath.Join(t.TempDir(), "does-not-exist")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()

	rootHandler(missingRoot).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "error") {
		t.Errorf("body %q should contain an error line when Scan fails", body)
	}
}
