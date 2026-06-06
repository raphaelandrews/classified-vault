# Classified Vault — Features

> **Pelican Town Archives** is a secure, department-scoped scroll management system built with Go. It combines a terminal-based UI (Bubble Tea) with a layered backend architecture, custom data structures, encryption at rest, a workflow engine, and cybersecurity hardening — all wrapped in a Stardew Valley theme.

---

## Tiered Access Control

Scrolls and villagers are assigned to departments with 6 clearance tiers. Access follows a 4-rule model:

1. **Public by default** — Tier 0 scrolls are visible to everyone
2. **Department scope** — Same department + sufficient tier grants access
3. **Mayor oversight** — Mayor's Office at Arcane tier+ sees all scrolls
4. **Arcane bypass** — Wizard's Tower sees `arcane`-tagged scrolls across departments

12 departments, each with 6 roles (3 universal + 3 themed). Role uniqueness enforced at the database level.

## Document Workflow Engine

Scrolls move through a state machine: `draft → review → frozen → archived → public`

- Role-gated transitions (Mayor required for freezing and public release)
- Document integrity verified via SHA-256 hash computed at freeze time
- Tampering detection on every view — shows `⚠ TAMPERING DETECTED` banner on hash mismatch

## Vault Encryption (AES-256-GCM)

- All scroll content encrypted at rest in SQLite
- PBKDF2 key derivation from `VAULT_KEY` environment variable
- Per-document random nonce; transparent encrypt/decrypt in service layer
- Supports demo: show raw encrypted DB vs decrypted in-app

## Custom Data Structures (Hand-Rolled)

Six production-grade data structures built from scratch, all thread-safe:

| Structure | Use Case | Operations |
|---|---|---|
| **AVL Tree** | Clearance-based document index | Insert O(log n), QueryUpTo O(log n + k) |
| **HashMap** | Session token cache (chaining) | Get/Set O(1) avg |
| **Doubly Linked List** | Audit log buffer | Append O(1), LastN O(n) |
| **LRU Cache** | Recently viewed scrolls per user | Get/Put O(1), TTL-based eviction |
| **Trie** | Search autocomplete from titles | Prefix search O(k) |
| **Max Heap** | Featured scrolls by score | Insert O(log n), ExtractMax O(log n) |

## Search & Discovery

- **Trie Autocomplete** — Type in the search bar, see matching scroll titles instantly via prefix tree
- **FTS5 Full-Text Search** — SQLite FTS5 indexes title + content; ranked results via `bm25()`. Accessible with Ctrl+F in the TUI
- **Featured Scrolls** — Max heap ranks scrolls by `tier × 10 + recency_days`; top entries displayed on demand

## Statistics Dashboard

Real-time town statistics with ASCII bar charts:
- Scroll distribution per clearance tier (visual bars)
- Top 5 departments by scroll count
- Most active scribe with scroll count
- Scrolls created this month, total counts
- Refreshable with `[r]` key

## Security Features

| Feature | Description |
|---|---|
| **Account Lockout** | 5 failed logins from same IP → 15-minute lockout with visual countdown |
| **Password Policy** | 8+ chars, requires letters + digits/symbols, blocks top 100 common passwords |
| **Session Warning** | Banner at TTL-5min: "Session expires in N minutes". Auto-refresh on activity |
| **Password Change** | In-TUI form: current password, new password, confirmation. Audit logged |
| **Login Rate Limiting** | 5 req/min per IP on `/auth/login`, returns 429 with `Retry-After` header |
| **Content Integrity** | SHA-256 hash computed at freeze, verified on every view |
| **RBAC Middleware** | `RequireAuth`, `RequireRole`, `RequireClearance`, `RequireAnyRole` chain |

## TUI Experience

- **6 color themes**: Gruvbox Dark, Catppuccin Mocha, Nord, Tokyo Night, Stardew Warm, Vault Classic
- **Breadcrumb trail**: `Dashboard > Scrolls > Scroll Title` — always visible navigation path
- **Markdown rendering**: Scroll content styled via glamour with terminal-friendly formatting
- **Export to file**: Save scrolls as `.md` to `./exports/` directory with `[x]` key
- **Server-side pagination**: Efficient `LIMIT/OFFSET` catalog browsing
- **Integrated signup**: Department picker with themed role names, password strength hints

## API Design

- Go 1.22+ `net/http` with path parameters (`{id}`)
- JSON request/response throughout
- Bearer token authentication via UUID session tokens
- Rate limiting on general API (100 req/min) and login (5 req/min)
- `X-Session-Expires` header for sliding session management
- `X-Total-Count` header for paginated endpoints

## Audit Trail

Every action is logged to the Town Ledger:
- Sign in/out, failed logins, account lockouts
- Scroll read, create, update, delete, status transitions
- Access denied events with tier/department details
- Villager registration, updates, dismissal
- Password changes

Stored in `audit_logs` table, buffered via doubly-linked list, viewable by the Mayor in the TUI.

## Deployment

- Single binary + SQLite database — no external dependencies
- Docker with 1GB persistent volume mount
- Render Blueprint support via `render.yaml`
- Graceful shutdown flushes audit buffer to DB
