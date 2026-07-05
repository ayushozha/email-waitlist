package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func requestFrom(remoteAddr string) *http.Request {
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = remoteAddr
	return r
}

func TestAdminAuth(t *testing.T) {
	handler := AdminAuth("s3cret")(okHandler())

	tests := []struct {
		name string
		key  string
		want int
	}{
		{"correct key", "s3cret", http.StatusOK},
		{"wrong key", "nope", http.StatusUnauthorized},
		{"empty key", "", http.StatusUnauthorized},
		{"prefix of key", "s3cre", http.StatusUnauthorized},
		{"key with suffix", "s3cret1", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/v1/projects", nil)
			if tt.key != "" {
				r.Header.Set("X-Admin-Key", tt.key)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestErrorJSONSetsCORSAndContentType(t *testing.T) {
	w := httptest.NewRecorder()
	errorJSON(w, http.StatusUnauthorized, `{"error":"x"}`)
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("auth errors must be readable cross-origin")
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", w.Header().Get("Content-Type"))
	}
}
