package jobs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/oarkflow/video2gif/internal/config"
	"github.com/oarkflow/video2gif/internal/converter"
)

// Status constants
const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
)

const (
	KindGIF   = "gif"
	KindVideo = "video"
)

// Job is the full job state stored in the queue.
type Job struct {
	ID           string                      `json:"id"`
	Kind         string                      `json:"kind"`
	Status       string                      `json:"status"`
	InputPath    string                      `json:"input_path"`
	OutputPath   string                      `json:"output_path"`
	FileName     string                      `json:"file_name"`
	Profile      config.GifProfile           `json:"profile"`
	CreatedAt    time.Time                   `json:"created_at"`
	StartedAt    *time.Time                  `json:"started_at,omitempty"`
	CompletedAt  *time.Time                  `json:"completed_at,omitempty"`
	Progress     float64                     `json:"progress"`
	Stage        string                      `json:"stage,omitempty"`
	Detail       string                      `json:"detail,omitempty"`
	Error        string                      `json:"error,omitempty"`
	DownloadName string                      `json:"download_name,omitempty"`
	ContentType  string                      `json:"content_type,omitempty"`
	Result       *converter.ConversionResult `json:"result,omitempty"`

	cutRanges    []config.ClipSegment
	durationHint float64
}

// Queue manages concurrent conversion jobs with a worker pool.
type Queue struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	work chan *Job
	cfg  *config.Config
	conv *converter.Converter
	wg   sync.WaitGroup
}

// New creates and starts a job queue with cfg.Queue.Workers workers.
func New(cfg *config.Config) *Queue {
	q := &Queue{
		jobs: make(map[string]*Job),
		work: make(chan *Job, cfg.Queue.MaxQueueSize),
		cfg:  cfg,
		conv: converter.NewConverter(cfg),
	}
	for i := 0; i < cfg.Queue.Workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	go q.janitor()
	return q
}

// Submit enqueues a new conversion job. Returns the job ID.
func (q *Queue) Submit(inputPath, outputPath, fileName string, profile config.GifProfile) (*Job, error) {
	job := &Job{
		ID:           uuid.New().String(),
		Kind:         KindGIF,
		Status:       StatusQueued,
		InputPath:    inputPath,
		OutputPath:   outputPath,
		FileName:     fileName,
		Profile:      profile,
		CreatedAt:    time.Now(),
		Stage:        "Queued",
		Detail:       "Waiting for an available worker",
		DownloadName: defaultGIFDownloadName(fileName, profile.Name),
		ContentType:  "image/gif",
	}

	q.mu.Lock()
	q.jobs[job.ID] = job
	q.mu.Unlock()

	select {
	case q.work <- job:
		log.Printf("[queue] Job %s submitted (%s)", job.ID, fileName)
		return job, nil
	default:
		q.mu.Lock()
		delete(q.jobs, job.ID)
		q.mu.Unlock()
		return nil, fmt.Errorf("queue is full (%d jobs)", q.cfg.Queue.MaxQueueSize)
	}
}

// SubmitEditedVideo enqueues a video export job that removes the provided cut ranges.
func (q *Queue) SubmitEditedVideo(inputPath, outputPath, fileName string, cutRanges []config.ClipSegment, durationHint float64) (*Job, error) {
	job := &Job{
		ID:           uuid.New().String(),
		Kind:         KindVideo,
		Status:       StatusQueued,
		InputPath:    inputPath,
		OutputPath:   outputPath,
		FileName:     fileName,
		CreatedAt:    time.Now(),
		Stage:        "Queued",
		Detail:       "Waiting for an available worker",
		DownloadName: defaultEditedDownloadName(fileName),
		ContentType:  "video/mp4",
		cutRanges:    append([]config.ClipSegment(nil), cutRanges...),
		durationHint: durationHint,
	}

	q.mu.Lock()
	q.jobs[job.ID] = job
	q.mu.Unlock()

	select {
	case q.work <- job:
		log.Printf("[queue] Video export job %s submitted (%s)", job.ID, fileName)
		return job, nil
	default:
		q.mu.Lock()
		delete(q.jobs, job.ID)
		q.mu.Unlock()
		return nil, fmt.Errorf("queue is full (%d jobs)", q.cfg.Queue.MaxQueueSize)
	}
}

// Get returns a job by ID.
func (q *Queue) Get(id string) (*Job, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	j, ok := q.jobs[id]
	return j, ok
}

// List returns all jobs sorted by creation time (newest first).
func (q *Queue) List() []*Job {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]*Job, 0, len(q.jobs))
	for _, j := range q.jobs {
		out = append(out, j)
	}
	return out
}

