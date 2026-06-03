package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := "ok"
		dbStatus := "connected"
		if err := db.Ping(); err != nil {
			status = "degraded"
			dbStatus = "disconnected"
		}
		uptime := time.Since(startupTime).Truncate(time.Second).String()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Fprintf(w, `{"status":"%s","db":"%s","uptime":"%s","requests":%d,"goroutines":%d,"memory_mb":%d}`,
			status, dbStatus, uptime, requestCount.Load(), runtime.NumGoroutine(), m.Alloc/1024/1024)
	}
}
