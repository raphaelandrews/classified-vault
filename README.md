# Pelican Town Archives

> Secure scroll management system for Pelican Town public services — department-scoped, tiered access control with a Stardew Valley theme.

## Quick Start

```bash
# 1. Seed the archives with villagers and scrolls
go run ./cmd/seed

# 2. Start the backend
go run ./cmd/server

# 3. Launch the TUI client
go run ./cmd/client
```

Default mayor credentials: `lewis` / `mayor123`

## Architecture

```
┌──────────────┐     HTTP/JSON     ┌──────────────────────────────────────┐
│  TUI Client  │ ◄──────────────► │          Go Backend                   │
│ (Bubble Tea) │                   │  (net/http, stdlib)                   │
└──────────────┘                   │  ┌─────────────────────────────────┐ │
                                   │  │  SQLite (WAL mode + FTS5)       │ │
                                   │  │  AES-256-GCM encrypted content  │ │
                                   │  └─────────────────────────────────┘ │
                                   └──────────────────────────────────────┘
```

## Access Control Model — Departments & Tiers

Every scroll belongs to a **department**. Every villager works for a **department**. Access is decided by 4 rules:

| # | Rule | Description |
|---|---|---|
| 1 | **Town Notice** | Tier 0 scrolls are visible to everyone, regardless of department |
| 2 | **Department Scope** | Same department AND villager tier ≥ scroll tier → access granted |
| 3 | **Mayor Oversight** | Mayor's Office with tier ≥ 4 can see ALL scrolls across all departments |
| 4 | **Arcane Bypass** | Wizard's Tower can see any scroll tagged `arcane` from any department |

### Access Tiers (6 levels)

| Tier | Name | Badge | Description |
|---|---|---|---|
| 0 | **TOWN NOTICE** | Gray | Town bulletin, festival announcements — visible to ALL |
| 1 | **GUILD SEALED** | Green | Community center, carpenter's shop, pier & docks |
| 2 | **COUNCIL SEALED** | Yellow | Adventurer's Guild, Harvey's Clinic |
| 3 | **VAULT SEALED** | Orange | Directors, Joja Corp, Mayor's Office budget |
| 4 | **ARCANE SEALED** | Purple | Wizard's Tower, Guildmaster, Qi's paranormal logs |
| 5 | **JUNIMO SCRIPT** | Red | Mayor, Archmage — highest secrecy |

### Departments (12 departments)

| Department | Special Ability |
|---|---|
| **Mayor's Office** | Tier 4+ sees ALL records (oversight override) |
| **Wizard's Tower** | Sees `arcane`-tagged records across departments |
| **Mr. Qi's Office** | X-Files investigations, secret notes, paranormal tracking |
| **Adventurer's Guild** | Mine safety, monster reports, expeditions |
| **Harvey's Clinic** | Medical records, health services |
| **Joja Corp** | Commercial records, warehouse logistics |
| **Community Center** | Bundles, festivals, forest spirit logs |
| **Carpenter's Shop** | Building permits, infrastructure, shortcuts |
| **Pier & Docks** | Legendary fish catalog, fishing competitions |
| **Museum** | Artifacts, lost books, public archives |
| **Bulletin Board** | Public notices, festival announcements |
| **Roving Trader** | Caravan merchant, unaffiliated scrolls |

### Roles (per department)

Each department has 6 roles — 3 universal + 3 department-specific:

| Universal Roles | Clearance | Capabilities |
|---|---|---|
| **Director** | VAULT SEALED (3) | Create, edit, archive scrolls. Manage department members |
| **Member** | COUNCIL SEALED (2) | Read department scrolls at tier ≤2, create drafts |
| **Visitor** | TOWN NOTICE (0) | Read public scrolls only |

Departments have unique themed roles: Mayor, Archmage, Doctor, Guildmaster, Curator, Agent, Harbormaster, etc.

## Workflow Engine

Scrolls follow a state machine lifecycle with role-gated transitions:

```
draft → review → frozen → archived → public
  ↑        ↑         ↑          ↑         ↑
  └────────┴─────────┴──────────┴─────────┘  (Mayor revert)
```

