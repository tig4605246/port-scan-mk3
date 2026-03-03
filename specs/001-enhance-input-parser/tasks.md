# Tasks: Enhanced Input Field Parsing

**Input**: Design documents from `/specs/001-enhance-input-parser/`  
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Test tasks are REQUIRED by constitution and feature spec (test-first + integration boundary coverage + mandatory e2e when scan pipeline/writer behavior is touched).

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no incomplete-task dependency)
- **[Story]**: User story label (`[US1]`, `[US2]`, `[US3]`)
- Every task includes concrete file path(s)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare feature-specific test fixtures and task scaffolding.

- [ ] T001 Create rich-input fixture directory and seed CSV samples in `tests/integration/testdata/rich_input/valid_mixed.csv`, `tests/integration/testdata/rich_input/invalid_rows.csv`, and `tests/integration/testdata/rich_input/dedup_context.csv`
- [ ] T002 [P] Create feature-specific integration test scaffold in `tests/integration/rich_input_parse_test.go`
- [ ] T003 [P] Create parser unit test scaffold in `pkg/input/rich_parser_test.go`
- [ ] T004 [P] Create pipeline-boundary integration test scaffold in `tests/integration/rich_input_pipeline_boundary_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core shared structures and helpers that block all user stories.

**⚠️ CRITICAL**: No user story implementation before this phase is complete.

- [ ] T005 Define shared rich-input data structures in `pkg/input/rich_types.go`
- [ ] T006 [P] Implement canonical header matching helpers (case/trim only) in `pkg/input/header_match.go`
- [ ] T007 [P] Implement parser validation error-code helpers in `pkg/input/validation_errors.go`
- [ ] T008 [P] Implement execution-key helper (`dst_ip:port/protocol`) in `pkg/task/execution_key.go`
- [ ] T009 Add shared rich-input integration scenario helper in `tests/integration/rich_input_scenario.go`

**Checkpoint**: Foundation complete; user stories can proceed.

---

## Phase 3: User Story 1 - Parse rich input rows into usable scan targets (Priority: P1) 🎯 MVP

**Goal**: Parse high-column input rows into usable scan targets while tolerating extra columns and header case/trim variance.

**Independent Test**: Use a mixed-column CSV and verify accepted rows generate expected parsed targets without requiring any other user story behavior.

### Tests for User Story 1 ⚠️

- [ ] T010 [P] [US1] Add unit tests for canonical-header match and alias rejection in `pkg/input/rich_parser_test.go`
- [ ] T011 [P] [US1] Add unit tests for required-10-field enforcement and protocol/decision/port acceptance in `pkg/input/rich_parser_test.go`
- [ ] T012 [P] [US1] Add integration test for mixed-column parse success in `tests/integration/rich_input_parse_test.go`

### Implementation for User Story 1

- [ ] T013 [US1] Implement rich-row CSV parser and header binding in `pkg/input/rich_parser.go`
- [ ] T014 [US1] Implement required-field/value validation for canonical 10 fields in `pkg/input/rich_parser.go`
- [ ] T015 [US1] Integrate rich parser entry path with existing input load flow in `pkg/input/cidr.go`
- [ ] T016 [US1] Build `ParsedTargetRecord` generation including `execution_key` in `pkg/input/rich_parser.go`
- [ ] T017 [US1] Wire parsed target usage into scan path entry in `pkg/scanapp/scan.go`
- [ ] T018 [US1] Keep input schema contract aligned with implemented behavior in `specs/001-enhance-input-parser/contracts/input-schema-contract.md`

**Checkpoint**: User Story 1 is independently functional and testable.

---

## Phase 4: User Story 2 - Retain operational context and deduplicated execution mapping (Priority: P2)

**Goal**: Preserve row-level background context while ensuring execution occurs once per dedup key and remains consistent through parser-output and writer-output contracts.

**Independent Test**: Provide rows sharing the same `dst_ip+port+protocol` but different `policy_id/reason`; verify one execution target with full row-level traceability and preserved writer output context fields.

### Tests for User Story 2 ⚠️

- [ ] T019 [P] [US2] Add unit tests for execution-key grouping and source-row mapping in `pkg/task/execution_key_test.go`
- [ ] T020 [P] [US2] Add integration test for dedup-with-context-retention behavior in `tests/integration/rich_input_mapping_test.go`
- [ ] T021 [P] [US2] Add boundary integration test for parser -> task expansion -> pipeline orchestration -> writer contract in `tests/integration/rich_input_pipeline_boundary_test.go` and `pkg/scanapp/scan_helpers_test.go`

### Implementation for User Story 2

- [ ] T022 [US2] Implement dedup collector by execution key in `pkg/task/execution_key.go`
- [ ] T023 [US2] Update task dispatch construction to execute once per dedup key in `pkg/scanapp/scan.go`
- [ ] T024 [US2] Extend writer record model for context fields (`service_label`, `decision`, `policy_id`, `reason`) in `pkg/writer/csv_writer.go` and `pkg/writer/csv_writer_contract_test.go`
- [ ] T025 [US2] Populate extended writer fields from parsed row context in `pkg/scanapp/scan.go`
- [ ] T026 [US2] Align parser-output and writer-output contract examples with implemented mapping in `specs/001-enhance-input-parser/contracts/parser-output-contract.md`

**Checkpoint**: User Stories 1 and 2 are both independently testable.

---

## Phase 5: User Story 3 - Reject invalid rows with clear reasons (Priority: P3)

**Goal**: Provide deterministic row rejection reasons and summary buckets for malformed input.

**Independent Test**: Run parser on mixed invalid rows and verify row-level errors, aggregated buckets, and no-usable-input stop behavior when all rows fail.

### Tests for User Story 3 ⚠️

- [ ] T027 [P] [US3] Add unit tests for src/dst IP-in-segment validation failures in `pkg/input/rich_parser_test.go`
- [ ] T028 [P] [US3] Add unit tests for rejection reason codes and summary bucket aggregation in `pkg/input/rich_parser_test.go`
- [ ] T029 [P] [US3] Add integration test for all-invalid-input stop behavior in `tests/integration/rich_input_rejection_test.go`

### Implementation for User Story 3

- [ ] T030 [US3] Implement src/dst containment validation in parser pipeline in `pkg/input/rich_parser.go`
- [ ] T031 [US3] Implement row-level rejection result + summary aggregator in `pkg/input/rich_parser.go` and `pkg/input/rich_types.go`
- [ ] T032 [US3] Surface parse summary and no-usable-input stop signal in scan flow outputs in `pkg/scanapp/scan.go`
- [ ] T033 [US3] Extend observability tests for summary/error evidence in `pkg/scanapp/scan_observability_test.go`
- [ ] T034 [US3] Align failure contract examples with implemented error outputs in `specs/001-enhance-input-parser/contracts/input-schema-contract.md`

**Checkpoint**: All user stories are independently functional and testable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Cross-story verification, compliance evidence, and documentation sync.

- [ ] T035 [P] Refresh/expand rich-input fixture coverage in `tests/integration/testdata/rich_input/`
- [ ] T036 Run targeted feature tests and record outcomes in `specs/001-enhance-input-parser/quickstart.md`
- [ ] T037 Measure and record SC-003 using `tests/integration/testdata/rich_input/invalid_rows.csv` (formula: corrected-invalid-rows/original-invalid-rows within 10 minutes) in `specs/001-enhance-input-parser/quickstart.md`
- [ ] T038 Measure and record SC-004 using mixed dataset (formula: successful-execution-keys/expected-execution-keys-from-valid-rows) in `specs/001-enhance-input-parser/quickstart.md`
- [ ] T039 Run full regression `go test ./...` and record evidence in `specs/001-enhance-input-parser/quickstart.md`
- [ ] T040 Run coverage gate `bash scripts/coverage_gate.sh` and record evidence in `specs/001-enhance-input-parser/quickstart.md`
- [ ] T041 Run mandatory `bash e2e/run_e2e.sh` for this feature (touches scan pipeline/writer behavior), store artifacts in `e2e/out/`, and record result in `specs/001-enhance-input-parser/quickstart.md` and `specs/001-enhance-input-parser/plan.md`
- [ ] T042 Add/update Go doc comments for public behavior in `pkg/input/rich_parser.go` and `pkg/task/execution_key.go` (inputs, outputs, failure modes)
- [ ] T043 Update release notes for feature delivery in `docs/release-notes/1.2.0.md`
- [ ] T044 [P] Sync final model/contract docs with implemented behavior in `specs/001-enhance-input-parser/data-model.md`, `specs/001-enhance-input-parser/contracts/input-schema-contract.md`, and `specs/001-enhance-input-parser/contracts/parser-output-contract.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: no dependency.
- **Phase 2 (Foundational)**: depends on Phase 1; blocks all user stories.
- **Phase 3 (US1)**: depends on Phase 2.
- **Phase 4 (US2)**: depends on Phase 2 and US1 parsed-target pipeline availability (T013-T017).
- **Phase 5 (US3)**: depends on Phase 2 and can proceed in parallel with late US2 tasks once parser baseline exists (T013-T014).
- **Phase 6 (Polish)**: depends on all selected stories complete.

