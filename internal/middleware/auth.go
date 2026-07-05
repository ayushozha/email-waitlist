package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const ProjectKey contextKey = "project"

func ProjectFromContext(ctx context.Context) *model.Project {
	p, _ := ctx.Value(ProjectKey).(*model.Project)
	return p
}

// Scope controls which project key types an endpoint accepts.
type Scope int

const (
	// ScopeSubscribe accepts the publishable key or the secret key.
	ScopeSubscribe Scope = iota
	// ScopeManage accepts only the secret key. Publishable keys are embedded
	// in browser code, so they must never grant access to subscriber data.
	ScopeManage
)

func APIKeyAuth(pool *pgxpool.Pool, scope Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				errorJSON(w, http.StatusUnauthorized, `{"error":"missing X-API-Key header"}`)
				return
			}

			if scope == ScopeManage && strings.HasPrefix(apiKey, model.PublicKeyPrefix) {
				errorJSON(w, http.StatusForbidden, `{"error":"publishable key cannot access management endpoints; use the secret key"}`)
				return
			}

			var (
				project *model.Project
				err     error
			)
			if scope == ScopeSubscribe {
				project, err = model.GetProjectForSubscribe(r.Context(), pool, apiKey)
			} else {
				project, err = model.GetProjectBySecretKey(r.Context(), pool, apiKey)
			}
			if err != nil {
				errorJSON(w, http.StatusUnauthorized, `{"error":"invalid API key"}`)
				return
			}

			ctx := context.WithValue(r.Context(), ProjectKey, project)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminAuth(adminKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Hash both sides so the comparison is constant-time and doesn't
			// leak the admin key's length.
			got := sha256.Sum256([]byte(r.Header.Get("X-Admin-Key")))
			want := sha256.Sum256([]byte(adminKey))
			if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
				errorJSON(w, http.StatusUnauthorized, `{"error":"unauthorized"}`)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// errorJSON writes a JSON error with a permissive CORS header so browsers can
// read failures produced before per-project CORS enforcement runs (auth and
// rate limiting). "*" is safe here — no credentials mode is involved.
func errorJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Add("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(body))
}
