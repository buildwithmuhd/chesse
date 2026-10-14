# Avant-Garde Chess

Browser chess arena with a Go backend for accounts, authenticated multiplayer, stored game history, and tournaments. Built from `project.md`.

## Access Rules

Authentication is the gate for everything that touches other people or stored data.

| Feature | Guest | Authenticated |
|---|---|---|
| Play vs AI | Yes | Yes |
| Game history and scores stored | No (nothing is persisted) | Yes |
| Challenge another player | No | Yes |
| Spectate live games | No | Yes |
| Host a tournament | No | Yes |
| Join a tournament | No | Yes |
| Voice chat | No | Yes |

Guest AI games run entirely in the browser. The backend never sees them, so there is nothing to score or store.

## Architecture

```
Browser (web client)
  chess.js rules, board UI, AI games (guest + authed)
  REST + WebSocket
        |
Go backend
  auth, users, challenges, games, history, tournaments, realtime hub
        |
PostgreSQL            (users, sessions, games, moves, tournaments)
```

- **Web client**: TypeScript, Vite, `chess.js` for client-side legality and UX.
- **Backend**: Go. The server is authoritative for every rated or challenged game; client-side legality is a UX convenience only.
- **Database**: PostgreSQL via `pgx` with SQL migrations.
- **Realtime**: WebSockets for moves, clocks, challenges, spectating, and tournament updates.
- **Voice**: WebRTC (LiveKit), with room tokens issued by the Go backend to authenticated users only.

## Authentication

- Email + password registration and login.
- Passwords hashed with argon2id.
- Short-lived access token (JWT) plus a rotating refresh token in an `HttpOnly`, `Secure`, `SameSite` cookie.
- Auth middleware guards all protected REST routes and the WebSocket upgrade.
- Rate limiting on login and registration.
- The web client treats "no session" as guest mode: AI play only, with challenge, tournament, history, and spectate entry points redirecting to sign in.

## Repository Layout

```
.
|-- chesse-frontend/         # Frontend (Vite + TypeScript)
|   |-- src/
|   |-- public/
|   `-- package.json
|-- chesse-backend/          # Go backend
|   |-- cmd/chesse-backend/ # Entry point
|   |-- internal/
|   |   |-- auth/           # Register, login, tokens, middleware
|   |   |-- users/
|   |   |-- games/          # Authoritative game state, history
|   |   |-- challenges/
|   |   |-- tournaments/    # Hosting, joining, pairings, standings
|   |   |-- realtime/       # WebSocket hub, spectating
|   |   `-- db/             # Migrations and queries
|   `-- go.mod
|-- chessground/            # Reference repo (lichess-org/chessground), not shipped
|-- lila/                   # Reference repo (lichess-org/lila), not shipped
|-- chesse/
|   `-- README.md           # Product plan
|-- docs/
|   `-- decisions.md
|-- chess-design.jpg        # UI inspiration           

```

## Reference Material

**`chessground/`** and **`lila/`** are cloned from `https://github.com/lichess-org/chessground.git` and `https://github.com/lichess-org/lila.git`. They are for study: board interaction patterns (drag/click, highlights, animation timing), game and tournament flow, Swiss/arena pairing logic, and data modeling.

> Licensing: chessground is GPL-3.0 and lila is AGPL-3.0. Use them to learn from, not to copy from. Copying their code into this project would pull this project under the same licenses. Write our own implementations.

**`chess-design.jpg`** is the visual direction for the UI: layout, spacing, palette, and overall mood. Match the feel; do not trace it pixel for pixel.



## Run

Backend:

```bash
cd chesse-backend
cp .env.example .env        # DATABASE_URL, JWT_SECRET, etc.
go run ./cmd/chesse-backend
```

Frontend:

```bash
cd chesse-frontend
npm.cmd install
npm.cmd run dev -- --port 5173
```

Open `http://127.0.0.1:5173/`.

## Verify

```bash
cd chesse-backend && go vet ./... && go test -race ./...
cd chesse-frontend && npm.cmd test && npm.cmd run build
```

## API Surface (planned)

Public:

- `GET /healthz`
- `GET /readyz`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`

Authenticated:

- `GET /me`
- `GET /me/games`, `GET /games/:id`
- `POST /challenges`, `POST /challenges/:id/accept`, `POST /challenges/:id/decline`
- `GET /games/live`, `WS /ws/games/:id` (play and spectate)
- `POST /tournaments`, `GET /tournaments`, `GET /tournaments/:id`
- `POST /tournaments/:id/join`, `POST /tournaments/:id/start`
- `POST /voice/token`

Guest AI games call none of these.

## Built So Far (frontend, browser-local)

- Legal chess via `chess.js`: check, checkmate, stalemate, castling, en passant, promotion, threefold repetition, insufficient material, FEN, and PGN.
- Responsive board with click-to-move, legal target and capture highlighting, promotion picker, undo, resign, orientation flip, and local/AI modes.
- Chess clocks for 5+0, 10+0, and 30+0.
- Move list, captured material, FEN and PGN export.
- Local tournament surface (players, pairings, result entry), to be replaced by the backend-driven version.
- Local microphone control with level meter as the voice-chat foundation.
- Local ambient music picker and generated move, check, capture, and game-end sounds.

## Build Order

1. Go server skeleton, Postgres, migrations, config.
2. Auth: register, login, refresh, middleware, rate limits.
3. Frontend auth flow and guest/authenticated gating (AI-only for guests).
4. Game persistence: store finished games and history for authenticated users.
5. Challenges and authoritative multiplayer over WebSockets.
6. Spectating.
7. Tournaments: create, join, pairings, standings.
8. Voice via LiveKit, with tokens from the Go backend.
9. Re-skin the UI with `chess-design.jpg` direction and the pixel packs.

## Current Boundary

The browser app is still local-only until steps 1-5 land. Until then there are no real accounts, no authoritative multiplayer, and no persistent tournaments.
