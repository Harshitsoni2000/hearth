package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hearth/internal/config"
)

// New builds an *http.Server for cfg, serving mux behind the logging
// middleware.
//
// WriteTimeout is deliberately left unset (zero). Hearth streams whole
// movie files over long-lived responses; a WriteTimeout would cut those
// streams off mid-playback. Do not add one.
func New(cfg config.Config, mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port),
		Handler:           logging(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code
// written by the wrapped handler, defaulting to 200 if WriteHeader is
// never called.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// logging wraps next, logging method, path, Range header (if present),
// response status, and duration for each request.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		rng := r.Header.Get("Range")
		log.Printf("%s %s range=%q status=%d duration=%s", r.Method, r.URL.Path, rng, rw.status, time.Since(start))
	})
}

// Run starts srv and blocks until it shuts down. It listens for
// os.Interrupt and syscall.SIGTERM, and on receiving one, gracefully
// shuts srv down with a 5-second timeout. It returns nil on a clean
// shutdown, or the error from ListenAndServe/Shutdown otherwise.
func Run(srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		return <-errCh
	}
}
