# Verification Quickstart

## Baseline Verification

Use these commands before structural changes to confirm the current contract is protected.

### CLI baseline

```bash
go test ./cmd/port-scan -run 'TestMainValidate_JSONOutput|TestMainValidate_CustomCIDRColumnNames|TestRunMain_ScanWritesCSV|TestRunMain_ScanWritesOpenedResultsCSV|TestScanApp_CancelSavesResumeState' -count=1
```

Expected:

- validate JSON output stays stable
- custom CIDR column names remain valid
- scan writes main and open-only CSV outputs
- canceled scans still save resume state

### Scanapp baseline

```bash
go test ./pkg/scanapp -run 'TestRun_WhenResumeStateFileProvided_ContinuesFromNextIndex|TestRun_WhenCanceledWithoutResumePath_SavesFallbackResumeState|TestRun_WhenPressureAPIFailsThreeTimes_ReturnsFatalErrorAndSavesResumeState|TestRun_WhenObservabilityJSONEnabled_EmitsProgressAndCompletionEvents' -count=1
```

Expected:

- resume restarts from the saved next index
- fallback resume path is written on cancellation
- pressure API fails hard on the third consecutive error
- progress and completion summary observability events remain present

### Integration baseline

```bash
go test ./tests/integration -run 'TestScanPipeline_PausesOnPressureAndResumes|TestResumeFlow_CompletesAllTargets' -count=1
```

Expected:

- scan pauses and resumes under pressure
- resume flow completes with zero duplicates and zero missing results

## US1 Verification

### Increment 1: Extract command handlers from `cmd/port-scan/main.go`

- Red command:
  - `go test ./cmd/port-scan -run 'TestHandleValidateCommand_WhenJSONValidationSucceeds_ReturnsExit0|TestHandleValidateCommand_WhenConfigParseFails_ReturnsExit2AndWritesStderr' -count=1`
- Red observation:
  - build failed with `undefined: handleValidateCommand`
- Green command:
  - `go test ./cmd/port-scan -run 'TestHandleValidateCommand_WhenJSONValidationSucceeds_ReturnsExit0|TestHandleValidateCommand_WhenConfigParseFails_ReturnsExit2AndWritesStderr' -count=1`
- Green observation:
  - helper-based validate tests passed
- Refactor note:
  - introduced `cmd/port-scan/command_handlers.go` and reduced `runMain` to command routing glue
- Post-refactor command:
  - `go test ./cmd/port-scan -count=1`
- Post-refactor observation:
  - all `cmd/port-scan` tests passed

### Increment 2: Define and extract scanapp preparation seams

- Red command:
  - `go test ./pkg/scanapp -run 'TestLoadRunInputs_WhenDependenciesInjected_UsesConfigPathsAndColumns|TestPrepareRunPlan_WhenDependenciesInjected_BuildsChunksRuntimesAndOutputPaths' -count=1`
- Red observation:
  - build failed with `undefined: runDependencies`, `undefined: loadRunInputs`, and `undefined: prepareRunPlan`
- Green command:
  - `go test ./pkg/scanapp -run 'TestLoadRunInputs_WhenDependenciesInjected_UsesConfigPathsAndColumns|TestPrepareRunPlan_WhenDependenciesInjected_BuildsChunksRuntimesAndOutputPaths' -count=1`
- Green observation:
  - seam tests passed after introducing preparation helpers
- Refactor note:
  - extracted input-loading and runtime-preparation seams into `pkg/scanapp/input_loader.go` and `pkg/scanapp/runtime_builder.go`
- Post-refactor command:
  - `go test ./pkg/scanapp ./cmd/port-scan ./tests/integration -count=1`
- Post-refactor observation:
  - all scanapp, command, and integration tests passed

### Increment 3: Strengthen US1 runtime and pipeline regressions

- Red command:
  - `go test ./pkg/scanapp -run TestRun_WhenCIDRColumnNamesBlank_UsesDefaultInputColumns -count=1`
  - `go test ./tests/integration -run TestScanPipeline_DefaultScenarioCompletesWithoutLoss -count=1`
- Red observation:
  - new tests were added to lock default-column behavior and baseline pipeline completion
- Green command:
  - `go test ./pkg/scanapp -run TestRun_WhenCIDRColumnNamesBlank_UsesDefaultInputColumns -count=1`
  - `go test ./tests/integration -run TestScanPipeline_DefaultScenarioCompletesWithoutLoss -count=1`
- Green observation:
  - both targeted regressions passed
- Refactor note:
  - no extra production change was needed beyond the extracted seams; the tests now guard the new boundary
- Post-refactor command:
  - `go test ./pkg/scanapp ./cmd/port-scan ./tests/integration -count=1`
- Post-refactor observation:
  - the full US1 protection set remained green

## US2 Verification

Record for each increment:

- Red command:
- Red observation:
- Green command:
- Green observation:
- Refactor note:
- Post-refactor command:
- Post-refactor observation:

## US3 Verification

Record for each increment:

- Red command:
- Red observation:
- Green command:
- Green observation:
- Refactor note:
- Post-refactor command:
- Post-refactor observation:

## Final Gates

```bash
go test ./...
bash scripts/coverage_gate.sh
bash e2e/run_e2e.sh
```

Record:

- `go test ./...` result:
- `bash scripts/coverage_gate.sh` result:
- `bash e2e/run_e2e.sh` result:
