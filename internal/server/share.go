package server

import (
	"time"

	"github.com/oarkflow/video2gif/internal/config"
)

type ShareComment struct {
	ID        string    `json:"id"`
	Time      float64   `json:"time"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Text      string    `json:"text"`
	Status    string    `json:"status,omitempty"`
	Author    string    `json:"author,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type ShareSession struct {
	ID         string               `json:"id"`
	FileName   string               `json:"file_name"`
	VideoPath  string               `json:"-"`
	CutRanges  []config.ClipSegment `json:"cut_ranges"`
	Comments   []ShareComment       `json:"comments"`
	CreatedAt  time.Time            `json:"created_at"`
	ExpiresAt  time.Time            `json:"expires_at"`
	CreatedBy  string               `json:"created_by,omitempty"`
	PublicView bool                 `json:"public_view"`
}