// Stats returns queue statistics.
func (q *Queue) Stats() map[string]int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	s := map[string]int{"queued": 0, "running": 0, "done": 0, "failed": 0, "total": len(q.jobs)}
	for _, j := range q.jobs {
		s[j.Status]++
	}
	return s
}

func (q *Queue) worker(id int) {
	defer q.wg.Done()
	log.Printf("[worker %d] started", id)
	for job := range q.work {
		q.process(job)
	}
	log.Printf("[worker %d] stopped", id)
}

func (q *Queue) process(job *Job) {
	now := time.Now()
	q.setStatus(job.ID, StatusRunning, func(j *Job) {
		j.StartedAt = &now
		j.Stage = "Starting"
		j.Detail = "Preparing job"
		j.Progress = 0.01
	})

	timeout := time.Duration(q.cfg.Queue.JobTimeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	progressFn := func(update converter.ProgressUpdate) {
		q.setStatus(job.ID, StatusRunning, func(j *Job) {
			j.Progress = update.Fraction
			if strings.TrimSpace(update.Stage) != "" {
				j.Stage = update.Stage
			}
			if strings.TrimSpace(update.Detail) != "" {
				j.Detail = update.Detail
			}
		})
	}

	var (
		result *converter.ConversionResult
		err    error
	)

	switch job.Kind {
	case KindVideo:
		result, err = converter.SaveEditedVideo(ctx, job.InputPath, job.OutputPath, job.cutRanges, job.durationHint, progressFn)
	case KindGIF:
		convJob := &converter.ConversionJob{
			ID:         job.ID,
			InputPath:  job.InputPath,
			OutputPath: job.OutputPath,
			Profile:    job.Profile,
			CreatedAt:  job.CreatedAt,
		}
		result, err = q.conv.ConvertWithProgress(ctx, convJob, progressFn)
	default:
		err = fmt.Errorf("unsupported job kind %q", job.Kind)
	}

	done := time.Now()
	if err != nil {
		_ = os.Remove(job.OutputPath)
		q.setStatus(job.ID, StatusFailed, func(j *Job) {
			j.CompletedAt = &done
			j.Error = err.Error()
			j.Progress = 0
			if strings.TrimSpace(j.Stage) == "" {
				j.Stage = "Failed"
			}
			if strings.TrimSpace(j.Detail) == "" {
				j.Detail = "Operation did not complete"
			}
		})
		log.Printf("[worker] Job %s FAILED: %v", job.ID, err)
	} else {
		q.setStatus(job.ID, StatusDone, func(j *Job) {
			j.CompletedAt = &done
			j.Result = result
			j.Progress = 1.0
			j.Stage = "Complete"
			if j.Kind == KindVideo {
				j.Detail = "Edited video ready for download"
			} else {
				j.Detail = "GIF ready for download"
			}
		})
		log.Printf("[worker] Job %s DONE", job.ID)
	}
	_ = os.Remove(job.InputPath)
}

func (q *Queue) setStatus(id, status string, fn func(*Job)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if j, ok := q.jobs[id]; ok {
		j.Status = status
		if fn != nil {
			fn(j)
		}
	}
}

// janitor removes completed/failed jobs older than MaxAgeHours.
func (q *Queue) janitor() {
	interval := time.Duration(q.cfg.Storage.CleanupIntervalMin) * time.Minute
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		maxAge := time.Duration(q.cfg.Storage.MaxAgeHours) * time.Hour
		cutoff := time.Now().Add(-maxAge)
		q.mu.Lock()
		for id, j := range q.jobs {
			if (j.Status == StatusDone || j.Status == StatusFailed) && j.CreatedAt.Before(cutoff) {
				_ = os.Remove(j.InputPath)
				_ = os.Remove(j.OutputPath)
				delete(q.jobs, id)
			}
		}
		q.mu.Unlock()
	}
}

// Shutdown drains the work channel and waits for workers.
func (q *Queue) Shutdown() {
	close(q.work)
	q.wg.Wait()
}

func defaultGIFDownloadName(fileName, profileName string) string {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if strings.TrimSpace(base) == "" {
		base = "video"
	}
	if strings.TrimSpace(profileName) == "" {
		profileName = "gif"
	}
	return fmt.Sprintf("%s_%s.gif", base, profileName)
}

func defaultEditedDownloadName(fileName string) string {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if strings.TrimSpace(base) == "" {
		base = "recording"
	}
	return base + "_edited.mp4"
}
