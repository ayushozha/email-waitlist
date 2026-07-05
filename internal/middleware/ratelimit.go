package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	limit      int
	window     time.Duration
	trustProxy bool
}

// NewRateLimiter limits requests per client IP. trustProxy must only be true
// when the service sits behind a reverse proxy that appends the real client
// IP to X-Forwarded-For — otherwise the header is client-controlled and would
// let anyone bypass the limit by rotating fake values.
func NewRateLimiter(requestsPerMinute int, trustProxy bool) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitor),
		limit:      requestsPerMinute,
		window:     time.Minute,
		trustProxy: trustProxy,
	}
	// Clean up old entries every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, v := range rl.visitors {
		if now.After(v.resetAt) {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists || now.After(v.resetAt) {
		rl.visitors[ip] = &visitor{count: 1, resetAt: now.Add(rl.window)}
		return true
	}

	v.count++
	return v.count <= rl.limit
}

func (rl *RateLimiter) clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if rl.trustProxy {
		// Only the right-most X-Forwarded-For entry was appended by our own
		// proxy; anything left of it is client-supplied and spoofable.
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if last := strings.TrimSpace(parts[len(parts)-1]); last != "" {
				ip = last
			}
		}
	}
	return ip
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(rl.clientIP(r)) {
				errorJSON(w, http.StatusTooManyRequests, `{"error":"rate limit exceeded, try again later"}`)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
