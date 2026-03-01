#!/usr/bin/env sh
set -eu
while true; do
  nc -l -p 8080 >/dev/null 2>&1 || true
done
