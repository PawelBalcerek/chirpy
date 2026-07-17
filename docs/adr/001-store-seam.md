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

An architecture review (conversation `096bd71f-a896-4d24-8cbd-46190772c2f8`) identified this as a deepening candidate —
the handlers were shallow consumers of a wide concrete type, and a thin interface seam would dramatically increase
testability.

---

## Decision

Introduce three domain-grouped interfaces in `handlers/store.go`, owned by the consumer package (`handlers/`) per Go
convention:

| Interface    | Methods                                                                   | Controllers (methods)                                                                 |
| ------------ | ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| `ChirpStore` | `CreateChirp`, `GetChirp`, `GetChirps`, `DeleteChirp`                     | `ChirpController` (`Create`, `Get`, `List`, `Delete`)                                 |
| `UserStore`  | `CreateUser`, `GetUser`, `UpdateUser`, `MakeUserChirpyRed`, `DeleteUsers` | `UserController` (`Create`, `Update`, `Login`) + `PolkaController.ReceiveWebhook`, `SystemController.Reset` |
| `TokenStore` | `CreateRefreshToken`, `GetRefreshToken`, `RevokeRefreshToken`             | `UserController` (`Login`, `Refresh`, `Revoke`)                                       |

`LoginHandler` depends on **both** `UserStore` and `TokenStore` as separate named fields — explicit dependency
declaration, no fat interface.

Interface method signatures mirror sqlc-generated code (same parameter and return types including `database.*` types) so
`*database.Queries` satisfies all three interfaces with zero glue code.

---

## Consequences

### Positive

- Handler unit tests run without a database — in-memory fakes (`fakeChirpStore`, etc.) implemented in `*_test.go` files.
- The seam is precise: each handler only depends on the methods it actually calls.
- `*database.Queries` satisfies all three interfaces structurally; no adapter needed.
- Surfaced and fixed a pre-existing profanity filter dead-code bug (`CreateChirpHandler` was saving `request.Body`
  instead of `cleanedBody`).

### Neutral

- `LoginHandler` field names changed from `DbQueries` to `UserStore`/`TokenStore`; `main.go` wiring updated accordingly.
- Other handler field `DbQueries` type changed from `*database.Queries` to the appropriate interface — no field rename
  needed.

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
