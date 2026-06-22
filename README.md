# Ishchi Bormi

An Uzbek daily-job / labor marketplace. Telegram-based passwordless login, MongoDB-backed Go API, Next.js (App Router + TS + Tailwind) frontend, and a Go Telegram bot for OTP delivery and notifications.

> **Realistic status note:** this repo is a complete, runnable skeleton of the platform described in the spec. All backend domains are implemented end-to-end (auth, users, categories, e'lons, applications, reviews, finance, chat over WebSocket, notifications, reports, admin). The frontend wires every screen to the API. Pixel-faithful 1:1 polish for every one of the 24 design screenshots is out of scope for a single shot — the design tokens and layout match the screenshots, but individual screens may need polish passes.

## Stack

| Layer    | Tech                                                            |
|----------|-----------------------------------------------------------------|
| Frontend | Next.js (App Router) + TypeScript + Tailwind CSS                |
| Backend  | Go (chi router, mongo-go-driver, gorilla/websocket, JWT)        |
| Bot      | Go (`go-telegram-bot-api/v5`)                                   |
| DB       | MongoDB 7 (TTL collection used for OTP — no Redis needed)       |
| Files    | Local disk volume for avatars                                   |
| Infra    | Docker + docker-compose, Makefile                               |

## Prerequisites

- Docker + Docker Compose (or Go 1.22+ and Node 20+ for local dev)
- A Telegram bot — create via [@BotFather](https://t.me/BotFather), get the token and username

## Quick start

```bash
cp .env.example .env
# edit .env: set TELEGRAM_BOT_TOKEN, TELEGRAM_BOT_USERNAME, and the JWT/secret values
docker compose up --build -d
# seed realistic demo data (≥15 records per collection)
docker compose exec backend /app/seed
# open
open http://localhost:3000
```

Default admin: `admin / Admin123!` (override via `ADMIN_SEED_USER` / `ADMIN_SEED_PASS`).

## Local dev (no Docker)

```bash
# Mongo (Docker)
docker run -d --name mongo -p 27017:27017 mongo:7

cp .env.example .env
make be-run    # backend on :8080
make bot-run   # Telegram bot (needs TELEGRAM_BOT_TOKEN)
make fe-run    # frontend on :3000
make seed      # populate demo data
```

## Login flow

1. Open `http://localhost:3000/login`.
2. Click **Telegram botga o'tish** — opens `t.me/<bot>?start=<token>`.
3. In the bot send `/start` and share your contact.
4. The bot replies with a 6-digit code.
5. Enter the code on `/login` → you're in.

In `OTP_DEV_RETURN=true` mode (the default in `.env.example`) the backend additionally returns the OTP in the issue response, so you can complete the login flow even without a real bot token.

## API surface (high level)

- `POST /api/auth/otp/request` → returns `{ tgToken }` and (in dev) `{ code }`
- `POST /api/auth/otp/verify` → returns `{ accessToken, refreshToken, user }`
- `POST /api/auth/refresh`
- `GET  /api/me`, `PATCH /api/me`
- `GET  /api/users/:id`, `GET /api/users?q=`
- `POST /api/users/:id/block`, `DELETE /api/users/:id/block`
- `GET  /api/categories`, `POST /api/categories`
- `GET  /api/elons`, `POST /api/elons`, `GET/PATCH/DELETE /api/elons/:id`, `POST /api/elons/:id/publish`
- `GET  /api/my/elons`
- `POST /api/elons/:id/apply`
- `GET  /api/my/applications`, `GET /api/my/elons/applications`
- `POST /api/applications/:id/accept|reject|cancel|confirm-done`
- `POST /api/applications/:id/review`
- `GET  /api/me/history`, `GET /api/me/finance`
- `GET  /api/conversations`, `POST /api/conversations`, `GET /api/conversations/:id/messages`, `POST /api/conversations/:id/messages`
- `GET  /api/notifications`, `POST /api/notifications/read-all`
- `POST /api/reports`
- `WS   /ws?token=...` — message + notification stream
- `POST /api/admin/login` + `GET/POST /api/admin/*`

## Folder layout

```
ishchibormi/
├── backend/         # Go API + WebSocket + seed
├── bot/             # Go Telegram bot
├── frontend/        # Next.js app
├── design/          # the 24 reference screenshots
├── docker-compose.yml
├── .env.example
├── Makefile
└── README.md
```

See each subfolder's source for details.
