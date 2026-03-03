package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/oarkflow/video2gif/internal/config"
	"github.com/oarkflow/video2gif/internal/jobs"
)

// Server wraps the HTTP router, queue, and config.
type Server struct {
	cfg        *config.Config
	queue      *jobs.Queue
	configPath string
}

// New creates a server.
func New(cfg *config.Config, configPath string) *Server {
	return &Server{
		cfg:        cfg,
		queue:      jobs.New(cfg),
		configPath: configPath,
	}
}

// Router builds and returns the HTTP mux.
func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	// Apply middleware
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)
	r.Use(recoveryMiddleware)

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/convert", s.handleConvert).Methods("POST")
	api.HandleFunc("/jobs", s.handleListJobs).Methods("GET")
	api.HandleFunc("/jobs/{id}", s.handleGetJob).Methods("GET")
	api.HandleFunc("/jobs/{id}/download", s.handleDownload).Methods("GET")
	api.HandleFunc("/jobs/{id}", s.handleDeleteJob).Methods("DELETE")
	api.HandleFunc("/profiles", s.handleListProfiles).Methods("GET")
	api.HandleFunc("/profiles/{name}", s.handleGetProfile).Methods("GET")
	api.HandleFunc("/profiles/{name}", s.handleUpdateProfile).Methods("PUT")
	api.HandleFunc("/config", s.handleGetConfig).Methods("GET")
	api.HandleFunc("/config", s.handleUpdateConfig).Methods("PUT")
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/stats", s.handleStats).Methods("GET")
	api.HandleFunc("/probe", s.handleProbe).Methods("POST")
	api.HandleFunc("/save-edited", s.handleSaveEdited).Methods("POST")

	// Serve the single-page web UI
	r.PathPrefix("/").HandlerFunc(s.handleUI)

	return r
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	s.queue.Shutdown()
}

type healthResponse struct {
	Status    string         `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
	Version   string         `json:"version"`
	Queue     map[string]int `json:"queue"`
}
