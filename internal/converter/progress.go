package converter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

// ProgressUpdate describes a single operation progress update.
type ProgressUpdate struct {
	Fraction float64
	Stage    string
	Detail   string
}

// ProgressFunc receives progress updates from long-running video operations.
type ProgressFunc func(ProgressUpdate)

type ffmpegProgressOptions struct {
	Label      string
	Stage      string
	Duration   float64
	Base       float64
	Span       float64
	OnProgress ProgressFunc
}

func runFFmpeg(ctx context.Context, args []string, label string) error {
	return runFFmpegWithProgress(ctx, args, ffmpegProgressOptions{Label: label})
}

func runFFmpegWithProgress(ctx context.Context, args []string, opts ffmpegProgressOptions) error {
	fullArgs := append([]string{"-progress", "pipe:1", "-nostats"}, args...)
	cmd := exec.CommandContext(ctx, "ffmpeg", fullArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("[%s] stdout pipe: %w", opts.Label, err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("[%s] start ffmpeg: %w", opts.Label, err)
	}

	progressErrCh := make(chan error, 1)
	go func() {
		progressErrCh <- parseFFmpegProgress(stdout, opts)
	}()

	waitErr := cmd.Wait()
	parseErr := <-progressErrCh
	if parseErr != nil {
		return fmt.Errorf("[%s] progress parse failed: %w", opts.Label, parseErr)
	}
	if waitErr != nil {
		errMsg := extractFFmpegError(stderr.String())
		return fmt.Errorf("[%s] %w\nffmpeg: %s", opts.Label, waitErr, errMsg)
	}
	if opts.OnProgress != nil && opts.Span > 0 {
		reportProgress(opts.OnProgress, opts.Base+opts.Span, opts.Stage, "Finishing output")
	}
	return nil
}

func parseFFmpegProgress(r io.Reader, opts ffmpegProgressOptions) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "out_time", "out_time_ms", "out_time_us":
			seconds := parseFFmpegOutTime(key, value)
			if seconds <= 0 || opts.OnProgress == nil || opts.Duration <= 0 || opts.Span <= 0 {
				continue
			}
			stageProgress := clampProgress(seconds / opts.Duration)
			detail := fmt.Sprintf("%s / %s", formatProgressTime(seconds), formatProgressTime(opts.Duration))
			reportProgress(opts.OnProgress, opts.Base+(stageProgress*opts.Span), opts.Stage, detail)
		case "progress":
			if value == "end" && opts.OnProgress != nil && opts.Span > 0 {
				reportProgress(opts.OnProgress, opts.Base+opts.Span, opts.Stage, "Finishing output")
			}
		}
	}

	return scanner.Err()
}

func parseFFmpegOutTime(key, value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "N/A" {
		return 0
	}
	switch key {
	case "out_time":
		return parseTimestampSeconds(value)
	case "out_time_ms", "out_time_us":
		raw, err := strconv.ParseFloat(value, 64)
		if err != nil || raw <= 0 {
			return 0
		}
		if raw >= 1000 {
			return raw / 1_000_000
		}
		return raw / 1000
	default:
		return 0
	}
}

func parseTimestampSeconds(raw string) float64 {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0
	}
	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0
	}
	return (hours * 3600) + (minutes * 60) + seconds
}

func formatProgressTime(seconds float64) string {
	totalMillis := int(math.Round(math.Max(0, seconds) * 1000))
	hours := totalMillis / 3600000
	minutes := (totalMillis % 3600000) / 60000
	secs := (totalMillis % 60000) / 1000
	millis := totalMillis % 1000
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
	}
	return fmt.Sprintf("%02d:%02d.%03d", minutes, secs, millis)
}

func clampProgress(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func reportProgress(fn ProgressFunc, fraction float64, stage, detail string) {
	if fn == nil {
		return
	}
	fn(ProgressUpdate{
		Fraction: clampProgress(fraction),
		Stage:    strings.TrimSpace(stage),
		Detail:   strings.TrimSpace(detail),
	})
}
