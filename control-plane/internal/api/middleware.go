package api

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	contextKeyRequestID contextKey = "request_id"
	contextKeyClaims   contextKey = "claims"
)

// Claims holds the verified JWT claims from an OIDC token.
type Claims struct {
	Subject string
	Email   string
	Groups  []string
	Raw     map[string]any
}

// claimsFromContext retrieves the verified claims from the request context.
// Returns nil if the request was not authenticated.
func claimsFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(contextKeyClaims).(*Claims)
	return v
}

// requestIDFromContext retrieves the request ID from the context.
func requestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKeyRequestID).(string)
	return v
}

// ── middlewareRequestID ────────────────────────────────────────────────────────

// middlewareRequestID injects a unique request ID into every request.
// The ID is returned in the X-Request-ID response header.
func middlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), contextKeyRequestID, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ── middlewareLogger ───────────────────────────────────────────────────────────

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// middlewareLogger logs every request with method, path, status, duration, and request ID.
func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", rw.bytes,
			"request_id", requestIDFromContext(r.Context()),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// ── middlewareRecover ──────────────────────────────────────────────────────────

// middlewareRecover recovers from panics and returns a 500 Problem response.
func middlewareRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", rec,
					"stack", string(debug.Stack()),
					"request_id", requestIDFromContext(r.Context()),
				)
				writeProblem(w, r, problemInternal(r, "an unexpected error occurred"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ── middlewareTimeout ──────────────────────────────────────────────────────────

// middlewareTimeout cancels the request context after the given duration.
func middlewareTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ── middlewareCORS ─────────────────────────────────────────────────────────────

// middlewareCORS sets CORS headers. origins is a comma-separated list of allowed origins.
// Pass "*" to allow all origins (dev only — not recommended for production).
func middlewareCORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := false
			for _, o := range origins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ── middlewareOIDC ─────────────────────────────────────────────────────────────

// OIDCConfig holds the configuration for OIDC authentication.
type OIDCConfig struct {
	// IssuerURL is the OIDC provider URL.
	// Examples:
	//   Keycloak: https://keycloak.example.com/realms/cosmonaut
	//   Dex:      https://dex.example.com
	//   Okta:     https://example.okta.com
	IssuerURL string

	// ClientID is the expected audience in the JWT.
	ClientID string

	// Enabled controls whether auth is enforced.
	// Set to false in development to skip authentication.
	Enabled bool
}

// middlewareOIDC validates Bearer tokens against an OIDC provider.
// Verified claims are injected into the request context.
// If cfg.Enabled is false, the middleware is a no-op (dev mode).
func middlewareOIDC(cfg OIDCConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	provider, err := oidc.NewProvider(context.Background(), cfg.IssuerURL)
	if err != nil {
		panic("failed to initialize OIDC provider: " + err.Error())
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeProblem(w, r, problemUnauthorized(r))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeProblem(w, r, problemBadRequest(r, "authorization header must be: Bearer <token>"))
				return
			}

			token, err := verifier.Verify(r.Context(), parts[1])
			if err != nil {
				slog.Warn("OIDC token verification failed",
					"err", err,
					"request_id", requestIDFromContext(r.Context()),
				)
				writeProblem(w, r, problemUnauthorized(r))
				return
			}

			// extract standard claims
			var raw map[string]any
			if err := token.Claims(&raw); err != nil {
				writeProblem(w, r, problemInternal(r, "failed to parse token claims"))
				return
			}

			claims := &Claims{
				Subject: token.Subject,
				Raw:     raw,
			}

			if email, ok := raw["email"].(string); ok {
				claims.Email = email
			}

			// groups claim — common in Keycloak and Dex
			if groups, ok := raw["groups"].([]any); ok {
				for _, g := range groups {
					if s, ok := g.(string); ok {
						claims.Groups = append(claims.Groups, s)
					}
				}
			}

			ctx := context.WithValue(r.Context(), contextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
