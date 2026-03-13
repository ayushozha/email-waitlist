package middleware

import (
	"net/http"
	"slices"

	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CORS(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if r.Method == http.MethodOptions {
				// For preflight, look up the project by API key if provided
				apiKey := r.Header.Get("X-API-Key")
				if apiKey != "" && origin != "" {
					project, err := model.GetProjectByAPIKey(r.Context(), pool, apiKey)
					if err == nil && isOriginAllowed(origin, project.AllowedOrigins) {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
				} else {
					// Allow preflight without API key (browser sends OPTIONS without custom headers sometimes)
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Admin-Key")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// For actual requests, check origin against project's allowed origins
			project := ProjectFromContext(r.Context())
			if project != nil && origin != "" {
				if len(project.AllowedOrigins) == 0 || isOriginAllowed(origin, project.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			} else if origin != "" {
				// Admin endpoints or no project context — allow
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Admin-Key")
			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true // no restrictions
	}
	return slices.Contains(allowed, "*") || slices.Contains(allowed, origin)
}
