package main

import (
	"fmt"
	"log"
	"net/http"

	"hearth/internal/config"
	"hearth/internal/library"
	"hearth/internal/media"
	"hearth/internal/server"
)

func main() {
	cfg := config.Parse()

	mux := http.NewServeMux()
	mux.Handle("GET /media/{path...}", media.Handler(cfg.Dir))
	mux.Handle("GET /api/files", library.JSONHandler(cfg.Dir))
	mux.HandleFunc("GET /", rootHandler(cfg.Dir))

	srv := server.New(cfg, mux)

	log.Printf("hearth serving %s on %s:%d", cfg.Dir, cfg.Addr, cfg.Port)
	if err := server.Run(srv); err != nil {
		log.Fatal(err)
	}
}

// rootHandler returns a handler that lists every media file under dir as a
// plaintext list of playable URLs, e.g. http://<host>/media/<path>.
func rootHandler(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		entries, err := library.Scan(dir)
		if err != nil {
			fmt.Fprintf(w, "error scanning library: %v\n", err)
			return
		}

		for _, e := range entries {
			fmt.Fprintf(w, "http://%s/media/%s\n", r.Host, e.Path)
		}
	}
}
