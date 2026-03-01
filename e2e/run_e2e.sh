#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT/e2e/out"
mkdir -p "$OUT_DIR"

HAS_DOCKER=0
if [[ "${E2E_SKIP_DOCKER:-0}" != "1" ]] && command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  HAS_DOCKER=1
fi

if [[ "$HAS_DOCKER" -eq 1 ]]; then
  docker compose -f "$ROOT/e2e/docker-compose.yml" up -d --build
  trap 'docker compose -f "$ROOT/e2e/docker-compose.yml" down -v' EXIT
else
  echo "docker compose unavailable, running logical e2e checks only" >&2
fi

go test ./tests/integration -v

go run ./e2e/report/cmd/generate -out "$OUT_DIR" -total 4 -open 2 -closed 1 -timeout 1

echo "e2e report generated at $OUT_DIR"
