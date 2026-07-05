package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestChainOrder locks in that chain applies middleware outermost-first.
// Auth must run before CORS (CORS reads the project from request context),
// so a regression here silently disables allowed_origins enforcement.
func TestChainOrder(t *testing.T) {
	var order []string
	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	}), mw("first"), mw("second"))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	want := []string{"first", "second", "handler"}
	for i := range want {
		if i >= len(order) || order[i] != want[i] {
			t.Fatalf("execution order = %v, want %v", order, want)
		}
	}
}
