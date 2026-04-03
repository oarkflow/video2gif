package converter

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
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
	MaxFileSize int64 // Target max output size in bytes; 0 = no limit
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
	return c.ConvertWithProgress(ctx, job, nil)
}

// ConvertWithProgress runs the full two-pass GIF conversion pipeline and emits progress updates.
// If MaxFileSize is set on the job, it will iteratively reduce quality to meet the target.
func (c *Converter) ConvertWithProgress(ctx context.Context, job *ConversionJob, onProgress ProgressFunc) (*ConversionResult, error) {
	result, err := c.convertOnce(ctx, job, onProgress)
	if err != nil {
		return nil, err
	}

	// If no size target or already under target, return immediately
	if job.MaxFileSize <= 0 || result.OutputSize <= job.MaxFileSize {
		return result, nil
	}

	// Iterative size reduction: up to 3 retries
	type adjustment struct {
		desc string
		fn   func(p *config.GifProfile)
	}

	adjustments := []adjustment{
		{"reducing colors", func(p *config.GifProfile) {
			p.Colors = p.Colors / 2
			if p.Colors < 16 {
				p.Colors = 16
			}
		}},
		{"lowering FPS", func(p *config.GifProfile) {
			p.FPS = math.Max(5, p.FPS*0.6)
		}},
		{"reducing resolution", func(p *config.GifProfile) {
			if p.Width > 0 {
				p.Width = int(float64(p.Width) * 0.7)
			} else {
				p.Width = 320
			}
			p.Height = -1 // keep aspect ratio
		}},
	}

	for i, adj := range adjustments {
		if result.OutputSize <= job.MaxFileSize {
			break
		}
		log.Printf("[%s] Output %s exceeds target %s — %s (retry %d/%d)",
			job.ID, FormatBytes(result.OutputSize), FormatBytes(job.MaxFileSize), adj.desc, i+1, len(adjustments))

		reportProgress(onProgress, 0.02, "Retrying", fmt.Sprintf("File too large (%s > %s), %s…", FormatBytes(result.OutputSize), FormatBytes(job.MaxFileSize), adj.desc))

		adj.fn(&job.Profile)
		_ = os.Remove(job.OutputPath)

		result, err = c.convertOnce(ctx, job, onProgress)
		if err != nil {
			return nil, err
		}
	}

	if job.MaxFileSize > 0 && result.OutputSize > job.MaxFileSize {
		log.Printf("[%s] Warning: could not meet target %s, final size is %s",
			job.ID, FormatBytes(job.MaxFileSize), FormatBytes(result.OutputSize))
	}

	return result, nil
}

