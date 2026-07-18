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
