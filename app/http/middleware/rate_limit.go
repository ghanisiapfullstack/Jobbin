package middleware

import (
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http"
)

type rateLimitEntry struct {
	count    int
	resetAt  time.Time
}

type RateLimit struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	limit   int
	window  time.Duration
}

func NewRateLimitMiddleware(limit int, window time.Duration) *RateLimit {
	r := &RateLimit{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
	}
	// Cleanup goroutine — hapus entries expired setiap 5 menit
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			r.mu.Lock()
			now := time.Now()
			for ip, entry := range r.entries {
				if now.After(entry.resetAt) {
					delete(r.entries, ip)
				}
			}
			r.mu.Unlock()
		}
	}()
	return r
}

func (r *RateLimit) Handle() http.Middleware {
	return func(ctx http.Context) {
		ip := ctx.Request().Ip()

		r.mu.Lock()
		entry, exists := r.entries[ip]
		now := time.Now()

		if !exists || now.After(entry.resetAt) {
			r.entries[ip] = &rateLimitEntry{
				count:   1,
				resetAt: now.Add(r.window),
			}
			r.mu.Unlock()
			ctx.Request().Next()
			return
		}

		if entry.count >= r.limit {
			r.mu.Unlock()
			ctx.Request().AbortWithStatusJson(429, http.Json{
				"message": "Terlalu banyak percobaan. Coba lagi dalam 1 menit.",
			})
			return
		}

		entry.count++
		r.mu.Unlock()
		ctx.Request().Next()
	}
}
