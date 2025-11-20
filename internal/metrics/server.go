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
	// Add CORS middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		promhttp.Handler().ServeHTTP(w, r)
	})

	mux.Handle("/metrics", handler)

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
