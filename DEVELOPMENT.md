## Local Development

**Prerequisites:** Go 1.24+, Node 22+, PostgreSQL 17+

```bash
# Start the database
docker compose up postgres -d

# Backend
cd server
cp .env.example .env.local   # edit .env.local
go run ./cmd/api

# Frontend
cd frontend
npm install
npm run dev
```

The frontend dev server runs on `http://localhost:5173` and proxies `/api` requests to the Go server on port 8080.

## Docker Compose

```bash
# Set required env vars
export JWT_SECRET=$(openssl rand -hex 32)
export APP_DOMAIN=http://localhost:8080

docker compose up --build
```

App is available at `http://localhost:8080`.

## Production Checklist

- [ ] `JWT_SECRET` — random string, **minimum 32 characters** (`openssl rand -hex 32`)
- [ ] `APP_DOMAIN` — set to the public domain (e.g. `https://arete.com`)
- [ ] `COOKIE_SECURE=true` — required when serving over HTTPS
- [ ] `DB_SSLMODE=require` — enable TLS for the database connection
- [ ] `DB_PASSWORD`
