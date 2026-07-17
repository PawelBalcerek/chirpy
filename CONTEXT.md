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
| **Chirp Body**       | The text content of a chirp. Must be 1–140 chars and sanitised of profane words.                                       |
| **User**             | A registered account holder with email + hashed password. Can be upgraded to Chirpy Red.                              |
| **Chirpy Red**       | A premium subscription status for a user, triggered by a Polka webhook event.                                         |
| **JWT**              | Short-lived (1 h) access token. Used to authenticate API calls.                                                        |
| **Refresh token**    | Long-lived (60 days) token. Exchanged for a new JWT. Can be revoked.                                                   |
| **Polka**            | External payment webhook source. Sends `user.upgraded` events.                                                         |
| **Profanity filter** | Replaces `kerfuffle`, `sharbert`, `fornax` (case-insensitive) with `****`.                                             |

---

## Architecture

```
main.go
  └─ registers routes → controller structs (handlers/)
        └─ middleware: RequireJWT (JWT auth), RequireApiKey (Polka auth)
        └─ depend directly on *database.Queries (internal/database/)
```

Key decisions documented in `docs/adr/`:

- [ADR 001](docs/adr/001-store-seam.md) — Store seam (superseded by controller refactor in main branch)

---

## Package map

| Package              | Role                                                        |
| -------------------- | ----------------------------------------------------------- |
| `main`               | App entry, DB init, route wiring                            |
| `handlers`           | HTTP controllers (ChirpController, UserController, PolkaController, SystemController) + middleware |
| `internal/auth`      | JWT creation/validation, password hashing, token generation |
| `internal/chirp`     | Chirp body validation and profanity filtering               |
| `internal/database`  | sqlc-generated DB access (`*database.Queries`)              |
| `metrics`            | Request counter, served via admin routes                    |

---

## Known bugs / follow-ups

| #   | Bug                                                                            | Status                            |
| --- | ------------------------------------------------------------------------------ | --------------------------------- |
| 1   | Profanity filter was dead code — `request.Body` saved instead of `cleanedBody` | **Fixed** in controller refactor  |
| 2   | `LoginHandler` silently drops `CreateRefreshToken` error                       | **Fixed** in controller refactor  |
| 3   | `ResetHandler` writes body before status code                                  | Open — follow-up issue            |
