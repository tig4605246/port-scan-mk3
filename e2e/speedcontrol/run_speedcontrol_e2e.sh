#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="$ROOT/e2e/out/speedcontrol"

mkdir -p "$OUT_DIR"
rm -f "$OUT_DIR"/report.md "$OUT_DIR"/report.html "$OUT_DIR"/raw_metrics.json

go test ./internal/testkit/speedcontrol ./tests/integration -run 'Analyze|Collector|SpeedControl' -count=1
go run ./e2e/report/cmd/generate-speedcontrol -out "$OUT_DIR"

echo "speed-control report generated at $OUT_DIR"

