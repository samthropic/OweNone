# OweNone API

Go API and PostgreSQL ledger for cross-group expense reconciliation.

## Architecture

- `cmd/api`: HTTP server with graceful shutdown and health checks.
- `cmd/seed`: idempotent demo-data loader.
- `internal/domain`: integer-money ledger projection and exact minimum-transfer search.
- `internal/app`: validation and use-case orchestration.
- `internal/store`: authorized PostgreSQL reads and transactional writes.
- `internal/httpapi`: versioned JSON handlers and middleware.
- `migrations`: embedded, ordered PostgreSQL migrations.

Expenses and completed settlements are the source of truth. Dashboard totals and settlement suggestions are recomputed from the ledger, so derived balances cannot drift from their underlying transactions. Each currency is reconciled independently; the current dashboard uses the user's preferred currency.

## Run Locally

Requires PostgreSQL 17 running locally. One-time database setup:

```bash
brew services start postgresql@17
export PATH="/opt/homebrew/opt/postgresql@17/bin:$PATH"
psql -d postgres -c "CREATE ROLE owenone LOGIN PASSWORD 'owenone' CREATEDB;"
createdb -O owenone owenone
```

Then seed and run:

```bash
export DATABASE_URL='postgres://owenone:owenone@localhost:5432/owenone?sslmode=disable'
go run ./cmd/seed
go run ./cmd/api
```

The API is then available at `http://localhost:8080`. Migrations run automatically at API startup and are serialized with a PostgreSQL advisory lock. `cmd/seed` also applies migrations before loading demo data, so it is safe to run first on an empty database.

The seeded demo user is **sarah@example.com** with password **password123**. The demo ledger spans several months of expenses across a handful of groups, with dozens of activities, varied balances (some owed to you, some you owe, some already settled), and a debt graph that visibly compresses into a smaller set of transfers. All demo data is denominated in **USD**. `go run ./cmd/seed` is idempotent — re-running it on an already-seeded database is safe.

New accounts created via `POST /api/v1/auth/signup` also default to **USD** as their preferred currency. The dashboard reconciles one currency at a time — the signed-in user's preferred currency — so a user whose preferred currency has no associated expenses will see an empty ledger. If you sign up and the dashboard shows nothing, verify that your preferred currency matches the currency of your group's expenses (update it via `POST /api/v1/profile`).

## Authentication

User accounts are created via `POST /api/v1/auth/signup` and authenticated via `POST /api/v1/auth/login`. Both endpoints return an opaque session token and the authenticated user record. All other API endpoints (except `/health/*`) require `Authorization: Bearer <token>`.

**Password hashing**: passwords are hashed with PBKDF2-HMAC-SHA256 (`crypto/pbkdf2`, ~210 000 iterations, 16-byte random salt). The stored value has the form `pbkdf2_sha256$<iterations>$<salt>$<key>`. No third-party crypto library is used.

**Sessions**: the raw session token is returned once at signup/login. Only its SHA-256 hash is persisted in a `sessions` table with a 30-day `expires_at`. The Next.js server stores the token in an `httpOnly`, `SameSite=Lax` cookie named `owenone_session` (flagged `Secure` in production), so it is never exposed to client JavaScript.

Session lifetime is 30 days from issuance. Production deployments would additionally need token rotation, per-IP rate limiting on login, email verification, and CSRF protection for browser clients.

Financial `POST` requests also require `Idempotency-Key`. Reusing a key with the same payload returns the original resource; reusing it with a different payload returns `409 Conflict`.

## Commands

```bash
go test ./...
go vet ./...
go run ./cmd/api
go run ./cmd/seed
```

The API contract is documented in [openapi.yaml](openapi.yaml).