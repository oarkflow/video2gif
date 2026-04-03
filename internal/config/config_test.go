package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNormalizesLegacyRuntimeDirsIntoStorage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configJSON := `{
		"server": {"host": "0.0.0.0", "port": 8080},
		"storage": {
			"upload_dir": "./uploads",
			"output_dir": "outputs",
			"temp_dir": "./tmp",
			"share_dir": "shares"
		},
		"queue": {"workers": 1, "max_queue_size": 10, "job_timeout_sec": 10},
		"auth": {"enabled": false},
		"sharing": {"enabled": true, "public_view": true},
		"default_profile": "balanced",
		"profiles": {
			"balanced": {
				"name": "balanced",
				"fps": 20,
				"width": 640,
				"height": -1,
				"colors": 256,
				"dither": "sierra2_4a",
				"loop": 0,
				"optimize_palette": true,
				"stats_mode": "diff",
				"speed_multiplier": 1.0
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Storage.UploadDir != defaultUploadDir {
		t.Fatalf("upload dir = %q, want %q", cfg.Storage.UploadDir, defaultUploadDir)
	}
	if cfg.Storage.OutputDir != defaultOutputDir {
		t.Fatalf("output dir = %q, want %q", cfg.Storage.OutputDir, defaultOutputDir)
	}
	if cfg.Storage.TempDir != defaultTempDir {
		t.Fatalf("temp dir = %q, want %q", cfg.Storage.TempDir, defaultTempDir)
	}
	if cfg.Storage.ShareDir != defaultShareDir {
		t.Fatalf("share dir = %q, want %q", cfg.Storage.ShareDir, defaultShareDir)
	}
	if cfg.Storage.JobStorePath != defaultJobStorePath {
		t.Fatalf("job store path = %q, want %q", cfg.Storage.JobStorePath, defaultJobStorePath)
	}
}

func TestApplyDefaultsKeepsCustomRuntimeDirs(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.Storage.UploadDir = "./custom/uploads"
	cfg.Storage.OutputDir = "./custom/outputs"
	cfg.Storage.TempDir = "./custom/tmp"
	cfg.Storage.ShareDir = "./custom/shares"
	cfg.Storage.JobStorePath = "./custom/jobs"

	cfg.applyDefaults()

	if cfg.Storage.UploadDir != "./custom/uploads" {
		t.Fatalf("upload dir = %q", cfg.Storage.UploadDir)
	}
	if cfg.Storage.OutputDir != "./custom/outputs" {
		t.Fatalf("output dir = %q", cfg.Storage.OutputDir)
	}
	if cfg.Storage.TempDir != "./custom/tmp" {
		t.Fatalf("temp dir = %q", cfg.Storage.TempDir)
	}
	if cfg.Storage.ShareDir != "./custom/shares" {
		t.Fatalf("share dir = %q", cfg.Storage.ShareDir)
	}
	if cfg.Storage.JobStorePath != "./custom/jobs" {
		t.Fatalf("job store path = %q", cfg.Storage.JobStorePath)
	}
}
