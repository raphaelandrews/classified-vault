# Classified Vault — Arquitetura do Projeto

> Sistema de Gerenciamento de Documentos Classificados  
> Stack: Go (net/http stdlib) + Go (Bubble Tea) + SQLite  
> Paradigma: REST API + TUI Client

---

## Sumário

1. [Visão Geral](#visão-geral)
2. [Estrutura de Diretórios](#estrutura-de-diretórios)
3. [Domínio e Entidades](#domínio-e-entidades)
4. [Estruturas de Dados Implementadas Manualmente](#estruturas-de-dados-implementadas-manualmente)
5. [Backend — Arquitetura](#backend--arquitetura)
6. [Backend — Rotas da API](#backend--rotas-da-api)
7. [Backend — Middleware e Segurança](#backend--middleware-e-segurança)
8. [Frontend TUI — Telas](#frontend-tui--telas)
9. [Frontend TUI — Modelo Bubble Tea](#frontend-tui--modelo-bubble-tea)
10. [Banco de Dados — Schema](#banco-de-dados--schema)
11. [Fluxo de Autenticação](#fluxo-de-autenticação)
12. [Fluxo de Acesso a Documentos](#fluxo-de-acesso-a-documentos)
13. [Sistema de Permissões RBAC](#sistema-de-permissões-rbac)
14. [Log de Auditoria](#log-de-auditoria)
15. [Deploy](#deploy)
16. [Ordem de Implementação](#ordem-de-implementação)
17. [Dependências](#dependências)
18. [Backend Stack Decisions](#backend-stack-decisions)
19. [O que Apresentar para a Banca](#o-que-apresentar-para-a-banca)
20. [Hardening & Known Gaps](#hardening--known-gaps)

---

## Visão Geral

O **Classified Vault** é um sistema de gerenciamento de documentos sigilosos inspirado em sistemas reais de controle de acesso governamental e corporativo. Cada documento possui um nível de classificação, e cada usuário possui um *clearance level* — um usuário só pode acessar documentos cujo nível seja igual ou inferior ao seu.

### Níveis de Classificação (ordem crescente de sigilo)

```
PUBLIC < RESTRICTED < CONFIDENTIAL < SECRET < TOP_SECRET
  0          1              2           3          4
```

### Perfis de Usuário (Roles)

| Role      | Clearance Máximo | Pode Gerenciar Usuários | Pode Criar Documentos | Pode Ver Auditoria |
|-----------|-----------------|------------------------|----------------------|-------------------|
| `admin`   | TOP_SECRET      | ✅                      | ✅                    | ✅                 |
| `analyst` | CONFIDENTIAL    | ❌                      | ✅                    | ❌                 |
| `viewer`  | RESTRICTED      | ❌                      | ❌                    | ❌                 |
| `intern`  | PUBLIC          | ❌                      | ❌                    | ❌                 |

---

## Estrutura de Diretórios

```
classified-vault/
├── cmd/
│   ├── server/
│   │   └── main.go                  # Entrypoint do backend
│   └── client/
│       └── main.go                  # Entrypoint da TUI
│
├── internal/
│   ├── domain/
│   │   ├── user.go                  # User entity
│   │   ├── document.go              # Document entity
│   │   ├── audit.go                 # AuditLog entity
│   │   └── clearance.go             # ClearanceLevel enum + helpers
│   │
│   ├── ds/                          # Hand-rolled data structures
│   │   ├── avl_tree.go              # AVL tree — document index by classification
│   │   ├── linked_list.go           # Doubly linked list — audit log buffer
│   │   └── hash_map.go              # HashMap — session cache
│   │
│   ├── auth/
│   │   ├── token.go                 # UUID v4 token generation & validation
│   │   ├── bcrypt.go                # Password hashing & verification
│   │   └── middleware.go            # RequireAuth, RequireRole, RequireAnyRole
│   │
│   ├── apperr/
│   │   └── apperr.go                # Structured AppError type
│   │
│   ├── repository/
│   │   ├── user_repo.go             # CRUD users in SQLite
│   │   ├── document_repo.go         # CRUD documents in SQLite
│   │   └── audit_repo.go            # Persist audit logs
│   │
│   ├── service/
│   │   ├── auth_service.go          # Login, logout, token refresh
│   │   ├── document_service.go      # Access with clearance verification
│   │   ├── user_service.go          # User management (admin only)
│   │   └── audit_service.go         # Audit log queries and persistence
│   │
│   ├── handler/
│       ├── auth_handler.go          # POST /auth/login, /auth/logout
│       ├── document_handler.go      # CRUD /api/documents
│       ├── user_handler.go          # CRUD /api/users (admin)
│       └── audit_handler.go         # GET /api/audit (admin)
│
│   ├── middleware/
│       ├── auth.go                  # RequireAuth, RequireRole, RequireAnyRole, RequireClearance
│       ├── logger.go                # slog-based request logging
│       ├── cors.go                  # CORS headers
│       ├── recovery.go              # Panic recovery
│       └── ratelimit.go             # Token-bucket rate limiter
│
├── tui/
│   ├── app.go                       # Root Bubble Tea model
│   ├── styles/
│   │   └── styles.go                # Centralized Lip Gloss styles
│   ├── screens/
│   │   ├── login.go                 # Login screen (Huh form)
│   │   ├── dashboard.go             # Post-login dashboard
│   │   ├── documents.go             # Accessible document list
│   │   ├── document_view.go         # Document viewer
│   │   ├── document_create.go       # Creation form (Huh)
│   │   ├── users.go                 # User management (admin)
│   │   ├── audit_log.go             # Audit log viewer (admin)
│   │   └── access_denied.go         # Permission denied screen
│   └── client/
│       └── api.go                   # HTTP client for the backend
│
├── migrations/
│   ├── 000_schema_migrations.sql    # Migration tracker table
│   ├── 001_create_users.sql
│   ├── 002_create_documents.sql
│   └── 003_create_audit_logs.sql
│
├── config/
│   └── config.go                    # Config via environment variables
│
├── docs/                            # Swagger-generated docs (gitignored)
├── .air.toml                        # Air hot-reload config
├── .env.example                     # Documented env vars
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile                       # Backend deploy
├── render.yaml                      # Render Blueprint (infra-as-code)
├── tmp/                             # Air build output (gitignored)
└── README.md
```

---

## Domínio e Entidades

### User

```go
// internal/domain/user.go

type ClearanceLevel int

const (
    ClearancePublic      ClearanceLevel = 0
    ClearanceRestricted  ClearanceLevel = 1
    ClearanceConfidential ClearanceLevel = 2
    ClearanceSecret      ClearanceLevel = 3
    ClearanceTopSecret   ClearanceLevel = 4
)

func (c ClearanceLevel) String() string {
    switch c {
    case ClearancePublic:       return "PUBLIC"
    case ClearanceRestricted:   return "RESTRICTED"
    case ClearanceConfidential: return "CONFIDENTIAL"
    case ClearanceSecret:       return "SECRET"
    case ClearanceTopSecret:    return "TOP SECRET"
    default:                    return "UNKNOWN"
    }
}

func (c ClearanceLevel) Label() string {
    labels := map[ClearanceLevel]string{
        ClearancePublic:       "⬜ PUBLIC",
        ClearanceRestricted:   "🟦 RESTRICTED",
        ClearanceConfidential: "🟨 CONFIDENTIAL",
        ClearanceSecret:       "🟧 SECRET",
        ClearanceTopSecret:    "🟥 TOP SECRET",
    }
    return labels[c]
}

type Role string

const (
    RoleAdmin   Role = "admin"
    RoleAnalyst Role = "analyst"
    RoleViewer  Role = "viewer"
    RoleIntern  Role = "intern"
)

type User struct {
    ID           string         `json:"id"`
    Username     string         `json:"username"`
    PasswordHash string         `json:"-"`
    Email        string         `json:"email"`
    Role         Role           `json:"role"`
    Clearance    ClearanceLevel `json:"clearance"`
    Active       bool           `json:"active"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
}

func MaxClearanceForRole(role Role) ClearanceLevel {
    switch role {
    case RoleAdmin:   return ClearanceTopSecret
    case RoleAnalyst: return ClearanceConfidential
    case RoleViewer:  return ClearanceRestricted
    case RoleIntern:  return ClearancePublic
    default:          return ClearancePublic
    }
}
```

### Document

```go
// internal/domain/document.go

type DocumentStatus string

const (
    StatusActive   DocumentStatus = "active"
    StatusArchived DocumentStatus = "archived"
    StatusRevoked  DocumentStatus = "revoked"
)

type Document struct {
    ID           string         `json:"id"`
    Title        string         `json:"title"`
    Content      string         `json:"content"`
    Classification ClearanceLevel `json:"classification"`
    Status       DocumentStatus `json:"status"`
    Tags         []string       `json:"tags"`
    CreatedBy    string         `json:"created_by"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
}
```

### AuditLog

```go
// internal/domain/audit.go

type AuditAction string

const (
    ActionLogin          AuditAction = "LOGIN"
    ActionLogout         AuditAction = "LOGOUT"
    ActionLoginFailed    AuditAction = "LOGIN_FAILED"
    ActionDocumentRead   AuditAction = "DOCUMENT_READ"
    ActionDocumentCreate AuditAction = "DOCUMENT_CREATE"
    ActionDocumentUpdate AuditAction = "DOCUMENT_UPDATE"
    ActionDocumentDelete AuditAction = "DOCUMENT_DELETE"
    ActionAccessDenied   AuditAction = "ACCESS_DENIED"
    ActionUserCreated    AuditAction = "USER_CREATED"
    ActionUserUpdated    AuditAction = "USER_UPDATED"
    ActionUserDeleted    AuditAction = "USER_DELETED"
)

type AuditLog struct {
    ID         string      `json:"id"`
    UserID     string      `json:"user_id"`
    Username   string      `json:"username"`
    Action     AuditAction `json:"action"`
    Resource   string      `json:"resource"`    // ex: "document:abc123"
    IPAddress  string      `json:"ip_address"`
    Success    bool        `json:"success"`
    Details    string      `json:"details"`
    Timestamp  time.Time   `json:"timestamp"`
}
```

---

## Estruturas de Dados Implementadas Manualmente

> Estas estruturas são implementadas do zero, sem usar as da stdlib, para atender ao requisito de "Estrutura de Dados" da disciplina. Cada uma tem uma justificativa de uso no sistema.

### 1. Hash Map — Cache de Sessões

Usado para armazenar tokens JWT ativos em memória com O(1) de lookup. Implementação com separate chaining para colisões.

```go
// internal/ds/hash_map.go

type Entry[V any] struct {
    Key   string
    Value V
    Next  *Entry[V]
}

type HashMap[V any] struct {
    buckets  []*Entry[V]
    size     int
    count    int
    mu       sync.RWMutex
}

func NewHashMap[V any](size int) *HashMap[V] {
    return &HashMap[V]{
        buckets: make([]*Entry[V], size),
        size:    size,
    }
}

func (h *HashMap[V]) hash(key string) int {
    hash := 0
    for _, c := range key {
        hash = (hash*31 + int(c)) % h.size
    }
    return hash
}

func (h *HashMap[V]) Set(key string, value V) {
    h.mu.Lock()
    defer h.mu.Unlock()

    idx := h.hash(key)
    entry := &Entry[V]{Key: key, Value: value}

    if h.buckets[idx] == nil {
        h.buckets[idx] = entry
    } else {
        current := h.buckets[idx]
        for current != nil {
            if current.Key == key {
                current.Value = value
                return
            }
            if current.Next == nil {
                current.Next = entry
                break
            }
            current = current.Next
        }
    }
    h.count++
}

func (h *HashMap[V]) Get(key string) (V, bool) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    idx := h.hash(key)
    current := h.buckets[idx]
    for current != nil {
        if current.Key == key {
            return current.Value, true
        }
        current = current.Next
    }
    var zero V
    return zero, false
}

func (h *HashMap[V]) Delete(key string) bool {
    h.mu.Lock()
    defer h.mu.Unlock()

    idx := h.hash(key)
    current := h.buckets[idx]

    if current == nil {
        return false
    }

    if current.Key == key {
        h.buckets[idx] = current.Next
        h.count--
        return true
    }

    for current.Next != nil {
        if current.Next.Key == key {
            current.Next = current.Next.Next
            h.count--
            return true
        }
        current = current.Next
    }
    return false
}

func (h *HashMap[V]) Count() int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.count
}
```

**Uso no sistema:**
```go
// Session armazenada no cache
type Session struct {
    UserID    string
    Username  string
    Role      Role
    Clearance ClearanceLevel
    ExpiresAt time.Time
}

var sessionCache = ds.NewHashMap[Session](256)
```

---

### 2. Lista Encadeada — Log de Auditoria em Memória

Append-only linked list. Novos eventos são adicionados na cabeça (O(1)). Nunca remove entradas — garante imutabilidade do histórico em memória antes de persistir no banco.

```go
// internal/ds/linked_list.go

type ListNode[T any] struct {
    Value T
    Next  *ListNode[T]
    Prev  *ListNode[T]
}

type LinkedList[T any] struct {
    Head  *ListNode[T]
    Tail  *ListNode[T]
    size  int
    mu    sync.RWMutex
}

func NewLinkedList[T any]() *LinkedList[T] {
    return &LinkedList[T]{}
}

// Append adiciona no final — mantém ordem cronológica
func (l *LinkedList[T]) Append(value T) {
    l.mu.Lock()
    defer l.mu.Unlock()

    node := &ListNode[T]{Value: value}
    if l.Tail == nil {
        l.Head = node
        l.Tail = node
    } else {
        node.Prev = l.Tail
        l.Tail.Next = node
        l.Tail = node
    }
    l.size++
}

// ToSlice retorna os N últimos elementos (para paginação no log)
func (l *LinkedList[T]) LastN(n int) []T {
    l.mu.RLock()
    defer l.mu.RUnlock()

    result := make([]T, 0, n)
    current := l.Tail
    for current != nil && len(result) < n {
        result = append([]T{current.Value}, result...)
        current = current.Prev
    }
    return result
}

func (l *LinkedList[T]) Size() int {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.size
}
```

---

### 3. Árvore AVL — Índice de Documentos por Classificação

Árvore de busca binária balanceada. Indexa documentos por `ClearanceLevel`, permitindo buscar eficientemente "todos os documentos com classificação <= X" (operação crucial para filtrar o que um usuário pode ver).

```go
// internal/ds/avl_tree.go

type AVLNode struct {
    Key      int    // ClearanceLevel como int
    DocIDs   []string
    Height   int
    Left     *AVLNode
    Right    *AVLNode
}

type AVLTree struct {
    Root *AVLNode
    mu   sync.RWMutex
}

func NewAVLTree() *AVLTree {
    return &AVLTree{}
}

func height(n *AVLNode) int {
    if n == nil {
        return 0
    }
    return n.Height
}

func balanceFactor(n *AVLNode) int {
    if n == nil {
        return 0
    }
    return height(n.Left) - height(n.Right)
}

func updateHeight(n *AVLNode) {
    lh := height(n.Left)
    rh := height(n.Right)
    if lh > rh {
        n.Height = lh + 1
    } else {
        n.Height = rh + 1
    }
}

func rotateRight(y *AVLNode) *AVLNode {
    x := y.Left
    t := x.Right
    x.Right = y
    y.Left = t
    updateHeight(y)
    updateHeight(x)
    return x
}

func rotateLeft(x *AVLNode) *AVLNode {
    y := x.Right
    t := y.Left
    y.Left = x
    x.Right = t
    updateHeight(x)
    updateHeight(y)
    return y
}

func insert(node *AVLNode, key int, docID string) *AVLNode {
    if node == nil {
        return &AVLNode{Key: key, DocIDs: []string{docID}, Height: 1}
    }

    if key < node.Key {
        node.Left = insert(node.Left, key, docID)
    } else if key > node.Key {
        node.Right = insert(node.Right, key, docID)
    } else {
        node.DocIDs = append(node.DocIDs, docID)
        return node
    }

    updateHeight(node)
    bf := balanceFactor(node)

    // LL
    if bf > 1 && key < node.Left.Key {
        return rotateRight(node)
    }
    // RR
    if bf < -1 && key > node.Right.Key {
        return rotateLeft(node)
    }
    // LR
    if bf > 1 && key > node.Left.Key {
        node.Left = rotateLeft(node.Left)
        return rotateRight(node)
    }
    // RL
    if bf < -1 && key < node.Right.Key {
        node.Right = rotateRight(node.Right)
        return rotateLeft(node)
    }

    return node
}

func (t *AVLTree) Insert(clearanceLevel int, docID string) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.Root = insert(t.Root, clearanceLevel, docID)
}

// QueryUpTo retorna todos os docIDs com classificação <= maxLevel
func (t *AVLTree) QueryUpTo(maxLevel int) []string {
    t.mu.RLock()
    defer t.mu.RUnlock()

    var result []string
    var traverse func(node *AVLNode)
    traverse = func(node *AVLNode) {
        if node == nil {
            return
        }
        if node.Key <= maxLevel {
            traverse(node.Left)
            result = append(result, node.DocIDs...)
            traverse(node.Right)
        } else {
            traverse(node.Left)
        }
    }
    traverse(t.Root)
    return result
}

func (t *AVLTree) Remove(clearanceLevel int, docID string) {
    t.mu.Lock()
    defer t.mu.Unlock()

    var removeFromNode func(node *AVLNode)
    removeFromNode = func(node *AVLNode) {
        if node == nil {
            return
        }
        if node.Key == clearanceLevel {
            newDocIDs := make([]string, 0)
            for _, id := range node.DocIDs {
                if id != docID {
                    newDocIDs = append(newDocIDs, id)
                }
            }
            node.DocIDs = newDocIDs
        } else if clearanceLevel < node.Key {
            removeFromNode(node.Left)
        } else {
            removeFromNode(node.Right)
        }
    }
    removeFromNode(t.Root)
}
```

---

## Backend — Arquitetura

```
cmd/server/main.go
    │
    ├── Config (env vars)
    ├── Init slog structured logger (JSON in production, text in dev)
    ├── SQLite connection with WAL mode + busy timeout (modernc.org/sqlite)
    ├── Run versioned migrations (schema_migrations tracker table)
    ├── Init data structures (HashMap, AVLTree, LinkedList)
    ├── Init repositories
    ├── Init services
    ├── Init handlers
    └── net/http Server
            ├── Middleware chain: Logging, CORS, RateLimiter, Recovery
            ├── /health          → Health check (DB status, uptime)
            ├── /auth/*          → AuthHandler (public)
            ├── /api/*           → RequireAuth middleware
            │       ├── /documents/*  → DocumentHandler
            │       ├── /users/*      → UserHandler + RequireRole(admin)
            │       └── /audit/*      → AuditHandler + RequireRole(admin)
            ├── /docs/*          → Swagger UI (swaggo/swag, dev only)
            ├── /debug/pprof/*   → pprof profiling endpoint (stdlib)
            └── Graceful shutdown (signal.NotifyContext, 10s drain)
```

### main.go

```go
// cmd/server/main.go

package main

import (
    "context"
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

    "classified-vault/config"
    "classified-vault/internal/domain"
    "classified-vault/internal/ds"
    "classified-vault/internal/auth"
    "classified-vault/internal/repository"
    "classified-vault/internal/service"
    "classified-vault/internal/handler"

    _ "github.com/joho/godotenv/autoload"
)

var startupTime = time.Now()

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

    sessionCache  := ds.NewHashMap[auth.Session](256)
    auditBuffer   := ds.NewLinkedList[domain.AuditLog]()
    documentIndex := ds.NewAVLTree()

    userRepo     := repository.NewUserRepository(db)
    documentRepo := repository.NewDocumentRepository(db)
    auditRepo    := repository.NewAuditRepository(db)

    docs, err := documentRepo.FindAll()
    if err != nil {
        slog.Error("failed to load document index", "error", err)
        os.Exit(1)
    }
    for _, doc := range docs {
        documentIndex.Insert(int(doc.Classification), doc.ID)
    }
    slog.Info("document index built", "count", len(docs))

    authService     := service.NewAuthService(userRepo, sessionCache, cfg)
    documentService := service.NewDocumentService(documentRepo, documentIndex, auditBuffer, auditRepo)
    userService     := service.NewUserService(userRepo, auditBuffer, auditRepo)
    auditService    := service.NewAuditService(auditRepo, auditBuffer)

    authHandler     := handler.NewAuthHandler(authService)
    documentHandler := handler.NewDocumentHandler(documentService)
    userHandler     := handler.NewUserHandler(userService)
    auditHandler    := handler.NewAuditHandler(auditService)

    mux := http.NewServeMux()

    // Public routes
    mux.HandleFunc("POST /auth/login",  authHandler.Login)
    mux.HandleFunc("POST /auth/logout", authHandler.Logout)

    // Protected routes
    api := auth.RequireAuth(sessionCache)

    mux.Handle("GET /api/me", api(authHandler.Me))

    mux.Handle("GET /api/documents",     api(documentHandler.List))
    mux.Handle("GET /api/documents/{id}", api(documentHandler.Get))
    mux.Handle("POST /api/documents",     api(auth.RequireAnyRole(domain.RoleAdmin, domain.RoleAnalyst)(documentHandler.Create)))
    mux.Handle("PUT /api/documents/{id}", api(auth.RequireAnyRole(domain.RoleAdmin, domain.RoleAnalyst)(documentHandler.Update)))
    mux.Handle("DELETE /api/documents/{id}", api(auth.RequireRole(domain.RoleAdmin)(documentHandler.Delete)))

    mux.Handle("GET /api/users",     api(auth.RequireRole(domain.RoleAdmin)(userHandler.List)))
    mux.Handle("POST /api/users",     api(auth.RequireRole(domain.RoleAdmin)(userHandler.Create)))
    mux.Handle("PUT /api/users/{id}", api(auth.RequireRole(domain.RoleAdmin)(userHandler.Update)))
    mux.Handle("DELETE /api/users/{id}", api(auth.RequireRole(domain.RoleAdmin)(userHandler.Delete)))

    mux.Handle("GET /api/audit", api(auth.RequireRole(domain.RoleAdmin)(auditHandler.List)))

    // Health check
    mux.HandleFunc("GET /health", healthHandler(db))

    // pprof (stdlib, no extra deps)
    mux.Handle("GET /debug/pprof/", http.DefaultServeMux)
    mux.Handle("GET /debug/pprof/{profile}", http.DefaultServeMux)

    // Middleware chain (outermost → innermost)
    handler := corsMiddleware(
        loggerMiddleware(
            recoveryMiddleware(
                rateLimitMiddleware(mux),
            ),
        ),
    )

    srv := &http.Server{
        Addr:    ":" + cfg.ServerPort,
        Handler: handler,
    }

    // Graceful shutdown
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

    // Flush audit buffer to DB before exiting
    slog.Info("flushing audit buffer", "count", auditBuffer.Size())
    for _, entry := range auditBuffer.LastN(auditBuffer.Size()) {
        auditRepo.Save(entry)
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

var requestCount atomic.Int64

func healthHandler(db repository.DB) http.HandlerFunc {
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
```

---

## Backend — Rotas da API

### Autenticação

| Método | Rota           | Auth | Descrição                        |
|--------|----------------|------|----------------------------------|
| POST   | /auth/login    | ❌    | Login, retorna JWT               |
| POST   | /auth/logout   | ✅    | Invalida token no cache          |
| GET    | /api/me        | ✅    | Dados do usuário logado          |

### Documentos

| Método | Rota               | Auth | Clearance         | Descrição                              |
|--------|--------------------|------|-------------------|----------------------------------------|
| GET    | /api/documents     | ✅    | Qualquer          | Lista docs acessíveis ao clearance     |
| GET    | /api/documents/:id | ✅    | >= classificação  | Lê documento (ou 403)                  |
| POST   | /api/documents     | ✅    | analyst+          | Cria documento                         |
| PUT    | /api/documents/:id | ✅    | analyst+ e autor  | Atualiza documento                     |
| DELETE | /api/documents/:id | ✅    | admin             | Remove documento                       |

### Usuários (admin only)

| Método | Rota            | Auth | Role  | Descrição          |
|--------|-----------------|------|-------|--------------------|
| GET    | /api/users      | ✅    | admin | Lista usuários     |
| POST   | /api/users      | ✅    | admin | Cria usuário       |
| PUT    | /api/users/:id  | ✅    | admin | Atualiza usuário   |
| DELETE | /api/users/:id  | ✅    | admin | Remove usuário     |

### Auditoria (admin only)

| Método | Rota        | Auth | Role  | Descrição               |
|--------|-------------|------|-------|-------------------------|
| GET    | /api/audit  | ✅    | admin | Lista logs de auditoria |

### Health & Debug

| Método | Rota                | Auth | Description                                                  |
|--------|---------------------|------|--------------------------------------------------------------|
| GET    | /health             | ❌    | JSON: status, DB connectivity, uptime, goroutines, memory    |
| GET    | /debug/pprof/       | ❌    | pprof index page (stdlib)                                   |
| GET    | /debug/pprof/{name} | ❌    | pprof profile (goroutine, heap, allocs, profile, trace, etc.)|

---

## Backend — Middleware e Segurança

### Auth Middleware (net/http)

```go
// internal/auth/middleware.go

type Session struct {
    UserID    string
    Username  string
    Role      domain.Role
    Clearance domain.ClearanceLevel
    ExpiresAt time.Time
}

// RequireAuth wraps an http.Handler; rejects unauthenticated requests.
func RequireAuth(cache *ds.HashMap[Session]) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            session, ok := cache.Get(token)
            if !ok || time.Now().After(session.ExpiresAt) {
                if ok {
                    cache.Delete(token)
                }
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), ctxKeySession, session)
            ctx = context.WithValue(ctx, ctxKeyToken, token)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RequireRole(role domain.Role) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            session := r.Context().Value(ctxKeySession).(Session)
            if session.Role != role {
                http.Error(w, `{"error":"permission denied"}`, http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func RequireAnyRole(roles ...domain.Role) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            session := r.Context().Value(ctxKeySession).(Session)
            for _, role := range roles {
                if session.Role == role {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            http.Error(w, `{"error":"permission denied"}`, http.StatusForbidden)
        })
    }
}

func RequireClearance(level domain.ClearanceLevel) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            session := r.Context().Value(ctxKeySession).(Session)
            if session.Clearance < level {
                http.Error(w, `{"error":"insufficient clearance"}`, http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func extractToken(r *http.Request) string {
    auth := r.Header.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    return ""
}
```

### Document Service — Verificação de Clearance

```go
// internal/service/document_service.go

func (s *DocumentService) GetByID(userSession auth.Session, docID string) (*domain.Document, error) {
    doc, err := s.repo.FindByID(docID)
    if err != nil {
        return nil, err
    }

    // VERIFICAÇÃO CENTRAL DE SEGURANÇA
    if userSession.Clearance < doc.Classification {
        // Registra tentativa negada no log
        s.logAudit(domain.AuditLog{
            UserID:   userSession.UserID,
            Username: userSession.Username,
            Action:   domain.ActionAccessDenied,
            Resource: "document:" + docID,
            Success:  false,
            Details:  fmt.Sprintf("clearance %s < %s", userSession.Clearance, doc.Classification),
        })
        return nil, ErrAccessDenied
    }

    // Registra acesso bem-sucedido
    s.logAudit(domain.AuditLog{
        UserID:   userSession.UserID,
        Username: userSession.Username,
        Action:   domain.ActionDocumentRead,
        Resource: "document:" + docID,
        Success:  true,
    })

    return doc, nil
}

func (s *DocumentService) List(userSession auth.Session) ([]*domain.Document, error) {
    // Usa a AVL Tree para buscar IDs acessíveis em O(log n)
    accessibleIDs := s.index.QueryUpTo(int(userSession.Clearance))
    return s.repo.FindByIDs(accessibleIDs)
}

func (s *DocumentService) logAudit(log domain.AuditLog) {
    log.ID = uuid.New().String()
    log.Timestamp = time.Now()
    // Append na lista encadeada em memória
    s.auditBuffer.Append(log)
    // Persiste no banco de forma assíncrona
    go s.auditRepo.Save(log)
}
```

---

## Banco de Dados — Schema

SQLite in WAL mode (`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;`) for concurrent reads during writes. Migrations are versioned via a `schema_migrations` tracker table — only unapplied versions run on startup.

```sql
-- migrations/000_schema_migrations.sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version   INTEGER PRIMARY KEY,
    name      TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- migrations/001_create_users.sql
CREATE TABLE IF NOT EXISTS users (
    id           TEXT PRIMARY KEY,
    username     TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email        TEXT UNIQUE NOT NULL,
    role         TEXT NOT NULL DEFAULT 'intern',
    clearance    INTEGER NOT NULL DEFAULT 0,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Default admin user (password: admin123 — change in production)
INSERT OR IGNORE INTO users (id, username, password_hash, email, role, clearance)
VALUES (
    'usr_admin_0000000001',
    'admin',
    '$2a$12$LJ3m4ys3Kk0mMfFq1B8qHeEqHqFm3jNrSPRQukPMtRzPVQqK9aGhu',
    'admin@vault.local',
    'admin',
    4
);

-- migrations/002_create_documents.sql
CREATE TABLE IF NOT EXISTS documents (
    id             TEXT PRIMARY KEY,
    title          TEXT NOT NULL,
    content        TEXT NOT NULL,
    classification INTEGER NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'active',
    tags           TEXT NOT NULL DEFAULT '[]',
    created_by     TEXT NOT NULL REFERENCES users(id),
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_documents_classification ON documents(classification);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);

-- migrations/003_create_audit_logs.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    username   TEXT NOT NULL,
    action     TEXT NOT NULL,
    resource   TEXT NOT NULL,
    ip_address TEXT,
    success    BOOLEAN NOT NULL DEFAULT TRUE,
    details    TEXT,
    timestamp  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_user_id  ON audit_logs(user_id);
```

---

## Fluxo de Autenticação

```
TUI Client                          Backend
    │                                   │
    │  POST /auth/login                 │
    │  { username, password }           │
    ├──────────────────────────────────►│
    │                                   │ 1. Busca user no SQLite
    │                                   │ 2. bcrypt.CompareHashAndPassword
    │                                   │ 3. Gera token UUID v4
    │                                   │ 4. Salva Session no HashMap
    │                                   │    key=token, value=Session{...}
    │                                   │ 5. Registra LOGIN no audit log
    │◄──────────────────────────────────│
    │  200 { token, user }              │
    │                                   │
    │  (armazena token em memória)      │
    │                                   │
    │  GET /api/documents               │
    │  Authorization: Bearer <token>    │
    ├──────────────────────────────────►│
    │                                   │ 1. Extrai token do header
    │                                   │ 2. HashMap.Get(token) → Session
    │                                   │ 3. Verifica expiração
    │                                   │ 4. AVLTree.QueryUpTo(clearance)
    │                                   │ 5. Retorna docs filtrados
    │◄──────────────────────────────────│
    │  200 [{ doc1 }, { doc2 }]        │
    │                                   │
    │  GET /api/documents/:secret_id    │
    ├──────────────────────────────────►│
    │                                   │ 1. Session.Clearance < SECRET
    │                                   │ 2. Registra ACCESS_DENIED
    │◄──────────────────────────────────│
    │  403 { error: "Acesso Negado" }  │
    │                                   │
    │  (exibe tela de Access Denied)   │
```

---

## Fluxo de Acesso a Documentos

```
Usuário tenta acessar documento classificado como CONFIDENTIAL (nível 2)

┌─────────────────────────────────────────────────────────────────┐
│                    VERIFICAÇÃO DE CLEARANCE                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Usuário clearance 0 (PUBLIC)    →  🚫 ACESSO NEGADO            │
│  Usuário clearance 1 (RESTRICTED) →  🚫 ACESSO NEGADO           │
│  Usuário clearance 2 (CONFIDENTIAL) → ✅ ACESSO PERMITIDO        │
│  Usuário clearance 3 (SECRET)    →  ✅ ACESSO PERMITIDO          │
│  Usuário clearance 4 (TOP SECRET) → ✅ ACESSO PERMITIDO          │
│                                                                  │
│  Regra: session.Clearance >= document.Classification            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Sistema de Permissões RBAC

```
                         PERMISSÕES POR ROLE
                         
┌────────────────┬────────┬──────────┬────────┬────────┐
│ Ação           │ intern │  viewer  │analyst │ admin  │
├────────────────┼────────┼──────────┼────────┼────────┤
│ Login          │  ✅    │    ✅    │   ✅   │   ✅   │
│ Ver docs PUBLIC│  ✅    │    ✅    │   ✅   │   ✅   │
│ Ver RESTRICTED │  ❌    │    ✅    │   ✅   │   ✅   │
│ Ver CONFIDENTIAL│ ❌    │    ❌    │   ✅   │   ✅   │
│ Ver SECRET     │  ❌    │    ❌    │   ❌   │   ✅   │
│ Ver TOP SECRET │  ❌    │    ❌    │   ❌   │   ✅   │
│ Criar docs     │  ❌    │    ❌    │   ✅   │   ✅   │
│ Gerenciar users│  ❌    │    ❌    │   ❌   │   ✅   │
│ Ver audit log  │  ❌    │    ❌    │   ❌   │   ✅   │
└────────────────┴────────┴──────────┴────────┴────────┘
```

---

## Log de Auditoria

Cada ação relevante gera uma entrada imutável no log. O admin pode visualizar em tempo real na TUI.

### Formato de uma entrada

```
[2025-06-02 14:32:01]  ✅  LOGIN          user:joao_silva          →  system
[2025-06-02 14:32:15]  ✅  DOCUMENT_READ  user:joao_silva          →  document:abc123 [RESTRICTED]
[2025-06-02 14:32:22]  🚫  ACCESS_DENIED  user:joao_silva          →  document:xyz789 [SECRET] (clearance insuficiente)
[2025-06-02 14:35:10]  ✅  LOGIN          user:maria_admin         →  system
[2025-06-02 14:35:45]  ✅  USER_CREATED   user:maria_admin         →  user:novo_analista [analyst/CONFIDENTIAL]
[2025-06-02 14:40:00]  🚫  LOGIN_FAILED   user:???                 →  system (username: hacker, senha errada)
```

---

## Frontend TUI — Telas

### Mapa de Navegação

```
                    ┌─────────────────┐
                    │   TELA LOGIN    │
                    │  ┌───────────┐  │
                    │  │ username  │  │
                    │  │ password  │  │
                    │  │  [Entrar] │  │
                    └──┴───────────┴──┘
                            │
              ┌─────────────┼──────────────┐
              │             │              │
        [intern/viewer]  [analyst]      [admin]
              │             │              │
              ▼             ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────────┐
        │DASHBOARD │  │DASHBOARD │  │  DASHBOARD   │
        │          │  │          │  │   (+ admin   │
        │Documentos│  │Documentos│  │    panel)    │
        │  PUBLIC  │  │  até     │  └──────────────┘
        │          │  │  CONF.   │         │
        └──────────┘  └──────────┘         │
              │             │         ┌────┴────┐
              ▼             ▼         ▼         ▼
        ┌──────────┐  ┌──────────┐ ┌──────┐ ┌───────┐
        │ LISTA DE │  │ LISTA DE │ │USERS │ │ AUDIT │
        │   DOCS   │  │   DOCS   │ │ MGMT │ │  LOG  │
        └──────────┘  └──────────┘ └──────┘ └───────┘
              │
              ▼
    ┌──────────────────┐
    │  ao tentar doc   │
    │  acima clearance │
    ├──────────────────┤
    │  🚫 ACESSO       │
    │     NEGADO       │
    └──────────────────┘
```

### Tela de Login

```
╔══════════════════════════════════════════════╗
║           🔒  CLASSIFIED VAULT               ║
║          Sistema de Acesso Seguro            ║
╠══════════════════════════════════════════════╣
║                                              ║
║  Usuário  ┌──────────────────────────────┐  ║
║           │ admin_                        │  ║
║           └──────────────────────────────┘  ║
║                                              ║
║  Senha    ┌──────────────────────────────┐  ║
║           │ ••••••••                      │  ║
║           └──────────────────────────────┘  ║
║                                              ║
║             [ Entrar ]  [ Sair ]             ║
║                                              ║
╚══════════════════════════════════════════════╝
```

### Dashboard (Admin)

```
╔══════════════════════════════════════════════════════════════════╗
║  🔒 CLASSIFIED VAULT          Logado: admin  [TOP SECRET] [admin] ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  ┌─────────────────────────────────────────────────────────┐    ║
║  │  DOCUMENTOS (32 acessíveis)                             │    ║
║  │  [D] Listar Documentos                                  │    ║
║  │  [N] Novo Documento                                     │    ║
║  └─────────────────────────────────────────────────────────┘    ║
║                                                                  ║
║  ┌─────────────────────────────────────────────────────────┐    ║
║  │  ADMINISTRAÇÃO                                          │    ║
║  │  [U] Gerenciar Usuários                                 │    ║
║  │  [A] Log de Auditoria                                   │    ║
║  └─────────────────────────────────────────────────────────┘    ║
║                                                                  ║
║  [Q] Logout                                                      ║
╚══════════════════════════════════════════════════════════════════╝
```

### Lista de Documentos

```
╔══════════════════════════════════════════════════════════════════╗
║  📁 DOCUMENTOS — clearance: CONFIDENTIAL                         ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  ▶ Relatório Anual de Segurança         ⬜ PUBLIC      2025-01-10║
║    Manual de Procedimentos Internos     🟦 RESTRICTED  2025-02-14║
║    Análise de Vulnerabilidades Q1       🟨 CONFIDENTIAL 2025-03-01║
║    ████████████████████████████████    🟧 SECRET      [BLOQUEADO]║
║    ████████████████████████████████    🟥 TOP SECRET  [BLOQUEADO]║
║                                                                  ║
║  [Enter] Abrir  [N] Novo  [/] Buscar  [Q] Voltar                 ║
╚══════════════════════════════════════════════════════════════════╝
```

### Tela de Acesso Negado

```
╔══════════════════════════════════════════════╗
║                                              ║
║                    🚫                        ║
║                                              ║
║           A C E S S O   N E G A D O          ║
║                                              ║
║   Você não possui clearance suficiente       ║
║   para acessar este documento.               ║
║                                              ║
║   Seu clearance:    🟨 CONFIDENTIAL          ║
║   Requerido:        🟧 SECRET                ║
║                                              ║
║   Esta tentativa foi registrada no log.      ║
║                                              ║
║              [ Voltar ]                      ║
║                                              ║
╚══════════════════════════════════════════════╝
```

### Log de Auditoria (Admin)

```
╔══════════════════════════════════════════════════════════════════╗
║  📋 LOG DE AUDITORIA — últimas 50 entradas          [Admin]      ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  14:40:00  🚫  LOGIN_FAILED    hacker          system            ║
║  14:35:45  ✅  USER_CREATED    maria_admin     user:new_analyst  ║
║  14:35:10  ✅  LOGIN           maria_admin     system            ║
║  14:32:22  🚫  ACCESS_DENIED   joao_silva      doc:xyz [SECRET]  ║
║  14:32:15  ✅  DOCUMENT_READ   joao_silva      doc:abc           ║
║  14:32:01  ✅  LOGIN           joao_silva      system            ║
║                                                                  ║
║  [R] Atualizar  [/] Filtrar por usuário  [Q] Voltar              ║
╚══════════════════════════════════════════════════════════════════╝
```

---

## Frontend TUI — Modelo Bubble Tea

```go
// tui/app.go

type Screen int

const (
    ScreenLogin Screen = iota
    ScreenDashboard
    ScreenDocumentList
    ScreenDocumentView
    ScreenDocumentCreate
    ScreenUserManagement
    ScreenAuditLog
    ScreenAccessDenied
)

type AppModel struct {
    screen       Screen
    session      *client.Session     // nil se não logado
    apiClient    *client.APIClient
    
    // Sub-modelos de cada tela
    loginModel   screens.LoginModel
    dashModel    screens.DashboardModel
    docListModel screens.DocumentListModel
    docViewModel screens.DocumentViewModel
    docCreateModel screens.DocumentCreateModel
    usersModel   screens.UsersModel
    auditModel   screens.AuditLogModel
    accessDeniedModel screens.AccessDeniedModel
    
    err         error
    width       int
    height      int
}

func (m AppModel) Init() tea.Cmd {
    return m.loginModel.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

    case client.LoginSuccessMsg:
        m.session = &msg.Session
        m.screen = ScreenDashboard
        return m, m.dashModel.Init()

    case client.LogoutMsg:
        m.session = nil
        m.screen = ScreenLogin
        return m, m.loginModel.Init()

    case client.AccessDeniedMsg:
        m.accessDeniedModel = screens.NewAccessDeniedModel(msg)
        m.screen = ScreenAccessDenied

    case client.NavigateMsg:
        return m.handleNavigation(msg)
    }

    return m.updateCurrentScreen(msg)
}

func (m AppModel) View() string {
    switch m.screen {
    case ScreenLogin:          return m.loginModel.View()
    case ScreenDashboard:      return m.dashModel.View()
    case ScreenDocumentList:   return m.docListModel.View()
    case ScreenDocumentView:   return m.docViewModel.View()
    case ScreenDocumentCreate: return m.docCreateModel.View()
    case ScreenUserManagement: return m.usersModel.View()
    case ScreenAuditLog:       return m.auditModel.View()
    case ScreenAccessDenied:   return m.accessDeniedModel.View()
    default:                   return ""
    }
}
```

---

## Config

```go
// config/config.go
package config

import (
    "os"
    "strconv"
    "time"

    _ "github.com/joho/godotenv/autoload" // loads .env automatically on import
)

type Config struct {
    DatabasePath string
    JWTSecret    string
    ServerPort   string
    ServerURL    string // used by TUI client
    SessionTTL   time.Duration
    Environment  string
}

func Load() *Config {
    return &Config{
        DatabasePath: getEnv("DATABASE_PATH", "./vault.db"),
        JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
        ServerPort:   getEnv("PORT", "8080"),
        ServerURL:    getEnv("SERVER_URL", "http://localhost:8080"),
        SessionTTL:   parseDuration(getEnv("SESSION_TTL", "8h")),
        Environment:  getEnv("ENV", "development"),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

## Deploy

### Render (recommended)

```yaml
# render.yaml — Blueprint spec
services:
  - type: web
    name: classified-vault
    env: docker
    dockerfilePath: ./Dockerfile
    healthCheckPath: /health
    envVars:
      - key: PORT
        value: 8080
      - key: ENV
        value: production
      - key: DATABASE_PATH
        value: /data/vault.db
    disk:
      name: vault-data
      mountPath: /data
      sizeGB: 1
```

```bash
# Deploy via Blueprint (render.yaml in repo)
# 1. Push to GitHub
# 2. Connect repo in Render dashboard → "Blueprints"
# 3. Set secrets: JWT_SECRET, ADMIN_PASSWORD
# Or deploy manually: New Web Service → Docker → point to repo
```

### Compilar client para Windows

```makefile
# Makefile

.PHONY: build-client-windows build-client-linux build-server

build-server:
	go build -o bin/server ./cmd/server

build-client-linux:
	go build -o bin/classified-vault ./cmd/client

build-client-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/classified-vault.exe ./cmd/client

build-all: build-server build-client-linux build-client-windows

run-server:
	./bin/server

run-client:
	./bin/classified-vault --server http://localhost:8080
```

---

## Ordem de Implementação

### Fase 1 — Fundação
- [x] Estrutura de diretórios e `go.mod`
- [x] Domínio: entidades User, Document, AuditLog, ClearanceLevel
- [x] Schema SQLite + migrations
- [x] Estruturas de dados: HashMap, LinkedList, AVLTree

### Fase 2 — Backend Core
- [x] `internal/apperr/` — structured AppError type
- [x] `internal/middleware/` — auth, logger, cors, recovery, ratelimit
- [x] `internal/handler/` — HTTP handlers (auth, document, user, audit)
- [x] `config/config.go` — add godotenv autoload
- [x] Repositórios (user, document, audit) com WAL mode connection
- [x] Versioned migration runner (`repository/migrate.go`)
- [x] Auth service (bcrypt, UUID v4 token, login, logout)
- [x] Document service com verificação de clearance
- [x] `cmd/server/main.go` — net/http server with health check, pprof, graceful shutdown
- [x] Swagger annotations on all handlers + `swag init`
- [x] `.air.toml` + `make dev` (hot-reload)
- [x] `.env.example` with documented env vars

### Fase 3 — Backend Complementar
- [x] User service (CRUD admin) — Create, Update, Delete with audit logging and auto-clearance
- [x] Audit service — List, ListByUser, async DB persistence
- [x] Integração AVLTree no document service — QueryUpTo, Insert, Remove on CRUD
- [x] Testes manuais via curl — smoke test script (`scripts/smoke_test.sh`) — all endpoints verified

### Fase 4 — TUI
- [x] Styles com Lip Gloss (paleta, borders, badges de clearance)
- [x] API client HTTP em Go
- [x] Tela de login (textinput, password, error handling)
- [x] Dashboard (role-aware panels, doc count)
- [x] Lista de documentos (cursor nav, clearance badges, blocked entries)
- [x] Visualização de documento (full content display)
- [x] Tela de Access Denied (reason, clearance comparison)
- [x] Formulário de criação (multi-step: title → content → classification → tags)
- [x] Gestão de usuários (list, add with role picker, delete)
- [x] Log de auditoria (timeline view, success/failure icons)

### Fase 5 — Deploy e Polimento
- [x] Deploy config — Dockerfile + `render.yaml` (Render Blueprint), one-click deploy
- [x] Compilar `.exe` para Windows — both server and TUI client cross-compile
- [x] Testar no Windows Terminal — ANSI-compatible TUI via Bubble Tea
- [x] README com instruções de uso — comprehensive English README
- [x] Popular banco com dados de demo — `cmd/seed/main.go` (4 users, 7 documents across all clearance levels)

---

## Dependências

### Backend (`go.mod`)

```
github.com/google/uuid              // ID generation
github.com/joho/godotenv            // .env file loader (autoload)
github.com/swaggo/http-swagger/v2   // Swagger UI endpoint (dev only)
github.com/swaggo/swag              // Swagger annotation generator
golang.org/x/crypto                 // bcrypt
modernc.org/sqlite                  // SQLite, pure Go (no CGO)
```

> No HTTP framework dependency. Go 1.22+ `net/http` handles routing with path params natively. Logging uses stdlib `log/slog`. Profiling via stdlib `net/http/pprof`. The total dependency count is kept minimal to emphasize the hand-written data structures.

### Dev Tools (not in `go.mod`)

```
github.com/cosmtrek/air             // Hot-reload on save: `air` rebuilds + restarts automatically
```

### Client TUI (`go.mod` — mesmo módulo ou separado)

```
github.com/charmbracelet/bubbletea  // TUI framework
github.com/charmbracelet/lipgloss   // Estilos
github.com/charmbracelet/bubbles    // Componentes prontos (list, textinput, spinner)
github.com/charmbracelet/huh        // Formulários
```

---

## Backend Stack Decisions

### Why net/http over Fiber/Gin/Echo

Go 1.22+ `net/http.ServeMux` supports `"GET /api/documents/{id}"` route patterns with path params. For a project of this scope (15 routes, ~5 middlewares), an external framework is unnecessary weight. Benefits:

| Decision | Rationale |
|---|---|
| **stdlib mux** | Zero HTTP dependencies. The banca sees direct Go, not a third-party abstraction. |
| **`log/slog`** | Structured, leveled logging from stdlib (Go 1.21+). No need for zerolog/zap. |
| **`context.Context`** | Native session propagation via `r.WithContext()` — no custom `c.Locals()` or framework-specific context. |
| **middleware as `func(http.Handler) http.Handler`** | Standard Go middleware pattern. Readers/Writers are stdlib interfaces. |
| **graceful shutdown** | `signal.NotifyContext` + `srv.Shutdown()` is 15 lines of stdlib. Fiber's graceful shutdown is essentially the same, so no advantage. |

### Robustness additions (built into Phase 2)

| Feature | Implementation |
|---|---|
| **Structured logging** | `slog.NewJSONHandler` in prod, `slog.NewTextHandler` in dev. All services log through the default logger. |
| **SQLite WAL mode** | 2 pragmas on connect: `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000`. Allows concurrent reads during writes. |
| **Versioned migrations** | `migrations/000_schema_migrations.sql` tracker table. `RunMigrations` reads `migrations/*.sql` sorted, skips already-applied versions. |
| **Graceful shutdown** | SIGINT/SIGTERM → stop accepting requests → flush audit buffer to DB → close server. 10s timeout. |
| **Swagger/OpenAPI** | `swaggo/swag` annotations on handlers. `/docs` serves interactive Swagger UI in dev mode. |
| **Structured errors** | `type AppError struct { Code int; Message string }` — handlers return this, a top-level response wrapper marshals to `{"error": msg}`. |
| **Health check** | `GET /health` returns `{"status":"ok","db":"connected","uptime":"3m42s","requests":152,"goroutines":7,"memory_mb":4}`. No auth required. |
| **pprof** | `GET /debug/pprof/goroutine?debug=1` gives live goroutine dump. stdlib, zero deps. |
| **godotenv** | `_ "github.com/joho/godotenv/autoload"` loads `.env` on import. No manual `export` needed in dev. |
| **Air** | `go install github.com/cosmtrek/air@latest && air` — hot-reload on file save. Rebuilds and restarts in < 1s. |

---

## O que Apresentar para a Banca

### Roteiro de Demo (10–15 minutos)

1. **Mostrar o código das estruturas de dados** — abrir `internal/ds/` e explicar AVL Tree, HashMap e LinkedList em 2 minutos. Isso diferencia o projeto.

2. **Login como `intern`** — mostrar que só vê documentos PUBLIC. Tentar acessar RESTRICTED → tela de "Acesso Negado".

3. **Login como `analyst`** — mostrar acesso até CONFIDENTIAL. Tentar SECRET → negado. Criar um novo documento.

4. **Login como `admin`** — acesso total. Abrir o log de auditoria e mostrar todas as tentativas negadas dos usuários anteriores registradas em tempo real. Criar um novo usuário, definir clearance, testar login com o novo usuário.

5. **Mostrar o deploy** — abrir o `.exe` no Windows apontando para o backend hospedado no Fly.io. Demonstrar que é um binário único sem instalação.

### Pontos para enfatizar
- As estruturas de dados não são da stdlib — foram implementadas do zero
- O log de auditoria é imutável (append-only list)
- A AVL Tree é usada para filtrar documentos acessíveis em O(log n)
- O HashMap de sessões evita consulta ao banco em cada request
- bcrypt para senhas, tokens seguros para sessões
- Zero dependências no executável do cliente
```

---

## Hardening & Known Gaps

> Items to address before or shortly after the core implementation. Checked items are resolved in the updated architecture.

### Minor Gaps — Fix Before Handlers

- [x] **`main.go` — silent error ignore**  
  Resolved. The new `main.go` checks `documentRepo.FindAll()` error with `slog.Error` + `os.Exit(1)`.

- [x] **Rename `auth/jwt.go` → `auth/token.go`**  
  Resolved. File is `internal/auth/token.go`, uses UUID v4 via `github.com/google/uuid`.

- [x] **AVL `Remove` never prunes empty nodes**  
  Resolved. Proper AVL delete with rebalancing. Empty nodes are now removed from the tree.

- [x] **RequireAnyRole middleware** — done in Phase 2.

- [x] **Input validation**  
  Resolved. `internal/validate/` package validates username (alphanumeric, 3–32), password (≥8), email, document title (≤256), content (≤64KB).

- [x] **Graceful shutdown** — done in Phase 2.

### Nice-to-add

- [x] **Seed script** (`cmd/seed/main.go`)  
  Creates 4 users (admin/analyst/viewer/intern) and 7 documents across all clearance levels. Run: `go run ./cmd/seed`

- [x] **Data structure unit tests** (`internal/ds/*_test.go`)  
  11 tests covering HashMap (get/set/delete/collisions), LinkedList (append/lastN), AVLTree (insert/query/remove/rebalance). All pass.

- [x] **Swagger annotations**  
  Annotations on main (title/description), auth handler (login), and document handler (list, get). Docs served at `/docs/`.

- [x] **Rebuild AVL index**  
  `AVLTree.RebuildIndex(docs)` method added. Can reconstruct the full index from database state.