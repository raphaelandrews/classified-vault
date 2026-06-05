# New Features Plan — Classified Vault

## Tier & Role Naming Overhaul

Current names have mixed themes (corporate + fantasy) and inconsistent patterns. Suggested renames:

### Clearance Tiers

| Was | Better | Rationale |
|-----|--------|-----------|
| PUBLIC NOTICE | **TOWN NOTICE** | More village-themed, shorter by 3 chars |
| COUNCIL EYES ONLY | **GUILD SEALED** | "SEALED" pattern for restricted tiers, shorter |
| GUILD BUSINESS | **COUNCIL SEALED** | Council at middle tier, consistent pattern |
| CORPORATE ACCESS | **VAULT SEALED** | Fantasy-appropriate, no corporate break |
| ARCANE KNOWLEDGE | **ARCANE SEALED** | Consistent "SEALED" suffix |
| JUNIMO SCRIPT | **JUNIMO SCRIPT** | Already perfect, unique top tier |

Pattern: 4 tiers end in "SEALED" — escalation is immediately readable. `TOWN NOTICE` and `JUNIMO SCRIPT` bookend the system as open/closed extremes.

### Roles (Department-Specific)

Each department defines its own role hierarchy. Three universal roles exist in every department. On top, each department adds 3 uniquely themed roles, for 6 total per department. Every role maps to a clearance tier.

#### Universal Roles (exist in every department)

| Role | Clearance | Capabilities |
|------|-----------|-------------|
| **Director** | VAULT SEALED (3) | Create, edit, archive scrolls. Manage department members |
| **Member** | COUNCIL SEALED (2) | Read department scrolls at tier ≤2, create draft scrolls |
| **Visitor** | TOWN NOTICE (0) | Read public scrolls only |

#### Department-Specific Roles

| Department | Role | Clearance |
|------------|------|-----------|
| **Mayor's Office** | Mayor | JUNIMO SCRIPT (5) |
| | Secretary | GUILD SEALED (1) |
| | Clerk | TOWN NOTICE (0) |
| **Wizard's Tower** | Archmage | JUNIMO SCRIPT (5) |
| | Enchanter | ARCANE SEALED (4) |
| | Apprentice | GUILD SEALED (1) |
| **Harvey's Clinic** | Doctor | VAULT SEALED (3) |
| | Nurse | COUNCIL SEALED (2) |
| | Medic | GUILD SEALED (1) |
| **Adventurer's Guild** | Guildmaster | ARCANE SEALED (4) |
| | Adventurer | COUNCIL SEALED (2) |
| | Scout | GUILD SEALED (1) |
| **Museum** | Curator | VAULT SEALED (3) |
| | Archivist | COUNCIL SEALED (2) |
| | Docent | GUILD SEALED (1) |
| **Joja Corp** | Manager | VAULT SEALED (3) |
| | Supervisor | COUNCIL SEALED (2) |
| | Clerk | TOWN NOTICE (0) |
| **Carpenter's Shop** | Master Builder | VAULT SEALED (3) |
| | Carpenter | COUNCIL SEALED (2) |
| | Apprentice | GUILD SEALED (1) |
| **Community Center** | Coordinator | VAULT SEALED (3) |
| | Organizer | COUNCIL SEALED (2) |
| | Volunteer | GUILD SEALED (1) |
| **Bulletin Board** | Editor | VAULT SEALED (3) |
| | Scribe | COUNCIL SEALED (2) |
| | Crier | GUILD SEALED (1) |
| **Mr. Qi's Office** | Agent | ARCANE SEALED (4) |
| | Operative | COUNCIL SEALED (2) |
| | Courier | GUILD SEALED (1) |
| **Pier & Docks** | Harbormaster | VAULT SEALED (3) |
| | Sailor | COUNCIL SEALED (2) |
| | Fisher | GUILD SEALED (1) |
| **Roving Trader** | Merchant | COUNCIL SEALED (2) |
| | Trader | GUILD SEALED (1) |
| | Hawker | TOWN NOTICE (0) |

All departments include the 3 universal roles on top (6 roles total per department).

#### Role Uniqueness Rules

- **Mayor**: Only one Mayor may exist globally. `UNIQUE` constraint on `role = 'Mayor'`
- **Director per department**: Only one Director per department. `UNIQUE` on `(department, role) WHERE role = 'Director'`
- **Lead per department**: Each department's unique top role (Doctor, Archmage, Guildmaster, etc.) is exclusive — only one person can hold that title per department. `UNIQUE` on `(department, role)` for non-universal roles
- Application-level validation in `UserService.Create()` and `UserService.Update()` with clear error messages
- Database-level enforcement via partial unique indexes (migration `007_unique_role_constraints.sql`)

Role exclusivity:
- **Medic** only exists in Harvey's Clinic
- **Enchanter / Archmage** only exist in Wizard's Tower
- **Fisher** only exists in Pier & Docks
- **Doctor / Nurse** only exist in Harvey's Clinic
- **Operative / Courier** only exist in Mr. Qi's Office
- **Mayor** only exists in Mayor's Office
- All other unique roles are department-scoped

