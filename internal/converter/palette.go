package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oarkflow/video2gif/internal/config"
)

// BuildPaletteFilter constructs the ffmpeg filtergraph for two-pass palette GIF generation.
// Two-pass is superior: pass 1 generates an optimal palette, pass 2 applies it.
func BuildPaletteFilter(p config.GifProfile) (pass1Filter, pass2Filter string) {
	segmentPrefix, inputLabel := buildSegmentSourceFilter(p)

	// Scale filter
	scaleFilter := buildScaleFilter(p)

	// FPS filter
	fpsFilter := fmt.Sprintf("fps=%.4f", clampFPS(p.FPS))

	// Speed filter (setpts)
	var speedFilter string
	if p.SpeedMultiplier != 0 && p.SpeedMultiplier != 1.0 {
		pts := 1.0 / p.SpeedMultiplier
		speedFilter = fmt.Sprintf(",setpts=%.4f*PTS", pts)
	}

	// Stats mode for palette
	statsMode := p.StatsMode
	if statsMode == "" {
		statsMode = "diff"
	}

	if p.OptimizePalette {
		// Two-pass:
		// Pass 1: video → scale → fps → palettegen
		pass1Main := fmt.Sprintf("%s%s,%s%s,palettegen=max_colors=%d:stats_mode=%s[palette]",
			inputLabel, fpsFilter, scaleFilter, speedFilter, p.Colors, statsMode)
		if segmentPrefix != "" {
			pass1Filter = segmentPrefix + ";" + pass1Main
		} else {
			pass1Filter = pass1Main
		}

		// Pass 2: video + palette → paletteuse
		ditherOpts := buildDitherOpts(p)
		pass2Main := fmt.Sprintf("%s%s,%s%s[vprep];[vprep][1:v]paletteuse=dither=%s%s",
			inputLabel, fpsFilter, scaleFilter, speedFilter, p.Dither, ditherOpts)
		if segmentPrefix != "" {
			pass2Filter = segmentPrefix + ";" + pass2Main
		} else {
			pass2Filter = pass2Main
		}
	} else {
		// Single-pass: simpler but lower quality
		pass1Filter = ""
		pass2Main := fmt.Sprintf(
			"%s%s,%s%s,split[a][b];[a]palettegen=max_colors=%d:stats_mode=%s[palette];[b][palette]paletteuse=dither=%s%s",
			inputLabel, fpsFilter, scaleFilter, speedFilter, p.Colors, statsMode, p.Dither, buildDitherOpts(p),
		)
		if segmentPrefix != "" {
			pass2Filter = segmentPrefix + ";" + pass2Main
		} else {
			pass2Filter = pass2Main
		}
	}

	return
}

func buildSegmentSourceFilter(p config.GifProfile) (prefix string, inputLabel string) {
	if len(p.KeepSegments) == 0 {
		return "", "[0:v]"
	}

	parts := make([]string, 0, len(p.KeepSegments)+1)
	concatInputs := make([]string, 0, len(p.KeepSegments))

	for i, seg := range p.KeepSegments {
		start := seg.Start
		end := seg.End
		if end <= start {
			continue
		}
		label := fmt.Sprintf("v%d", i)
		parts = append(parts, fmt.Sprintf("[0:v]trim=start=%.6f:end=%.6f,setpts=PTS-STARTPTS[%s]", start, end, label))
		concatInputs = append(concatInputs, fmt.Sprintf("[%s]", label))
	}

	switch len(concatInputs) {
	case 0:
		return "", "[0:v]"
	case 1:
		return parts[0], concatInputs[0]
	default:
		parts = append(parts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vsrc]", strings.Join(concatInputs, ""), len(concatInputs)))
		return strings.Join(parts, ";"), "[vsrc]"
	}
}

func buildScaleFilter(p config.GifProfile) string {
	w, h := p.Width, p.Height
	if w <= 0 {
		w = -1
	}
	if h <= 0 {
		h = -1
	}

	// Ensure dimensions are even numbers (required by many codecs/scalers)
	wStr := fmt.Sprintf("%d", w)
	hStr := fmt.Sprintf("%d", h)
	if w > 0 {
		wStr = fmt.Sprintf("trunc(%d/2)*2", w)
	}
	if h > 0 {
		hStr = fmt.Sprintf("trunc(%d/2)*2", h)
	}

	return fmt.Sprintf("scale=%s:%s:flags=lanczos", wStr, hStr)
}

func buildDitherOpts(p config.GifProfile) string {
	if p.Dither == "bayer" {
		return fmt.Sprintf(":bayer_scale=%d", clampBayerScale(p.BayerScale))
	}
	return ""
}

func clampFPS(fps float64) float64 {
	if fps < 1 {
		return 1
	}
	if fps > 60 {
		return 60
	}
	return fps
}

func clampBayerScale(s int) int {
	if s < 0 {
		return 0
	}
	if s > 5 {
		return 5
	}
	return s
}

// PalettePath returns the temp palette PNG path for a job.
func PalettePath(tempDir, jobID string) string {
	return filepath.Join(tempDir, fmt.Sprintf("palette_%s.png", jobID))
}

// CleanupPalette removes the temporary palette file.
func CleanupPalette(path string) {
	_ = os.Remove(path)
}
