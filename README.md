# Port Scan MK3

TCP port scanner CLI written in Go.

## Commands

- Validate inputs:
  - `go run ./cmd/port-scan validate -cidr-file <cidr.csv> -port-file <ports.csv> -format human`
  - `go run ./cmd/port-scan validate -cidr-file <cidr.csv> -port-file <ports.csv> -format json`
- Scan (pipeline wiring in place):
  - `go run ./cmd/port-scan scan -cidr-file <cidr.csv> -port-file <ports.csv>`
  - `go run ./cmd/port-scan scan -cidr-file <cidr.csv> -port-file <ports.csv> -cidr-ip-col source_ip -cidr-ip-cidr-col source_cidr`
  - `go run ./cmd/port-scan scan -cidr-file <cidr.csv> -port-file <ports.csv> -output out/scan_results.csv -resume out/resume_state.json`

## Output Contract

- Each scan run writes a timestamped batch pair:
  - `scan_results-YYYYMMDDTHHMMSSZ[-n].csv`
  - `opened_results-YYYYMMDDTHHMMSSZ[-n].csv`
- `opened_results-*` contains only `open` rows and keeps the same header as the main output.
- If started multiple times within the same second, `-n` is an increasing positive integer shared by both files.

## Resume Rules

- If `-resume <path>` is provided, state is loaded from and saved back to the same path.
- If `-resume` is omitted, fallback path is `<output-dir>/resume_state.json`.
- On cancellation or runtime failure, current state is persisted for resume.

## Tests

- Unit + integration:
  - `go test ./...`
- Coverage gate:
  - `bash scripts/coverage_gate.sh`
- E2E (Docker required):
  - `bash e2e/run_e2e.sh`

## Artifacts

- E2E report output:
  - `e2e/out/report.html`
  - `e2e/out/report.txt`
