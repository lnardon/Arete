## Local Development

**Prerequisites:** Go 1.26+, Node 22+, PostgreSQL 17+

```bash
# Start the database
docker compose up postgres -d

# Backend
cd server
cp .env.example .env.local   # edit .env.local — needs OPENAI_API_KEY
go run ./cmd/api

# Frontend
cd frontend
npm install
npm run dev
```

The frontend dev server runs on `http://localhost:5173` and proxies `/api` requests to the Go server on port 8080.

## AI Model (OpenAI)

The WhatsApp assistant's tool-calling loop (`internal/ai`) runs against the OpenAI API. You need an `OPENAI_API_KEY` — no other setup.

- **Default model:** `gpt-5.6-luna` — the cheapest current tier, still supports full function/tool calling. Change it via `OPENAI_MODEL` in `.env`/`.env.local`, no code changes needed. `gpt-5.6-terra` (balanced) or `gpt-5.6-sol` (frontier) are drop-in upgrades if replies need to be smarter — cost scales up accordingly.
- **Cost:** this is a paid API — every WhatsApp turn (including each tool call round-trip) is billed per token. For a personal habit-tracker bot doing simple CRUD calls, usage should be low, but keep an eye on your OpenAI usage dashboard, especially while testing.

## WhatsApp Assistant (Evolution API)

The WhatsApp integration talks to a self-hosted [Evolution API](https://doc.evolution-api.com) instance. Arete's Go server only sends messages and receives webhooks at runtime — provisioning the WhatsApp connection itself is a one-time manual step per environment.

### 1. Start Evolution API

```bash
export EVOLUTION_API_KEY=$(openssl rand -hex 32)
export EVOLUTION_WEBHOOK_SECRET=$(openssl rand -hex 32)
docker compose up evolution-postgres evolution-redis evolution-api -d
```

### 2. Create the instance and pair a WhatsApp number

Evolution API ships a web Manager UI — the easiest path is opening `http://localhost:8081/manager`, authenticating with `EVOLUTION_API_KEY`, creating an instance named `arete` (matching `EVOLUTION_INSTANCE_NAME`), and scanning the QR code with the WhatsApp account you want the assistant to run as.

If you'd rather script it, Evolution API also exposes this over REST (exact request/response shape can drift between versions — check `https://doc.evolution-api.com` if these don't match what's deployed):

```bash
curl -X POST http://localhost:8081/instance/create \
  -H "apikey: $EVOLUTION_API_KEY" -H "Content-Type: application/json" \
  -d '{"instanceName": "arete", "qrcode": true, "integration": "WHATSAPP-BAILEYS"}'
# Then fetch/refresh the QR code:
curl http://localhost:8081/instance/connect/arete -H "apikey: $EVOLUTION_API_KEY"
```

### 3. Confirm the webhook

`docker-compose.yml` already points Evolution API's global webhook at `http://app:8080/api/v1/webhooks/whatsapp?token=$EVOLUTION_WEBHOOK_SECRET`. Evolution API has no way to attach a custom header to outbound webhook calls, so the shared secret rides in the query string instead — `WhatsAppHandler.Webhook` checks it there, not in a header. This is weaker than a header (the token can end up in Evolution API's own request logs), which is why production also relies on `evolution-api` not being publicly reachable at all — see the Production Checklist.

Send yourself a test WhatsApp message from another phone and check the Go server logs for a `whatsapp:` log line to confirm delivery.

### 4. Backend env vars

Add to `server/.env.local` (see `server/.env.example`):

```
EVOLUTION_API_BASE_URL=http://localhost:8081
EVOLUTION_API_KEY=<same value as above>
EVOLUTION_INSTANCE_NAME=arete
EVOLUTION_WEBHOOK_SECRET=<same value as above>
OPENAI_API_KEY=<your OpenAI API key>
OPENAI_MODEL=gpt-5.6-luna
```

### 5. Frontend: the assistant's phone number

The Settings page shows the assistant's WhatsApp number in its linking instructions. Set it (not a secret — just the display number) in `frontend/.env.local`:

```
VITE_WHATSAPP_NUMBER=+1 555 0100
```

## Docker Compose

```bash
# Set required env vars
export JWT_SECRET=$(openssl rand -hex 32)
export APP_DOMAIN=http://localhost:8080

export EVOLUTION_API_KEY=$(openssl rand -hex 32)
export EVOLUTION_WEBHOOK_SECRET=$(openssl rand -hex 32)
export OPENAI_API_KEY=sk-...

docker compose up --build -d
```

App is available at `http://localhost:8080`.

## Production Checklist

- [ ] `JWT_SECRET` — random string, **minimum 32 characters** (`openssl rand -hex 32`)
- [ ] `APP_DOMAIN` — set to the public domain (e.g. `https://arete.com`)
- [ ] `COOKIE_SECURE=true` — required when serving over HTTPS
- [ ] `DB_SSLMODE=require` — enable TLS for the database connection
- [ ] `DB_PASSWORD`
- [ ] `EVOLUTION_API_KEY` / `EVOLUTION_WEBHOOK_SECRET` — random strings, generated fresh per environment
- [ ] `evolution-api` is **not** publicly routable — reachable only from `app` over a private network (e.g. Railway's internal networking). The `8081` port mapping in `docker-compose.yml` is for local QR-pairing/admin only.
- [ ] `OPENAI_API_KEY` set, with usage limits/alerts configured on the OpenAI dashboard sized for expected WhatsApp traffic — this is billed per token, unlike the rest of the stack.
