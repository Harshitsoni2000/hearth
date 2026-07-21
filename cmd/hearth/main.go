package main

import (
	"io"
	"log"
	"net/http"

	"hearth/internal/config"
	"hearth/internal/library"
	"hearth/internal/media"
	"hearth/internal/server"
	"hearth/internal/web"
)

func main() {
	cfg := config.Parse()

	mux := http.NewServeMux()
	mux.Handle("GET /media/{path...}", media.Handler(cfg.Dir))
	mux.Handle("GET /api/files", library.JSONHandler(cfg.Dir))
	mux.Handle("GET /", rootHandler(cfg.Dir))

	srv := server.New(cfg, mux)

	log.Printf("hearth serving %s on %s:%d", cfg.Dir, cfg.Addr, cfg.Port)
	for _, addr := range server.LANAddrs(cfg.Port) {
		log.Printf("available at %s", addr)
	}
	if err := server.Run(srv); err != nil {
		log.Fatal(err)
	}
}

func rootHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tree, err := library.BuildTree(root)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "<!DOCTYPE html><p>Could not read the media library.</p>")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := web.Render(w, tree); err != nil {
			log.Printf("render error: %v", err)
		}
	}
}
