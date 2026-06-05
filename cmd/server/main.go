package main

// @title           Pelican Town Archives API
// @version         1.0
// @description     Secure scroll management system for Pelican Town public services.
// @host            localhost:8080
// @BasePath        /

import (
	"context"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"classified-vault/config"
	"classified-vault/internal/auth"
	vaultcrypto "classified-vault/internal/crypto"
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

	vaultcrypto.InitVault(cfg.VaultKey, nil)
	slog.Info("vault encryption initialized")

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
	slog.Info("scroll index built", "count", len(docs))

	authService := service.NewAuthService(userRepo, sessionCache, cfg)
	documentService := service.NewDocumentService(documentRepo, documentIndex, auditBuffer, auditRepo)
	userService := service.NewUserService(userRepo, auditBuffer, auditRepo)
	auditService := service.NewAuditService(auditRepo, auditBuffer)

	if err := userService.SeedMayor(cfg.AdminPassword); err != nil {
		slog.Error("failed to seed mayor", "error", err)
		os.Exit(1)
	}

	authHandler := handler.NewAuthHandler(authService, sessionCache)
	documentHandler := handler.NewDocumentHandler(documentService)
	userHandler := handler.NewUserHandler(userService)
	auditHandler := handler.NewAuditHandler(auditService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/register", authHandler.Register)

	api := middleware.RequireAuth(sessionCache)

	mux.Handle("GET /api/me", api(http.HandlerFunc(authHandler.Me)))

	mux.Handle("GET /api/documents", api(http.HandlerFunc(documentHandler.List)))
	mux.Handle("GET /api/documents/{id}", api(http.HandlerFunc(documentHandler.Get)))
	mux.Handle("GET /api/catalog", api(http.HandlerFunc(documentHandler.Catalog)))
	mux.Handle("POST /api/documents", api(middleware.RequireAnyRole(domain.RoleMayor, domain.RoleKeeper)(http.HandlerFunc(documentHandler.Create))))
	mux.Handle("PUT /api/documents/{id}", api(middleware.RequireAnyRole(domain.RoleMayor, domain.RoleKeeper)(http.HandlerFunc(documentHandler.Update))))
	mux.Handle("DELETE /api/documents/{id}", api(middleware.RequireRole(domain.RoleMayor)(http.HandlerFunc(documentHandler.Delete))))
	mux.Handle("PUT /api/documents/{id}/transition", api(middleware.RequireAnyRole(domain.RoleMayor, domain.RoleKeeper)(http.HandlerFunc(documentHandler.Transition))))

	mux.Handle("GET /api/users", api(middleware.RequireRole(domain.RoleMayor)(http.HandlerFunc(userHandler.List))))
	mux.Handle("POST /api/users", api(middleware.RequireRole(domain.RoleMayor)(http.HandlerFunc(userHandler.Create))))
	mux.Handle("PUT /api/users/{id}", api(middleware.RequireRole(domain.RoleMayor)(http.HandlerFunc(userHandler.Update))))
	mux.Handle("DELETE /api/users/{id}", api(middleware.RequireRole(domain.RoleMayor)(http.HandlerFunc(userHandler.Delete))))

	mux.Handle("GET /api/audit", api(middleware.RequireRole(domain.RoleMayor)(http.HandlerFunc(auditHandler.List))))

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

	slog.Info("flushing town ledger", "count", auditBuffer.Size())
	for _, entry := range auditBuffer.LastN(auditBuffer.Size()) {
		e := entry
		auditRepo.Save(&e)
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forceful shutdown", "error", err)
	}
	slog.Info("server stopped")
}