#### Director vs Thematic Lead

When a department's top unique role (e.g. Doctor) has the same clearance as Director (tier 3):
- **Director** = administrative authority (manages department roster)
- **Thematic lead** = functional authority (creates/edits scrolls at tier 3)
- A single user can hold both if needed, but the distinction is intentional: a Doctor manages scrolls, the Director manages people

#### Implementation Notes

- Store roles as `(department, role_name, clearance)` triples in a `roles` table
- Access check becomes: `canAccess(user_dept, user_clearance, scroll_dept, scroll_tier)` — role name is cosmetic
- Signup form: pick department → see available roles for that department
- Admin panel: users shown as `Department / Role` badge
- Migration: `roles` table replacing the flat `Role` enum
- Backward compat: existing `mayor` → `Mayor`, `keeper` → `Director`, `villager` → `Member`, `associate` → `Visitor`

### Department

Add `Roving Trader` department — a proper "unaffiliated" option for the caravan merchant. Total departments: 12.

---

## TIER 1 — Major Demo Features

### 1. Scroll Freezing & Integrity System
- SHA-256 hash of content + metadata stored per scroll
- `frozen` workflow status locks content; hash stored in `ContentHash` field
- On every view: recompute hash, compare → `⚠ TAMPERING DETECTED` banner on mismatch
- Hash computed from: title + content + classification + department + tags JSON
- **DS used**: None (stdlib crypto/sha256)
- **Demo value**: Presenter modifies DB directly, system detects tampering in real-time

### 2. Document Workflow Engine
- Scroll lifecycle: `draft → review → frozen → archived → public`
- Transition permission table:
  | Transition | Required Role |
  |------------|---------------|
  | draft → review | Director or Mayor |
  | review → frozen | Mayor only |
  | frozen → archived | Director or Mayor |
  | archived → public | Mayor only |
  | any → draft | Mayor only (revert) |
- `frozen` = hash computed, content immutable, integrity verifiable
- `archived` = read-only, hidden from default view
- `public` = visible to all villagers regardless of department
- Workflow actions shown as `[w]` key in doc view screen
- Backend: `PUT /api/documents/{id}/transition` endpoint
- **DS used**: State transition map/matrix
- **Demo value**: State machine + RBAC enforcement + role-gated transitions

### 3. Vault Encryption (AES-GCM)
- All scroll content encrypted at rest in SQLite
- Master key derived via PBKDF2 from `VAULT_KEY` env var + random salt
- Per-document random nonce (IV); stored format: `base64(nonce + ciphertext + tag)`
- Transparent encrypt on write, decrypt on read in service layer
- Existing scrolls auto-encrypted on first access
- **DS used**: None (stdlib crypto/aes, crypto/cipher)
- **Demo value**: Show raw encrypted DB entries vs decrypted in-app

---

## TIER 2 — Data Structure Showcase

### 4. LRU Cache — Recently Viewed Scrolls
- Track last 10 viewed scrolls with O(1) access
- **Structure**: `HashMap[string]*Node + DoublyLinkedList` (both already built in `internal/ds/`)
- Cache stores `(docID, Document)` pairs with TTL eviction (1 hour)
- Dashboard shows "Recently Viewed" section with quick-access keys
- **Demo value**: Reuses custom DS, demonstrates O(1) get/put, LRU eviction policy

### 5. Trie — Search Autocomplete
- Prefix tree built from all scroll titles
- Rebuilt incrementally on document list load
- As user types in `/` search bar, dropdown shows matching titles
- Match highlighting on the filtered results
- **DS used**: Trie (new — `internal/ds/trie.go`)
- **Demo value**: Classic data structure application, instant visual feedback

### 6. Priority Queue — Featured Scrolls
- Rank scrolls by importance score: `score = tier × 10 + recency_days`
- Higher score = more prominent display
- Dashboard widget: "★ Featured Scrolls" showing top 3 by score
- **DS used**: Max Heap (new — `internal/ds/heap.go`)
- **Demo value**: Heap-based prioritization, sorting without full sort

---

## TIER 3 — Cybersecurity Enhancements

### 7. Account Lockout
- 5 failed login attempts from same IP → 15-minute lockout
- Per-IP tracking using in-memory `HashMap[ip, *LockoutState]`
- LockoutState: `{attempts, firstAttempt, lockedUntil}`
- Visual countdown timer shown on failed login
- New audit action: `VILLAGER_LOCKED`

### 8. Password Strength Policy
- Reject passwords < 8 characters on signup
- Check against top 100 common passwords (embedded list)
- Enforce mix: at least 1 letter + 1 digit or symbol
- Show visual requirements on signup form

### 9. Session Auto-Logout Warning
- Show toast/banner: "Session expires in 5 minutes" at TTL-5min
- Auto-refresh token on user activity (sliding window)
- Session expiry generates audit log entry

### 10. Password Change
- New TUI screen: "Change Password" accessible from dashboard
- Requires current password + new password + confirmation
- Backend endpoint: `PUT /api/me/password`
- Audit log entry for password changes

