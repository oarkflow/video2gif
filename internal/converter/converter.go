package converter

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oarkflow/video2gif/internal/config"
)

// Converter handles the full video-to-GIF pipeline.
type Converter struct {
	cfg *config.Config
}

// NewConverter creates a new Converter.
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{cfg: cfg}
}

// ConversionJob represents a single conversion request.
type ConversionJob struct {
	ID          string
	InputPath   string
	OutputPath  string
	Profile     config.GifProfile
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Progress    float64 // 0.0 – 1.0
	Error       string
	Status      string // queued | running | done | failed
}

// ConversionResult is returned on success.
type ConversionResult struct {
	OutputPath string
	OutputSize int64
	Duration   time.Duration
	FrameCount int
	VideoInfo  *VideoInfo
}

// Convert runs the full two-pass GIF conversion pipeline.
func (c *Converter) Convert(ctx context.Context, job *ConversionJob) (*ConversionResult, error) {
	start := time.Now()

	log.Printf("[%s] Starting conversion: %s → %s (profile: %s)",
		job.ID, filepath.Base(job.InputPath), filepath.Base(job.OutputPath), job.Profile.Name)

	// Probe input
	info, err := ProbeVideo(ctx, job.InputPath)
	if err != nil {
		return nil, fmt.Errorf("probe video: %w", err)
	}
	log.Printf("[%s] Input: %dx%d @ %.2ffps, duration=%.2fs, codec=%s",
		job.ID, info.Width, info.Height, info.FPS, info.Duration, info.Codec)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	var frameCount int

	if job.Profile.OptimizePalette {
		// Two-pass conversion
		frameCount, err = c.convertTwoPass(ctx, job)
	} else {
		// Single-pass conversion
		frameCount, err = c.convertSinglePass(ctx, job)
	}

	if err != nil {
		return nil, err
	}

	// Get output file size
	stat, err := os.Stat(job.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("stat output: %w", err)
	}

	elapsed := time.Since(start)
	log.Printf("[%s] Done in %s, output size: %s",
		job.ID, elapsed.Round(time.Millisecond), FormatBytes(stat.Size()))

	return &ConversionResult{
		OutputPath: job.OutputPath,
		OutputSize: stat.Size(),
		Duration:   elapsed,
		FrameCount: frameCount,
		VideoInfo:  info,
	}, nil
}

func (c *Converter) convertTwoPass(ctx context.Context, job *ConversionJob) (int, error) {
	palettePath := PalettePath(c.cfg.Storage.TempDir, job.ID)
	defer CleanupPalette(palettePath)

	p := job.Profile

	// ---- PASS 1: Generate palette PNG ----
	pass1Args := buildCommonArgs(job)
	pass1Filter, _ := BuildPaletteFilter(p)
	pass1Args = append(pass1Args,
		"-vf", pass1Filter,
		"-update", "1",
		"-y",
		palettePath,
	)

	log.Printf("[%s] Pass 1: generating palette...", job.ID)
	if err := runFFmpeg(ctx, pass1Args, job.ID+" pass1"); err != nil {
		return 0, fmt.Errorf("palette generation failed: %w", err)
	}

	// ---- PASS 2: Apply palette and produce GIF ----
	pass2Args := buildCommonArgs(job)
	pass2Args = append(pass2Args, "-i", palettePath) // second input: palette
	_, pass2Filter := BuildPaletteFilter(p)

	pass2Args = append(pass2Args,
		"-lavfi", pass2Filter,
	)

	// FPS + scale on video stream for pass2
	fpsFilter := fmt.Sprintf("fps=%.4f", clampFPS(p.FPS))
	scaleFilter := buildScaleFilter(p)
	var speedFilter string
	if p.SpeedMultiplier != 0 && p.SpeedMultiplier != 1.0 {
		pts := 1.0 / p.SpeedMultiplier
		speedFilter = fmt.Sprintf(",setpts=%.4f*PTS", pts)
	}
	_ = fpsFilter
	_ = scaleFilter
	_ = speedFilter // used inside BuildPaletteFilter

	pass2Args = append(pass2Args,
		"-loop", strconv.Itoa(p.Loop),
		"-y",
		job.OutputPath,
	)

	log.Printf("[%s] Pass 2: rendering GIF...", job.ID)
	if err := runFFmpeg(ctx, pass2Args, job.ID+" pass2"); err != nil {
		return 0, fmt.Errorf("GIF rendering failed: %w", err)
	}

	return estimateFrameCount(job.Profile, job.InputPath, ctx), nil
}

