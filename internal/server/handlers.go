package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/oarkflow/video2gif/internal/config"
	"github.com/oarkflow/video2gif/internal/converter"
	"github.com/oarkflow/video2gif/internal/jobs"
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

	// Accept explicit format override from form field
	if fmtVal := r.FormValue("format"); fmtVal != "" {
		profile.OutputFormat = fmtVal
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

	// Build output path with correct extension for output format
	outputPath := filepath.Join(s.cfg.Storage.OutputDir, jobID+profile.OutputFormatExt())

	// Parse optional max file size
	var maxFileSize int64
	if raw := r.FormValue("maxFileSize"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil && v > 0 {
			maxFileSize = v
		}
	}

	// Submit to queue
	job, err := s.queue.Submit(uploadPath, outputPath, header.Filename, profile, maxFileSize)
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

func (s *Server) handleJobSSE(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	job, ok := s.queue.Get(id)
	if !ok {
		jsonError(w, "job not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		jsonError(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send initial state immediately.
	initialType := "progress"
	switch job.Status {
	case "done":
		initialType = "completed"
	case "failed":
		initialType = "failed"
	}
	writeSSEEvent(w, flusher, initialType, job)

	// If the job is already terminal, close right away.
	if job.Status == "done" || job.Status == "failed" {
		return
	}

	ch := s.queue.Subscribe(id)
	defer s.queue.Unsubscribe(id, ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			writeSSEEvent(w, flusher, evt.Type, evt.Job)
			if evt.Type == "completed" || evt.Type == "failed" {
				return
			}
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, job *jobs.Job) {
	data, err := json.Marshal(map[string]any{
		"type": eventType,
		"job":  job,
	})
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
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
	s.queue.Delete(id)
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
	downloadName := job.DownloadName
	if strings.TrimSpace(downloadName) == "" {
		base := strings.TrimSuffix(job.FileName, filepath.Ext(job.FileName))
		downloadName = fmt.Sprintf("%s_%s%s", base, job.Profile.Name, job.Profile.OutputFormatExt())
	}
	contentType := strings.TrimSpace(job.ContentType)
	if contentType == "" {
		contentType = job.Profile.OutputContentType()
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	w.Header().Set("Cache-Control", "no-cache")

	stat, _ := f.Stat()
	if stat != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}

	http.ServeContent(w, r, downloadName, time.Now(), f)
}

func (s *Server) handleViewJob(w http.ResponseWriter, r *http.Request) {
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

	name := job.DownloadName
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(job.OutputPath)
	}
	contentType := strings.TrimSpace(job.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(name)))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, name))
	w.Header().Set("Cache-Control", "no-cache")

	stat, _ := f.Stat()
	if stat != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}

	http.ServeContent(w, r, name, time.Now(), f)
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
	cfg.ApplyDefaults()
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
	if err := os.MkdirAll(s.cfg.Storage.ShareDir, 0755); err != nil {
		jsonError(w, "failed to create share dir", http.StatusInternalServerError)
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

	job, err := s.queue.SubmitEditedVideo(inputPath, outputPath, header.Filename, cutRanges, durationHint)
	if err != nil {
		_ = os.Remove(inputPath)
		_ = os.Remove(outputPath)
		jsonError(w, "queue full: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	jsonOK(w, job, http.StatusAccepted)
}

func (s *Server) handleCreateShare(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.Sharing.Enabled || s.shares == nil {
		jsonError(w, "sharing is disabled", http.StatusServiceUnavailable)
		return
	}
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
	var comments []ShareComment
	if raw := r.FormValue("comments"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &comments); err != nil {
			jsonError(w, "invalid comments JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	createdBy := strings.TrimSpace(r.FormValue("created_by"))
	expiresAt, err := s.resolveShareExpiry(r.FormValue("expires_in_hours"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(s.cfg.Storage.TempDir, 0755); err != nil {
		jsonError(w, "failed to create temp dir", http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	inputPath := filepath.Join(s.cfg.Storage.TempDir, "share_src_"+id+ext)
	videoPath := filepath.Join(s.cfg.Storage.ShareDir, id+".mp4")
	dst, err := os.Create(inputPath)
	if err != nil {
		jsonError(w, "failed to create share file", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(inputPath)
		jsonError(w, "failed to write share file", http.StatusInternalServerError)
		return
	}
	dst.Close()
	defer os.Remove(inputPath)

	if _, err := converter.SaveEditedVideo(r.Context(), inputPath, videoPath, cutRanges, durationHint, nil); err != nil {
		os.Remove(videoPath)
		jsonError(w, "failed to build shared video: "+err.Error(), http.StatusBadRequest)
		return
	}

	inputDuration := durationHint
	if info, err := converter.ProbeVideo(r.Context(), inputPath); err == nil && info != nil && info.Duration > 0 {
		inputDuration = info.Duration
	}
	sharedComments := remapCommentsForCuts(comments, cutRanges, inputDuration)
	now := time.Now()
	for i := range sharedComments {
		if strings.TrimSpace(sharedComments[i].Author) == "" {
			sharedComments[i].Author = createdBy
		}
		if sharedComments[i].CreatedAt.IsZero() {
			sharedComments[i].CreatedAt = now
		}
	}

	base := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	if strings.TrimSpace(base) == "" {
		base = "shared-video"
	}
	sharedFileName := base + "_shared.mp4"

	sess := &ShareSession{
		ID:         id,
		FileName:   sharedFileName,
		VideoPath:  videoPath,
		CutRanges:  nil,
		Comments:   sharedComments,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		CreatedBy:  createdBy,
		PublicView: s.cfg.Sharing.PublicView,
	}
	if err := s.shares.Save(sess); err != nil {
		os.Remove(videoPath)
		jsonError(w, "failed to persist share: "+err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := "http://" + r.Host
	if r.TLS != nil {
		baseURL = "https://" + r.Host
	}

	jsonOK(w, map[string]any{
		"id":         id,
		"share_url":  baseURL + "/?share=" + id,
		"expires_at": expiresAt,
	}, http.StatusOK)
}

func (s *Server) handleGetShareMeta(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sess, ok, status, err := s.loadShareForRequest(id, r)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}
	if !ok {
		jsonError(w, "share not found", http.StatusNotFound)
		return
	}

	baseURL := "http://" + r.Host
	if r.TLS != nil {
		baseURL = "https://" + r.Host
	}
	jsonOK(w, map[string]any{
		"id":          sess.ID,
		"file_name":   sess.FileName,
		"cut_ranges":  sess.CutRanges,
		"comments":    sess.Comments,
		"video_url":   baseURL + "/api/v1/share/" + id + "/video",
		"created_at":  sess.CreatedAt,
		"expires_at":  sess.ExpiresAt,
		"created_by":  sess.CreatedBy,
		"public_view": sess.PublicView,
	}, http.StatusOK)
}

func (s *Server) handleGetShareVideo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sess, ok, status, err := s.loadShareForRequest(id, r)
	if err != nil {
		jsonError(w, err.Error(), status)
		return
	}
	if !ok {
		jsonError(w, "share not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(sess.VideoPath)
	if err != nil {
		jsonError(w, "shared video not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	downloadName := strings.TrimSpace(sess.FileName)
	if downloadName == "" {
		downloadName = "shared-video.webm"
	}
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(downloadName)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, downloadName))
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, downloadName, sess.CreatedAt, f)
}

func (s *Server) resolveShareExpiry(raw string) (time.Time, error) {
	hours := s.cfg.Sharing.DefaultExpiryHours
	if v := strings.TrimSpace(raw); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid share expiry")
		}
		hours = parsed
	}
	if hours <= 0 {
		return time.Time{}, fmt.Errorf("share expiry must be greater than zero")
	}
	if s.cfg.Sharing.MaxExpiryHours > 0 && hours > s.cfg.Sharing.MaxExpiryHours {
		return time.Time{}, fmt.Errorf("share expiry exceeds configured maximum")
	}
	return time.Now().Add(time.Duration(hours) * time.Hour), nil
}

func (s *Server) loadShareForRequest(id string, r *http.Request) (*ShareSession, bool, int, error) {
	if s.shares == nil {
		return nil, false, http.StatusNotFound, nil
	}
	sess, ok, err := s.shares.Get(id, time.Now())
	if err != nil {
		return nil, false, http.StatusInternalServerError, fmt.Errorf("failed to load share")
	}
	if !ok {
		return nil, false, http.StatusNotFound, nil
	}
	if sess.PublicView {
		return sess, true, http.StatusOK, nil
	}
	okAuth, _ := s.isAuthenticated(r)
	if !okAuth {
		return nil, false, http.StatusUnauthorized, fmt.Errorf("authentication required")
	}
	return sess, true, http.StatusOK, nil
}

func remapCommentsForCuts(comments []ShareComment, cuts []config.ClipSegment, duration float64) []ShareComment {
	if len(comments) == 0 {
		return nil
	}
	norm := normalizeClipSegments(cuts, duration)
	if len(norm) == 0 {
		out := make([]ShareComment, len(comments))
		copy(out, comments)
		return out
	}
	out := make([]ShareComment, 0, len(comments))
	for _, c := range comments {
		newTime, ok := remapTimeAfterCuts(c.Time, norm)
		if !ok {
			continue
		}
		cc := c
		cc.Time = newTime
		out = append(out, cc)
	}
	return out
}

func remapTimeAfterCuts(t float64, cuts []config.ClipSegment) (float64, bool) {
	removed := 0.0
	for _, c := range cuts {
		if t < c.Start {
			break
		}
		if t >= c.End {
			removed += c.End - c.Start
			continue
		}
		if t >= c.Start && t < c.End {
			return 0, false
		}
	}
	v := t - removed
	if v < 0 {
		v = 0
	}
	return v, true
}

func normalizeClipSegments(in []config.ClipSegment, duration float64) []config.ClipSegment {
	if len(in) == 0 {
		return nil
	}
	segs := make([]config.ClipSegment, 0, len(in))
	for _, s := range in {
		start, end := s.Start, s.End
		if end <= start {
			continue
		}
		if duration > 0 {
			if start < 0 {
				start = 0
			}
			if end > duration {
				end = duration
			}
		}
		if end > start {
			segs = append(segs, config.ClipSegment{Start: start, End: end})
		}
	}
	if len(segs) == 0 {
		return nil
	}
	sort.Slice(segs, func(i, j int) bool { return segs[i].Start < segs[j].Start })
	merged := []config.ClipSegment{segs[0]}
	for i := 1; i < len(segs); i++ {
		last := &merged[len(merged)-1]
		cur := segs[i]
		if cur.Start <= last.End {
			if cur.End > last.End {
				last.End = cur.End
			}
		} else {
			merged = append(merged, cur)
		}
	}
	return merged
}

func (s *Server) handleEstimate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Duration        float64 `json:"duration"`
		Width           int     `json:"width"`
		Height          int     `json:"height"`
		FPS             float64 `json:"fps"`
		Colors          int     `json:"colors"`
		Dither          string  `json:"dither"`
		SpeedMultiplier float64 `json:"speed_multiplier"`
		SourceWidth     int     `json:"source_width"`
		SourceHeight    int     `json:"source_height"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	info := &converter.VideoInfo{
		Duration: req.Duration,
		Width:    req.SourceWidth,
		Height:   req.SourceHeight,
		FPS:      req.FPS,
	}
	if info.Width <= 0 {
		info.Width = req.Width
	}
	if info.Height <= 0 {
		info.Height = req.Height
	}

	profile := &config.GifProfile{
		FPS:             req.FPS,
		Width:           req.Width,
		Height:          req.Height,
		Colors:          req.Colors,
		Dither:          req.Dither,
		SpeedMultiplier: req.SpeedMultiplier,
		Duration:        fmt.Sprintf("%.6f", req.Duration),
	}
	if profile.Colors < 2 {
		profile.Colors = 256
	}
	if profile.FPS <= 0 {
		profile.FPS = 20
	}
	if profile.SpeedMultiplier <= 0 {
		profile.SpeedMultiplier = 1.0
	}

	estimated := converter.EstimateOutputSize(info, profile)
	jsonOK(w, map[string]any{
		"estimated_size":       estimated,
		"estimated_size_human": converter.FormatBytes(estimated),
	}, http.StatusOK)
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
