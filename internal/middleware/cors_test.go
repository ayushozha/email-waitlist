package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ayush10/email-waitlist/internal/model"
)

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		origin  string
		allowed []string
		want    bool
	}{
		{"https://a.com", nil, true},
		{"https://a.com", []string{}, true},
		{"https://a.com", []string{"https://a.com"}, true},
		{"https://evil.com", []string{"https://a.com"}, false},
		{"https://evil.com", []string{"*"}, true},
	}
	for _, tt := range tests {
		if got := isOriginAllowed(tt.origin, tt.allowed); got != tt.want {
			t.Errorf("isOriginAllowed(%q, %v) = %v, want %v", tt.origin, tt.allowed, got, tt.want)
		}
	}
}

func corsRequest(t *testing.T, method, origin string, project *model.Project) *httptest.ResponseRecorder {
	t.Helper()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := httptest.NewRequest(method, "/api/v1/subscribe", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if project != nil {
		r = r.WithContext(context.WithValue(r.Context(), ProjectKey, project))
	}
	w := httptest.NewRecorder()
	CORS(next).ServeHTTP(w, r)
	return w
}

func TestCORSAllowsListedOrigin(t *testing.T) {
	p := &model.Project{AllowedOrigins: []string{"https://a.com"}}
	w := corsRequest(t, http.MethodPost, "https://a.com", p)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://a.com" {
		t.Errorf("ACAO = %q, want https://a.com", got)
	}
	if w.Header().Get("Vary") != "Origin" {
		t.Error("missing Vary: Origin")
	}
}

func TestCORSRejectsUnlistedOrigin(t *testing.T) {
	p := &model.Project{AllowedOrigins: []string{"https://a.com"}}
	w := corsRequest(t, http.MethodPost, "https://evil.com", p)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 — allowed_origins enforcement regressed", w.Code)
	}
}

func TestCORSEmptyAllowlistPermitsAll(t *testing.T) {
	p := &model.Project{AllowedOrigins: nil}
	w := corsRequest(t, http.MethodPost, "https://anywhere.com", p)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestCORSNonBrowserRequestPasses(t *testing.T) {
	p := &model.Project{AllowedOrigins: []string{"https://a.com"}}
	w := corsRequest(t, http.MethodPost, "", p) // no Origin header (curl etc.)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	w := corsRequest(t, http.MethodOptions, "https://a.com", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "PUT") {
		t.Errorf("Allow-Methods = %q, must include PUT for email-template upsert", got)
	}
}
