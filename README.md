# Chirpy

A small HTTP server written in Go, built while following [boot.dev's Learn HTTP Servers in Go course](https://www.boot.dev/courses/learn-http-servers-golang).

Chirpy is a simplified Twitter clone. It serves a static frontend, exposes a JSON API, and stores data in Postgres using queries generated with [sqlc](https://sqlc.dev/).

## Status

The project is incomplete and still follows the course chapters in order.

Implemented:

- Static file server for the frontend, mounted at `/app/`
- `GET /api/healthz`, a health check endpoint
- `POST /api/validate_chirp`, validates a chirp's content
- `GET /admin/metrics` and `POST /admin/reset`, track and reset file server request counts
- Users table schema and generated database queries

Not implemented yet: chirp creation and storage, authentication, and the remaining course endpoints.

## Running locally

Requires Go and a Postgres database.

1. Set the `DB_URL` environment variable to your Postgres connection string (see `.env`).
2. Start the server:

   ```sh
   go run .
   ```

The server listens on port 8080.
