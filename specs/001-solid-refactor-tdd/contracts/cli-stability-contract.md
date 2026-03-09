# CLI Stability Contract

## Protected Commands

The following commands are protected during this refactor:

- `port-scan validate`
- `port-scan scan`
- `port-scan --help`
- `port-scan -h`

No refactor increment may silently add, remove, or rename these command entrypoints.

## Protected Exit Codes

The refactor must preserve the currently observable exit-code behavior:

- `0`
  - help output
  - successful `validate`
  - successful `scan`
- `1`
  - invalid `validate` result where validation completed and returned `valid=false`
  - scan failure that is not a cancellation
- `2`
  - parse or usage errors, including unknown command and invalid flag/config input
- `130`
  - canceled scan when `scanapp.Run` returns `context.Canceled`

## Protected Output Semantics

### `validate`

- writes validation result to stdout using the requested `human` or `json` format
- writes parse/config/write errors to stderr
- supports custom CIDR column names through `-cidr-ip-col` and `-cidr-ip-cidr-col`

### `scan`

- writes batch results to timestamped output files
- keeps `scan_results-*.csv` and `opened_results-*.csv` behavior intact
- maps cancellation to `scan canceled` on stderr and exit code `130`
- preserves structured runtime visibility when JSON format is enabled
- preserves progress snapshots on stdout when progress reporting is enabled

### Help

- keeps usage text on stdout
- continues to advertise the current top-level commands and flag families

## Protected Operator Expectations

The refactor must not break these operator-facing expectations:

- the same valid inputs continue to produce a successful `validate`
- successful scans still emit usable main and open-only CSV artifacts
- canceled scans still produce resume state when the runtime conditions require it
- progress and completion visibility remain available for long-running scans
- `pkg/cli/output.go` continues to emit the existing `validate` human and JSON payload shapes

## Allowed Internal Changes

The following internal changes are allowed without being treated as contract changes:

- extracting helper functions or new collaborator files
- moving orchestration code out of `cmd/port-scan/main.go`
- moving scan sub-responsibilities out of `pkg/scanapp/scan.go`
- introducing narrow internal interfaces owned by the consuming package
- reorganizing tests to improve red/green coverage

## Disallowed Hidden Changes

The following changes require explicit spec and release-note treatment, not a silent refactor:

- changing exit-code semantics
- changing where errors or validation output are written
- changing file naming rules for scan artifacts
- removing required observability fields or lifecycle events
- changing resume behavior in a way operators would notice

## Baseline Test Mapping

- `cmd/port-scan/main_test.go`
  protects `validate` output format and custom column handling
- `cmd/port-scan/main_scan_test.go`
  protects `scan` exit behavior and output artifacts
- `tests/integration/scan_pipeline_test.go`
  protects pause/resume operator flow
- `tests/integration/resume_flow_test.go`
  protects resume correctness
- `pkg/scanapp/scan_observability_test.go`
  protects progress and completion observability payloads
- `pkg/scanapp/scan_test.go`
  protects resume save/skip decisions for interrupted vs. completed runs
