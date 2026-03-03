# Quickstart: IP-Aware Baseline Spec

## Prerequisites

- Go 1.24+
- Docker + Docker Compose (for e2e)
- Repository root: `/Users/xuxiping/tsmc/port-scan-mk3`

## 1) Validate Inputs

```bash
cd /Users/xuxiping/tsmc/port-scan-mk3
go run ./cmd/port-scan validate \
  -cidr-file e2e/inputs/cidr_normal.csv \
  -port-file e2e/inputs/ports.csv \
  -cidr-ip-col source_ip \
  -cidr-ip-cidr-col source_cidr \
  -format json
```

Expected:
- Exit code `0`
- JSON output contains `"valid": true`

## 2) Run Scan

```bash
go run ./cmd/port-scan scan \
  -cidr-file e2e/inputs/cidr_normal.csv \
  -port-file e2e/inputs/ports.csv \
  -output e2e/out/scan_results.csv \
  -cidr-ip-col source_ip \
  -cidr-ip-cidr-col source_cidr \
  -pressure-api http://localhost:8080/api/pressure \
  -pressure-interval 5s \
  -format human
```

Expected artifacts in `e2e/out/` (timestamped batch):
- `scan_results-YYYYMMDDTHHMMSSZ[-n].csv`
- `opened_results-YYYYMMDDTHHMMSSZ[-n].csv`

## 3) Resume Behavior

With explicit resume path:

```bash
go run ./cmd/port-scan scan \
  -cidr-file e2e/inputs/cidr_normal.csv \
  -port-file e2e/inputs/ports.csv \
  -output e2e/out/scan_results.csv \
  -resume e2e/out/resume_state.json
```

Rules:
- Provided `-resume` path is used for load + subsequent saves.
- Without `-resume`, default save path is `<output-dir>/resume_state.json`.

## 4) Execute Test Gates

```bash
go test ./...
bash scripts/coverage_gate.sh
bash e2e/run_e2e.sh
```

Expected:
- Unit + integration pass
- Coverage gate > 85%
- e2e scenarios all pass (normal, API 5xx, API timeout/connection failure)
