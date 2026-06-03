# Pelican Town Archives

> Secure scroll management system for Pelican Town public services — faction-scoped, tiered access control with a Stardew Valley theme.

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
┌──────────────┐     HTTP/JSON     ┌──────────────────┐
│  TUI Client  │ ◄──────────────► │   Go Backend      │
│ (Bubble Tea) │                   │  (net/http, stdlib)│
└──────────────┘                   │  ┌──────────────┐ │
                                   │  │   SQLite     │ │
                                   │  │  (WAL mode)  │ │
                                   │  └──────────────┘ │
                                   └──────────────────┘
```

## Access Control Model — Factions & Tiers

Every scroll belongs to a **faction**. Every villager works for a **faction**. Access is decided by 4 rules:

| # | Rule | Description |
|---|---|---|
| 1 | **Public Notice** | Tier 0 scrolls are visible to everyone, regardless of faction |
| 2 | **Faction Scope** | Same faction AND villager tier ≥ scroll tier → access granted |
| 3 | **Mayor Oversight** | Mayor's Office with tier ≥ 4 can see ALL scrolls across all factions |
| 4 | **Arcane Bypass** | Wizard's Tower can see any scroll tagged `arcane` from any faction |

### Access Tiers (6 levels)

| Tier | Name | Badge | Description |
|---|---|---|---|
| 0 | **Public Notice** | Gray | Town bulletin, festival announcements — visible to ALL |
| 1 | **Council Eyes Only** | Green | Community center, carpenter's shop, pier & docks |
| 2 | **Guild Business** | Yellow | Adventurer's Guild, Harvey's Clinic |
| 3 | **Corporate Access** | Orange | Joja Corp, Mayor's Office budget |
| 4 | **Arcane Knowledge** | Purple | Wizard's Tower, Mr. Qi's paranormal logs |
| 5 | **Junimo Script** | Red | Mayor's inner circle, Qi's X-Files — highest secrecy |

### Factions (11 departments)

| Faction | Special Ability |
|---|---|
| **Mayor's Office** | Tier 4+ sees ALL records (oversight override) |
| **Wizard's Tower** | Sees `arcane`-tagged records cross-faction |
| **Mr. Qi's Office** | X-Files investigations, secret notes, paranormal tracking |
| **Adventurer's Guild** | Mine safety, monster reports, expeditions |
| **Harvey's Clinic** | Medical records, health services |
| **Joja Corp** | Commercial records, warehouse logistics |
| **Community Center** | Bundles, festivals, forest spirit logs |
| **Carpenter's Shop** | Building permits, infrastructure, shortcuts |
| **Pier & Docks** | Legendary fish catalog, fishing competitions |
| **Museum / Library** | Artifacts, lost books, public archives |
| **Bulletin Board** | Public notices, festival announcements |

### System Roles (4 permission levels)

| Role | Create Scrolls | Manage Villagers | View Ledger | Default Tier |
|---|---|---|---|---|
| `mayor` | Yes (any faction) | Yes | Yes | 5 |
| `keeper` | Yes (own faction) | No | No | 4 |
| `villager` | No | No | No | 1 |
| `associate` | No | No | No | 0 |

## Seed Data — 14 Villagers & 46 Scrolls

### Villagers

| Username | Password | Faction | Tier | Role |
|---|---|---|---|---|
| `lewis` | `mayor123` | Mayor's Office | 5 | mayor |
| `marnie` | `deputy123` | Mayor's Office | 4 | mayor |
| `qi` | `qichallenge` | Mr. Qi's Office | 5 | keeper |
| `rasmodius` | `wizard123` | Wizard's Tower | 4 | keeper |
| `morris` | `joja123` | Joja Corp | 3 | keeper |
| `marlon` | `guild123` | Adventurer's Guild | 2 | keeper |
| `gil` | `guild123` | Adventurer's Guild | 1 | villager |
| `harvey` | `clinic123` | Harvey's Clinic | 2 | keeper |
| `junimo` | `bundle123` | Community Center | 1 | villager |
| `robin` | `build123` | Carpenter's Shop | 1 | villager |
| `willy` | `fishmaster` | Pier & Docks | 1 | villager |
| `gunther` | `museum123` | Museum | 0 | associate |
| `gus` | `saloon123` | Bulletin Board | 0 | associate |
| `krobus` | `voidshadow` | Pier & Docks | 0 | associate |

### Scroll Folders

- **War of the Worlds Broadcast** — Emergency broadcast of a fictional Gotoro invasion that caused town-wide panic, followed by a retraction. 47 villagers hid in the Community Center basement.
- **X-Files: Qi's Investigations** (5 scrolls) — Strange Capsule follow-up, Prismatic Entity observations, Grandpa's temporal anomaly, Wizard's identity investigation, weekly paranormal activity log
- **Secret Notes** (3 scrolls) — Solid Gold Lewis statue, Maple Syrup Bear, Qi Challenge
- **The Legendary Fish** (3 scrolls) — Legend (Mountain Lake), Glacierfish, Crimsonfish

### Easter Eggs

- Void Chicken origin study (Krobus shadow dimension testimony)
- Statue of Endless Fortune produced a gold coin from **year 2319**
- Mermaid's song at the Night Market is actually coded fishing coordinates
- Iridium Seam legal dispute between Joja Corp and Adventurer's Guild
- Grange Display contest: Morris tried entering Joja Cola cans, pyramid collapsed
- Ice Fishing competition: Pam fished while drinking Pale Ale, Lewis caught a driftwood splinter

### Access Scenarios (for presentation)

| Villager | Tries to read... | Result |
|---|---|---|
| `lewis` (Mayor, t5) | Any scroll in any faction | ✅ Mayor override — sees all 46 |
| `marnie` (Mayor, t4) | Any scroll in any faction | ✅ Mayor override — sees all 46 |
| `marlon` (Guild, t2) | Mine Monster Report (Guild, t2) | ✅ Same faction, sufficient tier |
| `marlon` (Guild, t2) | Medical Record (Clinic, t2) | ❌ Wrong faction |
| `marlon` (Guild, t2) | Skull Cavern Expedition (Guild, t2) | ✅ Same faction, sufficient tier |
| `gil` (Guild, t1) | Skull Cavern Expedition (Guild, t2) | ❌ Insufficient tier within faction |
| `rasmodius` (Wizard, t4) | Shadow Brute Census (Guild, t1, tagged `arcane`) | ✅ Arcane bypass |
| `rasmodius` (Wizard, t4) | Joja Expansion Plan (Joja, t3) | ❌ Wrong faction, no arcane tag |
| `qi` (Qi, t5) | X-Files scrolls (Qi's Office, t4–5) | ✅ Same faction |
| `qi` (Qi, t5) | Town Budget (Mayor's Office, t3) | ❌ Wrong faction — Qi's Office cannot see Mayor's Office scrolls |
| `willy` (Pier, t1) | Legendary Fish catalog (Pier, t1) | ✅ Same faction |
| `willy` (Pier, t1) | Void Chicken Study (Clinic, t2) | ❌ Wrong faction |
| `krobus` (Pier, t0) | Legendary Fish catalog (Pier, t1) | ❌ Same faction but insufficient tier |
| `krobus` (Pier, t0) | Ice Fishing Results (Pier, t0) | ✅ Public tier 0 |
| `gunther` (Museum, t0) | Lost Books Recovery (Museum, t0) | ✅ Public tier 0 |
| `gunther` (Museum, t0) | Bundle Progress (Community, t1) | ❌ Wrong faction, tier > 0 |

## API Endpoints

### Public

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | JSON health check (DB, uptime, memory) |
| `POST` | `/auth/login` | Sign in, returns session token |

### Authenticated (Bearer token required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/me` | Current villager info |
| `GET` | `/api/documents` | List accessible scrolls (faction + tier filtered) |
| `GET` | `/api/documents/{id}` | Get scroll (403 if sealed) |
| `POST` | `/api/documents` | Scribe new scroll (keeper+) |
| `PUT` | `/api/documents/{id}` | Amend scroll (keeper+) |
| `DELETE` | `/api/documents/{id}` | Destroy scroll (mayor only) |
| `GET` | `/api/catalog` | All metadata (titles, tiers, factions — no content) |

