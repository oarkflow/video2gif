package converter

import (
	"strings"
	"testing"

	"github.com/oarkflow/video2gif/internal/config"
)

func TestParseFFmpegProgressReportsFraction(t *testing.T) {
	raw := strings.NewReader("out_time=00:00:12.500000\nprogress=continue\nout_time=00:00:25.000000\nprogress=end\n")
	updates := make([]ProgressUpdate, 0, 2)

	err := parseFFmpegProgress(raw, ffmpegProgressOptions{
		Stage:    "Rendering GIF",
		Duration: 25,
		Base:     0.2,
		Span:     0.7,
		OnProgress: func(update ProgressUpdate) {
			updates = append(updates, update)
		},
	})
	if err != nil {
		t.Fatalf("parseFFmpegProgress error: %v", err)
	}
	if len(updates) < 2 {
		t.Fatalf("expected progress updates, got %d", len(updates))
	}
	first := updates[0]
	if first.Stage != "Rendering GIF" {
		t.Fatalf("unexpected stage: %q", first.Stage)
	}
	if first.Fraction <= 0.5 || first.Fraction >= 0.6 {
		t.Fatalf("expected mapped progress near 0.55, got %.4f", first.Fraction)
	}
	last := updates[len(updates)-1]
	if last.Fraction < 0.899 || last.Fraction > 0.901 {
		t.Fatalf("expected final mapped fraction 0.9, got %.4f", last.Fraction)
	}
}

func TestNormalizeSegmentsProducesExpectedKeeps(t *testing.T) {
	cuts := normalizeSegments([]config.ClipSegment{
		{Start: 5, End: 10},
		{Start: 0, End: 2},
		{Start: 8, End: 14},
		{Start: 30, End: 31},
	}, 40)

	keeps := inverseSegments(cuts, 40)
	if len(keeps) != 3 {
		t.Fatalf("expected 3 keep segments, got %d", len(keeps))
	}
	want := []config.ClipSegment{
		{Start: 2, End: 5},
		{Start: 14, End: 30},
		{Start: 31, End: 40},
	}
	for i := range want {
		if keeps[i] != want[i] {
			t.Fatalf("segment %d: got %+v want %+v", i, keeps[i], want[i])
		}
	}
	if got := sumSegmentDuration(keeps); got != 28 {
		t.Fatalf("expected keep duration 28, got %.2f", got)
	}
}
