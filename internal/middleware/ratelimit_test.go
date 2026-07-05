package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, false)
	for i := 1; i <= 3; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if rl.Allow("1.2.3.4") {
		t.Error("request 4 should be blocked")
	}
	if !rl.Allow("5.6.7.8") {
		t.Error("different IP should not share the limit")
	}
}

func TestClientIPIgnoresSpoofedXFFByDefault(t *testing.T) {
	rl := NewRateLimiter(1, false)
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "9.9.9.9:1234"
	r.Header.Set("X-Forwarded-For", "6.6.6.6")

	if got := rl.clientIP(r); got != "9.9.9.9" {
		t.Errorf("clientIP = %q, want 9.9.9.9 (X-Forwarded-For must be ignored without TRUST_PROXY)", got)
	}
}

func TestClientIPUsesLastXFFHopWhenTrusted(t *testing.T) {
	rl := NewRateLimiter(1, true)
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "10.0.0.1:1234" // the proxy
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 7.7.7.7")

	// Only the right-most entry was appended by our proxy; the left one is
	// client-supplied and spoofable.
	if got := rl.clientIP(r); got != "7.7.7.7" {
		t.Errorf("clientIP = %q, want 7.7.7.7 (right-most XFF hop)", got)
	}
}

func TestRateLimitMiddlewareBlocksAfterLimit(t *testing.T) {
	rl := NewRateLimiter(2, false)
	handler := rl.Middleware()(okHandler())

	for i := 1; i <= 2; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, requestFrom("1.1.1.1:5000"))
		if w.Code != 200 {
			t.Fatalf("request %d: status %d, want 200", i, w.Code)
		}
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestFrom("1.1.1.1:5000"))
	if w.Code != 429 {
		t.Errorf("status = %d, want 429", w.Code)
	}
}
