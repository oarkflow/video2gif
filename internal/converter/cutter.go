package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/oarkflow/video2gif/internal/config"
)

// SaveEditedVideo exports an MP4 with all cut ranges removed.
func SaveEditedVideo(ctx context.Context, inputPath, outputPath string, cutRanges []config.ClipSegment, durationHint float64) error {
	info, err := ProbeVideo(ctx, inputPath)
	if err != nil {
		return fmt.Errorf("probe video: %w", err)
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
		return fmt.Errorf("invalid video duration")
	}

	cuts := normalizeSegments(cutRanges, duration)
	keeps := inverseSegments(cuts, duration)
	if len(keeps) == 0 {
		return fmt.Errorf("all video content has been removed by cut ranges")
	}

	hasAudio, _ := hasAudioStream(ctx, inputPath)

	var args []string
	if hasAudio {
		filter := buildKeepConcatFilterWithAudio(keeps)
		args = []string{
			"-hide_banner", "-loglevel", "error",
			"-i", inputPath,
			"-filter_complex", filter,
			"-map", "[vout]",
			"-map", "[aout]",
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-crf", "18",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-b:a", "160k",
			"-movflags", "+faststart",
			"-y", outputPath,
		}
	} else {
		filter := buildKeepConcatFilter(keeps)
		args = []string{
			"-hide_banner", "-loglevel", "error",
			"-i", inputPath,
			"-filter_complex", filter,
			"-map", "[vout]",
			"-an",
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-crf", "18",
			"-pix_fmt", "yuv420p",
			"-movflags", "+faststart",
			"-y", outputPath,
		}
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("save edited video: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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

func buildKeepConcatFilter(keeps []config.ClipSegment) string {
	parts := make([]string, 0, len(keeps)+1)
	labels := make([]string, 0, len(keeps))
	for i, s := range keeps {
		label := fmt.Sprintf("v%d", i)
		parts = append(parts, fmt.Sprintf("[0:v]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS[%s]", s.Start, s.End, label))
		labels = append(labels, fmt.Sprintf("[%s]", label))
	}
	parts = append(parts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vtmp]", strings.Join(labels, ""), len(labels)))
	parts = append(parts, "[vtmp]format=yuv420p[vout]")
	return strings.Join(parts, ";")
}

func buildKeepConcatFilterWithAudio(keeps []config.ClipSegment) string {
	parts := make([]string, 0, len(keeps)*2+1)
	labels := make([]string, 0, len(keeps)*2)
	for i, s := range keeps {
		vLabel := fmt.Sprintf("v%d", i)
		aLabel := fmt.Sprintf("a%d", i)
		parts = append(parts, fmt.Sprintf("[0:v]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS[%s]", s.Start, s.End, vLabel))
		parts = append(parts, fmt.Sprintf("[0:a]atrim=start=%.6f:end=%.6f,asetpts=PTS-STARTPTS[%s]", s.Start, s.End, aLabel))
		labels = append(labels, fmt.Sprintf("[%s][%s]", vLabel, aLabel))
	}
	parts = append(parts, fmt.Sprintf("%sconcat=n=%d:v=1:a=1[vtmp][atmp]", strings.Join(labels, ""), len(keeps)))
	parts = append(parts, "[vtmp]format=yuv420p[vout]")
	parts = append(parts, "[atmp]aresample=async=1[aout]")
	return strings.Join(parts, ";")
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
