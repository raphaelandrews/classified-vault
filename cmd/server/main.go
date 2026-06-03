package main

// @title           Classified Vault API
// @version         1.0
// @description     Secure classified document management system.
// @host            localhost:8080
// @BasePath        /

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"classified-vault/config"
	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/ds"
	"classified-vault/internal/handler"
	"classified-vault/internal/middleware"
	"classified-vault/internal/repository"
	"classified-vault/internal/service"
)

var (
	startupTime  = time.Now()
	requestCount atomic.Int64
)

func main() {
	cfg := config.Load()

	logger := setupLogger(cfg.Environment)
	slog.SetDefault(logger)

	db, err := repository.Connect(cfg.DatabasePath)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.RunMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database initialized", "path", cfg.DatabasePath)

	sessionCache := ds.NewHashMap[auth.Session](256)
	auditBuffer := ds.NewLinkedList[domain.AuditLog]()
	documentIndex := ds.NewAVLTree()

	userRepo := repository.NewUserRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	docs, err := documentRepo.FindAll()
	if err != nil {
		slog.Error("failed to load document index", "error", err)
		os.Exit(1)
	}
	for _, doc := range docs {
		documentIndex.Insert(int(doc.Classification), doc.ID)
	}
	slog.Info("document index built", "count", len(docs))

	authService := service.NewAuthService(userRepo, sessionCache, cfg)
	documentService := service.NewDocumentService(documentRepo, documentIndex, auditBuffer, auditRepo)
	userService := service.NewUserService(userRepo, auditBuffer, auditRepo)
	auditService := service.NewAuditService(auditRepo, auditBuffer)

	authHandler := handler.NewAuthHandler(authService, sessionCache)
	documentHandler := handler.NewDocumentHandler(documentService)
	userHandler := handler.NewUserHandler(userService)
	auditHandler := handler.NewAuditHandler(auditService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	api := middleware.RequireAuth(sessionCache)

	mux.Handle("GET /api/me", api(http.HandlerFunc(authHandler.Me)))

	mux.Handle("GET /api/documents", api(http.HandlerFunc(documentHandler.List)))
	mux.Handle("GET /api/documents/{id}", api(http.HandlerFunc(documentHandler.Get)))
	mux.Handle("POST /api/documents", api(middleware.RequireAnyRole(domain.RoleAdmin, domain.RoleAnalyst)(http.HandlerFunc(documentHandler.Create))))
	mux.Handle("PUT /api/documents/{id}", api(middleware.RequireAnyRole(domain.RoleAdmin, domain.RoleAnalyst)(http.HandlerFunc(documentHandler.Update))))
	mux.Handle("DELETE /api/documents/{id}", api(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(documentHandler.Delete))))

	mux.Handle("GET /api/users", api(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(userHandler.List))))
	mux.Handle("POST /api/users", api(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(userHandler.Create))))
	mux.Handle("PUT /api/users/{id}", api(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(userHandler.Update))))
	mux.Handle("DELETE /api/users/{id}", api(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(userHandler.Delete))))

	mux.Handle("GET /api/audit", api(middleware.RequireRole(domain.RoleAdmin)(http.HandlerFunc(auditHandler.List))))

	mux.HandleFunc("GET /health", healthHandler(db))

	mux.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: middlewareChain(mux),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("flushing audit buffer", "count", auditBuffer.Size())
	for _, entry := range auditBuffer.LastN(auditBuffer.Size()) {
		e := entry
		auditRepo.Save(&e)
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forceful shutdown", "error", err)
	}
	slog.Info("server stopped")
}

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
