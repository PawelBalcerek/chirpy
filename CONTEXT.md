# Chirpy — Domain Context

> Single-context domain model. Agents and LLMs: read this file before touching any handler or database code.

---

## What Chirpy is

Chirpy is a Go REST API that implements a small Twitter-like micro-posting service. Users post short messages called
**chirps** (≤ 140 characters). Profane words are replaced with `****` before storage. Users authenticate via JWT
(short-lived) and refresh tokens (long-lived).

---

## Ubiquitous language

| Term                 | Definition                                                                                                             |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| **Chirp**            | A short user message (≤ 140 chars). Profanity-filtered before storage.                                                 |
| **User**             | An account with email + hashed password. Can be upgraded to Chirpy Red.                                                |
| **Chirpy Red**       | A premium user status, triggered by a Polka webhook event.                                                             |
| **JWT**              | Short-lived (1 h) access token. Used to authenticate API calls.                                                        |
| **Refresh token**    | Long-lived (60 days) token. Exchanged for a new JWT. Can be revoked.                                                   |
| **Polka**            | External payment webhook source. Sends `user.upgraded` events.                                                         |
| **Store seam**       | The interface boundary (`ChirpStore`, `UserStore`, `TokenStore`) between handlers and the database layer. See ADR 001. |
| **Profanity filter** | Replaces `kerfuffle`, `sharbert`, `fornax` (case-insensitive) with `****`.                                             |

---

## Architecture

```
main.go
  └─ registers routes → handler structs (handlers/)
        └─ depend on store interfaces (ChirpStore / UserStore / TokenStore)
              └─ satisfied in production by *database.Queries (internal/database/)
              └─ satisfied in tests by in-memory fakes (*_test.go)
```

Key decisions documented in `docs/adr/`:

- [ADR 001](docs/adr/001-store-seam.md) — Store seam: `ChirpStore`, `UserStore`, `TokenStore` interfaces in `handlers/`

---

## Package map

| Package             | Role                                                        |
| ------------------- | ----------------------------------------------------------- |
| `main`              | App entry, DB init, route wiring                            |
| `handlers`          | HTTP handlers + store interfaces + fakes (test only)        |
| `internal/auth`     | JWT creation/validation, password hashing, token generation |
| `internal/database` | sqlc-generated DB access (`*database.Queries`)              |
| `metrics`           | Request counter, served via admin routes                    |

---

## Known bugs / follow-ups

| #   | Bug                                                                            | Status                            |
| --- | ------------------------------------------------------------------------------ | --------------------------------- |
| 1   | Profanity filter was dead code — `request.Body` saved instead of `cleanedBody` | **Fixed** in Store seam changeset |
| 2   | `LoginHandler` silently drops `CreateRefreshToken` error                       | Open — follow-up issue            |
| 3   | `ResetHandler` writes body before status code                                  | Open — follow-up issue            |
