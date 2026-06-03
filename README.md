# Classified Vault

> Secure classified document management system with role-based access control.

## Quick Start

```bash
# 1. Clone & setup
cp .env.example .env

# 2. Start the backend
go run ./cmd/server

# 3. Seed demo data (separate terminal)
go run ./cmd/seed

# 4. Launch the TUI client
go run ./cmd/client
```

Default admin credentials: `admin` / `admin123`

## Architecture

```
┌──────────────┐     HTTP/JSON     ┌──────────────────┐
│  TUI Client  │ ◄──────────────► │   Go Backend      │
│ (Bubble Tea) │                   │  (net/http, stdlib)│
└──────────────┘                   │  ┌──────────────┐ │
                                   │  │   SQLite     │ │
                                   │  │  (WAL mode)  │ │
                                   │  └──────────────┘ │
                                   └──────────────────┘
```

## Security Model — Clearance Levels

| Level | Value | Accessible to |
|---|---|---|
| PUBLIC | 0 | Everyone |
| RESTRICTED | 1 | Viewer, Analyst, Admin |
| CONFIDENTIAL | 2 | Analyst, Admin |
| SECRET | 3 | Admin only |
| TOP SECRET | 4 | Admin only |

## Roles & Permissions

| Role | Max Clearance | Create Docs | Manage Users | View Audit |
|---|---|---|---|---|
| `admin` | TOP SECRET | Yes | Yes | Yes |
| `analyst` | CONFIDENTIAL | Yes | No | No |
| `viewer` | RESTRICTED | No | No | No |
| `intern` | PUBLIC | No | No | No |

## API Endpoints

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | JSON health check (DB, uptime, memory) |
| `POST` | `/auth/login` | Login, returns session token |

### Authenticated (Bearer token required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/me` | Current user info |
| `GET` | `/api/documents` | List accessible documents (clearance-filtered) |
| `GET` | `/api/documents/{id}` | Get document (403 if insufficient clearance) |
| `POST` | `/api/documents` | Create document (analyst+) |
| `PUT` | `/api/documents/{id}` | Update document (analyst+) |
| `DELETE` | `/api/documents/{id}` | Delete document (admin only) |

### Admin only

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/users` | List all users |
| `POST` | `/api/users` | Create user |
| `PUT` | `/api/users/{id}` | Update user |
| `DELETE` | `/api/users/{id}` | Delete user |
| `GET` | `/api/audit` | View audit log |

### Dev only

| Method | Path | Description |
|---|---|---|
| `GET` | `/docs/` | Swagger API docs |
| `GET` | `/debug/pprof/` | Go profiler |

## Makefile Commands

```bash
make dev           # Hot-reload with Air
make build         # Build server
make build-client  # Build TUI client
make run           # Build & run server
make test          # Run unit tests
make smoke         # Run curl smoke tests
make seed          # Populate with demo data
make build-exe     # Cross-compile server for Windows
make build-release # All binaries, all platforms
make clean         # Remove binaries and DB
```

## Deploy

### Render

Push to GitHub, connect repo in Render dashboard → "Blueprints". It reads `render.yaml` — builds via Docker, mounts a 1GB disk for SQLite, and sets up HTTPS automatically.

```bash
# Manual deploy via Docker
docker build -t classified-vault .
docker run -p 8080:8080 -v $(pwd)/data:/data classified-vault
```

Secrets to configure in dashboard:
- `JWT_SECRET` (required for session security)
- `ADMIN_PASSWORD` (default: admin123)

## Project Structure

```
├── cmd/
│   ├── server/        # Backend entrypoint
│   ├── client/        # TUI entrypoint
│   └── seed/          # Database seeder
├── internal/
│   ├── domain/        # Entities (User, Document, AuditLog)
│   ├── ds/            # Hand-written data structures
│   │   ├── avl_tree.go
│   │   ├── hash_map.go
│   │   └── linked_list.go
│   ├── auth/          # Token + bcrypt
│   ├── middleware/     # HTTP middleware (auth, cors, logger)
│   ├── repository/    # SQLite CRUD
│   ├── service/       # Business logic
│   ├── handler/       # HTTP handlers
│   ├── validate/      # Input validation
│   └── apperr/        # Structured errors
├── tui/
│   ├── app.go         # Bubble Tea root model
│   ├── client/        # HTTP API client
│   ├── screens/       # UI screens
│   └── styles/        # Lip Gloss styles
├── migrations/        # SQL schema files
├── scripts/           # Smoke test script
├── Dockerfile
├── render.yaml
└── Makefile
```

## Data Structures (Hand-rolled)

| Structure | Use Case | Complexity |
|---|---|---|
| AVL Tree | Document index by clearance level | Insert: O(log n), Query: O(log n + k) |
| HashMap | Session token cache | Get/Set: O(1) average |
| Doubly Linked List | Audit log buffer | Append: O(1), LastN: O(n) |

## Tech Stack

- **Backend**: Go 1.22+ `net/http` (stdlib routing with path params)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **TUI**: Bubble Tea + Lip Gloss + Bubbles
- **Logging**: stdlib `log/slog` (structured JSON in production)
- **Docs**: Swagger via `swaggo/swag`
- **Deploy**: Docker, Render
