#!/usr/bin/env bash
set -euo pipefail
set -m   # job control: each background job gets its own process group,
         # so cleanup() can kill `go run`'s forked child too, not just the launcher.
cd "$(dirname "$0")"

SERVER_DIR="chamadoServer"
FRONT_DIR="chamadoFront"
FRONT_PORT="${FRONT_PORT:-8080}"

if [ ! -f "$SERVER_DIR/.env" ]; then
	cp "$SERVER_DIR/.env.example" "$SERVER_DIR/.env"
	echo "Created $SERVER_DIR/.env from .env.example — edit it (passwords, JWT secret) before exposing this to the internet."
fi

set -a
source "$SERVER_DIR/.env"
set +a

echo "==> Starting Postgres..."
(cd "$SERVER_DIR" && docker compose up -d)

echo "==> Waiting for Postgres to become healthy..."
until [ "$(docker inspect -f '{{.State.Health.Status}}' chamado-postgres 2>/dev/null)" = "healthy" ]; do
	sleep 1
done
echo "Postgres is ready."

PIDS=()
cleanup() {
	echo
	echo "==> Shutting down..."
	for pid in "${PIDS[@]}"; do
		# `go run` forks a separate compiled-binary child that survives killing
		# just the launcher, so kill direct children explicitly too. The
		# negative-PID form (process group, from `set -m`) is a second layer
		# that also works when a job spawns a whole subtree.
		for child in $(pgrep -P "$pid" 2>/dev/null); do
			kill "$child" 2>/dev/null || true
		done
		kill -- "-$pid" 2>/dev/null || true
		kill "$pid" 2>/dev/null || true
	done
}
trap cleanup EXIT INT TERM

echo "==> Starting API server on ${SERVER_PORT:-:8083}..."
(cd "$SERVER_DIR/chamadoApi" && exec go run ./cmd/api) &
PIDS+=($!)

API_PORT="${SERVER_PORT:-:8083}"
API_PORT="${API_PORT##*:}"

echo "==> Waiting for the API to be ready..."
until curl -s -o /dev/null "http://localhost:${API_PORT}/login"; do
	sleep 1
done

echo "==> Ensuring a default admin account exists..."
sql_escape() { printf '%s' "$1" | sed "s/'/''/g"; }
ADMIN_NAME_SQL="$(sql_escape "${ADMIN_NAME:-Admin Geral}")"
ADMIN_PASSWORD_SQL="$(sql_escape "${ADMIN_PASSWORD:-senha123}")"
ADMIN_TEAM_SQL="$(sql_escape "${ADMIN_TEAM:-TLM}")"
docker exec chamado-postgres psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-chamado-db}" \
	-c "INSERT INTO usuarios (name, password, team, role) VALUES ('${ADMIN_NAME_SQL}', '${ADMIN_PASSWORD_SQL}', '${ADMIN_TEAM_SQL}', 'admin') ON CONFLICT (name) DO NOTHING;" \
	> /dev/null

echo "==> Starting frontend on 0.0.0.0:${FRONT_PORT}..."
"./$FRONT_DIR/serve.sh" "$FRONT_PORT" &
PIDS+=($!)

IP="$(hostname -I 2>/dev/null | awk '{print $1}')"

echo
echo "======================================================"
echo " Chamado is running."
echo "   Local:   http://localhost:${FRONT_PORT}/index.html"
if [ -n "$IP" ]; then
	echo "   Network: http://${IP}:${FRONT_PORT}/index.html"
fi
echo " API listening on port ${SERVER_PORT:-:8083}"
echo " Admin login: ${ADMIN_NAME:-Admin Geral} / ${ADMIN_PASSWORD:-senha123}"
echo " Press Ctrl+C to stop everything."
echo "======================================================"
echo

wait
