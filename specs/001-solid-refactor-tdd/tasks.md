# Tasks: SOLID Refactor and TDD Enforcement

**Input**: Design documents from `/specs/001-solid-refactor-tdd/`
**Prerequisites**: [plan.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/plan.md) (required), [spec.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/spec.md) (required for user stories), research.md, data-model.md, contracts/

**Tests**: Test tasks are REQUIRED by constitution and by this feature spec. Every story starts with
failing tests, and changed runtime contracts must be protected by unit, integration, and e2e
verification where applicable.

**Architecture**: Generated tasks MUST preserve SOLID boundaries. Each refactor increment must make
responsibility splits, interface ownership, and dependency direction explicit in code or feature
artifacts before merge.

**Organization**: Tasks are grouped by user story so each story can be implemented, validated, and
reviewed as an independent increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Every task includes an exact file path

## Path Conventions

- Single Go project rooted at `cmd/`, `pkg/`, `tests/`, `docs/`, and `specs/`
- Runtime refactor work is concentrated in `cmd/port-scan/` and `pkg/scanapp/`
- Cross-boundary regression protection stays in `tests/integration/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the feature-local artifacts that define refactor scope, test evidence rules,
and operator-facing contracts before production changes begin

- [ ] T001 Capture refactor decisions and hotspots in `specs/001-solid-refactor-tdd/research.md`
- [ ] T002 [P] Define CLI stability expectations in `specs/001-solid-refactor-tdd/contracts/cli-stability-contract.md`
- [ ] T003 [P] Define red/green/refactor evidence requirements in `specs/001-solid-refactor-tdd/contracts/tdd-evidence-contract.md`
- [ ] T004 [P] Define runtime responsibility boundaries in `specs/001-solid-refactor-tdd/contracts/runtime-boundaries.md`
- [ ] T005 Create refactor increment data model in `specs/001-solid-refactor-tdd/data-model.md`
- [ ] T006 Create verification quickstart for validate/scan/refactor evidence in `specs/001-solid-refactor-tdd/quickstart.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish the baseline regression net and shared seams that every user story depends on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T007 [P] Add baseline validate command regression coverage in `cmd/port-scan/main_test.go`
- [ ] T008 [P] Add baseline scan command regression coverage in `cmd/port-scan/main_scan_test.go`
- [ ] T009 [P] Add orchestration helper coverage for refactor seams in `pkg/scanapp/scan_helpers_test.go`
- [ ] T010 [P] Add scan orchestration baseline regression coverage in `pkg/scanapp/scan_test.go`
- [ ] T011 [P] Add observability baseline regression coverage in `pkg/scanapp/scan_observability_test.go`
- [ ] T012 [P] Add resume and pipeline baseline regression coverage in `tests/integration/resume_flow_test.go`
- [ ] T013 [P] Add validate/scan pipeline baseline regression coverage in `tests/integration/scan_pipeline_test.go`
- [ ] T014 Define shared scanapp collaborator seams in `pkg/scanapp/scan.go`

**Checkpoint**: Baseline contract and regression net are ready; story work can now proceed safely

---

## Phase 3: User Story 1 - Clarify Responsibility Boundaries (Priority: P1) 🎯 MVP

**Goal**: Split the current command entry and scan orchestration flow into smaller, single-purpose
units without changing operator-visible behavior

**Independent Test**: Run focused command and orchestration tests to prove `validate`/`scan`
behavior is stable while `cmd/port-scan/main.go` and `pkg/scanapp/scan.go` delegate to extracted
collaborators

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T015 [P] [US1] Add failing composition-root regression cases in `cmd/port-scan/main_test.go`
- [ ] T016 [P] [US1] Add failing runtime-build regression cases in `pkg/scanapp/scan_test.go`
- [ ] T017 [P] [US1] Add failing runtime-boundary integration coverage in `tests/integration/scan_pipeline_test.go`

### Implementation for User Story 1

- [ ] T018 [P] [US1] Extract command composition helpers into `cmd/port-scan/command_handlers.go`
- [ ] T019 [P] [US1] Extract input loading responsibilities into `pkg/scanapp/input_loader.go`
- [ ] T020 [P] [US1] Extract runtime building responsibilities into `pkg/scanapp/runtime_builder.go`
- [ ] T021 [US1] Refactor orchestration entrypoint delegation in `pkg/scanapp/scan.go`
- [ ] T022 [US1] Reduce `runMain` and `runScan` to composition glue in `cmd/port-scan/main.go`
- [ ] T023 [US1] Record US1 red/green/refactor evidence in `specs/001-solid-refactor-tdd/quickstart.md`

