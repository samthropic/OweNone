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
	frontend/           Next.js 16 application
```

## Quick Start

Requirements:

- PostgreSQL 17
- Go 1.25 or newer
- Node.js 24 LTS (see `frontend/.nvmrc`)

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
cp frontend/.env.example frontend/.env.local
cd frontend
nvm use
npm install
npm run dev
```

Open `http://localhost:3000` — you land on the marketing page; click **Sign up** to create an account, or log in with the seeded demo user **sarah@example.com / password123** (demo data and new accounts both use **USD**). The frontend serves `/landing` (marketing), `/login`, `/signup`, `/` (authenticated dashboard, redirects to `/login` when signed out), `/friends` (friends tab: add or remove friends by email), `/groups` (groups tab: create, edit or delete groups; open a group to see its expenses and add a new one), `/activity` (the activity tab: full history and where you add an expense), `/settle-up` (the settle up tab: record the payments that clear your balances), and `/profile` (your profile).

The API listens on `http://localhost:8080`; readiness is available at `/health/ready`. See [backend/README.md](backend/README.md) for architecture, direct Go commands, and the authentication system.

On later sessions PostgreSQL is already running, so only steps 3 and 4 are needed. Stop the database with `brew services stop postgresql@17` when you are not developing.

## Validation

```bash
cd backend && go test ./... && go vet ./...
cd ../frontend && npm run build
```