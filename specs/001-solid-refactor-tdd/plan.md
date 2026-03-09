# Implementation Plan: SOLID Refactor and TDD Enforcement

**Branch**: `[001-solid-refactor-tdd]` | **Date**: 2026-03-07 | **Spec**: [/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/spec.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/spec.md)
**Input**: Feature specification from `/specs/001-solid-refactor-tdd/spec.md`

## Summary

本計畫將把目前混合在 `cmd/port-scan` 與 `pkg/scanapp` 內的多重責任重新拆分為明確、
可測試、可逐步交付的邊界，並用憲章要求的 SOLID 原則重建依賴方向。重點不是重寫
產品功能，而是在不破壞既有操作契約的前提下，把 CLI 路由、掃描協調、任務派發、
觀測性輸出、恢復狀態與外部互動隔離成單一責任的組件。

技術路徑採嚴格 test-first。每個重構增量都先建立失敗測試，再做最小結構調整，並以
單元、整合、必要時 e2e 驗證來保護 validate、scan、progress、resume 等既有行為。

## Technical Context

**Language/Version**: Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`)
**Primary Dependencies**: Go standard library (`context`, `flag`, `io`, `net`, `net/http`, `os`, `sync`, `syscall`, `time`, `encoding/json`), existing internal packages (`pkg/config`, `pkg/input`, `pkg/logx`, `pkg/pipeline`, `pkg/ratelimit`, `pkg/scanapp`, `pkg/scanner`, `pkg/speedctrl`, `pkg/state`, `pkg/task`, `pkg/writer`)
**Storage**: Local filesystem (input CSV, output CSV, resume JSON, spec artifacts)
**Testing**: `go test ./...`, targeted package tests under `cmd/port-scan` and `pkg/*`, integration tests under `tests/integration`, coverage gate `bash scripts/coverage_gate.sh`, e2e gate `bash e2e/run_e2e.sh` when scan workflow behavior is touched
**Target Platform**: CLI on Linux/macOS; Docker Compose available for e2e verification
**Project Type**: Library-first CLI application
**Performance Goals**: Preserve existing operator-visible behavior for `validate` and `scan`; maintain deterministic task dispatch, progress reporting, and resume behavior while reducing structural coupling
**Constraints**: No unapproved CLI/output contract drift; every increment MUST start with failing tests; `cmd/port-scan` stays as composition glue only; package boundaries MUST avoid cyclic dependencies and oversized interfaces; e2e remains mandatory when scan flow, writer flow, or recovery behavior changes
**Scale/Scope**: Focus on high-risk orchestration boundaries first, especially `cmd/port-scan/main.go` and `pkg/scanapp/scan.go` (~973 lines), then ripple into dependent packages only where contract clarity requires it

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Gate Review

- **I. Library-First Design**: PASS
  重構主體集中在 `pkg/*`，`cmd/port-scan` 只保留命令路由與 I/O 接線。
- **II. CLI Contract-First**: PASS
  本計畫以維持 `validate|scan` 行為穩定為前提，任何契約變更都必須被單獨標示。
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS
  每個增量都以失敗測試開場，並要求保留 red/green/refactor 證據。
- **IV. Integration Coverage for Contract Boundaries**: PASS
  `cmd/config/scanapp/scanner/state/writer` 的邊界重整需要整合測試保護。
- **V. Isolated End-to-End Verification**: PASS
  若重構觸及 scan pipeline、writer 或 resume 行為，必須執行 Docker Compose e2e。
- **VI. Observability by Default**: PASS
  progress、completion、error、resume 相關訊號都屬於受保護契約。
- **VII. Versioning and Release Evidence**: PASS
  預設為 PATCH；若發現使用者可見契約改動，需升級評估與 release notes 說明。
- **VIII. SOLID Structural Boundaries**: PASS
  計畫核心就是把責任、抽象 ownership、依賴方向與擴充點重新對齊憲章。
- **Technology Stack Requirements**: PASS
  維持 Go 1.24 與既有標準庫/內部 package，不引入不必要依賴。
- **Quality Gates**: PASS
  已規劃 `go test ./...`、coverage gate 與條件式 e2e gate。

**Gate Result**: PASS（無需違規豁免）

## Project Structure

### Documentation (this feature)

```text
specs/001-solid-refactor-tdd/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli-stability-contract.md
│   ├── tdd-evidence-contract.md
│   └── runtime-boundaries.md
└── tasks.md                  # 將由 /speckit.tasks 產生
```

### Source Code (repository root)

```text
cmd/
└── port-scan/

pkg/
├── cli/
├── config/
├── input/
├── logx/
├── pipeline/
├── ratelimit/
├── scanapp/
├── scanner/
├── speedctrl/
├── state/
├── task/
└── writer/

tests/
└── integration/

e2e/
├── docker-compose.yml
├── run_e2e.sh
├── mock-pressure-api/
├── mock-target-open/
├── mock-target-closed/
└── out/
```

**Structure Decision**: 維持單一 Go CLI 專案。優先把協調與 I/O glue 分離到小型、
consumer-owned 的 package 邊界，避免在 `cmd/port-scan` 或 `pkg/scanapp` 聚集解析、
外部互動、流程控制、觀測性與狀態持久化等多重責任。

## Phase 0: Outline & Research

### Research Inputs

- **Unknowns from Technical Context**:
  - 如何把 `pkg/scanapp/scan.go` 拆成更小的協調元件，而不改變既有 `scan` 指令外部行為。
  - 哪些責任邊界需要消費端擁有的窄介面，哪些情況應直接依賴 concrete type 以避免假抽象。
  - 如何定義可審查的 TDD 證據格式，讓 reviewer 能確認每個增量確實經過 red/green/refactor。
- **Dependency best-practice tasks**:
  - Go 中以小介面與 composition 重構大型流程函式的做法。
  - context cancellation、pause gate、rate limit、worker fan-out 的邊界劃分原則。
  - CLI composition root 與 domain/service orchestration 的依賴方向實務。
- **Integration pattern tasks**:
  - `cmd -> config -> scanapp -> scanner/state/writer` 的最小契約切分路徑。
  - progress/completion/resume 訊號在拆分後仍保持一致的驗證模式。
  - 何時以 integration test 足夠，何時必須升級到 e2e 才能保護 operator contract。

### Phase 0 Output

- [research.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/research.md)

## Phase 1: Design & Contracts

### Design Artifacts

- [data-model.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/data-model.md)
- [quickstart.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/quickstart.md)
- Contracts:
  - [cli-stability-contract.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/contracts/cli-stability-contract.md)
  - [tdd-evidence-contract.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/contracts/tdd-evidence-contract.md)
  - [runtime-boundaries.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-solid-refactor-tdd/contracts/runtime-boundaries.md)

### Agent Context Update

- Run: `.specify/scripts/bash/update-agent-context.sh codex`
- Purpose: 同步本 feature 的 SOLID 與 TDD 約束到 agent context，同時保留手動維護區塊。

## Post-Design Constitution Check

- **I. Library-First Design**: PASS（核心責任重新安置在 `pkg/*`，CLI 保持 glue）
- **II. CLI Contract-First**: PASS（先定義穩定契約，再做內部拆分）
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS（quickstart 將明列每個增量的 red/green/refactor）
- **IV. Integration Coverage for Contract Boundaries**: PASS（邊界重整至少覆蓋 `cmd/scanapp/scanner/state/writer`）
- **V. Isolated End-to-End Verification**: PASS（scan flow/resume/writer 受影響時必跑 e2e）
- **VI. Observability by Default**: PASS（progress/completion/error/resume 訊號列為受保護契約）
- **VII. Versioning and Release Evidence**: PASS（預設 PATCH；若出現契約差異需升級評估）
- **VIII. SOLID Structural Boundaries**: PASS（明確驗證 responsibility、interface ownership、dependency direction）
- **Technology Stack Requirements**: PASS（Go 1.24 與既有依賴維持不變）
- **Quality Gates**: PASS（`go test ./...`、coverage、條件式 e2e 已納入）

**Post-Design Gate Result**: PASS

## Phase 2: Task Planning Approach

- 以可獨立驗證的重構增量拆任務，建議順序如下：
  1. 建立 operator contract baseline tests，先鎖定 `validate`/`scan` 既有對外行為。
  2. 把 `cmd/port-scan/main.go` 收斂為純 composition root，移除與掃描流程耦合的責任。
  3. 拆分 `pkg/scanapp/scan.go` 為明確協調元件，例如輸入載入、runtime 構建、dispatch、
     result aggregation、progress reporting、resume persistence、pressure control。
  4. 對新邊界補足單元與整合測試，要求每個增量都保留 red/green 證據。
  5. 在掃描流程契約受影響時執行 e2e，最後更新 release notes 與驗證證據。
- `/speckit.tasks` 必須把每個增量拆成 test-first 的最小步驟，避免單一大型 refactor commit。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
