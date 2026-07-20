package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"hearth/internal/config"
)

func TestNewSetsAddrFromConfig(t *testing.T) {
	cfg := config.Config{Addr: "127.0.0.1", Port: 9090}
	mux := http.NewServeMux()

	srv := New(cfg, mux)

	if srv.Addr != "127.0.0.1:9090" {
		t.Errorf("Addr = %q, want %q", srv.Addr, "127.0.0.1:9090")
	}
}

func TestNewSetsReadHeaderTimeout(t *testing.T) {
	srv := New(config.Config{Addr: "0.0.0.0", Port: 8080}, http.NewServeMux())

	if srv.ReadHeaderTimeout != 10*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want %v", srv.ReadHeaderTimeout, 10*time.Second)
	}
}

func TestNewDoesNotSetWriteTimeout(t *testing.T) {
	srv := New(config.Config{Addr: "0.0.0.0", Port: 8080}, http.NewServeMux())

	if srv.WriteTimeout != 0 {
		t.Errorf("WriteTimeout = %v, want 0 (unset) so long video streams are not killed", srv.WriteTimeout)
	}
}

func TestNewWrapsHandlerWithLoggingMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	srv := New(config.Config{Addr: "0.0.0.0", Port: 8080}, mux)

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("Range", "bytes=0-99")
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}

	logged := logBuf.String()
	for _, want := range []string{http.MethodGet, "/hello", "bytes=0-99", "418"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log output %q does not contain %q", logged, want)
		}
	}
}

func TestLoggingResponseWriterDefaultsTo200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/noop", func(w http.ResponseWriter, r *http.Request) {})
	srv := New(config.Config{Addr: "0.0.0.0", Port: 8080}, mux)

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest(http.MethodGet, "/noop", nil)
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	if !strings.Contains(logBuf.String(), "200") {
		t.Errorf("log output %q does not contain default status 200", logBuf.String())
	}
}

func TestRunShutsDownGracefullyOnSignal(t *testing.T) {
	mux := http.NewServeMux()
	srv := New(config.Config{Addr: "127.0.0.1", Port: 0}, mux)

	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(srv)
	}()

	// Give ListenAndServe a moment to start before signaling shutdown.
	time.Sleep(50 * time.Millisecond)

	self, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if err := self.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("Signal: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run() = %v, want nil on clean shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after shutdown signal")
	}
}
