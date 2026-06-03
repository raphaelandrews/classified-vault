package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"classified-vault/internal/middleware"
)

func setupLogger(env string) *slog.Logger {
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func middlewareChain(h http.Handler) http.Handler {
	rl := middleware.NewRateLimiter(100, 1*time.Minute)
	h = rl.Handler(h)
	h = middleware.CORS(h)
	h = middleware.Logger(h)
	h = middleware.Recovery(h)
	h = requestCounter(h)
	return h
}

func requestCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		next.ServeHTTP(w, r)
	})
}
