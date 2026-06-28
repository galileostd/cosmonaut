package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/galileostd/cosmonaut/internal/db"
)

// ── session store ─────────────────────────────────────────────────────────────

type session struct {
	UserID    string
	Username  string
	Role      string
	ExpiresAt time.Time
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session
}

func newSessionStore() *sessionStore {
	s := &sessionStore{sessions: make(map[string]session)}
	go s.pruneLoop()
	return s
}

func (s *sessionStore) create(userID, username, role string, ttl time.Duration) string {
	token := mustRandHex(32)
	s.mu.Lock()
	s.sessions[token] = session{
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Now().Add(ttl),
	}
	s.mu.Unlock()
	return token
}

func (s *sessionStore) get(token string) (session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || time.Now().After(sess.ExpiresAt) {
		return session{}, false
	}
	return sess, true
}

func (s *sessionStore) delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *sessionStore) pruneLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for token, sess := range s.sessions {
			if now.After(sess.ExpiresAt) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}
}

func mustRandHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("cosmonaut: failed to generate random token: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// ── request / response types ──────────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

type meResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// handleLogin authenticates a user with username + password.
// POST /api/v1/auth/login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if prob := decodeJSON(r, &req); prob != nil {
		writeProblem(w, r, prob)
		return
	}

	if req.Username == "" || req.Password == "" {
		writeProblem(w, r, problemUnprocessable(r, "username and password are required"))
		return
	}

	var user db.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeProblem(w, r, problemUnauthorized(r))
			return
		}
		writeProblem(w, r, problemInternal(r, "database error"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeProblem(w, r, problemUnauthorized(r))
		return
	}

	const sessionTTL = 24 * time.Hour
	token := s.sessions.create(user.ID, user.Username, user.Role, sessionTTL)

	slog.Info("user logged in", "username", user.Username, "role", user.Role)

	writeJSON(w, http.StatusOK, loginResponse{
		Token:     token,
		Username:  user.Username,
		Role:      user.Role,
		ExpiresAt: time.Now().Add(sessionTTL),
	})
}

// handleLogout invalidates the current session token.
// POST /api/v1/auth/logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token != "" {
		s.sessions.delete(token)
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleMe returns the currently authenticated user.
// GET /api/v1/auth/me
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		writeProblem(w, r, problemUnauthorized(r))
		return
	}

	sess, ok := s.sessions.get(token)
	if !ok {
		writeProblem(w, r, problemUnauthorized(r))
		return
	}

	var user db.User
	if err := s.db.Where("id = ?", sess.UserID).First(&user).Error; err != nil {
		writeProblem(w, r, problemInternal(r, "database error"))
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
	})
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	// fallback: cookie (for browser UI)
	if cookie, err := r.Cookie("cosmonaut_session"); err == nil {
		return cookie.Value
	}
	return ""
}
