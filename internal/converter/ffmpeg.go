package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CheckFFmpeg verifies ffmpeg and ffprobe are available and returns their versions.
func CheckFFmpeg() (ffmpegVer, ffprobeVer string, err error) {
	for bin, dest := range map[string]*string{"ffmpeg": &ffmpegVer, "ffprobe": &ffprobeVer} {
		out, e := exec.Command(bin, "-version").Output()
		if e != nil {
			return "", "", fmt.Errorf("%s not found in PATH: %w", bin, e)
		}
		lines := strings.SplitN(string(out), "\n", 2)
		if len(lines) > 0 {
			*dest = strings.TrimSpace(lines[0])
		}
	}
	return
}

// VideoInfo contains metadata probed from a video file.
type VideoInfo struct {
	Duration   float64 `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FPS        float64 `json:"fps"`
	Codec      string  `json:"codec"`
	BitrateBps int64   `json:"bitrate_bps"`
	SizeBytes  int64   `json:"size_bytes"`
	Format     string  `json:"format"`
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType  string `json:"codec_type"`
		CodecName  string `json:"codec_name"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
		Duration   string `json:"duration"`
	} `json:"streams"`
	Format struct {
		Duration   string `json:"duration"`
		Size       string `json:"size"`
		BitRate    string `json:"bit_rate"`
		FormatName string `json:"format_name"`
	} `json:"format"`
}

// ProbeVideo extracts metadata from any video file using ffprobe.
func ProbeVideo(ctx context.Context, path string) (*VideoInfo, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		path,
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w\nstderr: %s", err, stderr.String())
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &probe); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	info := &VideoInfo{}

	// Format-level info
	if d, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		info.Duration = d
	}
	if s, err := strconv.ParseInt(probe.Format.Size, 10, 64); err == nil {
		info.SizeBytes = s
	}
	if b, err := strconv.ParseInt(probe.Format.BitRate, 10, 64); err == nil {
		info.BitrateBps = b
	}
	info.Format = probe.Format.FormatName

	// Stream-level video info
	for _, s := range probe.Streams {
		if s.CodecType == "video" {
			info.Width = s.Width
			info.Height = s.Height
			info.Codec = s.CodecName
			// Parse "30000/1001" style FPS
			info.FPS = parseRationalFPS(s.RFrameRate)
			// Stream duration overrides format if present
			if d, err := strconv.ParseFloat(s.Duration, 64); err == nil && d > 0 {
				info.Duration = d
			}
			break
		}
	}

	if info.Duration <= 0 {
		if d, err := probeDurationFormat(ctx, path); err == nil && d > 0 {
			info.Duration = d
		}
	}

	return info, nil
}

func probeDurationFormat(ctx context.Context, path string) (float64, error) {
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
	raw := strings.TrimSpace(stdout.String())
	if raw == "" || strings.EqualFold(raw, "N/A") {
		return 0, fmt.Errorf("duration unavailable")
	}
	d, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration <= 0")
	}
	return d, nil
}

func parseRationalFPS(r string) float64 {
	parts := strings.Split(r, "/")
	if len(parts) != 2 {
		f, _ := strconv.ParseFloat(r, 64)
		return f
	}
	num, err1 := strconv.ParseFloat(parts[0], 64)
	den, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || den == 0 {
		return 0
	}
	return num / den
}