**Checkpoint**: User Story 1 is independently testable and provides the MVP structural cleanup

---

## Phase 4: User Story 2 - Enforce Test-First Change Delivery (Priority: P2)

**Goal**: Make refactor increments reviewable through explicit failing-test evidence and narrower
workflow seams around dispatch and pause/pressure coordination

**Independent Test**: Verify one full refactor increment from failing test to passing test to
cleanup, and confirm dispatch/pause behavior remains intact after seam extraction

### Tests for User Story 2 ⚠️

- [ ] T024 [P] [US2] Add failing dispatch-seam regression cases in `pkg/scanapp/scan_test.go`
- [ ] T025 [P] [US2] Add failing pause and pressure coordination cases in `pkg/scanapp/scan_observability_test.go`
- [ ] T026 [P] [US2] Add failing dispatcher integration coverage in `tests/integration/scan_pipeline_test.go`

### Implementation for User Story 2

- [ ] T027 [P] [US2] Extract task dispatch coordination into `pkg/scanapp/task_dispatcher.go`
- [ ] T028 [P] [US2] Extract pause and pressure monitoring into `pkg/scanapp/pressure_monitor.go`
- [ ] T029 [US2] Refactor orchestration flow to use extracted dispatcher and monitors in `pkg/scanapp/scan.go`
- [ ] T030 [US2] Encode reviewer evidence workflow in `specs/001-solid-refactor-tdd/contracts/tdd-evidence-contract.md`
- [ ] T031 [US2] Update runtime seam ownership notes in `specs/001-solid-refactor-tdd/contracts/runtime-boundaries.md`
- [ ] T032 [US2] Record US2 red/green/refactor evidence in `specs/001-solid-refactor-tdd/quickstart.md`

**Checkpoint**: User Story 2 is independently testable and reviewers can verify strict TDD flow

---

## Phase 5: User Story 3 - Preserve Operational Confidence During Refactor (Priority: P3)

**Goal**: Keep progress, completion, output writing, and resume flows trustworthy after the
orchestration refactor

**Independent Test**: Exercise scan, observability, and resume flows through targeted unit and
integration tests, then confirm end-to-end verification still reflects operator expectations

### Tests for User Story 3 ⚠️

- [ ] T033 [P] [US3] Add failing result aggregation and progress cases in `pkg/scanapp/scan_observability_test.go`
- [ ] T034 [P] [US3] Add failing resume persistence cases in `pkg/scanapp/scan_test.go`
- [ ] T035 [P] [US3] Add failing resume/output integration coverage in `tests/integration/resume_flow_test.go`

### Implementation for User Story 3

- [ ] T036 [P] [US3] Extract result aggregation and progress reporting into `pkg/scanapp/result_aggregator.go`
- [ ] T037 [P] [US3] Extract resume persistence handling into `pkg/scanapp/resume_manager.go`
- [ ] T038 [US3] Refactor orchestration completion and recovery flow in `pkg/scanapp/scan.go`
- [ ] T039 [US3] Preserve operator-facing output and visibility hooks in `pkg/cli/output.go`
- [ ] T040 [US3] Update CLI stability and runtime contract docs in `specs/001-solid-refactor-tdd/contracts/cli-stability-contract.md`
- [ ] T041 [US3] Record US3 red/green/refactor evidence in `specs/001-solid-refactor-tdd/quickstart.md`

**Checkpoint**: All user stories are independently testable and operator confidence is preserved

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, and cleanup that affect multiple user stories

