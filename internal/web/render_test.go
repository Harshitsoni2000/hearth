package web

import (
	"strings"
	"testing"

	"hearth/internal/library"
)

func TestRender_FileHasDataPath(t *testing.T) {
	tree := []library.Node{
		{Name: "Dune.mkv", RelPath: "Movies/Dune.mkv", IsDir: false},
	}
	var sb strings.Builder
	if err := Render(&sb, tree); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), `data-path="Movies/Dune.mkv"`) {
		t.Error("missing correct data-path")
	}
}

func TestRender_FolderCollapsedByDefault(t *testing.T) {
	tree := []library.Node{
		{Name: "Movies", IsDir: true, Children: []library.Node{
			{Name: "Dune.mkv", RelPath: "Movies/Dune.mkv"},
		}},
	}
	var sb strings.Builder
	if err := Render(&sb, tree); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), `class="hearth-children collapsed"`) {
		t.Error("folder children should be collapsed by default")
	}
}

func TestRender_EmptyShowsMessage(t *testing.T) {
	var sb strings.Builder
	if err := Render(&sb, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sb.String(), "No media found") {
		t.Error("expected empty-state message")
	}
}

func TestRender_EscapesSpecialChars(t *testing.T) {
	tree := []library.Node{
		{Name: "A & B.mp4", RelPath: "A & B.mp4", IsDir: false},
	}
	var sb strings.Builder
	if err := Render(&sb, tree); err != nil {
		t.Fatal(err)
	}
	out := sb.String()
	if strings.Contains(out, ">A & B.mp4<") {
		t.Error("name ampersand not escaped")
	}
	if !strings.Contains(out, "A &amp; B.mp4") {
		t.Error("expected escaped ampersand")
	}
}
