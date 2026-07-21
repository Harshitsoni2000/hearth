package library

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hearth/internal/config"
)

// Node is one entry in the media tree. A Node is either a directory
// (IsDir == true, Children populated, RelPath empty) or a media file
// (IsDir == false, RelPath set, Children nil).
type Node struct {
	Name     string // base name shown to the user, e.g. "Dune.mkv" or "Movies"
	RelPath  string // for files only: slash-separated path relative to root, e.g. "Movies/Dune.mkv"
	IsDir    bool
	Children []Node // for directories only
}

// isMediaFile reports whether name has one of the allowed media extensions.
// The check is case-insensitive.
func isMediaFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := config.VideoExtensions[ext]
	return ok
}

// BuildTree walks root and returns the media tree: the top-level nodes
// that are the contents of root. Directories containing no media file at
// any depth are omitted. Within every directory, subdirectories come
// first (alphabetical, case-insensitive), then media files (alphabetical,
// case-insensitive). root must be an absolute path.
func BuildTree(root string) ([]Node, error) {
	return buildDir(root, root)
}

func buildDir(dir string, root string) ([]Node, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs []Node
	var files []Node

	for _, e := range entries {
		full := filepath.Join(dir, e.Name())

		if e.IsDir() {
			children, err := buildDir(full, root)
			if err != nil {
				return nil, err
			}
			// PRUNE: skip a directory that has no media anywhere inside it.
			if len(children) == 0 {
				continue
			}
			dirs = append(dirs, Node{
				Name:     e.Name(),
				IsDir:    true,
				Children: children,
			})
			continue
		}

		// It is a file. Keep it only if it is a media file.
		if !isMediaFile(e.Name()) {
			continue
		}
		rel, err := filepath.Rel(root, full)
		if err != nil {
			return nil, err
		}
		files = append(files, Node{
			Name:    e.Name(),
			RelPath: filepath.ToSlash(rel),
			IsDir:   false,
		})
	}

	// SORT: directories first (already collected separately), each group
	// alphabetical case-insensitive.
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Directories first, then files.
	result := append(dirs, files...)
	return result, nil
}
