#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/docs"

PORT="${1:-8080}"

echo "Serving chamadoFront on 0.0.0.0:${PORT} (reachable from other machines on this port)"
exec python3 -m http.server "$PORT" --bind 0.0.0.0
