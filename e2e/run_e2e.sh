#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT/e2e/out"
INPUT_DIR="$ROOT/e2e/inputs"
COMPOSE_FILE="$ROOT/e2e/docker-compose.yml"
mkdir -p "$OUT_DIR"
mkdir -p "$INPUT_DIR"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for e2e test" >&2
  exit 1
fi
if ! docker compose version >/dev/null 2>&1; then
  echo "docker compose is required for e2e test" >&2
  exit 1
fi

rm -f "$OUT_DIR/scan_results.csv" "$OUT_DIR/report.html" "$OUT_DIR/report.txt"

cat > "$INPUT_DIR/cidr.csv" <<'EOF'
fab_name,cidr,cidr_name
fab-open,172.28.0.10/32,mock-target-open
fab-closed,172.28.0.11/32,mock-target-closed
EOF

cat > "$INPUT_DIR/ports.csv" <<'EOF'
8080/tcp
EOF

docker compose -f "$COMPOSE_FILE" down -v --remove-orphans >/dev/null 2>&1 || true
docker compose -f "$COMPOSE_FILE" up -d --build mock-target-open mock-target-closed
trap 'docker compose -f "$COMPOSE_FILE" down -v --remove-orphans' EXIT

OPEN_READY=0
for _ in {1..30}; do
  if docker compose -f "$COMPOSE_FILE" exec -T mock-target-open sh -lc "netstat -lnt | grep -q ':8080'" >/dev/null 2>&1; then
    OPEN_READY=1
    break
  fi
  sleep 1
done
if [[ "$OPEN_READY" -ne 1 ]]; then
  echo "mock-target-open did not become ready on port 8080" >&2
  exit 1
fi

docker compose -f "$COMPOSE_FILE" run --rm scanner scan \
  -cidr-file /inputs/cidr.csv \
  -port-file /inputs/ports.csv \
  -output /out/scan_results.csv \
  -workers 2 \
  -delay 0ms \
  -timeout 200ms \
  -disable-api=true \
  -log-level error

go test ./tests/integration -v

go run ./e2e/report/cmd/generate -out "$OUT_DIR" -csv "$OUT_DIR/scan_results.csv"

OPEN_COUNT=$(awk -F= '/^Open=/{print $2}' "$OUT_DIR/report.txt")
CLOSED_COUNT=$(awk -F= '/^Closed=/{print $2}' "$OUT_DIR/report.txt")
TIMEOUT_COUNT=$(awk -F= '/^Timeout=/{print $2}' "$OUT_DIR/report.txt")

if [[ "${OPEN_COUNT:-0}" -lt 1 ]]; then
  echo "e2e assertion failed: expected at least 1 open result" >&2
  exit 1
fi
if [[ $(( ${CLOSED_COUNT:-0} + ${TIMEOUT_COUNT:-0} )) -lt 1 ]]; then
  echo "e2e assertion failed: expected at least 1 non-open result" >&2
  exit 1
fi

echo "e2e report generated at $OUT_DIR"
