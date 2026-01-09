package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// Server provides HTTP endpoints for health checks and metrics
type Server struct {
	server    *http.Server
	ready     int32 // 0 = not ready, 1 = ready
	startTime time.Time

	// Metrics callbacks - set these to provide live data
	GetConnectedPlayers func() int
	GetActiveStages     func() int
}

// NewServer creates a new health server
func NewServer(addr string) *Server {
	s := &Server{
		startTime: time.Now(),
	}
	atomic.StoreInt32(&s.ready, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return s
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	log.Printf("Health server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// SetReady sets the ready state
func (s *Server) SetReady(ready bool) {
	if ready {
		atomic.StoreInt32(&s.ready, 1)
	} else {
		atomic.StoreInt32(&s.ready, 0)
	}
}

// handleHealth returns basic health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status": "healthy",
		"uptime": time.Since(s.startTime).String(),
	}

	_ = json.NewEncoder(w).Encode(response) // Ignore encode errors
}

// handleReady returns readiness status for Kubernetes/Docker
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&s.ready) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready")) // Ignore write errors
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready")) // Ignore write errors
}

// handleMetrics returns Prometheus-compatible metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	// Basic metrics
	players := 0
	stages := 0

	if s.GetConnectedPlayers != nil {
		players = s.GetConnectedPlayers()
	}
	if s.GetActiveStages != nil {
		stages = s.GetActiveStages()
	}

	uptime := time.Since(s.startTime).Seconds()

	// Prometheus format
	metrics := `# HELP jinwoo_connected_players Current number of connected players
# TYPE jinwoo_connected_players gauge
jinwoo_connected_players %d

# HELP jinwoo_active_stages Number of active map stages
# TYPE jinwoo_active_stages gauge
jinwoo_active_stages %d

# HELP jinwoo_uptime_seconds Server uptime in seconds
# TYPE jinwoo_uptime_seconds counter
jinwoo_uptime_seconds %.0f
`
	_, _ = w.Write([]byte(fmt.Sprintf(metrics, players, stages, uptime))) // Ignore write errors
}