// convertOnce runs a single conversion attempt.
func (c *Converter) convertOnce(ctx context.Context, job *ConversionJob, onProgress ProgressFunc) (*ConversionResult, error) {
	start := time.Now()

	format := job.Profile.NormalizedOutputFormat()
	formatLabel := strings.ToUpper(format)

	log.Printf("[%s] Starting %s conversion: %s → %s (profile: %s)",
		job.ID, formatLabel, filepath.Base(job.InputPath), filepath.Base(job.OutputPath), job.Profile.Name)
	reportProgress(onProgress, 0.02, "Probing input", "Inspecting source video")

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

	switch format {
	case "webp":
		frameCount, err = c.convertWebP(ctx, job, info.Duration, onProgress)
	case "apng":
		frameCount, err = c.convertAPNG(ctx, job, info.Duration, onProgress)
	default:
		// GIF: existing two-pass or single-pass pipeline
		if job.Profile.OptimizePalette {
			frameCount, err = c.convertTwoPass(ctx, job, info.Duration, onProgress)
		} else {
			frameCount, err = c.convertSinglePass(ctx, job, info.Duration, onProgress)
		}
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
	reportProgress(onProgress, 1.0, "Complete", formatLabel+" ready")

	return &ConversionResult{
		OutputPath: job.OutputPath,
		OutputSize: stat.Size(),
		Duration:   elapsed,
		FrameCount: frameCount,
		VideoInfo:  info,
	}, nil
}

func (c *Converter) convertTwoPass(ctx context.Context, job *ConversionJob, duration float64, onProgress ProgressFunc) (int, error) {
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
	if err := runFFmpegWithProgress(ctx, pass1Args, ffmpegProgressOptions{
		Label:      job.ID + " pass1",
		Stage:      "Generating palette",
		Duration:   duration,
		Base:       0.08,
		Span:       0.37,
		OnProgress: onProgress,
	}); err != nil {
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
	if err := runFFmpegWithProgress(ctx, pass2Args, ffmpegProgressOptions{
		Label:      job.ID + " pass2",
		Stage:      "Rendering GIF",
		Duration:   duration,
		Base:       0.45,
		Span:       0.5,
		OnProgress: onProgress,
	}); err != nil {
		return 0, fmt.Errorf("GIF rendering failed: %w", err)
	}

	return estimateFrameCount(job.Profile, job.InputPath, ctx), nil
}

func (c *Converter) convertSinglePass(ctx context.Context, job *ConversionJob, duration float64, onProgress ProgressFunc) (int, error) {
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
	if err := runFFmpegWithProgress(ctx, args, ffmpegProgressOptions{
		Label:      job.ID,
		Stage:      "Rendering GIF",
		Duration:   duration,
		Base:       0.08,
		Span:       0.87,
		OnProgress: onProgress,
	}); err != nil {
		return 0, fmt.Errorf("GIF rendering failed: %w", err)
	}

	return estimateFrameCount(p, job.InputPath, ctx), nil
}

// convertWebP produces an animated WebP using libwebp.
func (c *Converter) convertWebP(ctx context.Context, job *ConversionJob, duration float64, onProgress ProgressFunc) (int, error) {
	p := job.Profile

	// Build video filter: segments + fps + scale + speed (no palette needed)
	vf := buildDirectFilter(p)

	args := buildCommonArgs(job)
	args = append(args,
		"-vf", vf,
		"-c:v", "libwebp",
		"-loop", strconv.Itoa(p.Loop),
		"-an", // no audio
	)

	// WebP quality: default 75
	quality := p.WebPQuality
	if quality <= 0 {
		quality = 75
	}
	if quality > 100 {
		quality = 100
	}
	args = append(args, "-quality", strconv.Itoa(quality))

	if p.WebPLossless {
		args = append(args, "-lossless", "1")
	}

	args = append(args, "-y", job.OutputPath)

	log.Printf("[%s] Rendering WebP (quality=%d, lossless=%v)...", job.ID, quality, p.WebPLossless)
	if err := runFFmpegWithProgress(ctx, args, ffmpegProgressOptions{
		Label:      job.ID,
		Stage:      "Rendering WebP",
		Duration:   duration,
		Base:       0.08,
		Span:       0.87,
		OnProgress: onProgress,
	}); err != nil {
		return 0, fmt.Errorf("WebP rendering failed: %w", err)
	}

	return estimateFrameCount(p, job.InputPath, ctx), nil
}

// convertAPNG produces an animated PNG.
func (c *Converter) convertAPNG(ctx context.Context, job *ConversionJob, duration float64, onProgress ProgressFunc) (int, error) {
	p := job.Profile

	// Build video filter: segments + fps + scale + speed (no palette needed)
	vf := buildDirectFilter(p)

	args := buildCommonArgs(job)
	args = append(args,
		"-vf", vf,
		"-f", "apng",
		"-plays", strconv.Itoa(p.Loop), // 0=infinite
		"-an", // no audio
		"-y",
		job.OutputPath,
	)

	log.Printf("[%s] Rendering APNG...", job.ID)
	if err := runFFmpegWithProgress(ctx, args, ffmpegProgressOptions{
		Label:      job.ID,
		Stage:      "Rendering APNG",
		Duration:   duration,
		Base:       0.08,
		Span:       0.87,
		OnProgress: onProgress,
	}); err != nil {
		return 0, fmt.Errorf("APNG rendering failed: %w", err)
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

func extractFFmpegError(stderr string) string {
	rawLines := strings.Split(stderr, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, l := range rawLines {
		t := strings.TrimSpace(l)
		if t != "" {
			lines = append(lines, t)
		}
	}
	if len(lines) == 0 {
		return ""
	}

	isKeyLine := func(line string) bool {
		switch {
		case strings.HasPrefix(line, "Error"),
			strings.HasPrefix(line, "Invalid"),
			strings.HasPrefix(line, "No such"),
			strings.HasPrefix(line, "Unable"),
			strings.HasPrefix(line, "Stream specifier"),
			strings.HasPrefix(line, "Input link"),
			strings.HasPrefix(line, "Filter"),
			strings.HasPrefix(line, "Impossible"),
			strings.HasPrefix(line, "Cannot"),
			strings.HasPrefix(line, "Failed"),
			strings.HasPrefix(line, "Could not"),
			strings.HasPrefix(line, "Option"):
			return true
		default:
			return false
		}
	}

	for i := len(lines) - 1; i >= 0; i-- {
		if !isKeyLine(lines[i]) {
			continue
		}
		start := i - 2
		if start < 0 {
			start = 0
		}
		return strings.Join(lines[start:i+1], " | ")
	}

	// Fallback: last few non-empty lines for context.
	start := len(lines) - 5
	if start < 0 {
		start = 0
	}
	return strings.Join(lines[start:], " | ")
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

// EstimateOutputSize estimates the GIF output size in bytes based on conversion parameters.
// Formula: frames * width * height * bytesPerPixel * compressionFactor
// This is a rough heuristic; actual GIF sizes vary with content complexity.
func EstimateOutputSize(info *VideoInfo, profile *config.GifProfile) int64 {
	if info == nil || profile == nil {
		return 0
	}

	// Determine effective duration
	dur := effectiveDurationSeconds(info.Duration, *profile)
	if dur <= 0 {
		return 0
	}

	fps := clampFPS(profile.FPS)
	frames := fps * dur

	// Determine effective resolution
	w := float64(info.Width)
	h := float64(info.Height)
	if profile.Width > 0 {
		w = float64(profile.Width)
	}
	if profile.Height > 0 {
		h = float64(profile.Height)
	}
	// If only width is set, scale height proportionally (and vice versa)
	if profile.Width > 0 && profile.Height <= 0 && info.Height > 0 && info.Width > 0 {
		h = float64(profile.Width) * float64(info.Height) / float64(info.Width)
	} else if profile.Height > 0 && profile.Width <= 0 && info.Width > 0 && info.Height > 0 {
		w = float64(profile.Height) * float64(info.Width) / float64(info.Height)
	}

	// Bytes per pixel based on color count (GIF uses indexed color, 1 byte per pixel max)
	// With fewer colors, LZW compresses better
	colors := float64(profile.Colors)
	if colors < 2 {
		colors = 2
	}
	if colors > 256 {
		colors = 256
	}
	// Bits per pixel for the palette depth
	bitsPerPixel := math.Ceil(math.Log2(colors))
	bytesPerPixel := bitsPerPixel / 8.0

	// GIF LZW compression factor (empirical: typical GIF achieves ~40-60% compression)
	compressionFactor := 0.5

	// Dither increases entropy (harder to compress)
	if profile.Dither != "none" {
		compressionFactor += 0.1
	}

	estimated := frames * w * h * bytesPerPixel * compressionFactor
	return int64(estimated)
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
