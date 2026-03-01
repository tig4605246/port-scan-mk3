#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

go test ./... -coverprofile=coverage.out
COVER=$(go tool cover -func=coverage.out | awk '/total:/ {print substr($3, 1, length($3)-1)}')
awk -v c="$COVER" 'BEGIN { if (c+0 < 85) { exit 1 } }'
echo "coverage gate passed: ${COVER}%"