| Transition | Required Role |
|---|---|
| draft → review | Director or Mayor |
| review → frozen | Mayor only |
| frozen → archived | Director or Mayor |
| archived → public | Mayor only |
| any → draft (revert) | Mayor only |

**Frozen** scrolls have their SHA-256 content hash computed and stored. On every view, the hash is verified — any mismatch shows `⚠ TAMPERING DETECTED`.

## Scroll Freezing & Integrity

- SHA-256 hash of title + content + tier + department + tags
- Frozen status locks content; hash stored in `content_hash`
- Every view recomputes and compares hash
- Direct DB modification is detected in real-time

## Vault Encryption (AES-256-GCM)

- All scroll content encrypted at rest in SQLite
- Master key derived via PBKDF2 from `VAULT_KEY` env var + static salt
- Per-document random nonce (IV); format: `base64(nonce + ciphertext + tag)`
- Transparent encrypt on write, decrypt on read in service layer

## Security Features

- **Account Lockout**: 5 failed logins from same IP → 15-minute lockout with countdown
- **Password Policy**: 8+ chars, letters + numbers/symbols, blocked common passwords
- **Session Warning**: "Session expires in N minutes" banner at TTL-5min, auto-refresh on activity
- **Password Change**: In-TUI password change with current password verification
- **Login Rate Limiting**: 5 req/min per IP on `/auth/login`, returns 429 with `Retry-After`

## Data Structures (Custom, Hand-Rolled)

| Structure | Use Case | Complexity |
|---|---|---|
| **AVL Tree** | Scroll index by access tier | Insert: O(log n), Query: O(log n + k) |
| **HashMap** | Session token cache | Get/Set: O(1) average |
| **Doubly Linked List** | Town ledger buffer | Append: O(1), LastN: O(n) |
| **LRU Cache** | Recently viewed scrolls (per-user, TTL eviction) | Get/Put: O(1), Eviction: O(1) |
| **Trie** | Search autocomplete from scroll titles | Search: O(k) for prefix length k |
| **Max Heap** | Featured scrolls by importance score | ExtractMax: O(log n), Insert: O(log n) |

## TUI Features

- **6 themes**: Gruvbox Dark, Catppuccin Mocha, Nord, Tokyo Night, Stardew Warm, Vault Classic
- **Breadcrumb navigation**: `Dashboard > Scrolls > Scroll Title` trail
- **Markdown rendering**: Scroll content rendered via glamour
- **ASCII bar charts**: Scroll distribution by tier on dashboard
- **FTS5 content search**: Full-text search across titles and content (Ctrl+F)
- **Server-side pagination**: Catalog browsing with LIMIT/OFFSET
- **Export to file**: Save scrolls as `.md` to `./exports/` directory
- **Stats dashboard**: Tier counts, department distribution, most active scribe, monthly totals
- **Recently viewed**: Per-user LRU cache tracks last viewed scrolls

## API Endpoints

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | JSON health check (DB, uptime, memory) |
| `POST` | `/auth/login` | Sign in, returns session token (rate limited: 5/min) |
| `POST` | `/auth/logout` | Sign out |
| `POST` | `/auth/register` | Register new villager |

### Authenticated (Bearer token required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/me` | Current villager info |
| `PUT` | `/api/me/refresh` | Extend session TTL |
| `PUT` | `/api/me/password` | Change password (current + new + confirm) |
| `GET` | `/api/me/recent` | Recently viewed scrolls (LRU cache) |
| `GET` | `/api/documents` | List accessible scrolls (tier + department filtered) |
| `GET` | `/api/documents/search?q=...` | FTS5 full-text search across titles + content |
| `GET` | `/api/documents/autocomplete?q=...` | Trie-based prefix search on titles |
| `GET` | `/api/documents/featured` | Top 5 featured scrolls by importance score (max heap) |
| `GET` | `/api/documents/{id}` | Get scroll (403 if sealed) |
| `GET` | `/api/documents/{id}/export` | Export scroll as Markdown file |
| `GET` | `/api/catalog?limit=&offset=` | Paginated metadata (titles, tiers, departments) |
| `GET` | `/api/stats` | Document statistics (tier counts, dept counts, trends) |
| `POST` | `/api/documents` | Scribe new scroll (Director or Mayor) |
| `PUT` | `/api/documents/{id}` | Amend scroll (Director or Mayor) |
| `PUT` | `/api/documents/{id}/transition` | Advance workflow state |
| `DELETE` | `/api/documents/{id}` | Destroy scroll (Mayor only) |

