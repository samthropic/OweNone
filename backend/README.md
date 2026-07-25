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

The seeded dashboard user ID is `10000000-0000-0000-0000-000000000001`. The demo ledger contains six obligations across three groups that compress to one £12 payment, cancelling £22 of offsetting debt.

## Development Identity

Authenticated development endpoints require `X-User-ID`. The Next.js server supplies this header from `OWENONE_USER_ID`, so it is not controlled by browser input. This boundary is intentionally replaceable: production must validate a real session or access token and derive the user ID from its subject rather than accept this header from public clients.

Financial `POST` requests also require `Idempotency-Key`. Reusing a key with the same payload returns the original resource; reusing it with a different payload returns `409 Conflict`.

## Commands

```bash
go test ./...
go vet ./...
go run ./cmd/api
go run ./cmd/seed
```

The API contract is documented in [openapi.yaml](openapi.yaml).