package library

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildTree_PrunesEmptyDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Keep", "movie.mp4"))
	// "Drop" contains only a non-media file and an empty subdir; must be pruned.
	writeFile(t, filepath.Join(root, "Drop", "readme.txt"))
	if err := os.MkdirAll(filepath.Join(root, "Drop", "Nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	tree, err := BuildTree(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 1 {
		t.Fatalf("want 1 top node (Keep), got %d: %+v", len(tree), tree)
	}
	if tree[0].Name != "Keep" {
		t.Fatalf("want Keep, got %q", tree[0].Name)
	}
}

func TestBuildTree_DirsBeforeFilesAlphabetical(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "zebra.mp4"))
	writeFile(t, filepath.Join(root, "apple.mkv"))
	writeFile(t, filepath.Join(root, "Beta", "x.mov"))
	writeFile(t, filepath.Join(root, "Alpha", "y.avi"))

	tree, err := BuildTree(root)
	if err != nil {
		t.Fatal(err)
	}
	// Expected order: Alpha (dir), Beta (dir), apple.mkv (file), zebra.mp4 (file).
	want := []struct {
		name  string
		isDir bool
	}{
		{"Alpha", true},
		{"Beta", true},
		{"apple.mkv", false},
		{"zebra.mp4", false},
	}
	if len(tree) != len(want) {
		t.Fatalf("want %d nodes, got %d: %+v", len(want), len(tree), tree)
	}
	for i, w := range want {
		if tree[i].Name != w.name || tree[i].IsDir != w.isDir {
			t.Errorf("node %d: want {%q dir=%v}, got {%q dir=%v}",
				i, w.name, w.isDir, tree[i].Name, tree[i].IsDir)
		}
	}
}

func TestBuildTree_RelPathIsSlashAndRelative(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "A", "B", "clip.webm"))

	tree, err := BuildTree(root)
	if err != nil {
		t.Fatal(err)
	}
	// tree[0] = A, A.Children[0] = B, B.Children[0] = clip.webm
	got := tree[0].Children[0].Children[0].RelPath
	if got != "A/B/clip.webm" {
		t.Fatalf("want RelPath A/B/clip.webm, got %q", got)
	}
}

func TestBuildTree_IgnoresNonMedia(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Show", "ep.mkv"))
	writeFile(t, filepath.Join(root, "Show", "poster.jpg"))
	writeFile(t, filepath.Join(root, "Show", "notes.txt"))

	tree, err := BuildTree(root)
	if err != nil {
		t.Fatal(err)
	}
	children := tree[0].Children
	if len(children) != 1 || children[0].Name != "ep.mkv" {
		t.Fatalf("want only ep.mkv, got %+v", children)
	}
}
