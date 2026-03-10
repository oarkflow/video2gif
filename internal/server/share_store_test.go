package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oarkflow/video2gif/internal/config"
)

func TestShareStorePersistsSessionsAndCleansExpired(t *testing.T) {
	dir := t.TempDir()
	store, err := NewShareStore(dir)
	if err != nil {
		t.Fatalf("create share store: %v", err)
	}

	videoPath := filepath.Join(dir, "share-1.mp4")
	if err := os.WriteFile(videoPath, []byte("video"), 0644); err != nil {
		t.Fatalf("write share video: %v", err)
	}

	session := &ShareSession{
		ID:         "share-1",
		FileName:   "demo.mp4",
		VideoPath:  videoPath,
		Comments:   []ShareComment{{ID: "c1", Text: "note", Author: "alex"}},
		CreatedAt:  time.Now().Add(-time.Hour),
		ExpiresAt:  time.Now().Add(time.Hour),
		CreatedBy:  "alex",
		PublicView: true,
	}
	if err := store.Save(session); err != nil {
		t.Fatalf("save share session: %v", err)
	}

	reloaded, err := NewShareStore(dir)
	if err != nil {
		t.Fatalf("reload share store: %v", err)
	}
	got, ok, err := reloaded.Get("share-1", time.Now())
	if err != nil || !ok {
		t.Fatalf("expected persisted share session, ok=%v err=%v", ok, err)
	}
	if got.CreatedBy != "alex" || len(got.Comments) != 1 || got.Comments[0].Author != "alex" {
		t.Fatalf("unexpected persisted share session: %+v", got)
	}

	expiredVideo := filepath.Join(dir, "share-expired.mp4")
	if err := os.WriteFile(expiredVideo, []byte("video"), 0644); err != nil {
		t.Fatalf("write expired share video: %v", err)
	}
	if err := reloaded.Save(&ShareSession{
		ID:         "share-expired",
		FileName:   "expired.mp4",
		VideoPath:  expiredVideo,
		CreatedAt:  time.Now().Add(-48 * time.Hour),
		ExpiresAt:  time.Now().Add(-time.Hour),
		PublicView: true,
	}); err != nil {
		t.Fatalf("save expired share: %v", err)
	}
	if err := reloaded.CleanupExpired(time.Now()); err != nil {
		t.Fatalf("cleanup expired shares: %v", err)
	}
	if _, err := os.Stat(expiredVideo); !os.IsNotExist(err) {
		t.Fatalf("expected expired video to be deleted, err=%v", err)
	}
}

func TestPublicShareRouteBypassesAuth(t *testing.T) {
	cfg := config.Default()
	cfg.Storage.UploadDir = filepath.Join(t.TempDir(), "uploads")
	cfg.Storage.OutputDir = filepath.Join(t.TempDir(), "outputs")
	cfg.Storage.TempDir = filepath.Join(t.TempDir(), "tmp")
	cfg.Storage.ShareDir = filepath.Join(t.TempDir(), "shares")
	cfg.Auth.Enabled = true
	cfg.Auth.PasswordSHA256 = hashPassword("secret-pass")
	cfg.Sharing.Enabled = true
	cfg.Sharing.PublicView = true

	srv := New(cfg, filepath.Join(t.TempDir(), "config.json"))
	defer srv.Shutdown()

	videoPath := filepath.Join(cfg.Storage.ShareDir, "public-share.mp4")
	if err := os.WriteFile(videoPath, []byte("video"), 0644); err != nil {
		t.Fatalf("write shared video: %v", err)
	}
	if err := srv.shares.Save(&ShareSession{
		ID:         "public-share",
		FileName:   "public.mp4",
		VideoPath:  videoPath,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
		CreatedBy:  "alex",
		PublicView: true,
	}); err != nil {
		t.Fatalf("save public share: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/share/public-share", nil)
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected public share metadata route to succeed without auth, got %d", rec.Code)
	}
}
