package converter

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveEditedVideo_VariableResolutionInput(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found in PATH")
	}

	dir := t.TempDir()
	seg1 := filepath.Join(dir, "seg1.ts")
	seg2 := filepath.Join(dir, "seg2.ts")
	concatTS := filepath.Join(dir, "varres.ts")
	out := filepath.Join(dir, "out.mp4")

	makeSeg := func(dst string, size string, tone string) {
		t.Helper()
		cmd := exec.Command(
			"ffmpeg",
			"-hide_banner", "-loglevel", "error",
			"-y",
			"-f", "lavfi", "-i", "testsrc=size="+size+":rate=30",
			"-f", "lavfi", "-i", "sine=frequency="+tone+":sample_rate=48000",
			"-t", "2",
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-g", "30",
			"-keyint_min", "30",
			"-bf", "0",
			"-c:a", "aac",
			"-ar", "48000",
			"-f", "mpegts",
			dst,
		)
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("ffmpeg segment (%s) failed: %v\n%s", size, err, string(b))
		}
	}

	makeSeg(seg1, "640x360", "440")
	makeSeg(seg2, "360x640", "880")

	outFile, err := os.Create(concatTS)
	if err != nil {
		t.Fatalf("create concat ts: %v", err)
	}
	for _, src := range []string{seg1, seg2} {
		f, err := os.Open(src)
		if err != nil {
			_ = outFile.Close()
			t.Fatalf("open segment: %v", err)
		}
		if _, err := io.Copy(outFile, f); err != nil {
			_ = f.Close()
			_ = outFile.Close()
			t.Fatalf("concat segment: %v", err)
		}
		_ = f.Close()
	}
	if err := outFile.Close(); err != nil {
		t.Fatalf("close concat ts: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	if _, err := SaveEditedVideo(ctx, concatTS, out, nil, 0, nil); err != nil {
		t.Fatalf("SaveEditedVideo failed: %v", err)
	}
	if st, err := os.Stat(out); err != nil || st.Size() == 0 {
		t.Fatalf("expected non-empty output mp4: statErr=%v size=%d", err, func() int64 {
			if st == nil {
				return 0
			}
			return st.Size()
		}())
	}
}

