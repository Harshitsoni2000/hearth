package media

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// newTestServer starts a real httptest.Server behind a ServeMux carrying the
// production route pattern, so r.PathValue("path") is populated the same way
// it is in main.go.
func newTestServer(t *testing.T, root string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("GET /media/{path...}", Handler(root))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// knownContent returns a deterministic byte slice of the given size so tests
// can assert on exact lengths and exact byte ranges.
func knownContent(size int) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(i % 256)
	}
	return b
}

func TestHandlerFullAndRangeRequests(t *testing.T) {
	root := t.TempDir()
	content := knownContent(500)
	if err := os.WriteFile(filepath.Join(root, "movie.mkv"), content, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	srv := newTestServer(t, root)

	tests := []struct {
		name           string
		path           string
		rangeHeader    string
		wantStatus     int
		wantBody       []byte
		wantContentRng string
	}{
		{
			name:       "full GET returns 200 and full length",
			path:       "movie.mkv",
			wantStatus: http.StatusOK,
			wantBody:   content,
		},
		{
			name:           "range 0-99 returns 206 with first 100 bytes",
			path:           "movie.mkv",
			rangeHeader:    "bytes=0-99",
			wantStatus:     http.StatusPartialContent,
			wantBody:       content[0:100],
			wantContentRng: "bytes 0-99/500",
		},
		{
			name:           "range from 100 to end returns 206 starting at offset 100",
			path:           "movie.mkv",
			rangeHeader:    "bytes=100-",
			wantStatus:     http.StatusPartialContent,
			wantBody:       content[100:],
			wantContentRng: "bytes 100-499/500",
		},
		{
			name:        "out of range Range returns 416",
			path:        "movie.mkv",
			rangeHeader: "bytes=999999999-",
			wantStatus:  http.StatusRequestedRangeNotSatisfiable,
		},
		{
			name:       "missing file returns 404",
			path:       "does-not-exist.mkv",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/media/"+tt.path, nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Do: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantContentRng != "" {
				if got := resp.Header.Get("Content-Range"); got != tt.wantContentRng {
					t.Errorf("Content-Range = %q, want %q", got, tt.wantContentRng)
				}
			}

			if tt.wantBody != nil {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("ReadAll: %v", err)
				}
				if len(body) != len(tt.wantBody) {
					t.Fatalf("body length = %d, want %d", len(body), len(tt.wantBody))
				}
				if !bytes.Equal(body, tt.wantBody) {
					t.Errorf("body bytes do not match expected range")
				}
			}
		})
	}
}

// TestHandlerPathTraversalNeutralized documents the actual, current behavior
// for a traversal attempt like /media/../../secret: net/http's ServeMux
// cleans the URL path and redirects (301) to the cleaned path before the
// request ever reaches media.Handler. The client here follows that redirect,
// lands on a path with no registered route, and gets 404 — the traversal
// never escapes root and the handler's own belt-and-suspenders 403 branch is
// never exercised through the real route. Confirmed with the project owner
// that 404 (not 403) is the expected, accepted behavior.
func TestHandlerPathTraversalNeutralized(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "movie.mkv"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	srv := newTestServer(t, root)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/media/../../secret", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandlerDirectoryNotServed(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	srv := newTestServer(t, root)

	resp, err := http.Get(srv.URL + "/media/sub")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandlerSetsContentTypeByExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType string
	}{
		{name: "mkv gets matroska content type", filename: "movie.mkv", wantType: "video/x-matroska"},
		{name: "mp4 gets mp4 content type", filename: "movie.mp4", wantType: "video/mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, tt.filename), []byte("content"), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			srv := newTestServer(t, root)

			resp, err := http.Get(fmt.Sprintf("%s/media/%s", srv.URL, tt.filename))
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if got := resp.Header.Get("Content-Type"); got != tt.wantType {
				t.Errorf("Content-Type = %q, want %q", got, tt.wantType)
			}
		})
	}
}
