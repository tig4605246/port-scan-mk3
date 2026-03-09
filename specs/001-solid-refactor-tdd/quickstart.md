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

### Increment 1: Protect dispatcher and pressure-monitor seams

- Red command:
  - `go test ./pkg/scanapp -run 'TestDispatchTasks_WhenRuntimeReady_EmitsTasksAndAdvancesNextIndex|TestPollPressureAPI_WhenJSONLoggerEnabled_EmitsPauseResumeMessages' -count=1`
- Red observation:
  - first run failed at compile time because the new seam-protection tests introduced a missing `ratelimit` import
- Green command:
  - `go test ./pkg/scanapp -run 'TestDispatchTasks_WhenRuntimeReady_EmitsTasksAndAdvancesNextIndex|TestPollPressureAPI_WhenJSONLoggerEnabled_EmitsPauseResumeMessages' -count=1`
- Green observation:
  - targeted dispatcher and pressure-monitor regression tests passed
- Refactor note:
  - baseline tests now guard queue emission, next-index advancement, and pause/resume log behavior before function extraction
- Post-refactor command:
  - `go test ./pkg/scanapp -run 'TestDispatchTasks_WhenRuntimeReady_EmitsTasksAndAdvancesNextIndex|TestPollPressureAPI_WhenJSONLoggerEnabled_EmitsPauseResumeMessages|TestPollPressureAPI_WhenPressureCrossesThreshold_TogglesPauseAndLogsTransition|TestStartManualPauseMonitor_WhenManualPauseChanges_LogsStateTransitions' -count=1`
- Post-refactor observation:
  - dispatcher and monitor protections remained green

### Increment 2: Extract dispatcher and pressure monitor files

- Red command:
  - `go test ./pkg/scanapp -run 'TestDispatchTasks_WhenRuntimeReady_EmitsTasksAndAdvancesNextIndex|TestPollPressureAPI_WhenJSONLoggerEnabled_EmitsPauseResumeMessages|TestPollPressureAPI_WhenPressureCrossesThreshold_TogglesPauseAndLogsTransition|TestStartManualPauseMonitor_WhenManualPauseChanges_LogsStateTransitions' -count=1`
- Red observation:
  - after file movement, compile failed because `scan.go` still referenced removed imports for `http` and `strconv`
- Green command:
  - `go test ./pkg/scanapp -run 'TestDispatchTasks_WhenRuntimeReady_EmitsTasksAndAdvancesNextIndex|TestPollPressureAPI_WhenJSONLoggerEnabled_EmitsPauseResumeMessages|TestPollPressureAPI_WhenPressureCrossesThreshold_TogglesPauseAndLogsTransition|TestStartManualPauseMonitor_WhenManualPauseChanges_LogsStateTransitions' -count=1`
- Green observation:
  - all targeted dispatcher and monitor tests passed after import cleanup
- Refactor note:
  - moved dispatch logic to `pkg/scanapp/task_dispatcher.go` and pause/pressure logic to `pkg/scanapp/pressure_monitor.go`
- Post-refactor command:
  - `go test ./pkg/scanapp ./cmd/port-scan ./tests/integration -count=1`
- Post-refactor observation:
  - scanapp, command, and integration suites all remained green

## US3 Verification

### Increment 1: Protect result aggregation and resume persistence seams

- Red command:
  - `go test ./pkg/scanapp -run 'TestEmitScanResultEvents_WhenProgressStepReached_EmitsProgressSnapshot|TestEmitCompletionSummary_WhenResultsMixed_EmitsOutcomeBreakdown|TestPersistResumeState_WhenRuntimeIncomplete_SavesResumeSnapshot|TestPersistResumeState_WhenRunCompletesCleanly_SkipsWrite' -count=1`
  - `go test ./tests/integration -run TestResumeFlow_WhenEnabled_PreservesOperatorVisibleOutcomes -count=1`
- Red observation:
  - the first `pkg/scanapp` run failed because the new completion-summary assertion assumed `error_cause=context_deadline_exceeded`, while the current observability contract emits `error_cause=deadline_exceeded`
- Green command:
  - `go test ./pkg/scanapp -run 'TestEmitScanResultEvents_WhenProgressStepReached_EmitsProgressSnapshot|TestEmitCompletionSummary_WhenResultsMixed_EmitsOutcomeBreakdown|TestPersistResumeState_WhenRuntimeIncomplete_SavesResumeSnapshot|TestPersistResumeState_WhenRunCompletesCleanly_SkipsWrite' -count=1`
  - `go test ./tests/integration -run TestResumeFlow_WhenEnabled_PreservesOperatorVisibleOutcomes -count=1`
- Green observation:
  - helper-focused observability and resume tests passed, and the resume integration outcome stayed duplicate-free and missing-free
- Refactor note:
  - extracted result writing, progress/completion emission, and resume persistence into `pkg/scanapp/result_aggregator.go` and `pkg/scanapp/resume_manager.go`, leaving `scan.go` as orchestration glue
- Post-refactor command:
  - `go test ./pkg/scanapp ./cmd/port-scan ./tests/integration -count=1`
- Post-refactor observation:
  - scanapp, command, and integration suites all remained green after the US3 extraction

## Final Gates

```bash
go test ./...
bash scripts/coverage_gate.sh
bash e2e/run_e2e.sh
```

Record:

- `go test ./...` result:
  - PASS on 2026-03-09; `cmd/port-scan`, `pkg/scanapp`, and `tests/integration` all remained green in the full-repo run
- `bash scripts/coverage_gate.sh` result:
  - PASS on 2026-03-09; coverage gate reported `86.8%`
- `bash e2e/run_e2e.sh` result:
  - PASS on 2026-03-09; e2e integration suite passed and report output was generated under `e2e/out`

## Quickstart Validation

- Verified on 2026-03-09 that every command recorded in this quickstart maps to an existing test or
  executable gate in the branch
- Replayed the current branch gates (`go test ./...`, coverage gate, and e2e gate) after the final
  polish cleanup to confirm the recorded evidence still reflects the checked-in code
