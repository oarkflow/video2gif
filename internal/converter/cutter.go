package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oarkflow/video2gif/internal/config"
)

// SaveEditedVideo exports an MP4 with all cut ranges removed.
func SaveEditedVideo(ctx context.Context, inputPath, outputPath string, cutRanges []config.ClipSegment, durationHint float64, onProgress ProgressFunc) (*ConversionResult, error) {
	startedAt := time.Now()
	reportProgress(onProgress, 0.02, "Probing input", "Inspecting source video")

	info, err := ProbeVideo(ctx, inputPath)
	if err != nil {
		return nil, fmt.Errorf("probe video: %w", err)
	}
	duration := info.Duration
	if duration <= 0 {
		if d, e := probeDurationRaw(ctx, inputPath); e == nil && d > 0 {
			duration = d
		}
	}
	if duration <= 0 && durationHint > 0 {
		duration = durationHint
	}
	if duration <= 0 {
		return nil, fmt.Errorf("invalid video duration")
	}

	cuts := normalizeSegments(cutRanges, duration)
	keeps := inverseSegments(cuts, duration)
	if len(keeps) == 0 {
		return nil, fmt.Errorf("all video content has been removed by cut ranges")
	}
	outputDuration := sumSegmentDuration(keeps)
	if outputDuration <= 0 {
		outputDuration = duration
	}
	targetFPS := normalizeEditFPS(info.FPS)
	gop := strconv.Itoa(int(math.Max(24, math.Round(targetFPS*2))))
	targetW, targetH := normalizeEditDims(info.Width, info.Height)

	hasAudio, _ := hasAudioStream(ctx, inputPath)

	var args []string
	if hasAudio {
		filter := buildKeepConcatFilterWithAudio(keeps, targetFPS, targetW, targetH)
		args = []string{
			"-hide_banner", "-loglevel", "error",
			"-i", inputPath,
			"-filter_complex", filter,
			"-map", "[vout]",
			"-map", "[aout]",
			"-c:v", "libx264",
			"-preset", "medium",
			"-crf", "15",
			"-profile:v", "high",
			"-g", gop,
			"-keyint_min", gop,
			"-pix_fmt", "yuv420p",
			"-fps_mode", "cfr",
			"-c:a", "aac",
			"-b:a", "192k",
			"-ar", "48000",
			"-movflags", "+faststart",
			"-y", outputPath,
		}
	} else {
		filter := buildKeepConcatFilter(keeps, targetFPS, targetW, targetH)
		args = []string{
			"-hide_banner", "-loglevel", "error",
			"-i", inputPath,
			"-filter_complex", filter,
			"-map", "[vout]",
			"-an",
			"-c:v", "libx264",
			"-preset", "medium",
			"-crf", "15",
			"-profile:v", "high",
			"-g", gop,
			"-keyint_min", gop,
			"-pix_fmt", "yuv420p",
			"-fps_mode", "cfr",
			"-movflags", "+faststart",
			"-y", outputPath,
		}
	}

	reportProgress(onProgress, 0.08, "Removing cut ranges", fmt.Sprintf("Keeping %s of %s", formatProgressTime(outputDuration), formatProgressTime(duration)))
	if err := runFFmpegWithProgress(ctx, args, ffmpegProgressOptions{
		Label:      "save edited video",
		Stage:      "Removing cut ranges",
		Duration:   outputDuration,
		Base:       0.08,
		Span:       0.88,
		OnProgress: onProgress,
	}); err != nil {
		return nil, fmt.Errorf("save edited video: %w (keeps=%d audio=%t fps=%.2f target=%dx%d)", err, len(keeps), hasAudio, targetFPS, targetW, targetH)
	}

	stat, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("stat output: %w", err)
	}
	elapsed := time.Since(startedAt)
	reportProgress(onProgress, 1.0, "Complete", "Edited video ready")

	return &ConversionResult{
		OutputPath: outputPath,
		OutputSize: stat.Size(),
		Duration:   elapsed,
		VideoInfo:  info,
	}, nil
}

func normalizeSegments(in []config.ClipSegment, duration float64) []config.ClipSegment {
	segments := make([]config.ClipSegment, 0, len(in))
	for _, s := range in {
		start := s.Start
		end := s.End
		if end <= start {
			continue
		}
		if start < 0 {
			start = 0
		}
		if end > duration {
			end = duration
		}
		if end > start {
			segments = append(segments, config.ClipSegment{Start: start, End: end})
		}
	}
	if len(segments) == 0 {
		return segments
	}
	sort.Slice(segments, func(i, j int) bool { return segments[i].Start < segments[j].Start })

	merged := []config.ClipSegment{segments[0]}
	for i := 1; i < len(segments); i++ {
		last := &merged[len(merged)-1]
		cur := segments[i]
		if cur.Start <= last.End {
			if cur.End > last.End {
				last.End = cur.End
			}
			continue
		}
		merged = append(merged, cur)
	}
	return merged
}

