# PersonaForge Backend

Go + Gin backend for persona chat (Gemini) with Google auth and guest mode.

## Requirements

- Go 1.21+
- Docker (optional, recommended for Postgres)

## Configuration

The server **requires** these env vars (see `internal/config/config.go`):

- `DATABASE_URL`
- `JWT_SECRET`
- `GEMINI_API_KEY`
- `GOOGLE_CLIENT_ID`

Optional:

- `ENV` (default: `development`)
- `PORT` (default: `8080`)
- `GEMINI_MODEL` (default: `gemini-2.5-flash`)
- `JWT_EXPIRY_MINUTES` (default: `30`)

For local development, use **`.env.local`**.
For production deployments, use **`.env.prod`**.

This repo includes `env.example` as a template (copy it to `.env.local` / `.env.prod` and fill values).

## Run locally

With Postgres running and env vars set:

```bash
make run
```

## Run with Docker (Postgres + API)

```bash
make docker-local
```

The API will be available at:

- `http://localhost:8080/health`
- `http://localhost:8080/docs/index.html`

## Production Docker compose

```bash
make docker-prod
```

## API behavior (guest vs authenticated)

- **Default personas**: always visible.
- **Guest personas**:
  - Create an anonymous session via `POST /api/auth/anonymous`
  - Create **exactly 1** custom persona by calling `POST /api/personas` with `X-Session-ID: <session_id>`
  - Guests **cannot** access `/api/chat/history` or `/api/insight`
- **Authenticated users**:
  - Login via `POST /api/auth/google` (send Google `id_token`)
  - Use returned JWT (`Authorization: Bearer <token>`) for protected endpoints
  - Can create multiple personas

## Tests

```bash
make test
```

Tests are endpoint-focused (Gin + httptest) and cover:

- auth routes
- guest persona limit
- chat history/insight auth restrictions

## Lint

```bash
make lint
```

This runs `golangci-lint` via Docker for a consistent setup.

## Swagger docs

Swagger JSON/YAML and the `docs` Go package are generated via `swag` and checked into `docs/`.
If you change route annotations, regenerate with:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.3 init -g main.go -o docs
```


