package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server provides Prometheus metrics endpoint
type Server struct {
	server *http.Server
}

// NewServer creates a new metrics server
func NewServer(addr string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Add reset endpoint
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}

		Reset()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metrics reset successfully\n"))
	})

	return &Server{
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// StartServer is a convenience function to start metrics server
func StartServer(addr string) error {
	server := NewServer(addr)
	return server.Start()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
