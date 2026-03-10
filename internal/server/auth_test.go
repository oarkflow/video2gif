package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/oarkflow/video2gif/internal/config"
)

func TestAuthMiddlewareAndLoginFlow(t *testing.T) {
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

	router := srv.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated private request, got %d", rec.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{"password": "secret-pass"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected successful login, got %d body=%s", loginRec.Code, loginRec.Body.String())
	}
	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected auth cookie to be set")
	}

	privateReq := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
	privateReq.AddCookie(cookies[0])
	privateRec := httptest.NewRecorder()
	router.ServeHTTP(privateRec, privateReq)
	if privateRec.Code != http.StatusOK {
		t.Fatalf("expected authenticated private request to succeed, got %d", privateRec.Code)
	}

	publicReq := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	publicRec := httptest.NewRecorder()
	router.ServeHTTP(publicRec, publicReq)
	if publicRec.Code != http.StatusOK {
		t.Fatalf("expected public health request to succeed, got %d", publicRec.Code)
	}
}
