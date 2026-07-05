package middleware

import (
	"net/http"
	"slices"
)

// CORS enforces each project's allowed_origins and answers preflights.
//
// It must sit *inside* APIKeyAuth in the chain (auth first), because the
// project — and therefore its origin allowlist — is read from the request
// context. Preflight OPTIONS requests never carry custom headers like
// X-API-Key, so they are answered permissively; enforcement happens on the
// actual request, which is rejected outright when its Origin isn't allowed.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Add("Vary", "Origin")

		if r.Method == http.MethodOptions {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Admin-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if origin != "" {
			project := ProjectFromContext(r.Context())
			if project != nil && !isOriginAllowed(origin, project.AllowedOrigins) {
				errorJSON(w, http.StatusForbidden, `{"error":"origin not allowed for this project"}`)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true // no restrictions
	}
	return slices.Contains(allowed, "*") || slices.Contains(allowed, origin)
}