---

## TIER 4 — UX & Polish

### 11. Document Statistics Dashboard
- ASCII bar chart of scrolls per clearance tier
- Count per department
- Most active scribe (most scrolls created)
- Scrolls created this month
- Refreshable with `[r]` key

### 12. Document Content Search (FTS5)
- Enable SQLite FTS5 extension
- Full-text index on `title + content`
- Search results show snippet with `**match**` highlighted
- `/` in doc list also searches content (not just metadata)

### 13. Breadcrumb Navigation Trail
- `Dashboard > Scrolls > The Mayor's Decree`
- Styled bar above content on each screen
- Shows current position in navigation hierarchy

### 14. Export Scroll to File
- Save scroll as `.md` (Markdown) file to `./exports/` directory
- Uses glamour markdown rendering for formatting
- Optional password-protect with AES encryption on the file
- Output path shown in status

---

## TIER 5 — Infrastructure

### 15. Repository Pagination
- Add `LIMIT/OFFSET` to `FindAll` / `FindAllMetadata` methods
- Configurable page sizes per endpoint
- Return total count for accurate page calculation
- Avoids loading all records into memory

### 16. Login Rate Limiting
- Stricter per-IP rate limiter for `/auth/login` (5 req/min)
- Separate from general API rate limiter
- Returns 429 with retry-after header

---

## Implementation Plan

### Phase 0 — Naming Overhaul (Day 1) ✅
- [x] Rename clearance tiers in `domain/clearance.go`
- [x] Create `roles` table migration
- [x] Seed default roles per department
- [x] Update `User` domain: replace `Role` enum with `(Department, RoleName)` pair
- [x] Update `DocumentStatus`: rename `StatusSealed` → `StatusFrozen`
- [x] Update all service/handler/TUI references
- [x] Update seed data
- [x] Add role uniqueness constraints (Mayor global, Director per-department)

### Phase 1 — Integrity + Workflow (Days 2–3) ✅
- [x] **Freezing (#1)**: Add `ContentHash` column migration, `ComputeHash()`, `VerifyIntegrity()`, update repository, service, TUI doc view
- [x] **Workflow (#2)**: Add `StatusReview`, `StatusFrozen`, `StatusPublic` to enum, create transition matrix, add `Transition()` to service, add `PUT /api/documents/{id}/transition` endpoint, `[w]` key in TUI doc view
- **Tests**: hash mismatch detection, transition permission gates

### Phase 2 — Encryption (Day 4) ✅
- [x] **Vault Encryption (#3)**: Add `crypto/aes` encrypt/decrypt to service layer, encrypt on Create/Update, decrypt on GetByID/List, migration to encrypt existing content, env var `VAULT_KEY`

### Phase 3 — Data Structures (Days 5–6)
- **LRU Cache (#4)**: Implement `LRUCache` using existing `HashMap` + `LinkedList`, integrate into `GetByID`, add "Recently Viewed" to dashboard TUI
- **Trie (#5)**: Implement `Trie` in `internal/ds/trie.go`, build from scroll titles on load, wire into search textinput for autocomplete dropdown
- **Priority Queue (#6)**: Implement `MaxHeap` in `internal/ds/heap.go`, compute scores for all scrolls, show top 3 on dashboard
- **Tests**: LRU eviction ordering, trie prefix search, heap extract-max ordering

### Phase 4 — Cybersecurity (Days 7–8)
- **Account Lockout (#7)**: Add `LockoutState` tracker in `AuthService`, check on login, return lockout duration, show countdown in TUI, audit log entry
- **Password Policy (#8)**: Add `validatePassword()` with rules, embed top 100 list, show requirements on signup form
- **Session Warning (#9)**: Add session TTL check in middleware, return `X-Session-Expires` header, TUI shows toast at 5min mark, sliding refresh
- **Password Change (#10)**: Add `PUT /api/me/password` endpoint, new TUI screen `ChangePasswordModel`, audit log
- **Tests**: lockout after 5 failures, weak password rejection, session expiry flow

### Phase 5 — UX Polish (Days 9–10)
- **Stats Dashboard (#11)**: Add stats endpoint `GET /api/stats`, ASCII bar chart widget, department counts
- **FTS5 Search (#12)**: Enable FTS5 in SQLite, create index, add content search to catalog query
- **Breadcrumbs (#13)**: Add `breadcrumb` field to TUI app model, render styled bar, update on navigation
- **Export (#14)**: Add `[x]` key on doc view, write `.md` to `./exports/`, optional AES password prompt
- **Tests**: stats computation, FTS5 query results, breadcrumb state transitions

### Phase 6 — Infrastructure (Days 11–12)
- **Pagination (#15)**: Add `LIMIT ? OFFSET ?` to repository queries, `Count()` methods, update handler pagination, update TUI paginator to use server-side paging
- **Rate Limiting (#16)**: Add per-IP rate limiter for `/auth/login`, return 429 + `Retry-After` header, TUI shows "Too many attempts"
- **Final polish**: End-to-end demo script, update README, add screenshots
