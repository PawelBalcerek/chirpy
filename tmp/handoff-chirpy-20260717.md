# Handoff — Chirpy Architecture Review

**Date:** 2026-07-17  
**Workspace:** `/Users/pawelbalcerek/etc/projects/bootdev/chirpy`  
**Conversation:** `096bd71f-a896-4d24-8cbd-46190772c2f8`

---

## Context

Chirpy is a Go REST API (Twitter-like micro-posting service) built with:
- `net/http` stdlib router (no framework)
- `sqlc` for code-generating database queries against PostgreSQL
- `internal/auth` for JWT, password hashing, bearer/API-key extraction
- A flat `handlers/` package with **one struct per HTTP endpoint** (10 structs total)

The codebase has **no `CONTEXT.md`**, **no `docs/adr/`** directory, and **no handler-level tests** — only `internal/auth` has tests (bearer, JWT, password).

---

## What was done

Ran the `/improve-codebase-architecture` skill. Produced an HTML architecture review with 6 deepening candidates:

### Architecture review report
**Location:** `/var/folders/7s/1d68gzv524b9dwzj3n_cpzs80000gn/T/architecture-review-20260717-103905.html`

### Candidates identified

| # | Candidate | Strength | Summary |
|---|-----------|----------|---------|
| 1 | Collapse 10 handler structs → 2 deep Chirp + User modules | **Strong** | Handler-per-endpoint is shallow; interface ≈ implementation |
| 2 | Extract JWT auth middleware | **Strong** | `GetBearerToken → ValidateJWT` copy-pasted 4× with inconsistent error handling |
| 3 | Introduce a Store seam (interface) for `*database.Queries` | **Strong** | Concrete sqlc type everywhere → zero handler tests without live Postgres |
| 4 | Fix profanity filter bug + deepen chirp validation | Worth exploring | `cleanedBodyElements` is computed but discarded; raw `request.Body` saved to DB |
| 5 | Module-owned routing | Worth exploring | `main.go` manually wires 12 routes with repeated field assignments |
| 6 | Config module | Speculative | 4 identical `os.Getenv → log.Fatalf` blocks |

**Top recommendation:** Candidate 3 (Store seam) — prerequisite for testability of all other candidates.

---

## What was NOT done

- **User has not yet picked a candidate** to explore. The next step per the skill is to run the `/grilling` skill on whichever candidate(s) they choose.
- No code changes have been made.
- No `CONTEXT.md` or ADRs have been created yet. The `/domain-modeling` skill should be invoked once design decisions crystallize during grilling.

---

## Key findings to carry forward

1. **Bug — profanity filter is dead code** (`handlers/chirps.go` lines 74–81): `cleanedBodyElements` is built but line 83 uses `request.Body` (the unfiltered original). This is a real bug, not just an architecture concern.

2. **Inconsistent auth error handling**: `RefreshHandler` title-cases the error message via `cases.Title()`, while `CreateChirpHandler`, `DeleteChirpHandler`, and `UpdateUserHandler` use a raw string. The `handleAuthorizationError` helper in `error.go` title-cases, but `RefreshHandler` duplicates this logic inline.

3. **`LoginHandler` silently drops the refresh token creation error** (line 73): `h.DbQueries.CreateRefreshToken(r.Context(), params)` return value is discarded.

4. **`internal/database/` is entirely sqlc-generated** — `DO NOT EDIT`. The Store interface (Candidate 3) would be defined in a new file, not by editing the generated code. `*database.Queries` already satisfies whatever interface you extract from it.

5. **`ResetHandler` writes body before status code** (line 18–19): `w.Write()` before `w.WriteHeader()` — Go will default to 200 on first `Write`, making the subsequent `WriteHeader(403)` a no-op.

---

## Suggested skills

| Skill | When to use |
|-------|-------------|
| `/grilling` | **Immediately** — once user picks a candidate, grill them on constraints, dependencies, what sits behind the seam, what tests survive |
| `/codebase-design` | Use its vocabulary (module, interface, depth, seam, adapter, leverage, locality) throughout all design discussions — already loaded in this session |
| `/domain-modeling` | Create `CONTEXT.md` and ADRs as decisions crystallize during grilling |
| `/tdd` | After grilling, when implementing the chosen deepening — especially for Candidate 3 (Store seam) where you'll be writing the in-memory adapter and first handler tests |
| `/code-review` | After implementation, review the changes against the architectural decisions made during grilling |

---

## File map (quick reference)

```
chirpy/
├── main.go                          # App entry, DB init, all 12 route registrations
├── handlers/
│   ├── chirps.go                    # CreateChirp, GetChirp, GetChirps, DeleteChirp (200 LOC)
│   ├── users.go                     # CreateUser, UpdateUser (110 LOC)
│   ├── login.go                     # LoginHandler (95 LOC)
│   ├── refresh.go                   # RefreshHandler (59 LOC)
│   ├── revoke.go                    # RevokeHandler (28 LOC)
│   ├── polka.go                     # PolkaWebhookHandler (61 LOC)
│   ├── reset.go                     # ResetHandler (32 LOC)
│   ├── healthz.go                   # HealthCheckHandler (15 LOC)
│   ├── metrics.go                   # MetricsHandler (29 LOC)
│   ├── error.go                     # handleError, handleAuthorizationError
│   └── json.go                      # writeJSON helper
├── internal/
│   ├── auth/                        # JWT, password, bearer, API key, refresh token
│   │   ├── jwt.go + jwt_test.go     # MakeJWT, ValidateJWT (well tested)
│   │   ├── password.go + test       # HashPassword, CheckPasswordHash (well tested)
│   │   ├── bearer.go + test         # GetBearerToken (tested)
│   │   ├── apikey.go                # GetApiKey (untested)
│   │   └── refresh.go               # MakeRefreshToken (untested)
│   └── database/                    # sqlc-generated (DO NOT EDIT)
│       ├── db.go                    # Queries struct, New(), DBTX interface
│       ├── models.go                # Chirp, User, RefreshToken structs
│       ├── chirps.sql.go            # CreateChirp, GetChirp, GetChirps, DeleteChirp
│       ├── users.sql.go             # CreateUser, GetUser, UpdateUser, MakeUserChirpyRed, DeleteUsers
│       └── refresh_tokens.sql.go    # CreateRefreshToken, GetRefreshToken, RevokeRefreshToken
├── metrics/metrics.go               # Metrics struct (atomic counter + middleware)
├── sql/
│   ├── queries/                     # Source SQL for sqlc
│   └── schema/                      # Goose migrations (001–005)
└── sqlc.yaml                        # sqlc config
```
