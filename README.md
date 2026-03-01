# Port Scan MK3

TCP port scanner CLI written in Go.

## Commands

- Validate inputs:
  - `go run ./cmd/port-scan validate -cidr-file <cidr.csv> -port-file <ports.csv> -format human`
  - `go run ./cmd/port-scan validate -cidr-file <cidr.csv> -port-file <ports.csv> -format json`
- Scan (pipeline wiring in place):
  - `go run ./cmd/port-scan scan -cidr-file <cidr.csv> -port-file <ports.csv>`

## Tests

- Unit + integration:
  - `go test ./...`
- Coverage gate:
  - `bash scripts/coverage_gate.sh`
- E2E (Docker optional):
  - `bash e2e/run_e2e.sh`
  - `E2E_SKIP_DOCKER=1 bash e2e/run_e2e.sh`

## Artifacts

- E2E report output:
  - `e2e/out/report.html`
  - `e2e/out/report.txt`
