package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	count    int
	windowAt time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup(5 * time.Minute)
	return rl
}

func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(r.RemoteAddr) {
			http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]

	if !ok || now.After(v.windowAt) {
		rl.visitors[ip] = &visitor{count: 1, windowAt: now.Add(rl.window)}
		return true
	}

	if v.count >= rl.limit {
		return false
	}

	v.count++
	return true
}

func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.After(v.windowAt) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