### Mayor only

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/users` | List all villagers |
| `POST` | `/api/users` | Register new villager |
| `PUT` | `/api/users/{id}` | Update villager |
| `DELETE` | `/api/users/{id}` | Dismiss villager |
| `GET` | `/api/audit?limit=&offset=` | Paginated town ledger |

### Dev only

| Method | Path | Description |
|---|---|---|
| `GET` | `/docs/` | Swagger API docs |
| `GET` | `/debug/pprof/` | Go profiler |

## Project Structure

```
├── cmd/
│   ├── client/main.go          # TUI client entrypoint
│   ├── seed/main.go            # Database seeder (14 villagers, 46+ scrolls)
│   └── server/
│       ├── main.go             # Server entrypoint + route registry
│       ├── middleware.go       # Middleware chain + logger setup
│       └── health.go           # Health check endpoint
├── config/config.go            # Environment configuration
├── internal/
│   ├── apperr/apperr.go        # Typed application errors
│   ├── auth/                   # Token generation, bcrypt, session types
│   ├── crypto/vault.go         # AES-256-GCM encryption/decryption
│   ├── domain/                 # Models: User, Document, Clearance, Audit, Workflow
│   ├── ds/                     # Custom data structures
│   │   ├── avl_tree.go         # Thread-safe AVL tree for clearance index
│   │   ├── hash_map.go         # Generic chaining hash map for sessions
│   │   ├── linked_list.go      # Doubly-linked list for audit buffer
│   │   ├── lru_cache.go        # LRU cache with TTL for recently viewed
│   │   ├── trie.go             # Prefix trie for search autocomplete
│   │   └── heap.go             # Max heap for featured scroll scoring
│   ├── handler/                # HTTP handlers (auth, docs, users, audit, stats)
│   ├── middleware/              # Auth, CORS, logging, rate limiting, recovery
│   ├── repository/             # SQLite data access layer
│   ├── service/                # Business logic, access control, encryption
│   └── validate/               # Input validation (password, username, docs)
├── migrations/                 # SQLite schema migrations (001-009)
├── tui/
│   ├── app.go                  # Main Bubble Tea model + screen router
│   ├── client/                 # HTTP API client for TUI
│   └── screens/                # UI screens (login, dashboard, docs, users, etc.)
│   ├── styles/styles.go        # Lipgloss styles + clearance/department badges
│   └── themes/themes.go        # 6 color themes
├── Dockerfile
├── Makefile
└── render.yaml                 # Render.com deployment config
```

## Tech Stack

- **Backend**: Go 1.22+ `net/http` (stdlib routing with path params)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO, WAL mode, FTS5)
- **Encryption**: AES-256-GCM via `crypto/aes` + `crypto/cipher`, PBKDF2 key derivation
- **TUI**: Bubble Tea + Lip Gloss + Bubbles + Glamour Markdown
- **Logging**: stdlib `log/slog` (structured JSON in production)
- **Docs**: Swagger via `swaggo/swag`
- **Deploy**: Docker, Render

## Deploy

### Docker

```bash
docker build -t pelican-town-archives .
docker run -p 8080:8080 -v $(pwd)/data:/data pelican-town-archives
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VAULT_KEY` | `pelican-town-master-key` | AES-256 master encryption key |
| `INITIAL_MAYOR_PASSWORD` | `mayor123` | Default mayor password on first run |
| `SESSION_TTL` | `8h` | Session token expiry duration |
| `PORT` | `8080` | HTTP server port |
| `DATABASE_PATH` | `./vault.db` | SQLite database file path |
