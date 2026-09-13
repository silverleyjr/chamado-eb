#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f .env ]; then
	cp .env.example .env
	echo "Created .env from .env.example — edit it if you need different values, then re-run."
fi

set -a
source .env
set +a

echo "Starting postgres container..."
docker compose up -d

echo "Waiting for postgres to become healthy..."
until [ "$(docker inspect -f '{{.State.Health.Status}}' chamado-postgres 2>/dev/null)" = "healthy" ]; do
	sleep 1
done
echo "Postgres is ready."

echo "Starting API server on ${SERVER_PORT:-:8083}..."
cd chamadoApi
exec go run ./cmd/api
