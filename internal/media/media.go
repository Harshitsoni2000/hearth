package media

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var extraMIMETypes = map[string]string{
	".mkv":  "video/x-matroska",
	".mp4":  "video/mp4",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".webm": "video/webm",
	".m4v":  "video/mp4",
}

// Handler returns an http.Handler that serves files out of root, the
// absolute media directory. It expects the request path to be available
// via r.PathValue("path") (route: GET /media/{path...}).
func Handler(root string) http.Handler {
	// Wrap the closure as an http.HandlerFunc so it satisfies http.Handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pull the {path...} wildcard segment out of the route, e.g. "Movies/Dune.mkv".
		reqPath := r.PathValue("path")

		// Prefix with "/" and clean: any leading ".." segments collapse against
		// this synthetic root instead of climbing above it, neutralizing traversal.
		clean := filepath.Clean("/" + reqPath)
		// Join the neutralized path onto the real media root to get the file to serve.
		fullPath := filepath.Join(root, clean)

		// Belt-and-suspenders check: confirm fullPath is root itself or nested under
		// it (root + separator prefix) before doing anything with it on disk.
		if fullPath != root && !strings.HasPrefix(fullPath, root+string(os.PathSeparator)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Stat the target so we can tell "doesn't exist" and "is a directory"
		// apart from a servable file.
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		// Open the file for reading; ServeContent needs an io.ReadSeeker.
		f, err := os.Open(fullPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		// Ensure the file descriptor is released once the request is handled.
		defer f.Close()

		// Look up the lowercased extension in our override table, since Go's
		// built-in MIME sniffing doesn't know these video container types.
		if mimeType, ok := extraMIMETypes[strings.ToLower(filepath.Ext(fullPath))]; ok {
			// Set Content-Type explicitly before ServeContent has a chance to guess.
			w.Header().Set("Content-Type", mimeType)
		}

		// Hand off to the stdlib: it handles Range/If-Range/Accept-Ranges and
		// conditional requests, and streams straight from the file without
		// buffering it all in memory.
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	})
}
