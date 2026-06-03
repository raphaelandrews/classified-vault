package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += int64(n)
	return n, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrap := &responseWriter{ResponseWriter: w, status: 0}

		next.ServeHTTP(wrap, r)

		duration := time.Since(start)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrap.status,
			"duration_ms", duration.Milliseconds(),
			"size", wrap.size,
			"remote_addr", r.RemoteAddr,
		)
	})
}