### Mayor only

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/users` | List all villagers |
| `POST` | `/api/users` | Register new villager |
| `PUT` | `/api/users/{id}` | Update villager |
| `DELETE` | `/api/users/{id}` | Dismiss villager |
| `GET` | `/api/audit` | View town ledger |

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
docker build -t pelican-town-archives .
docker run -p 8080:8080 -v $(pwd)/data:/data pelican-town-archives
```

Secrets to configure in dashboard:
- `JWT_SECRET` (required for session security)
- `INITIAL_MAYOR_PASSWORD` (default: `mayor123`)

## Project Structure

```
├── cmd/
│   ├── server/        # Backend entrypoint
│   ├── client/        # TUI entrypoint
│   └── seed/          # Database seeder (14 villagers + 46 scrolls)
├── internal/
│   ├── domain/        # Entities (Villager, Scroll, AuditLog)
│   │   ├── user.go
│   │   ├── document.go
│   │   ├── audit.go
│   │   └── clearance.go   # Tiers, Roles, Factions
│   ├── ds/            # Hand-written data structures
│   │   ├── avl_tree.go
│   │   ├── hash_map.go
│   │   └── linked_list.go
│   ├── auth/          # Token + bcrypt + session
│   ├── middleware/     # HTTP middleware (auth, cors, logger)
│   ├── repository/    # SQLite CRUD
│   ├── service/       # Business logic (faction-scoped access control)
│   ├── handler/       # HTTP handlers
│   ├── validate/      # Input validation
│   └── apperr/        # Structured errors
├── tui/
│   ├── app.go         # Bubble Tea root model
│   ├── client/        # HTTP API client
│   ├── screens/       # UI screens (Gruvbox Material Dark Hard palette)
│   └── styles/        # Lip Gloss styles + faction/tier badges
├── migrations/        # SQL schema files
├── scripts/           # Smoke test script
├── Dockerfile
├── render.yaml
└── Makefile
```

## Data Structures (Hand-rolled)

| Structure | Use Case | Complexity |
|---|---|---|
| AVL Tree | Scroll index by access tier | Insert: O(log n), Query: O(log n + k) |
| HashMap | Session token cache | Get/Set: O(1) average |
| Doubly Linked List | Town ledger buffer | Append: O(1), LastN: O(n) |

## Tech Stack

- **Backend**: Go 1.22+ `net/http` (stdlib routing with path params)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **TUI**: Bubble Tea + Lip Gloss + Bubbles — Gruvbox Material Dark Hard palette
- **Logging**: stdlib `log/slog` (structured JSON in production)
- **Docs**: Swagger via `swaggo/swag`
- **Deploy**: Docker, Render
