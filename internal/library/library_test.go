package library

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestScanReturnsOnlyVideoFiles(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "Movies", "Dune.mkv"), "video-a")
	mustWriteFile(t, filepath.Join(root, "Shows", "Ep1.mp4"), "video-b")
	mustWriteFile(t, filepath.Join(root, "notes.txt"), "not a video")

	entries, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(entries), entries)
	}

	want := []FileEntry{
		{Name: "Dune.mkv", Path: "Movies/Dune.mkv"},
		{Name: "Ep1.mp4", Path: "Shows/Ep1.mp4"},
	}

	for i, w := range want {
		got := entries[i]
		if got.Name != w.Name {
			t.Errorf("entries[%d].Name = %q, want %q", i, got.Name, w.Name)
		}
		if got.Path != w.Path {
			t.Errorf("entries[%d].Path = %q, want %q", i, got.Path, w.Path)
		}
		if got.Size <= 0 {
			t.Errorf("entries[%d].Size = %d, want > 0", i, got.Size)
		}
		if got.ModTime.IsZero() {
			t.Errorf("entries[%d].ModTime is zero", i)
		}
	}
}

func TestScanSortsByPath(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "z.mp4"), "z")
	mustWriteFile(t, filepath.Join(root, "a.mp4"), "a")
	mustWriteFile(t, filepath.Join(root, "m.mp4"), "m")

	entries, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	gotOrder := []string{entries[0].Path, entries[1].Path, entries[2].Path}
	wantOrder := []string{"a.mp4", "m.mp4", "z.mp4"}
	for i := range wantOrder {
		if gotOrder[i] != wantOrder[i] {
			t.Errorf("entries[%d].Path = %q, want %q", i, gotOrder[i], wantOrder[i])
		}
	}
}

func TestScanUsesForwardSlashesInPath(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "Sub", "Dir", "movie.mkv"), "x")

	entries, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "Sub/Dir/movie.mkv" {
		t.Errorf("Path = %q, want %q", entries[0].Path, "Sub/Dir/movie.mkv")
	}
}

func TestScanNonexistentRoot(t *testing.T) {
	_, err := Scan(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected error for nonexistent root, got nil")
	}
}

func TestJSONHandlerReturnsVideoEntriesOnly(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "Dune.mkv"), "video-a")
	mustWriteFile(t, filepath.Join(root, "Ep1.mp4"), "video-b")
	mustWriteFile(t, filepath.Join(root, "notes.txt"), "not a video")

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	rec := httptest.NewRecorder()

	JSONHandler(root).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var entries []FileEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to unmarshal response body: %v; body=%s", err, rec.Body.String())
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(entries), entries)
	}
}

func TestJSONHandlerScanError(t *testing.T) {
	missingRoot := filepath.Join(t.TempDir(), "missing")

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	rec := httptest.NewRecorder()

	JSONHandler(missingRoot).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
