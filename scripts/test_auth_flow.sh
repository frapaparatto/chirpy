#!/usr/bin/env bash
#
# test_auth_flow.sh — walk through the JWT/refresh-token lifecycle against a
# local chirpy server, printing full request/response headers at each step.
#
# Flow:
#   1. POST /api/users    - create the user
#   2. POST /api/login    - log in, get an access token (JWT) + refresh token
#   3. POST /api/chirps   - create a chirp using the access token
#   4. POST /api/refresh  - exchange the refresh token for a new access token
#   5. POST /api/chirps   - create a chirp using the *refreshed* access token
#   6. POST /api/revoke   - revoke the refresh token
#   7. POST /api/refresh  - confirm the revoked refresh token is rejected
#
# Requires: curl, jq
#
# Usage: ./scripts/test_auth_flow.sh
#   BASE_URL=http://localhost:8080 EMAIL=walt@breakingbad.com PASSWORD=123456 ./scripts/test_auth_flow.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="${EMAIL:-walt@breakingbad.com}"
PASSWORD="${PASSWORD:-123456}"
CHIRP_BODY="${CHIRP_BODY:-Hello world, this is a test chirp}"

divider() { printf '\n\033[1;36m==== %s ====\033[0m\n' "$1"; }

# curl -v mixes verbose logging (headers) with the response body on the same
# stderr/stdout streams in a way that's awkward to separate cleanly, so we do
# it in two passes: one with -v (headers only, discard body) and one plain
# (body only, for jq). Sets $RESPONSE_BODY for the caller to inspect/parse.
run_step() {
    local title="$1"; shift
    divider "$title"

    echo "--- headers (request + response) ---"
    curl -sS -v -o /dev/null "$@" 2>&1 | grep -E '^(> |< |\* Connected|\* Trying)' || true

    echo
    echo "--- body ---"
    RESPONSE_BODY="$(curl -sS "$@")"
    echo "$RESPONSE_BODY" | jq . 2>/dev/null || echo "$RESPONSE_BODY"
}

command -v jq >/dev/null || { echo "jq is required but not installed" >&2; exit 1; }

# 1. Create user
run_step "1. POST /api/users (create user)" \
    -X POST "$BASE_URL/api/users" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"

# 2. Login
run_step "2. POST /api/login (log in)" \
    -X POST "$BASE_URL/api/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"

ACCESS_TOKEN="$(echo "$RESPONSE_BODY" | jq -r '.token')"
REFRESH_TOKEN="$(echo "$RESPONSE_BODY" | jq -r '.refresh_token')"

if [[ -z "$ACCESS_TOKEN" || "$ACCESS_TOKEN" == "null" ]]; then
    echo "Login failed, no access token received. Aborting." >&2
    exit 1
fi

echo
echo "Access token:  $ACCESS_TOKEN"
echo "Refresh token: $REFRESH_TOKEN"

# 3. Create a chirp with the access token
run_step "3. POST /api/chirps (create chirp with access token)" \
    -X POST "$BASE_URL/api/chirps" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{\"body\":\"$CHIRP_BODY\"}"

# 4. Refresh the access token
run_step "4. POST /api/refresh (get new access token)" \
    -X POST "$BASE_URL/api/refresh" \
    -H "Authorization: Bearer $REFRESH_TOKEN"

NEW_ACCESS_TOKEN="$(echo "$RESPONSE_BODY" | jq -r '.token')"
echo
echo "New access token: $NEW_ACCESS_TOKEN"

# 5. Create another chirp with the refreshed access token
run_step "5. POST /api/chirps (create chirp with refreshed access token)" \
    -X POST "$BASE_URL/api/chirps" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
    -d "{\"body\":\"$CHIRP_BODY (round 2, via refreshed token)\"}"

# 6. Revoke the refresh token
run_step "6. POST /api/revoke (revoke refresh token)" \
    -X POST "$BASE_URL/api/revoke" \
    -H "Authorization: Bearer $REFRESH_TOKEN"

# 7. Confirm the revoked refresh token can no longer be used
run_step "7. POST /api/refresh (expect 401, token was revoked)" \
    -X POST "$BASE_URL/api/refresh" \
    -H "Authorization: Bearer $REFRESH_TOKEN"

divider "Done"