- [ ] T042 [P] Refresh feature research and design notes in `specs/001-solid-refactor-tdd/research.md`
- [ ] T043 [P] Review SOLID boundary cleanup across `cmd/port-scan/main.go` and `pkg/scanapp/scan.go`
- [ ] T044 [P] Extend cross-story regression coverage in `tests/integration/scenario_extra_test.go`
- [ ] T045 Document refactor guidance for contributors in `README.md`
- [ ] T046 Run `go test ./...` and record evidence in `specs/001-solid-refactor-tdd/quickstart.md`
- [ ] T047 Run `bash scripts/coverage_gate.sh` and record evidence in `specs/001-solid-refactor-tdd/quickstart.md`
- [ ] T048 Run `bash e2e/run_e2e.sh` and record evidence in `specs/001-solid-refactor-tdd/quickstart.md`
- [ ] T049 Update release impact notes in `docs/release-notes/1.2.0.md`
- [ ] T050 Run quickstart validation in `specs/001-solid-refactor-tdd/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; begin immediately
- **Foundational (Phase 2)**: Depends on Setup; blocks all story work
- **User Story 1 (Phase 3)**: Depends on Foundational; establishes the first safe refactor seam
- **User Story 2 (Phase 4)**: Depends on User Story 1 because dispatcher and pause/pressure
  extraction build on the US1 orchestration split
- **User Story 3 (Phase 5)**: Depends on User Story 1 and can run after US2 if team prefers a
  single orchestration stream; all three stories must complete before Polish
- **Polish (Phase 6)**: Depends on all selected user stories

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other stories; this is the MVP
- **User Story 2 (P2)**: Builds on the US1 runtime seam extraction
- **User Story 3 (P3)**: Builds on US1 and validates operator-facing stability after deeper refactor

### Within Each User Story

- Write the failing tests first and verify the failure is for the intended reason
- Extract the narrow boundary or collaborator next
- Refactor the orchestration entrypoint only after the new collaborator exists
- Record red/green/refactor evidence before closing the story
- Complete the story-specific independent test before moving on

### Parallel Opportunities

- Phase 1 documentation tasks `T002`-`T005` can run in parallel
- Phase 2 regression tasks `T007`-`T013` can run in parallel because they touch different tests
- In US1, `T018`-`T020` can run in parallel after the failing tests exist
- In US2, `T027` and `T028` can run in parallel after the failing tests exist
- In US3, `T036` and `T037` can run in parallel after the failing tests exist
- In Polish, `T042`-`T045` can run in parallel before final verification commands

---

## Parallel Example: User Story 1

```bash
# Launch US1 failing-test tasks together:
Task: "T015 [US1] Add failing composition-root regression cases in cmd/port-scan/main_test.go"
Task: "T016 [US1] Add failing runtime-build regression cases in pkg/scanapp/scan_test.go"
Task: "T017 [US1] Add failing runtime-boundary integration coverage in tests/integration/scan_pipeline_test.go"

# Launch US1 extraction tasks together after tests are red:
Task: "T018 [US1] Extract command composition helpers into cmd/port-scan/command_handlers.go"
Task: "T019 [US1] Extract input loading responsibilities into pkg/scanapp/input_loader.go"
Task: "T020 [US1] Extract runtime building responsibilities into pkg/scanapp/runtime_builder.go"
```

## Parallel Example: User Story 2

```bash
# Launch US2 failing-test tasks together:
Task: "T024 [US2] Add failing dispatch-seam regression cases in pkg/scanapp/scan_test.go"
Task: "T025 [US2] Add failing pause and pressure coordination cases in pkg/scanapp/scan_observability_test.go"
Task: "T026 [US2] Add failing dispatcher integration coverage in tests/integration/scan_pipeline_test.go"

# Launch US2 extraction tasks together after tests are red:
Task: "T027 [US2] Extract task dispatch coordination into pkg/scanapp/task_dispatcher.go"
Task: "T028 [US2] Extract pause and pressure monitoring into pkg/scanapp/pressure_monitor.go"
```

## Parallel Example: User Story 3

```bash
# Launch US3 failing-test tasks together:
Task: "T033 [US3] Add failing result aggregation and progress cases in pkg/scanapp/scan_observability_test.go"
Task: "T034 [US3] Add failing resume persistence cases in pkg/scanapp/scan_test.go"
Task: "T035 [US3] Add failing resume/output integration coverage in tests/integration/resume_flow_test.go"

# Launch US3 extraction tasks together after tests are red:
Task: "T036 [US3] Extract result aggregation and progress reporting into pkg/scanapp/result_aggregator.go"
Task: "T037 [US3] Extract resume persistence handling into pkg/scanapp/resume_manager.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. Stop and validate command and orchestration stability with the US1 independent test
5. Use the new seams as the baseline for deeper refactor work

### Incremental Delivery

1. Finish Setup + Foundational to create the regression net and contract docs
2. Deliver User Story 1 to establish the first SOLID boundary split
3. Deliver User Story 2 to enforce reviewable TDD and dispatch/monitor seams
4. Deliver User Story 3 to protect operator confidence, resume flow, and observability
5. Finish Polish with full verification evidence and release impact notes

### Parallel Team Strategy

1. Team completes Setup and Foundational together
2. One developer drives US1 because it defines the primary seams
3. After US1 lands, one developer can take US2 while another prepares US3 test work
4. Rejoin for final verification and release-note updates in Phase 6

---

## Notes

- Every task uses the required checklist format: checkbox, task ID, optional `[P]`, required story
  label inside story phases, and an exact file path
- User Story 1 is the recommended MVP scope
- The prerequisites script reported a duplicate `001-*` prefix in `specs/`; task generation used
  `FEATURE_DIR=/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd`
