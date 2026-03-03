# Implementation Plan: Enhanced Input Field Parsing

**Branch**: `[001-enhance-input-parser]` | **Date**: 2026-03-03 | **Spec**: [/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/spec.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/spec.md)
**Input**: Feature specification from `/specs/001-enhance-input-parser/spec.md`

## Summary

本計畫將擴充輸入解析契約，使系統可從高欄位輸入檔穩定解析 10 個關鍵欄位，並把資料列轉為可執行掃描目標與可追蹤背景語意。規格要求同時滿足嚴格欄位完整性（10 欄全必填）、來源/目標網段一致性驗證、以及執行層去重（`dst_ip + port + protocol`）但保留列級映射。

技術路徑採 library-first：核心改動落在 `pkg/input`、`pkg/task`、`pkg/scanapp` 的資料契約與邊界整合，再由 `cmd/port-scan` 保持既有 CLI 入口。驗證策略以 test-first 推進，覆蓋單元、整合與必要 e2e gate。

## Technical Context

**Language/Version**: Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`)  
**Primary Dependencies**: Go standard library (`encoding/csv`, `encoding/json`, `strings`, `net`, `strconv`, `context`, `time`, `os`), existing internal packages (`pkg/input`, `pkg/task`, `pkg/scanapp`, `pkg/writer`)  
**Storage**: Local filesystem (input CSV, output CSV, resume JSON)  
**Testing**: `go test ./...`, targeted integration tests under `tests/integration`, coverage gate `bash scripts/coverage_gate.sh`, e2e gate `bash e2e/run_e2e.sh` when pipeline behavior changes  
**Target Platform**: CLI on Linux/macOS with Docker Compose available for e2e  
**Project Type**: Library-first CLI application  
**Performance Goals**: 單次處理 100k 列輸入時維持列級可追蹤解析結果，且對合法列的目標生成正確率 100%  
**Constraints**: 10 個 canonical 欄位全必填；欄位名稱匹配僅容忍大小寫與前後空白；僅接受 `protocol=tcp` 與 `decision in {accept,deny}`；執行層去重鍵固定為 `dst_ip + port + protocol`  
**Scale/Scope**: 單檔 1~100k 列；允許大量重複目標列；每列必須保留原始策略背景並可回溯至執行結果

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Gate Review

- **I. Library-First Design**: PASS  
  改動聚焦 `pkg/*` 可重用邏輯，CLI 僅做參數與輸出接線。
- **II. CLI Contract-First**: PASS  
  不新增命令，需檢視是否影響既有輸入契約與輸出摘要文本/JSON。
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS  
  先新增失敗測試覆蓋欄位完整性、去重映射與驗證錯誤。
- **IV. Integration Coverage for Contract Boundaries**: PASS  
  需更新 parser/task/scanapp 邊界整合案例。
- **V. Isolated End-to-End Verification**: PASS  
  需評估並執行 e2e gate（若掃描分派行為受影響則必跑）。
- **VI. Observability by Default**: PASS  
  需保留/擴充解析摘要與列級錯誤可觀測訊號。
- **VII. Versioning and Release Evidence**: PASS  
  規劃 MINOR 變更並更新 release notes。
- **Technology Stack Requirements**: PASS  
  保持 Go 1.24 與標準庫 `net` 掃描路徑，不引入不必要依賴。
- **Quality Gates**: PASS  
  已規劃 `go test ./...`、coverage gate、條件性 e2e gate。

**Gate Result**: PASS（無豁免）

## Project Structure

### Documentation (this feature)

```text
specs/001-enhance-input-parser/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── input-schema-contract.md
│   └── parser-output-contract.md
└── tasks.md                 # 將由 /speckit.tasks 產生
```

### Source Code (repository root)

```text
cmd/
└── port-scan/

pkg/
├── config/
├── input/
├── pipeline/
├── scanapp/
├── task/
└── writer/

tests/
└── integration/

e2e/
├── docker-compose.yml
├── run_e2e.sh
└── report/
```

**Structure Decision**: 維持單一 Go 專案結構。核心規則與資料模型在 `pkg/input` 與 `pkg/task`，掃描分派整合在 `pkg/scanapp`，`cmd/port-scan` 不承載業務邏輯。

## Phase 0: Outline & Research

### Research Inputs

- **Unknowns from Technical Context**:
  - 100k 列輸入的解析策略與記憶體邊界最佳實務。
  - 去重後如何保留列級背景映射且不破壞可追蹤性。
  - 欄位匹配寬鬆策略（大小寫/空白）與誤匹配風險控管。
- **Dependency best-practice tasks**:
  - Go `encoding/csv` 在高欄位資料的安全解析與錯誤分級。
  - Go 測試分層策略（unit/integration/e2e）對 parser 契約變更的覆蓋。
- **Integration pattern tasks**:
  - parser -> task -> scanapp 的最小契約變更路徑。
  - 列級錯誤摘要與 observability 一致性輸出模式。

### Phase 0 Output

- [research.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/research.md)

## Phase 1: Design & Contracts

### Design Artifacts

- [data-model.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/data-model.md)
- [quickstart.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/quickstart.md)
- Contracts:
  - [input-schema-contract.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/contracts/input-schema-contract.md)
  - [parser-output-contract.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-enhance-input-parser/contracts/parser-output-contract.md)

### Agent Context Update

- Run: `.specify/scripts/bash/update-agent-context.sh codex`
- Purpose: 同步本 feature 的新契約重點到 agent context，且保留手動維護段落。

## Post-Design Constitution Check

- **I. Library-First Design**: PASS（資料模型與契約集中在 `pkg/*`）
- **II. CLI Contract-First**: PASS（已定義輸入契約與輸出摘要影響）
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS（quickstart 明確先測試後實作）
- **IV. Integration Coverage for Contract Boundaries**: PASS（已定義 parser/task/scanapp 邊界契約）
- **V. Isolated End-to-End Verification**: PASS（已定義何時必跑 e2e 與驗證點）
- **VI. Observability by Default**: PASS（已定義列級錯誤與摘要輸出契約）
- **VII. Versioning and Release Evidence**: PASS（已規劃 MINOR + release notes）
- **Technology Stack Requirements**: PASS（維持 Go 1.24 與既有依賴）
- **Quality Gates**: PASS（`go test ./...`、coverage gate、e2e gate 皆納入）

**Post-Design Gate Result**: PASS

## Phase 2: Task Planning Approach

- 以 test-first 拆解任務：
  1. parser 欄位匹配與必填驗證（unit tests first）
  2. 列級驗證結果模型與摘要（unit + integration）
  3. execution-key 去重與列級映射（integration）
  4. scanapp 邊界整合與 observability 證據（integration/e2e as needed）
  5. release notes 與回歸 gate
- `/speckit.tasks` 需輸出可直接執行的最小可驗證任務序列。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