func (c *Converter) convertSinglePass(ctx context.Context, job *ConversionJob) (int, error) {
	p := job.Profile
	_, singleFilter := BuildPaletteFilter(p)

	args := buildCommonArgs(job)
	args = append(args,
		"-lavfi", singleFilter,
		"-loop", strconv.Itoa(p.Loop),
		"-y",
		job.OutputPath,
	)

	log.Printf("[%s] Single pass: rendering GIF...", job.ID)
	if err := runFFmpeg(ctx, args, job.ID); err != nil {
		return 0, fmt.Errorf("GIF rendering failed: %w", err)
	}

	return estimateFrameCount(p, job.InputPath, ctx), nil
}

// buildCommonArgs builds the base ffmpeg arguments shared between passes.
func buildCommonArgs(job *ConversionJob) []string {
	p := job.Profile
	args := []string{"-hide_banner", "-loglevel", "info"}

	useSegments := len(p.KeepSegments) > 0

	// Seek (fast seek before input, then fine seek after)
	if !useSegments && p.StartTime != "" {
		args = append(args, "-ss", p.StartTime)
	}

	args = append(args, "-i", job.InputPath)

	// Duration limit
	if !useSegments && p.Duration != "" {
		args = append(args, "-t", p.Duration)
	}

	return args
}

// runFFmpeg executes an ffmpeg command and streams stderr for logging.
func runFFmpeg(ctx context.Context, args []string, label string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Extract meaningful error from stderr
		errMsg := extractFFmpegError(stderr.String())
		return fmt.Errorf("[%s] %w\nffmpeg: %s", label, err, errMsg)
	}
	return nil
}

func extractFFmpegError(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "Error") || strings.HasPrefix(line, "Invalid") ||
			strings.HasPrefix(line, "No such") || strings.HasPrefix(line, "Unable") {
			return line
		}
	}
	if len(lines) > 3 {
		return strings.Join(lines[len(lines)-3:], " | ")
	}
	return stderr
}

func estimateFrameCount(p config.GifProfile, inputPath string, ctx context.Context) int {
	info, err := ProbeVideo(ctx, inputPath)
	if err != nil || info.Duration == 0 {
		return 0
	}
	duration := effectiveDurationSeconds(info.Duration, p)
	if duration <= 0 {
		return 0
	}
	return int(math.Round(clampFPS(p.FPS) * duration))
}

func effectiveDurationSeconds(full float64, p config.GifProfile) float64 {
	if full <= 0 {
		return 0
	}
	if len(p.KeepSegments) > 0 {
		segments := make([]config.ClipSegment, 0, len(p.KeepSegments))
		for _, seg := range p.KeepSegments {
			if seg.End <= seg.Start {
				continue
			}
			start := math.Max(0, seg.Start)
			end := math.Min(full, seg.End)
			if end > start {
				segments = append(segments, config.ClipSegment{Start: start, End: end})
			}
		}
		if len(segments) == 0 {
			return 0
		}
		sort.Slice(segments, func(i, j int) bool { return segments[i].Start < segments[j].Start })
		total := 0.0
		curStart := segments[0].Start
		curEnd := segments[0].End
		for i := 1; i < len(segments); i++ {
			if segments[i].Start <= curEnd {
				if segments[i].End > curEnd {
					curEnd = segments[i].End
				}
				continue
			}
			total += curEnd - curStart
			curStart = segments[i].Start
			curEnd = segments[i].End
		}
		total += curEnd - curStart
		return applySpeed(total, p.SpeedMultiplier)
	}

	segmentDuration := full
	if p.Duration != "" {
		if v, err := parseTimeLike(p.Duration); err == nil && v > 0 {
			segmentDuration = v
		}
	}
	return applySpeed(segmentDuration, p.SpeedMultiplier)
}

func applySpeed(duration, speed float64) float64 {
	if speed <= 0 {
		return duration
	}
	return duration / speed
}

func parseTimeLike(v string) (float64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, fmt.Errorf("empty time")
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f, nil
	}
	parts := strings.Split(v, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("invalid time format")
	}
	total := 0.0
	mult := 1.0
	for i := len(parts) - 1; i >= 0; i-- {
		val, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return 0, err
		}
		total += val * mult
		mult *= 60
	}
	return total, nil
}

// FormatBytes returns a human-readable file size.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
