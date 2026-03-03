package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is the root configuration structure loaded from config.json
type Config struct {
	Server         ServerConfig          `json:"server"`
	Storage        StorageConfig         `json:"storage"`
	Queue          QueueConfig           `json:"queue"`
	DefaultProfile string                `json:"default_profile"`
	Profiles       map[string]GifProfile `json:"profiles"`
}

type ServerConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ReadTimeoutSec  int    `json:"read_timeout_sec"`
	WriteTimeoutSec int    `json:"write_timeout_sec"`
	MaxUploadBytes  int64  `json:"max_upload_bytes"`
}

type StorageConfig struct {
	UploadDir          string `json:"upload_dir"`
	OutputDir          string `json:"output_dir"`
	TempDir            string `json:"temp_dir"`
	MaxAgeHours        int    `json:"max_age_hours"`
	CleanupIntervalMin int    `json:"cleanup_interval_min"`
}

type QueueConfig struct {
	Workers       int `json:"workers"`
	MaxQueueSize  int `json:"max_queue_size"`
	JobTimeoutSec int `json:"job_timeout_sec"`
}

type ClipSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// GifProfile defines all conversion parameters
type GifProfile struct {
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	FPS             float64       `json:"fps"`                     // Output framerate (1–60)
	Width           int           `json:"width"`                   // Width in px; -1 = keep aspect
	Height          int           `json:"height"`                  // Height in px; -1 = keep aspect
	Colors          int           `json:"colors"`                  // 2–256 palette colors
	Dither          string        `json:"dither"`                  // none | bayer | sierra2 | sierra2_4a | floyd_steinberg
	BayerScale      int           `json:"bayer_scale"`             // 0–5, only used with bayer dither
	Loop            int           `json:"loop"`                    // 0=infinite, -1=no loop, N=N times
	OptimizePalette bool          `json:"optimize_palette"`        // Use two-pass palette generation
	StatsMode       string        `json:"stats_mode"`              // full | diff (for palette)
	StartTime       string        `json:"start_time"`              // e.g. "00:00:05" or "5.5"
	Duration        string        `json:"duration"`                // e.g. "10" (seconds)
	SpeedMultiplier float64       `json:"speed_multiplier"`        // 0.25–4.0
	KeepSegments    []ClipSegment `json:"keep_segments,omitempty"` // Segments to keep and concatenate
}

// Load reads and validates configuration from a JSON file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	cfg.applyDefaults()
	return &cfg, nil
}

// Default returns a safe default configuration.
func Default() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0", Port: 8080,
			ReadTimeoutSec: 300, WriteTimeoutSec: 300,
			MaxUploadBytes: 500 * 1024 * 1024,
		},
		Storage: StorageConfig{
			UploadDir: "./uploads", OutputDir: "./outputs", TempDir: "./tmp",
			MaxAgeHours: 24, CleanupIntervalMin: 30,
		},
		Queue:          QueueConfig{Workers: 4, MaxQueueSize: 100, JobTimeoutSec: 600},
		DefaultProfile: "balanced",
		Profiles: map[string]GifProfile{
			"balanced": {
				Name: "balanced", FPS: 20, Width: 640, Height: -1,
				Colors: 256, Dither: "sierra2_4a", BayerScale: 2,
				Loop: 0, OptimizePalette: true, StatsMode: "diff",
				SpeedMultiplier: 1.0,
			},
		},
	}
	cfg.applyDefaults()
	return cfg
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be 1–65535")
	}
	if c.Queue.Workers < 1 {
		return fmt.Errorf("queue.workers must be >= 1")
	}
	for name, p := range c.Profiles {
		if p.FPS <= 0 || p.FPS > 60 {
			return fmt.Errorf("profile %q: fps must be 1–60", name)
		}
		if p.Colors < 2 || p.Colors > 256 {
			return fmt.Errorf("profile %q: colors must be 2–256", name)
		}
		validDithers := map[string]bool{
			"none": true, "bayer": true, "sierra2": true,
			"sierra2_4a": true, "floyd_steinberg": true,
		}
		if !validDithers[p.Dither] {
			return fmt.Errorf("profile %q: unknown dither %q", name, p.Dither)
		}
		if p.SpeedMultiplier == 0 {
			c.Profiles[name] = func(pp GifProfile) GifProfile {
				pp.SpeedMultiplier = 1.0
				return pp
			}(p)
		}
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Server.ReadTimeoutSec == 0 {
		c.Server.ReadTimeoutSec = 300
	}
	if c.Server.WriteTimeoutSec == 0 {
		c.Server.WriteTimeoutSec = 300
	}
	if c.Server.MaxUploadBytes == 0 {
		c.Server.MaxUploadBytes = 500 * 1024 * 1024
	}
	if c.Storage.UploadDir == "" {
		c.Storage.UploadDir = "./uploads"
	}
	if c.Storage.OutputDir == "" {
		c.Storage.OutputDir = "./outputs"
	}
	if c.Storage.TempDir == "" {
		c.Storage.TempDir = "./tmp"
	}
	if c.Queue.Workers == 0 {
		c.Queue.Workers = 4
	}
	if c.Queue.JobTimeoutSec == 0 {
		c.Queue.JobTimeoutSec = 600
	}
	if c.DefaultProfile == "" {
		c.DefaultProfile = "balanced"
	}
	// Ensure dirs exist
	for _, d := range []string{c.Storage.UploadDir, c.Storage.OutputDir, c.Storage.TempDir} {
		_ = os.MkdirAll(d, 0755)
	}
}

// GetProfile returns a profile by name, falling back to default.
func (c *Config) GetProfile(name string) (GifProfile, bool) {
	if name == "" {
		name = c.DefaultProfile
	}
	p, ok := c.Profiles[name]
	return p, ok
}

// Save writes the current config back to disk.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