func inverseSegments(cuts []config.ClipSegment, duration float64) []config.ClipSegment {
	if len(cuts) == 0 {
		return []config.ClipSegment{{Start: 0, End: duration}}
	}
	keeps := make([]config.ClipSegment, 0, len(cuts)+1)
	cursor := 0.0
	for _, c := range cuts {
		if c.Start > cursor {
			keeps = append(keeps, config.ClipSegment{Start: cursor, End: c.Start})
		}
		if c.End > cursor {
			cursor = c.End
		}
	}
	if cursor < duration {
		keeps = append(keeps, config.ClipSegment{Start: cursor, End: duration})
	}
	out := make([]config.ClipSegment, 0, len(keeps))
	for _, s := range keeps {
		if s.End-s.Start >= 0.05 {
			out = append(out, s)
		}
	}
	return out
}

func sumSegmentDuration(segments []config.ClipSegment) float64 {
	total := 0.0
	for _, s := range segments {
		if s.End > s.Start {
			total += s.End - s.Start
		}
	}
	return total
}

func buildKeepConcatFilter(keeps []config.ClipSegment, fps float64, targetW, targetH int) string {
	parts := make([]string, 0, len(keeps)+1)
	labels := make([]string, 0, len(keeps))
	vNorm := videoNormalizeFilter(targetW, targetH)
	for i, s := range keeps {
		label := fmt.Sprintf("v%d", i)
		parts = append(parts, fmt.Sprintf("[0:v]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS%s[%s]", s.Start, s.End, vNorm, label))
		labels = append(labels, fmt.Sprintf("[%s]", label))
	}
	parts = append(parts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vtmp]", strings.Join(labels, ""), len(labels)))
	parts = append(parts, fmt.Sprintf("[vtmp]fps=%.4f,format=yuv420p[vout]", fps))
	return strings.Join(parts, ";")
}

func buildKeepConcatFilterWithAudio(keeps []config.ClipSegment, fps float64, targetW, targetH int) string {
	parts := make([]string, 0, len(keeps)*2+1)
	labels := make([]string, 0, len(keeps)*2)
	vNorm := videoNormalizeFilter(targetW, targetH)
	aNorm := audioNormalizeFilter()
	for i, s := range keeps {
		vLabel := fmt.Sprintf("v%d", i)
		aLabel := fmt.Sprintf("a%d", i)
		parts = append(parts, fmt.Sprintf("[0:v]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS%s[%s]", s.Start, s.End, vNorm, vLabel))
		parts = append(parts, fmt.Sprintf("[0:a]atrim=start=%.6f:end=%.6f,asetpts=PTS-STARTPTS%s[%s]", s.Start, s.End, aNorm, aLabel))
		labels = append(labels, fmt.Sprintf("[%s][%s]", vLabel, aLabel))
	}
	parts = append(parts, fmt.Sprintf("%sconcat=n=%d:v=1:a=1[vtmp][atmp]", strings.Join(labels, ""), len(keeps)))
	parts = append(parts, fmt.Sprintf("[vtmp]fps=%.4f,format=yuv420p[vout]", fps))
	parts = append(parts, "[atmp]aresample=async=1:first_pts=0[aout]")
	return strings.Join(parts, ";")
}

func normalizeEditDims(w, h int) (int, int) {
	if w <= 0 || h <= 0 {
		return 0, 0
	}
	if w%2 != 0 {
		w++
	}
	if h%2 != 0 {
		h++
	}
	return w, h
}

func videoNormalizeFilter(targetW, targetH int) string {
	if targetW <= 0 || targetH <= 0 {
		return ",setsar=1,format=yuv420p"
	}
	return fmt.Sprintf(",scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1,format=yuv420p", targetW, targetH, targetW, targetH)
}

func audioNormalizeFilter() string {
	// Normalize audio format per-segment so concat doesn't fail if the stream changes layout/rate mid-recording.
	return ",aresample=48000,aformat=sample_fmts=fltp:channel_layouts=stereo"
}

func normalizeEditFPS(fps float64) float64 {
	if math.IsNaN(fps) || math.IsInf(fps, 0) || fps <= 0 {
		return 30
	}
	if fps < 10 {
		return 30
	}
	if fps > 60 {
		return 60
	}
	return fps
}

func hasAudioStream(ctx context.Context, path string) (bool, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		path,
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("ffprobe streams: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	var out struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Channels  int    `json:"channels"`
			BitRate   string `json:"bit_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return false, err
	}
	for _, s := range out.Streams {
		if s.CodecType != "audio" {
			continue
		}
		if s.Channels > 0 {
			return true, nil
		}
		if b, err := strconv.ParseInt(s.BitRate, 10, 64); err == nil && b > 0 {
			return true, nil
		}
		return true, nil
	}
	return false, nil
}

func probeDurationRaw(ctx context.Context, path string) (float64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe duration: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	v := strings.TrimSpace(stdout.String())
	if v == "" || strings.EqualFold(v, "N/A") {
		return 0, fmt.Errorf("duration unavailable")
	}
	d, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration <= 0")
	}
	return d, nil
}
