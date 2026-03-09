# Runtime Boundaries

## Current Mixed Responsibilities

`pkg/scanapp/scan.go` currently mixes these runtime concerns:

- input loading from files
- chunk/runtime construction
- worker startup and shutdown
- task dispatch and pacing
- pause and pressure monitoring
- result aggregation and progress reporting
- resume save/load coordination
- logger construction and event emission
- helper parsing and utility behavior

`cmd/port-scan/main.go` currently mixes these CLI concerns:

- command dispatch
- parse and validate entry wiring
- direct input validation execution
- scan cancellation wiring
- exit-code translation

## Target Boundaries

### CLI Composition Root

Owns:

- command selection
- config parsing handoff
- stdout/stderr wiring
- exit-code mapping

Must not own:

- CIDR/port file parsing rules
- scan orchestration internals
- resume policy
- pressure or worker lifecycle logic

### Input Loader

Owns:

- opening CIDR and port files
- delegating parsing to `pkg/input`
- returning parsed records or file/parse errors

### Runtime Builder

Owns:

- chunk loading or construction
- runtime target grouping
- pre-dispatch runtime state assembly

### Task Dispatcher

Owns:

- iterating chunk indexes
- respecting gate and rate-limit constraints
- emitting work items into the worker queue

### Pressure Monitor

Owns:

- manual pause monitoring
- pressure API polling
- fatal pressure failure cutoff policy

### Result Aggregator

Owns:

- writing result records
- counting open/close/timeout outcomes
- progress emission
- completion summary emission

### Resume Manager

Owns:

- deciding whether resume state must be saved
- selecting the effective resume path
- persisting chunk state snapshots

## Dependency Direction

The desired direction for the first refactor wave is:

- `cmd/port-scan` -> `pkg/config`, `pkg/cli`, `pkg/scanapp`, `pkg/input`, `pkg/state`
- `pkg/scanapp` -> `pkg/input`, `pkg/task`, `pkg/state`, `pkg/scanner`, `pkg/writer`, `pkg/speedctrl`, `pkg/ratelimit`, `pkg/logx`
- extracted scanapp collaborators -> lower-level packages only; they must not depend back on `cmd/port-scan`

Rules:

- consumers define narrow seams when a seam is needed
- avoid shared mega-interfaces that mix unrelated runtime responsibilities
- new collaborator files may depend on `scanapp` local types, but not on CLI glue

## Boundary Ownership Rules

- if only one caller needs a seam, the caller owns the seam
- prefer concrete collaborators over interfaces until substitution is actually needed
- do not create a collaborator whose only job is to forward to another function
- a new file must reduce responsibility concentration, not just move lines around

## First-Wave File Targets

The first extraction wave is expected to land in:

- `cmd/port-scan/command_handlers.go`
- `pkg/scanapp/input_loader.go`
- `pkg/scanapp/runtime_builder.go`
- `pkg/scanapp/task_dispatcher.go`
- `pkg/scanapp/pressure_monitor.go`
- `pkg/scanapp/result_aggregator.go`
- `pkg/scanapp/resume_manager.go`

## Review Questions

- Did the change reduce the set of reasons a file can change?
- Did it remove CLI knowledge from runtime orchestration or vice versa?
- Did it introduce a narrow seam with a clear owner?
- Did it leave progress, completion, and resume behavior observable and testable?
