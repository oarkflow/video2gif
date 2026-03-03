package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/oarkflow/video2gif/internal/config"
	"github.com/oarkflow/video2gif/internal/converter"
)

// Allowed video MIME types
var allowedMIME = map[string]bool{
	"video/mp4": true, "video/x-m4v": true, "video/quicktime": true,
	"video/x-msvideo": true, "video/x-matroska": true, "video/webm": true,
	"video/mpeg": true, "video/ogg": true, "video/3gpp": true, "video/3gpp2": true,
	"video/x-flv": true, "video/x-ms-wmv": true, "application/octet-stream": true,
}

var allowedExt = map[string]bool{
	".mp4": true, ".m4v": true, ".mov": true, ".avi": true,
	".mkv": true, ".webm": true, ".mpeg": true, ".mpg": true,
	".ogv": true, ".3gp": true, ".flv": true, ".wmv": true,
	".ts": true, ".mts": true, ".m2ts": true,
}

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Server.MaxUploadBytes)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "request too large or malformed", http.StatusBadRequest)
		return
	}

	// Validate file
	file, header, err := r.FormFile("video")
	if err != nil {
		jsonError(w, "missing 'video' field in form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExt[ext] {
		jsonError(w, fmt.Sprintf("unsupported file extension: %s", ext), http.StatusBadRequest)
		return
	}

	// Determine profile
	profileName := r.FormValue("profile")
	if profileName == "" {
		profileName = s.cfg.DefaultProfile
	}

	profile, ok := s.cfg.GetProfile(profileName)
	if !ok {
		jsonError(w, fmt.Sprintf("unknown profile: %s", profileName), http.StatusBadRequest)
		return
	}

	// Allow per-request profile overrides via JSON field "params"
	if raw := r.FormValue("params"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &profile); err != nil {
			jsonError(w, "invalid params JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		profile.Name = "custom"
		if len(profile.KeepSegments) > 0 {
			// keep_segments takes precedence over legacy single-range trimming.
			profile.StartTime = ""
			profile.Duration = ""
		}
	}

	// Save upload
	jobID := uuid.New().String()
	uploadPath := filepath.Join(s.cfg.Storage.UploadDir, jobID+ext)

	if err := os.MkdirAll(s.cfg.Storage.UploadDir, 0755); err != nil {
		jsonError(w, "failed to create upload directory", http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(uploadPath)
	if err != nil {
		jsonError(w, "failed to save upload", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(uploadPath)
		jsonError(w, "failed to write upload", http.StatusInternalServerError)
		return
	}
	dst.Close()

	// Build output path
	outputPath := filepath.Join(s.cfg.Storage.OutputDir, jobID+".gif")

	// Submit to queue
	job, err := s.queue.Submit(uploadPath, outputPath, header.Filename, profile)
	if err != nil {
		os.Remove(uploadPath)
		jsonError(w, "queue full: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	jsonOK(w, job, http.StatusAccepted)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	job, ok := s.queue.Get(id)
	if !ok {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	jsonOK(w, job, http.StatusOK)
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.queue.List(), http.StatusOK)
}

func (s *Server) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	job, ok := s.queue.Get(id)
	if !ok {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	// Clean up files
	_ = os.Remove(job.InputPath)
	_ = os.Remove(job.OutputPath)
	jsonOK(w, map[string]string{"deleted": id}, http.StatusOK)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	job, ok := s.queue.Get(id)
	if !ok {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}
	if job.Status != "done" {
		jsonError(w, "job not completed yet, status: "+job.Status, http.StatusConflict)
		return
	}

	f, err := os.Open(job.OutputPath)
	if err != nil {
		jsonError(w, "output file not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Construct download filename
	base := strings.TrimSuffix(job.FileName, filepath.Ext(job.FileName))
	downloadName := fmt.Sprintf("%s_%s.gif", base, job.Profile.Name)

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	w.Header().Set("Cache-Control", "no-cache")

	stat, _ := f.Stat()
	if stat != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}

	http.ServeContent(w, r, downloadName, time.Now(), f)
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.cfg.Profiles, http.StatusOK)
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	p, ok := s.cfg.Profiles[name]
	if !ok {
		jsonError(w, "profile not found", http.StatusNotFound)
		return
	}
	jsonOK(w, p, http.StatusOK)
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var p config.GifProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	p.Name = name
	s.cfg.Profiles[name] = p
	if err := s.cfg.Save(s.configPath); err != nil {
		log.Printf("Warning: could not save config: %v", err)
	}
	jsonOK(w, p, http.StatusOK)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.cfg, http.StatusOK)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := cfg.Validate(); err != nil {
		jsonError(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}
	*s.cfg = cfg
	if err := s.cfg.Save(s.configPath); err != nil {
		log.Printf("Warning: could not save config: %v", err)
	}
	jsonOK(w, s.cfg, http.StatusOK)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	_, _, err := converter.CheckFFmpeg()
	status := "ok"
	if err != nil {
		status = "degraded: ffmpeg unavailable"
	}
	jsonOK(w, healthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Queue:     s.queue.Stats(),
	}, http.StatusOK)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.queue.Stats(), http.StatusOK)
}

func (s *Server) handleProbe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Server.MaxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "request too large", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("video")
	if err != nil {
		jsonError(w, "missing 'video' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	tmpPath := filepath.Join(s.cfg.Storage.TempDir, "probe_"+uuid.New().String()+ext)
	defer os.Remove(tmpPath)

	if err := os.MkdirAll(s.cfg.Storage.TempDir, 0755); err != nil {
		jsonError(w, "failed to create temp dir", http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(tmpPath)
	if err != nil {
		jsonError(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}
	io.Copy(dst, file)
	dst.Close()

	info, err := converter.ProbeVideo(r.Context(), tmpPath)
	if err != nil {
		jsonError(w, "probe failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, info, http.StatusOK)
}

func (s *Server) handleSaveEdited(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.Server.MaxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "request too large or malformed", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		jsonError(w, "missing 'video' field in form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".webm"
	}
	if !allowedExt[ext] {
		jsonError(w, fmt.Sprintf("unsupported file extension: %s", ext), http.StatusBadRequest)
		return
	}

	var cutRanges []config.ClipSegment
	if raw := r.FormValue("cut_ranges"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &cutRanges); err != nil {
			jsonError(w, "invalid cut_ranges JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	var durationHint float64
	if raw := strings.TrimSpace(r.FormValue("duration_hint")); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 {
			durationHint = v
		}
	}

	if err := os.MkdirAll(s.cfg.Storage.TempDir, 0755); err != nil {
		jsonError(w, "failed to create temp dir", http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	inputPath := filepath.Join(s.cfg.Storage.TempDir, "edit_in_"+id+ext)
	outputPath := filepath.Join(s.cfg.Storage.TempDir, "edit_out_"+id+".mp4")
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	dst, err := os.Create(inputPath)
	if err != nil {
		jsonError(w, "failed to create temp input", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		jsonError(w, "failed to write temp input", http.StatusInternalServerError)
		return
	}
	dst.Close()

	if err := converter.SaveEditedVideo(r.Context(), inputPath, outputPath, cutRanges, durationHint); err != nil {
		jsonError(w, "save failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	out, err := os.Open(outputPath)
	if err != nil {
		jsonError(w, "failed to open output file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	base := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	if base == "" {
		base = "recording"
	}
	downloadName := base + "_edited.mp4"
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	w.Header().Set("Cache-Control", "no-cache")
	if st, err := out.Stat(); err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", st.Size()))
	}
	http.ServeContent(w, r, downloadName, time.Now(), out)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}

// ---- helpers ----

func jsonOK(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
