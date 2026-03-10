package server

import (
	"log"
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
	shares     *ShareStore
	stopCh     chan struct{}
}

// New creates a server.
func New(cfg *config.Config, configPath string) *Server {
	shareStore, err := NewShareStore(cfg.Storage.ShareDir)
	if err != nil {
		log.Printf("Warning: could not initialize share store: %v", err)
	}
	s := &Server{
		cfg:        cfg,
		queue:      jobs.New(cfg),
		configPath: configPath,
		shares:     shareStore,
		stopCh:     make(chan struct{}),
	}
	if s.shares != nil {
		if err := s.shares.CleanupExpired(time.Now()); err != nil {
			log.Printf("Warning: could not clean expired shares: %v", err)
		}
		go s.runShareCleanup()
	}
	return s
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
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/auth/status", s.handleAuthStatus).Methods("GET")
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	api.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	api.HandleFunc("/share/{id}", s.handleGetShareMeta).Methods("GET")
	api.HandleFunc("/share/{id}/video", s.handleGetShareVideo).Methods("GET")

	privateAPI := api.NewRoute().Subrouter()
	privateAPI.Use(s.authMiddleware)
	privateAPI.HandleFunc("/convert", s.handleConvert).Methods("POST")
	privateAPI.HandleFunc("/jobs", s.handleListJobs).Methods("GET")
	privateAPI.HandleFunc("/jobs/{id}", s.handleGetJob).Methods("GET")
	privateAPI.HandleFunc("/jobs/{id}/download", s.handleDownload).Methods("GET")
	privateAPI.HandleFunc("/jobs/{id}", s.handleDeleteJob).Methods("DELETE")
	privateAPI.HandleFunc("/profiles", s.handleListProfiles).Methods("GET")
	privateAPI.HandleFunc("/profiles/{name}", s.handleGetProfile).Methods("GET")
	privateAPI.HandleFunc("/profiles/{name}", s.handleUpdateProfile).Methods("PUT")
	privateAPI.HandleFunc("/config", s.handleGetConfig).Methods("GET")
	privateAPI.HandleFunc("/config", s.handleUpdateConfig).Methods("PUT")
	privateAPI.HandleFunc("/stats", s.handleStats).Methods("GET")
	privateAPI.HandleFunc("/probe", s.handleProbe).Methods("POST")
	privateAPI.HandleFunc("/save-edited", s.handleSaveEdited).Methods("POST")
	privateAPI.HandleFunc("/share", s.handleCreateShare).Methods("POST")

	// Serve the single-page web UI
	r.PathPrefix("/").HandlerFunc(s.handleUI)

	return r
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	close(s.stopCh)
	s.queue.Shutdown()
}

func (s *Server) runShareCleanup() {
	ticker := time.NewTicker(time.Duration(s.cfg.Storage.CleanupIntervalMin) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if s.shares != nil {
				if err := s.shares.CleanupExpired(time.Now()); err != nil {
					log.Printf("Warning: share cleanup failed: %v", err)
				}
			}
		case <-s.stopCh:
			return
		}
	}
}

type healthResponse struct {
	Status    string         `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
	Version   string         `json:"version"`
	Queue     map[string]int `json:"queue"`
}
