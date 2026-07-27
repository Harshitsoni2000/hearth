package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Dir  string
	Port int
	Addr string
}

// VideoExtensions maps each supported video file extension (lowercase, with
// leading dot) to the MIME type Hearth should serve it as. It is the single
// source of truth for which files count as media, shared by internal/library
// (listing) and internal/media (serving).
var VideoExtensions = map[string]string{
	".mkv":  "video/x-matroska",
	".mp4":  "video/mp4",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".webm": "video/webm",
	".m4v":  "video/mp4",
	".pdf":  "application/pdf",
}

func Parse() Config {
	dir := flag.String("dir", ".", "directory to serve media from")
	port := flag.Int("port", 8080, "port to listen on")
	addr := flag.String("addr", "0.0.0.0", "address to listen on")
	flag.Parse()

	info, err := os.Stat(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: -dir %q: %v\n", *dir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: -dir %q is not a directory\n", *dir)
		os.Exit(1)
	}

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: resolving absolute path for -dir %q: %v\n", *dir, err)
		os.Exit(1)
	}

	return Config{
		Dir:  absDir,
		Port: *port,
		Addr: *addr,
	}
}
