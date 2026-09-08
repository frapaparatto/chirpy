# Chirpy

A small HTTP server written in Go, built while following [boot.dev's Learn HTTP Servers in Go course](https://www.boot.dev/courses/learn-http-servers-golang).

Chirpy is a simplified Twitter clone. It serves a static frontend, exposes a JSON API, and stores data in Postgres using queries generated with [sqlc](https://sqlc.dev/).

## Learning context

The course provides the chapters; the decisions inside each chapter are mine. Where it leaves something open (how to shape a validation error, where profanity filtering happens, or what a reset endpoint is allowed to touch), I make that call.

## Status

Implemented:

- Static file server for the frontend, mounted at `/app/`
- `GET /api/healthz`, a health check
- `GET /api/chirps`: list chirps, oldest first by default; supports filtering by `author_id` and reversing order with `sort=desc`
- `GET /api/chirps/{chirpID}`: fetch one chirp by UUID
- `POST /api/chirps`: create a chirp (requires a JWT); 140-character limit and profanity filtering
- `DELETE /api/chirps/{chirpID}`: delete a chirp (requires a JWT; only the chirp's author may delete it)
- `POST /api/users`: create a user with an Argon2id-hashed password
- `PUT /api/users`: update the authenticated user's email/password (requires a JWT)
- `POST /api/login`: verify an email/password pair and issue a JWT plus a refresh token
- `POST /api/refresh`: exchange a valid refresh token for a new JWT
- `POST /api/revoke`: revoke a refresh token
- `POST /api/polka/webhooks`: upgrade a user to Chirpy Red, authenticated with an API key
- `GET /admin/metrics` and `POST /admin/reset`: view and reset file-server hit counts; reset also clears the users table and only works when `PLATFORM=dev`

Known gaps, not yet addressed:

- No automated tests
- The database connection is opened but never pinged or closed, and the server exits via `log.Fatal` rather than shutting down cleanly
- Request bodies aren't size-limited

## Architecture

- `main.go`: composition root that loads env, opens the DB, builds the mux, and starts the server.
- `config.go`: `Config`, the state shared across handlers (db handle, hit counter, platform).
- `health.go`, `users.go`, `login.go`, `chirps.go`, `metrics.go`, `reset.go`: one file per handler group.
- `internal/auth`: password hashing, JWT issuance/validation, refresh tokens, and bearer/API-key header parsing.
- `refresh.go`, `webhooks.go`: refresh/revoke handlers and the Polka webhook handler.
- `json.go`: `respondWithJSON` / `respondWithError`, the only place handlers write a response.
- `internal/database`: sqlc-generated query code. Regenerate with `sqlc generate`; never hand-edit.
- `sql/schema`, `sql/queries`: goose migrations and the queries sqlc compiles them from.
- `app/`: static frontend served at `/app/`.

## Running locally

Requires Go 1.27, Postgres, and [goose](https://github.com/pressly/goose) for migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`).

1. Set `DB_URL` and `PLATFORM` (a `.env` file is loaded automatically).
2. Apply migrations:

   ```sh
   goose -dir sql/schema postgres "$DB_URL" up
   ```

3. Start the server:

   ```sh
   go run .
   ```

The server listens on port 8080.

## Configuration

| Variable  | Purpose                                                              |
| --------- | --------------------------------------------------------------------|
| `DB_URL`  | Postgres connection string. Required; the server exits without it.  |
| `PLATFORM`| Set to `dev` to allow `POST /admin/reset`; any other value gets 403. |
| `JWT_TOKEN`| Secret key used to sign and validate access-token JWTs.             |
| `API_KEY` | Shared key the Polka webhook must present to upgrade a user.        |

## API

| Method | Path                     | Auth       | Description                                                        |
| ------ | ------------------------ | ---------- | -------------------------------------------------------------------|
| GET    | `/api/healthz`           | -          | Liveness check, returns `OK`                                       |
| GET    | `/api/chirps`            | -          | List chirps, oldest first; filter with `?author_id=`, reverse with `?sort=desc` |
| GET    | `/api/chirps/{chirpID}`  | -          | Fetch one chirp by UUID, 404 if missing                            |
| POST   | `/api/chirps`            | JWT        | Create a chirp; 140-char limit, filters 3 known words               |
| DELETE | `/api/chirps/{chirpID}`  | JWT        | Delete a chirp; 403 unless the caller is the chirp's author         |
| POST   | `/api/users`             | -          | Create a user; stores an Argon2id-hashed password                   |
| PUT    | `/api/users`             | JWT        | Update the authenticated user's email and password                 |
| POST   | `/api/login`             | -          | Verify email and password; returns the user, a JWT, and a refresh token |
| POST   | `/api/refresh`           | Refresh token | Exchange a valid refresh token for a new JWT                     |
| POST   | `/api/revoke`            | Refresh token | Revoke a refresh token                                            |
| POST   | `/api/polka/webhooks`    | API key    | Upgrade a user to Chirpy Red                                        |
| GET    | `/admin/metrics`         | -          | HTML page with the file-server hit count                            |
| POST   | `/admin/reset`           | -          | Dev only: clears users, resets hit count                            |

JWT and refresh-token auth are both passed as `Authorization: Bearer <token>`; the API key is passed as `Authorization: ApiKey <key>`.

Errors are a `{"error": "message"}` body with a matching 4xx/5xx status.

## Todo

- Complete the remaining course chapters.
- Add automated tests for the application's core behavior and edge cases.
- Document key design decisions with clear, focused code comments.
- Improve the documentation for clarity, completeness, and ease of use.
- Strengthen the repository and codebase through ongoing refactoring, cleanup, and quality improvements.
