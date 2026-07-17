# Handoff — Chirpy Store Seam Implementation

**Date:** 2026-07-17  
**Workspace:** `/Users/pawelbalcerek/etc/projects/bootdev/chirpy`  
**Prior conversation (architecture review):** `096bd71f-a896-4d24-8cbd-46190772c2f8`  
**This conversation (grilling):** `db178851-622d-4c2c-9d83-c3d930ecd3ba`

---

## Context

Chirpy is a Go REST API (Twitter-like micro-posting). An architecture review identified 6 deepening candidates. The user chose **Candidate 3: Introduce a Store seam (interface) for `*database.Queries`** to unlock handler-level testing without a live Postgres database.

A grilling session resolved all design decisions. **No code changes have been made yet.**

---

## Decisions (from grilling)

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Interface location | `handlers/` package | Go consumer-owns-interface idiom |
| 2 | Interface shape | Three domain-grouped interfaces: `ChirpStore` (4 methods), `UserStore` (5 methods), `TokenStore` (3 methods) | Precision at manageable scale |
| 3 | Cross-domain handlers | Multiple fields — e.g. `LoginHandler` gets both `UserStore` + `TokenStore` | Explicit dependency declaration |
| 4 | In-memory fakes | Unexported, in `handlers/*_test.go` files, three separate fake structs | Keep fakes close to tests, no extra packages |
| 5 | Type coupling | Mirror sqlc-generated signatures — use `database.*` param/model types in interface methods | `*database.Queries` satisfies interfaces with zero glue code |
| 6 | First test target | `CreateChirpHandler` | Exercises profanity filter + auth; surfaces known bug |
| 7 | Profanity filter bug | Fix in the same changeset (red → green TDD) | Small fix, proves seam value |
| 8 | Documentation | Create `docs/adr/001-store-seam.md` + initial `CONTEXT.md` | Seed ADR convention, capture "why" |

---

## Interface design (reference)

The three interfaces should contain these methods (signatures mirror sqlc-generated code in `internal/database/`):

### `ChirpStore` (4 methods)
- `CreateChirp(ctx, database.CreateChirpParams) (database.Chirp, error)`
- `GetChirp(ctx, uuid.UUID) (database.Chirp, error)`
- `GetChirps(ctx, uuid.UUID) ([]database.Chirp, error)`
- `DeleteChirp(ctx, uuid.UUID) error`

### `UserStore` (5 methods)
- `CreateUser(ctx, database.CreateUserParams) (database.User, error)`
- `GetUser(ctx, string) (database.User, error)`
- `UpdateUser(ctx, database.UpdateUserParams) (database.User, error)`
- `MakeUserChirpyRed(ctx, uuid.UUID) (database.User, error)`
- `DeleteUsers(ctx) error`

### `TokenStore` (3 methods)
- `CreateRefreshToken(ctx, database.CreateRefreshTokenParams) (database.RefreshToken, error)`
- `GetRefreshToken(ctx, string) (database.RefreshToken, error)`
- `RevokeRefreshToken(ctx, string) error`

---

## Files to create or modify

### New files
- `handlers/store.go` — three interface definitions
- `handlers/chirps_test.go` — `fakeChirpStore` + `CreateChirpHandler` tests (first test)
- `docs/adr/001-store-seam.md` — architectural decision record
- `CONTEXT.md` — initial domain model / project context

### Modified files
- **All handler structs** — change `DbQueries *database.Queries` → appropriate interface field(s):
  - `chirps.go`: 4 structs → `ChirpStore` field (+ `DeleteChirpHandler` already only uses `ChirpStore`)
  - `users.go`: 2 structs → `UserStore` field
  - `login.go`: `LoginHandler` → `UserStore` + `TokenStore` fields
  - `refresh.go`: `RefreshHandler` → `TokenStore` field
  - `revoke.go`: `RevokeHandler` → `TokenStore` field
  - `reset.go`: `ResetHandler` → `UserStore` field
  - `polka.go`: `PolkaWebhookHandler` → `UserStore` field
- **`main.go`** — update struct field names when wiring (value stays `dbQueries`, it satisfies all three interfaces)
- **`handlers/chirps.go` line ~83** — fix profanity filter bug: use `cleanedBody` instead of `request.Body`

---

## Known bugs to address

1. **Profanity filter is dead code** (`handlers/chirps.go:74-83`): `cleanedBodyElements` is built but `request.Body` (unfiltered) is saved. Fix in this changeset.
2. **`LoginHandler` silently drops `CreateRefreshToken` error** (`login.go:73`): out of scope for this changeset but worth a follow-up issue.
3. **`ResetHandler` writes body before status code** (`reset.go:18-19`): out of scope, follow-up issue.

---

## What was NOT done

- No code changes — purely a design grilling session.
- No ADRs or `CONTEXT.md` created yet — to be created during implementation.
- No issues filed for bugs #2 and #3 above.

---

## Suggested skills

| Skill | When to use |
|-------|-------------|
| `/tdd` | **Immediately** — implement the Store seam test-first: write `CreateChirpHandler` test (red, exposing profanity bug), introduce interfaces + fakes, fix bug (green), refactor |
| `/codebase-design` | Reference its vocabulary (module, interface, depth, seam, adapter, leverage, locality) when writing the ADR and during implementation |
| `/domain-modeling` | Create `CONTEXT.md` and `docs/adr/001-store-seam.md` as part of the implementation |
| `/code-review` | After implementation, review changes against the decisions in this handoff |

---

## File map (quick reference)

```
chirpy/
├── main.go                          # App entry, DB init, all 12 route registrations
├── handlers/
│   ├── store.go                     # NEW — ChirpStore, UserStore, TokenStore interfaces
│   ├── chirps.go                    # MODIFY — use ChirpStore, fix profanity bug
│   ├── chirps_test.go              # NEW — fakeChirpStore + CreateChirpHandler tests
│   ├── users.go                     # MODIFY — use UserStore
│   ├── login.go                     # MODIFY — use UserStore + TokenStore
│   ├── refresh.go                   # MODIFY — use TokenStore
│   ├── revoke.go                    # MODIFY — use TokenStore
│   ├── polka.go                     # MODIFY — use UserStore
│   ├── reset.go                     # MODIFY — use UserStore
│   ├── healthz.go                   # unchanged
│   ├── metrics.go                   # unchanged
│   ├── error.go                     # unchanged
│   └── json.go                      # unchanged
├── internal/
│   ├── auth/                        # unchanged
│   └── database/                    # sqlc-generated (DO NOT EDIT)
├── docs/
│   └── adr/
│       └── 001-store-seam.md        # NEW — ADR
├── CONTEXT.md                       # NEW — domain model
└── ...
```
