package jobs

import (
	"context"
	"fmt"
	"log"
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

// Job is the full job state stored in the queue.
type Job struct {
	ID          string              `json:"id"`
	Status      string              `json:"status"`
	InputPath   string              `json:"input_path"`
	OutputPath  string              `json:"output_path"`
	FileName    string              `json:"file_name"`
	Profile     config.GifProfile   `json:"profile"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Progress    float64             `json:"progress"`
	Error       string              `json:"error,omitempty"`
	Result      *converter.ConversionResult `json:"result,omitempty"`
}

// Queue manages concurrent conversion jobs with a worker pool.
type Queue struct {
	mu      sync.RWMutex
	jobs    map[string]*Job
	work    chan *Job
	cfg     *config.Config
	conv    *converter.Converter
	wg      sync.WaitGroup
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
		ID:         uuid.New().String(),
		Status:     StatusQueued,
		InputPath:  inputPath,
		OutputPath: outputPath,
		FileName:   fileName,
		Profile:    profile,
		CreatedAt:  time.Now(),
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
	q.setStatus(job.ID, StatusRunning, func(j *Job) { j.StartedAt = &now })

	timeout := time.Duration(q.cfg.Queue.JobTimeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	convJob := &converter.ConversionJob{
		ID:         job.ID,
		InputPath:  job.InputPath,
		OutputPath: job.OutputPath,
		Profile:    job.Profile,
		CreatedAt:  job.CreatedAt,
	}

	result, err := q.conv.Convert(ctx, convJob)

	done := time.Now()
	if err != nil {
		q.setStatus(job.ID, StatusFailed, func(j *Job) {
			j.CompletedAt = &done
			j.Error = err.Error()
			j.Progress = 0
		})
		log.Printf("[worker] Job %s FAILED: %v", job.ID, err)
	} else {
		q.setStatus(job.ID, StatusDone, func(j *Job) {
			j.CompletedAt = &done
			j.Result = result
			j.Progress = 1.0
		})
		log.Printf("[worker] Job %s DONE", job.ID)
	}
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
