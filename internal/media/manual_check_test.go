package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandler(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "small.mkv"), []byte("0123456789abcdefghij"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "big.mkv"), make([]byte, 200), 0o644); err != nil {
		t.Fatal(err)
	}

	h := Handler(root)

	tests := []struct {
		name        string
		pathValue   string
		rangeHeader string
		wantStatus  int
		wantBodyLen int
	}{
		{
			name:        "full file",
			pathValue:   "sub/small.mkv",
			wantStatus:  http.StatusOK,
			wantBodyLen: 20,
		},
		{
			name:        "range within bounds",
			pathValue:   "sub/big.mkv",
			rangeHeader: "bytes=0-99",
			wantStatus:  http.StatusPartialContent,
			wantBodyLen: 100,
		},
		{
			name:        "range clamped to file size",
			pathValue:   "sub/small.mkv",
			rangeHeader: "bytes=0-99",
			wantStatus:  http.StatusPartialContent,
			wantBodyLen: 20,
		},
		{
			// filepath.Clean("/"+reqPath) neutralizes ".." before the
			// prefix check ever runs, so the joined path is clamped
			// to root/etc/passwd, which doesn't exist. The traversal
			// never reaches outside root; the visible result is 404.
			name:       "path traversal neutralized",
			pathValue:  "../../../../etc/passwd",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing file",
			pathValue:  "sub/does-not-exist.mkv",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "directory not served",
			pathValue:  "sub",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/media/"+tt.pathValue, nil)
			req.SetPathValue("path", tt.pathValue)
			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantBodyLen != 0 && rr.Body.Len() != tt.wantBodyLen {
				t.Errorf("body len = %d, want %d", rr.Body.Len(), tt.wantBodyLen)
			}
		})
	}
}
