# Decisions

## 2026-10-14

- Backend lives in `chesse-backend/` and uses the requested entry point `cmd/chesse-backend`.
- The existing Vite frontend remains in `chesse-frontend/` for now. The spec names `chesse-frontend/`, but moving the current app is outside step 1 and would add churn before auth/backend integration.
- Reference repos are siblings of the app folders: `chessground/` and `lila/`. They are study-only because their GPL-3.0 and AGPL-3.0 licenses are incompatible with copying code into this project.
- The UI reference asset exists as `chess-design.jpg`, not `chess-design.jpeg`.
- No license or credits file was found under `chesse/public/assets/`; do not ship those pixel assets publicly until their license is confirmed.
