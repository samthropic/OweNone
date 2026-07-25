# OweNone frontend

Next.js 16 frontend for OweNone. The home page is backed by the Go API; it is
not a standalone static page.

## Requirements

- Node.js 24 LTS (the version is pinned in `.nvmrc`)
- npm
- The OweNone API running locally on port 8080, backed by PostgreSQL 17

Node 23 is end-of-life and is not supported for this project. With `nvm`:

```bash
nvm install
nvm use
```

## Local setup

Start and seed the backend first:

```bash
cd ../backend
export DATABASE_URL='postgres://owenone:owenone@localhost:5432/owenone?sslmode=disable'
go run ./cmd/seed   # first run only
go run ./cmd/api
```

Verify it is up:

```bash
curl --fail http://localhost:8080/health/ready
```

Then start the frontend in a second terminal:

```bash
cd owenone-frontend
cp .env.example .env.local
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). If the API is not running,
the `/` route fails because it loads `/api/v1/dashboard` during server rendering.
The `/landing` route does not require the API.

## Resource troubleshooting

Next.js compilation adds several hundred megabytes of resident memory. On a
16 GB Mac that is already swapping, this can make the whole system appear to
freeze even when Next itself is healthy. Before starting it:

```bash
sysctl vm.swapusage
ps -axo pid,rss,command -r | head
```

Close memory-heavy browser tabs or editor windows when swap use is already
high. If the Turbopack cache becomes unusually large or behaves as if stale,
stop Next and remove `.next`; it is generated output and will be recreated.

## Configuration

`OWENONE_API_URL` defaults to `http://localhost:8080`. `OWENONE_USER_ID`
defaults to the seeded demo user. See `.env.example`.

## Validation

```bash
npm run lint
npx tsc --noEmit
```
