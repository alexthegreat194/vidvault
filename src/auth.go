package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const pinAuthCookieName = "vv_pin"

// maxPinBytes caps PIN length for constant-time hashing without huge allocations.
const maxPinBytes = 256

func pinEqual(got, want string) bool {
	if want == "" {
		return false
	}
	g := sha256.Sum256([]byte(got))
	w := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(g[:], w[:]) == 1
}

func (s *server) pinSessionToken(gen uint64) string {
	mac := hmac.New(sha256.New, s.authSecret)
	mac.Write([]byte("unlock:"))
	mac.Write([]byte(strconv.FormatUint(gen, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *server) authCookieOK(r *http.Request) bool {
	if s.pin == "" {
		return true
	}
	c, err := r.Cookie(pinAuthCookieName)
	if err != nil || c.Value == "" {
		return false
	}
	s.authMu.RLock()
	gen := s.authGen
	s.authMu.RUnlock()
	want := s.pinSessionToken(gen)
	if len(c.Value) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(want)) == 1
}

func authPublicWithoutSession(r *http.Request) bool {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/":
		return true
	case r.Method == http.MethodGet && r.URL.Path == "/favicon.svg":
		return true
	case r.Method == http.MethodGet && r.URL.Path == "/api/auth/status":
		return true
	case r.Method == http.MethodPost && r.URL.Path == "/api/auth/pin":
		return true
	default:
		return false
	}
}

func (s *server) authDenied(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	if r.URL.Path == "/video" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized\n"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}

func (s *server) applyPinGate(rw *statusWriter, r *http.Request) (handled bool) {
	if s.pin == "" {
		return false
	}
	if s.authCookieOK(r) {
		return false
	}
	if authPublicWithoutSession(r) {
		return false
	}
	s.authDenied(rw, r)
	return true
}

func (s *server) setPinAuthCookie(w http.ResponseWriter) {
	s.authMu.RLock()
	gen := s.authGen
	s.authMu.RUnlock()
	http.SetCookie(w, &http.Cookie{
		Name:     pinAuthCookieName,
		Value:    s.pinSessionToken(gen),
		Path:     "/",
		MaxAge:   86400 * 365 * 10,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *server) clearPinAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     pinAuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.pin == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"pinRequired":   false,
			"authenticated": true,
		})
		return
	}
	ok := s.authCookieOK(r)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"pinRequired":   true,
		"authenticated": ok,
	})
}

func (s *server) handleAuthPin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.pin == "" {
		http.Error(w, "pin not configured", http.StatusBadRequest)
		return
	}
	var body struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(body.Pin) > maxPinBytes {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !pinEqual(body.Pin, s.pin) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false})
		return
	}
	s.setPinAuthCookie(w)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *server) handleAuthLock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.pin == "" {
		http.Error(w, "pin not configured", http.StatusBadRequest)
		return
	}
	s.authMu.Lock()
	s.authGen++
	s.authMu.Unlock()
	s.clearPinAuthCookie(w)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func initPinAuth(s *server, pin string) error {
	s.pin = pin
	if pin == "" {
		return nil
	}
	s.authSecret = make([]byte, 32)
	if _, err := rand.Read(s.authSecret); err != nil {
		return err
	}
	s.authGen = 1
	return nil
}
