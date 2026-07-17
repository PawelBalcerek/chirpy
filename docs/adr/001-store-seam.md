# ADR 001 — Introduce Store seam interfaces in `handlers/`

**Date:** 2026-07-17  
**Status:** Accepted  
**Deciders:** @PawelBalcerek

---

## Context

Chirpy's handler structs held `*database.Queries` (a concrete sqlc-generated type) as a direct field. This made every
handler test require a live Postgres database, meaning:

- No unit tests could run without a running DB.
- CI would need DB infrastructure just to test business logic (auth checks, profanity filtering, etc.).
- The handlers were tightly coupled to the persistence implementation rather than its behaviour contract.

An architecture review identified this as a deepening candidate — the handlers were shallow consumers of a wide concrete
type, and a thin interface seam would dramatically increase testability.

---

## Decision

Introduce three domain-grouped interfaces in `handlers/store.go`, owned by the consumer package (`handlers/`) per Go
convention:

| Interface    | Methods                                                                   | Controllers (methods)                                                                                       |
| ------------ | ------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `ChirpStore` | `CreateChirp`, `GetChirp`, `GetChirps`, `DeleteChirp`                     | `ChirpController` (`Create`, `Get`, `List`, `Delete`)                                                       |
| `UserStore`  | `CreateUser`, `GetUser`, `UpdateUser`, `MakeUserChirpyRed`, `DeleteUsers` | `UserController` (`Create`, `Update`, `Login`) + `PolkaController.ReceiveWebhook`, `SystemController.Reset` |
| `TokenStore` | `CreateRefreshToken`, `GetRefreshToken`, `RevokeRefreshToken`             | `UserController` (`Login`, `Refresh`, `Revoke`)                                                             |

`UserController` depends on **both** `UserStore` and `TokenStore` as separate named fields — explicit dependency
declaration, no fat interface.

Interface method signatures mirror sqlc-generated code (same parameter and return types including `database.*` types) so
`*database.Queries` satisfies all three interfaces with zero glue code.

---

## Consequences

### Positive

- Handler unit tests run without a database. Fakes live in `handlers/fakes_test.go` as **function-field structs** (e.g.
  `fakeChirpStore.CreateChirpFunc func(...)`) — each method delegates to its field when set, returns a zero value
  otherwise. This lets individual tests wire only the methods they exercise.
- Context injection for auth-protected handlers is done via `handlers.WithUserIDContext` in `handlers/export_test.go` —
  a `package handlers` file compiled only during `go test`, giving tests access to the unexported `userIDContextKey`
  without exposing it to production callers.
- The seam is precise: each handler only depends on the methods it actually calls.
- `*database.Queries` satisfies all three interfaces structurally; no adapter needed.

### Neutral

- All handler `DbQueries *database.Queries` fields replaced with named interface fields matching the interface type:
  - `ChirpController.DbQueries` → `ChirpStore ChirpStore`
  - `UserController.DbQueries` → `UserStore UserStore` + `TokenStore TokenStore`
  - `SystemController.DbQueries` → `UserStore UserStore`
  - `PolkaController.DbQueries` → `UserStore UserStore`
- `main.go` wiring updated accordingly; value stays `dbQueries` (`*database.Queries` satisfies all three interfaces).

### Negative / Trade-offs

- Handler files still import `github.com/PawelBalcerek/chirpy/internal/database` for param/model types. The seam breaks
  the concrete-method dependency, not the type dependency. A further step (domain value types) could remove that
  coupling but is out of scope.

---

## Alternatives considered

| Alternative                               | Rejected because                                                                                |
| ----------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Single `Store` interface (all 12 methods) | Fat interface — each handler depends on methods it never calls; fakes must implement everything |
| Interface in `internal/database/`         | Breaks Go consumer-owns-interface convention; couples the db package to handler concerns        |
| Integration tests only                    | Keeps DB as hard dependency; slow, fragile, blocks CI without infrastructure                    |
