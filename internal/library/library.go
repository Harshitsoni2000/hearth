package library

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hearth/internal/config"
)

// FileEntry describes one video file found under a library root.
type FileEntry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

// Scan walks root and returns every video file found, sorted by Path.
func Scan(root string) ([]FileEntry, error) {
	var entries []FileEntry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := config.VideoExtensions[strings.ToLower(filepath.Ext(path))]; !ok {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		entries = append(entries, FileEntry{
			Name:    d.Name(),
			Path:    filepath.ToSlash(rel),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}

// JSONHandler returns an http.Handler that scans root on every request and
// writes the resulting FileEntry list as a JSON array.
func JSONHandler(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries, err := Scan(root)
		if err != nil {
			http.Error(w, "failed to scan library", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}