### User Story Dependency Graph

- **US1 (P1)**: Foundation story; no story dependency.
- **US2 (P2)**: Depends on US1 parser output contract being available.
- **US3 (P3)**: Depends on foundational parser baseline; does not depend on writer extensions from US2.

Graph (story level): `US1 -> US2`; `US1 -> US3`

### Within Each User Story

- Tests first: create and run failing tests before implementation tasks.
- Implement models/helpers before scanapp wiring.
- Complete story-specific integration checks before moving to polish.

### Parallel Opportunities

- Setup: T002/T003/T004 in parallel.
- Foundational: T006/T007/T008 in parallel after T005 starts.
- US1: T010/T011/T012 parallel; then implementation sequence.
- US2: T019/T020/T021 parallel; T024 can run parallel with T023 once dedup shape fixed.
- US3: T027/T028/T029 parallel; T033 parallel with T032.
- Polish: T035/T044 parallel; verification tasks mostly sequential for evidence clarity.

---

## Parallel Example: User Story 1

```bash
# Parallel test authoring (US1)
Task: "T010 [US1] in pkg/input/rich_parser_test.go"
Task: "T011 [US1] in pkg/input/rich_parser_test.go"
Task: "T012 [US1] in tests/integration/rich_input_parse_test.go"
```

## Parallel Example: User Story 2

