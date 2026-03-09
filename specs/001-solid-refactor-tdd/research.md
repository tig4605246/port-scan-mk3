# SOLID Refactor Research

## Current Hotspots

### `cmd/port-scan/main.go`

Current concerns mixed in one file:

- command routing for `validate` and `scan`
- config parsing and exit-code mapping
- direct input validation flow for `validate`
- scan startup wiring and cancellation handling
- user-facing usage text

Why this is a hotspot:

- small CLI changes currently require touching command dispatch and validation flow together
- composition-root responsibilities are not isolated from operator behavior rules
- later tests for exit codes and stderr/stdout behavior have no dedicated seam

### `pkg/scanapp/scan.go`

Current concerns mixed in one file:

- input loading (`readCIDRFile`, `readPortFile`)
- resume/chunk construction (`loadOrBuildChunks`)
- runtime assembly (`buildRuntime`, group builders, helper parsing)
- worker lifecycle and task dispatch
- pause and pressure monitoring
- result aggregation and progress reporting
- resume persistence and completion handling
- logger construction and observability event emission
- low-level utility helpers

Why this is a hotspot:

- one file carries most orchestration risk for `scan`
- behavioral and structural changes are hard to separate in review
- extracting one responsibility currently risks touching unrelated concerns

## Existing Contract Evidence

The current tests already reveal protected behavior that must stay stable during refactor.

### CLI and command behavior

- `cmd/port-scan/main_test.go`
  - `validate` supports `json` output
  - custom CIDR column names remain accepted
- `cmd/port-scan/main_scan_test.go`
  - `scan` returns exit code `0` on success
  - timestamped `scan_results-*.csv` and `opened_results-*.csv` are written
  - cancellation saves resume state

### Runtime and observability behavior

- `pkg/scanapp/scan_test.go`
  - resume restarts from `NextIndex`
  - fallback resume path is written when canceled
  - pressure API fails hard on the third consecutive error
- `pkg/scanapp/scan_observability_test.go`
  - JSON logging must include `target`, `port`, `state_transition`, `error_cause`
  - progress and completion summary events are mandatory

### Integration behavior

- `tests/integration/scan_pipeline_test.go`
  - scan pauses on pressure and resumes to completion
- `tests/integration/resume_flow_test.go`
  - resume flow completes all targets with zero duplicates and zero missing results

## Refactor Order

The first wave stays intentionally narrow.

1. Document protected contracts and target boundaries (`T001-T006`)
2. Strengthen baseline tests around existing behavior (`T007-T014`)
3. Extract composition-root helpers from `cmd/port-scan/main.go` (`US1`)
4. Extract runtime-building seams from `pkg/scanapp/scan.go` (`US1`)
5. Extract dispatch and pressure seams only after baseline protection is in place (`US2`)
6. Extract result aggregation and resume persistence after the orchestration seams exist (`US3`)

## Deferred Areas

These are intentionally out of the first refactor wave unless a direct seam requires them.

- `pkg/scanner/`
  - currently focused and already covered by dedicated tests
- `pkg/state/`
  - keep the persistence contract stable; only touch through orchestration seams
- `pkg/writer/`
  - protect via existing output behavior, not a writer redesign
- `pkg/config/`
  - parsing rules are stable; the first wave should consume config, not redesign it

## Architectural Decisions

- Prefer small, consumer-owned seams over broad shared interfaces.
- Do not introduce a generic "service layer" unless a clear consumer needs the abstraction.
- Keep `cmd/port-scan` as a composition root only; no new domain logic belongs there.
- Keep `pkg/scanapp` as the orchestration package, but break orchestration into narrowly scoped
  collaborators instead of one monolithic file.

## Review Heuristics

Use these checks during each refactor increment:

- Does the changed unit have one primary reason to change?
- Did the increment protect a known contract before restructuring code?
- Did the refactor add a seam that reduced coupling, or just move complexity around?
- Can a reviewer understand the change without mentally simulating the entire scan flow?
