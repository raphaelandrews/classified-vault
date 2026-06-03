# Classified Vault — Arquitetura do Projeto

> Sistema de Gerenciamento de Documentos Classificados  
> Stack: Go (Fiber) + Go (Bubble Tea) + SQLite  
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
18. [O que Apresentar para a Banca](#o-que-apresentar-para-a-banca)
19. [Hardening & Known Gaps](#hardening--known-gaps)

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
│   │   ├── user.go                  # Entidade User
│   │   ├── document.go              # Entidade Document
│   │   ├── audit.go                 # Entidade AuditLog
│   │   └── clearance.go             # Enum ClearanceLevel + helpers
│   │
│   ├── ds/                          # Estruturas de dados implementadas manualmente
│   │   ├── avl_tree.go              # Árvore AVL para índice de documentos
│   │   ├── linked_list.go           # Lista encadeada para log de auditoria
│   │   └── hash_map.go              # Hash map para cache de sessões
│   │
│   ├── auth/
│   │   ├── jwt.go                   # Geração e validação de JWT
│   │   ├── bcrypt.go                # Hash e verificação de senha
│   │   └── middleware.go            # Middleware de autenticação e autorização
│   │
│   ├── repository/
│   │   ├── user_repo.go             # CRUD de usuários no SQLite
│   │   ├── document_repo.go         # CRUD de documentos no SQLite
│   │   └── audit_repo.go            # Persistência do log de auditoria
│   │
│   ├── service/
│   │   ├── auth_service.go          # Lógica de login, registro, refresh
│   │   ├── document_service.go      # Lógica de acesso com verificação de clearance
│   │   ├── user_service.go          # Gestão de usuários (admin only)
│   │   └── audit_service.go         # Consulta e persistência de logs
│   │
│   └── handler/
│       ├── auth_handler.go          # POST /auth/login, /auth/logout
│       ├── document_handler.go      # CRUD /documents
│       ├── user_handler.go          # CRUD /users (admin)
│       └── audit_handler.go         # GET /audit (admin)
│
├── tui/
│   ├── app.go                       # Root model do Bubble Tea
│   ├── styles/
│   │   └── styles.go                # Lip Gloss styles centralizados
│   ├── screens/
│   │   ├── login.go                 # Tela de login (Huh form)
│   │   ├── dashboard.go             # Dashboard pós-login
│   │   ├── documents.go             # Lista de documentos acessíveis
│   │   ├── document_view.go         # Visualização de documento
│   │   ├── document_create.go       # Formulário de criação (Huh)
│   │   ├── users.go                 # Gestão de usuários (admin)
│   │   ├── audit_log.go             # Visualização do log (admin)
│   │   └── access_denied.go         # Tela de permissão negada
│   └── client/
│       └── api.go                   # HTTP client para o backend
│
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_documents.sql
│   └── 003_create_audit_logs.sql
│
├── config/
│   └── config.go                    # Config via variáveis de ambiente
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile                       # Para deploy do backend
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
    ├── SQLite connection (modernc.org/sqlite)
    ├── Run migrations
    ├── Init data structures (HashMap, AVLTree, LinkedList)
    ├── Init repositories
    ├── Init services
    ├── Init handlers
    └── Fiber app
            ├── Middleware: Logger, CORS, RateLimiter
            ├── /auth/*       → AuthHandler (público)
            ├── /api/*        → RequireAuth middleware
            │       ├── /documents/*  → DocumentHandler
            │       ├── /users/*      → UserHandler + RequireAdmin
            │       └── /audit/*      → AuditHandler + RequireAdmin
            └── Listen :8080
```

### main.go

```go
// cmd/server/main.go

package main

import (
    "log"
    "os"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/limiter"

    "classified-vault/config"
    "classified-vault/internal/ds"
    "classified-vault/internal/repository"
    "classified-vault/internal/service"
    "classified-vault/internal/handler"
    "classified-vault/internal/auth"
)

func main() {
    cfg := config.Load()

    db, err := repository.Connect(cfg.DatabasePath)
    if err != nil {
        log.Fatal("failed to connect to database:", err)
    }
    defer db.Close()

    if err := repository.RunMigrations(db); err != nil {
        log.Fatal("failed to run migrations:", err)
    }

    // Estruturas de dados globais
    sessionCache  := ds.NewHashMap[auth.Session](256)
    auditBuffer   := ds.NewLinkedList[domain.AuditLog]()
    documentIndex := ds.NewAVLTree()

    // Repositórios
    userRepo     := repository.NewUserRepository(db)
    documentRepo := repository.NewDocumentRepository(db)
    auditRepo    := repository.NewAuditRepository(db)

    // Popular AVL Tree com documentos existentes
    docs, _ := documentRepo.FindAll()
    for _, doc := range docs {
        documentIndex.Insert(int(doc.Classification), doc.ID)
    }

    // Serviços
    authService     := service.NewAuthService(userRepo, sessionCache, cfg)
    documentService := service.NewDocumentService(documentRepo, documentIndex, auditBuffer, auditRepo)
    userService     := service.NewUserService(userRepo, auditBuffer, auditRepo)
    auditService    := service.NewAuditService(auditRepo, auditBuffer)

    // Handlers
    authHandler     := handler.NewAuthHandler(authService)
    documentHandler := handler.NewDocumentHandler(documentService)
    userHandler     := handler.NewUserHandler(userService)
    auditHandler    := handler.NewAuditHandler(auditService)

    // Fiber
    app := fiber.New(fiber.Config{
        ErrorHandler: handler.ErrorHandler,
    })

    app.Use(logger.New())
    app.Use(cors.New())
    app.Use(limiter.New(limiter.Config{
        Max: 100,
    }))

    // Rotas públicas
    app.Post("/auth/login",  authHandler.Login)
    app.Post("/auth/logout", authHandler.Logout)

    // Rotas protegidas
    api := app.Group("/api", auth.RequireAuth(sessionCache))

    api.Get("/me", authHandler.Me)

    docs := api.Group("/documents")
    docs.Get("/",       documentHandler.List)
    docs.Get("/:id",    documentHandler.Get)
    docs.Post("/",      documentHandler.Create)
    docs.Put("/:id",    documentHandler.Update)
    docs.Delete("/:id", documentHandler.Delete)

    users := api.Group("/users", auth.RequireRole(domain.RoleAdmin))
    users.Get("/",       userHandler.List)
    users.Post("/",      userHandler.Create)
    users.Put("/:id",    userHandler.Update)
    users.Delete("/:id", userHandler.Delete)

    audit := api.Group("/audit", auth.RequireRole(domain.RoleAdmin))
    audit.Get("/", auditHandler.List)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Fatal(app.Listen(":" + port))
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

---

## Backend — Middleware e Segurança

### JWT Middleware

```go
// internal/auth/middleware.go

type Session struct {
    UserID    string
    Username  string
    Role      domain.Role
    Clearance domain.ClearanceLevel
    ExpiresAt time.Time
}

func RequireAuth(cache *ds.HashMap[Session]) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" {
            return fiber.ErrUnauthorized
        }

        session, ok := cache.Get(token)
        if !ok {
            return fiber.ErrUnauthorized
        }

        if time.Now().After(session.ExpiresAt) {
            cache.Delete(token)
            return fiber.ErrUnauthorized
        }

        // Injeta sessão no contexto para os handlers
        c.Locals("session", session)
        c.Locals("token", token)
        return c.Next()
    }
}

func RequireRole(role domain.Role) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := c.Locals("session").(Session)
        if session.Role != role {
            return fiber.NewError(fiber.StatusForbidden, "Permissão Negada")
        }
        return c.Next()
    }
}

func RequireClearance(level domain.ClearanceLevel) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := c.Locals("session").(Session)
        if session.Clearance < level {
            return fiber.NewError(fiber.StatusForbidden, "Clearance Insuficiente")
        }
        return c.Next()
    }
}

func extractToken(c *fiber.Ctx) string {
    auth := c.Get("Authorization")
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

```sql
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

-- Usuário admin padrão (senha: admin123 — trocar em produção)
INSERT OR IGNORE INTO users (id, username, password_hash, email, role, clearance)
VALUES (
    'usr_admin_000',
    'admin',
    '$2a$12$...',  -- bcrypt de 'admin123'
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
    tags           TEXT NOT NULL DEFAULT '[]',  -- JSON array
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

type Config struct {
    DatabasePath string
    JWTSecret    string
    ServerPort   string
    ServerURL    string // usado pelo client TUI
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

### Fly.io (recomendado)

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

```toml
# fly.toml
app = "classified-vault"
primary_region = "gru"  # São Paulo

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8080"
  ENV  = "production"

[mounts]
  source      = "vault_data"
  destination = "/app/data"

[http_service]
  internal_port = 8080
  force_https   = true

[[vm]]
  memory = "256mb"
  cpus   = 1
```

```bash
# Deploy
fly auth login
fly launch
fly secrets set JWT_SECRET="sua-chave-secreta-aqui"
fly secrets set DATABASE_PATH="/app/data/vault.db"
fly deploy
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
- [ ] Repositórios (user, document, audit)
- [ ] Auth service (bcrypt, JWT/token, login, logout)
- [ ] Middleware RequireAuth e RequireRole
- [ ] Document service com verificação de clearance
- [ ] Handlers e rotas Fiber

### Fase 3 — Backend Complementar
- [ ] User service (CRUD admin)
- [ ] Audit service
- [ ] Integração AVLTree no document service
- [ ] Testes manuais via curl/Insomnia

### Fase 4 — TUI
- [ ] Styles com Lip Gloss (paleta, borders, badges de clearance)
- [ ] API client HTTP em Go
- [ ] Tela de login (Huh form)
- [ ] Dashboard
- [ ] Lista de documentos (com filtro por clearance visual)
- [ ] Visualização de documento
- [ ] Tela de Access Denied
- [ ] Formulário de criação (analyst+)
- [ ] Gestão de usuários (admin)
- [ ] Log de auditoria (admin)

### Fase 5 — Deploy e Polimento
- [ ] Deploy no Fly.io
- [ ] Compilar `.exe` para Windows
- [ ] Testar no Windows Terminal
- [ ] README com instruções de uso
- [ ] Popular banco com dados de demo para apresentação

---

## Dependências

### Backend (`go.mod`)

```
github.com/gofiber/fiber/v2         // HTTP framework
github.com/golang-jwt/jwt/v5        // JWT (opcional, pode usar UUID simples)
github.com/google/uuid              // IDs
golang.org/x/crypto                 // bcrypt
modernc.org/sqlite                  // SQLite puro Go (sem CGO)
```

### Client TUI (`go.mod` — mesmo módulo ou separado)

```
github.com/charmbracelet/bubbletea  // TUI framework
github.com/charmbracelet/lipgloss   // Estilos
github.com/charmbracelet/bubbles    // Componentes prontos (list, textinput, spinner)
github.com/charmbracelet/huh        // Formulários
```

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

> Items to address before or shortly after the core implementation. Split into critical fixes and nice-to-have polish.

### Minor Gaps — Fix Before Handlers

- [ ] **`main.go:695` — silent error ignore**  
  `docs, _ := documentRepo.FindAll()` discards the error. Startup cannot proceed without a properly built AVL index, so use `log.Fatal` on failure.

- [ ] **Rename `auth/jwt.go` → `auth/token.go`**  
  The system uses UUID v4 tokens + HashMap lookup, not actual JWT. Rename the file to avoid confusion during the presentation.

- [ ] **AVL `Remove` never prunes empty nodes**  
  When the last `docID` is removed from a node, the node stays in the tree with an empty `DocIDs` slice forever. It will still be traversed by `QueryUpTo`, returning an empty slice from that node — no correctness bug, but wasted work and leaked tree nodes. Add a `deleteNode` helper that properly removes the node from the tree and rebalances.

- [ ] **Add `RequireAnyRole(roles ...Role)` middleware**  
  Current `RequireRole` checks `==` for a **single** role. Routes like `POST /api/documents` and `PUT /api/documents/:id` need "analyst or admin". Replace or augment with a variadic variant:
  ```go
  func RequireAnyRole(roles ...domain.Role) fiber.Handler {
      return func(c *fiber.Ctx) error {
          session := c.Locals("session").(Session)
          for _, r := range roles {
              if session.Role == r {
                  return c.Next()
              }
          }
          return fiber.NewError(fiber.StatusForbidden, "Permissão Negada")
      }
  }
  ```

- [ ] **Input validation**  
  No validation is specified for any endpoint. Add at minimum:
  - Username format (alphanumeric, 3–32 chars)
  - Password minimum length (≥ 8 chars)
  - Document title/content max length (e.g. 256 / 64 KB)
  - Validate in handlers (light) or services (thorough)

### Nice-to-add

- [ ] **Seed script** (`scripts/seed.go`)  
  Populate the database with demo data for the presentation: one user per role, several documents at each classification level, and pre-baked audit log entries showing access denied attempts. Makes the live demo seamless.

- [ ] **Data structure unit tests** (`internal/ds/*_test.go`)  
  Basic table-driven tests for AVL insert/balance, HashMap get/set/delete/collision, LinkedList append/lastN. Great for demo — open the terminal, run `go test ./internal/ds/ -v`, and show the tree rebalancing live.

- [ ] **Graceful shutdown**  
  Use `signal.NotifyContext` in `main.go` to trap SIGINT/SIGTERM. On shutdown, flush the in-memory linked-list audit buffer to the database so no log entries are lost.
  ```go
  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()
  // ... app setup ...
  go func() {
      <-ctx.Done()
      auditBuffer.FlushTo(auditRepo) // persist remaining entries
      app.Shutdown()
  }()
  ```

- [ ] **Rebuild AVL index from database on startup**  
  Currently the index is populated once at startup from `documentRepo.FindAll()`. If the service layer directly inserts/removes from the index alongside the DB (which the plan shows), this is fine. But if the DB is ever modified externally, the index becomes stale. A `RebuildIndex()` method would be a safety net.