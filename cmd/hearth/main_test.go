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

func TestRootHandlerRendersTree(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "Dune.mkv"), "video-a")
	mustWriteFile(t, filepath.Join(root, "Shows", "Ep1.mp4"), "video-b")
	mustWriteFile(t, filepath.Join(root, "notes.txt"), "not a video")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	rootHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}

	wantPaths := []string{
		`data-path="Dune.mkv"`,
		`data-path="Shows/Ep1.mp4"`,
	}
	body := rec.Body.String()
	for _, want := range wantPaths {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
	if strings.Contains(body, "notes.txt") {
		t.Errorf("body should not list non-video file notes.txt")
	}
}

func TestRootHandlerEmptyLibrary(t *testing.T) {
	root := t.TempDir()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	rootHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); strings.Contains(body, `data-path="`) {
		t.Errorf("body = %q, should contain no data-path attribute for an empty library", body)
	}
}

func TestRootHandlerBuildTreeErrorServesHTML500(t *testing.T) {
	missingRoot := filepath.Join(t.TempDir(), "does-not-exist")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	rootHandler(missingRoot).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Could not read the media library") {
		t.Errorf("body %q should contain the error message", body)
	}
}
