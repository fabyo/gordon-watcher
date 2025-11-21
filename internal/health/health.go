package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/fabyo/gordon-watcher/internal/logger"
)

// Status represents health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// HealthResponse is the health check response
type HealthResponse struct {
	Status    Status            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// Server provides health check endpoints
type Server struct {
	addr      string
	server    *http.Server
	startTime time.Time
	logger    *logger.Logger

	mu     sync.RWMutex
	ready  bool
	checks map[string]func() error
}

// NewServer creates a new health check server
func NewServer(addr string, log *logger.Logger) *Server {
	s := &Server{
		addr:      addr,
		startTime: time.Now(),
		logger:    log,
		ready:     false,
		checks:    make(map[string]func() error),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readyHandler)
	mux.HandleFunc("/live", s.liveHandler)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return s
}

// Start starts the health check server
func (s *Server) Start() error {
	s.logger.Info("Starting health check server", "addr", s.addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// SetReady sets the ready status
func (s *Server) SetReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = ready
}

// AddCheck adds a health check
func (s *Server) AddCheck(name string, check func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checks[name] = check
}

// healthHandler handles /health endpoint
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	checks := s.checks
	s.mu.RUnlock()

	// Run all checks
	checkResults := make(map[string]string)
	allHealthy := true

	for name, check := range checks {
		if err := check(); err != nil {
			checkResults[name] = err.Error()
			allHealthy = false
		} else {
			checkResults[name] = "ok"
		}
	}

	status := StatusHealthy
	if !allHealthy {
		status = StatusDegraded
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    time.Since(s.startTime).String(),
		Checks:    checkResults,
	}

	statusCode := http.StatusOK
	if status != StatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

// readyHandler handles /ready endpoint (Kubernetes readiness probe)
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	ready := s.ready
	s.mu.RUnlock()

	if ready {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
	}
}

// liveHandler handles /live endpoint (Kubernetes liveness probe)
func (s *Server) liveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("alive"))
}