```bash
# Parallel validation tasks (US2)
Task: "T019 [US2] in pkg/task/execution_key_test.go"
Task: "T020 [US2] in tests/integration/rich_input_mapping_test.go"
Task: "T021 [US2] in tests/integration/rich_input_pipeline_boundary_test.go"
```

## Parallel Example: User Story 3

```bash
# Parallel failure-path tests (US3)
Task: "T027 [US3] in pkg/input/rich_parser_test.go"
Task: "T028 [US3] in pkg/input/rich_parser_test.go"
Task: "T029 [US3] in tests/integration/rich_input_rejection_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Deliver Phase 3 (US1) end-to-end.
3. Validate US1 independently via T012 + targeted tests.
4. Demo/deploy parser capability baseline.

### Incremental Delivery

1. US1: rich parse to usable targets.
2. US2: context retention + dedup mapping + output contract consistency.
3. US3: strict rejection semantics + summary observability.
4. Polish: gates + release evidence.

### Parallel Team Strategy

1. Team completes Setup + Foundational together.
2. After US1 baseline lands, split:
   - Engineer A: US2 dedup + writer context path
   - Engineer B: US3 failure semantics + observability path
3. Rejoin for Phase 6 verification and release artifacts.

---

## Notes

- `[P]` tasks are file-isolated and can run concurrently.
- `[USx]` labels ensure story-level traceability.
- Every story includes explicit independent test criteria.
- Keep commits small: one logical behavior slice per commit.
- Do not skip test-first ordering mandated by constitution.
