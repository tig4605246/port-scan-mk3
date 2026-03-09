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

Record for each increment:

- Red command:
- Red observation:
- Green command:
- Green observation:
- Refactor note:
- Post-refactor command:
- Post-refactor observation:

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
