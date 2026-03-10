package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const authCookieName = "video2gif_session"

type loginRequest struct {
	Password string `json:"password"`
}

type authStatusResponse struct {
	Enabled       bool      `json:"enabled"`
	Authenticated bool      `json:"authenticated"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.authEnabled() {
			next.ServeHTTP(w, r)
			return
		}
		ok, _ := s.isAuthenticated(r)
		if !ok {
			jsonError(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	ok, expiresAt := s.isAuthenticated(r)
	jsonOK(w, authStatusResponse{
		Enabled:       s.authEnabled(),
		Authenticated: !s.authEnabled() || ok,
		ExpiresAt:     expiresAt,
	}, http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.authEnabled() {
		jsonOK(w, authStatusResponse{Enabled: false, Authenticated: true}, http.StatusOK)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if subtle.ConstantTimeCompare([]byte(hashPassword(req.Password)), []byte(s.effectivePasswordHash())) != 1 {
		jsonError(w, "invalid password", http.StatusUnauthorized)
		return
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.Auth.SessionTTLHours) * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    s.signedSessionValue(expiresAt),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(r),
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})

	jsonOK(w, authStatusResponse{
		Enabled:       true,
		Authenticated: true,
		ExpiresAt:     expiresAt,
	}, http.StatusOK)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(r),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	jsonOK(w, authStatusResponse{Enabled: s.authEnabled(), Authenticated: false}, http.StatusOK)
}

func (s *Server) authEnabled() bool {
	if !s.cfg.Auth.Enabled {
		return false
	}
	return s.effectivePasswordHash() != ""
}

func (s *Server) isAuthenticated(r *http.Request) (bool, time.Time) {
	if !s.authEnabled() {
		return true, time.Time{}
	}
	cookie, err := r.Cookie(authCookieName)
	if err != nil || cookie.Value == "" {
		return false, time.Time{}
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return false, time.Time{}
	}
	expUnix, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false, time.Time{}
	}
	expiresAt := time.Unix(expUnix, 0)
	if time.Now().After(expiresAt) {
		return false, time.Time{}
	}
	expected := s.signSession(parts[0])
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expected)) != 1 {
		return false, time.Time{}
	}
	return true, expiresAt
}

func (s *Server) signedSessionValue(expiresAt time.Time) string {
	exp := strconv.FormatInt(expiresAt.Unix(), 10)
	return exp + "." + s.signSession(exp)
}

func (s *Server) signSession(exp string) string {
	mac := hmac.New(sha256.New, []byte(s.effectivePasswordHash()))
	mac.Write([]byte(exp))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Server) effectivePasswordHash() string {
	if envHash := strings.TrimSpace(os.Getenv("VIDEO2GIF_PASSWORD_SHA256")); envHash != "" {
		return strings.ToLower(envHash)
	}
	if envPassword := os.Getenv("VIDEO2GIF_PASSWORD"); envPassword != "" {
		return hashPassword(envPassword)
	}
	return strings.ToLower(strings.TrimSpace(s.cfg.Auth.PasswordSHA256))
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func requestIsSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}
