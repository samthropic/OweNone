# OweNone

Single repository containing both backend and frontend apps.

## About

Drowning in group IOU's?
We untangle the mess.

OweNone transforms fragmented group expenses into the minimum number of payments across your entire relationship graph with real-time recomputation, cross-group netting, and settlement paths that actually make sense.

## Repository Layout

```text
/
	backend/            Go API, PostgreSQL migrations, and settlement engine
	owenone-frontend/   Next.js 16 application
```

## Quick Start

Requirements:

- PostgreSQL 17
- Go 1.25 or newer
- Node.js 24 LTS (see `owenone-frontend/.nvmrc`)

### 1. Start PostgreSQL

```bash
brew install postgresql@17
brew services start postgresql@17
export PATH="/opt/homebrew/opt/postgresql@17/bin:$PATH"
```

### 2. Create the database

One-time setup:

```bash
psql -d postgres -c "CREATE ROLE owenone LOGIN PASSWORD 'owenone' CREATEDB;"
createdb -O owenone owenone
```

### 3. Seed and run the API

```bash
cd backend
export DATABASE_URL='postgres://owenone:owenone@localhost:5432/owenone?sslmode=disable'
go run ./cmd/seed   # runs migrations, then loads demo data
go run ./cmd/api
```

### 4. Run the frontend

In a second terminal:

```bash
cp owenone-frontend/.env.example owenone-frontend/.env.local
cd owenone-frontend
nvm use
npm install
npm run dev
```

Open `http://localhost:3000`. The seeded user is Sarah Johnson and the demo graph contains six debts across three groups that reduce to one £12 payment.

The API listens on `http://localhost:8080`; readiness is available at `/health/ready`. See [backend/README.md](backend/README.md) for architecture, direct Go commands, and the development identity boundary.

On later sessions PostgreSQL is already running, so only steps 3 and 4 are needed. Stop the database with `brew services stop postgresql@17` when you are not developing.

## Validation

```bash
cd backend && go test ./... && go vet ./...
cd ../owenone-frontend && npm run build
```