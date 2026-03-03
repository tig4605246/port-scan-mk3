# Implementation Plan: IP-Aware Baseline Specification

**Branch**: `[001-ip-aware-baseline-spec]` | **Date**: 2026-03-02 | **Spec**: [/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/spec.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/spec.md)
**Input**: Feature specification from `/specs/001-ip-aware-baseline-spec/spec.md`

## Summary

本計畫將既有 Port Scan MK3 CLI 的 IP-aware 規格完整落地為可執行設計基準：以 `ip/ip_cidr` 欄位模型驅動掃描、保留 fail-fast 驗證矩陣、強化 resume 行為，並把輸出契約明確化為 UTC 時間戳批次檔（含同秒衝突序號規則）。

技術路徑採「增量擴充」：以現有 `pkg/input`、`pkg/scanapp`、`pkg/writer`、`pkg/state` 為核心，不重寫 pipeline，優先鎖定規格契約、資料模型、介面合約與可驗證流程。

## Technical Context

**Language/Version**: Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`)  
**Primary Dependencies**: Go standard library (`flag`, `net`, `encoding/csv`, `encoding/json`, `context`, `os`, `time`), `golang.org/x/term`, `golang.org/x/sys`  
**Storage**: 本機檔案系統（CSV 輸入/輸出、JSON resume state）  
**Testing**: `go test ./...`, `tests/integration/*`, `bash e2e/run_e2e.sh`, `bash scripts/coverage_gate.sh`，並驗證 `e2e/out/` 報告產物  
**Target Platform**: CLI on Linux/macOS; e2e 需 Docker Compose 隔離網路  
**Project Type**: Library-first CLI application  
**Performance Goals**: 啟動前 100% fail-fast 攔截、掃描任務數誤差 0、open-only 輸出與主輸出 open 筆數 100% 一致、API 連續 3 次失敗必定終止  
**Constraints**: 欄位名精確且區分大小寫、`ip` 目標不得超出 `ip_cidr`、輸出檔名必須 UTC `YYYYMMDDTHHMMSSZ` 且同秒衝突用遞增序號、`-resume` 路徑需讀寫一致  
**Scale/Scope**: IPv4 掃描；以 `ip_cidr` 分組、每組由 `ip` 展開唯一位址與 ports 形成任務；需覆蓋 normal/5xx/timeout e2e 場景

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Gate Review

- **I. Library-First Design**: PASS  
  本次規劃以 `pkg/*` 模組為主，保持可獨立測試與可重用。
- **II. CLI Contract-First**: PASS  
  介面仍為 `port-scan validate|scan`，保留 human/json text I/O。
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS  
  後續 `/speckit.tasks` 將要求 Red-Green-Refactor 與先寫失敗測試。
- **IV. Integration Coverage for Contract Boundaries**: PASS  
  已規劃 `tests/integration` 驗證契約與流程連動。
- **V. Isolated End-to-End Verification**: PASS  
  Docker Compose 隔離環境，含正常與 API 異常情境，且要求 `e2e/out/` 報告產物。
- **VI. Observability by Default**: PASS  
  規格已要求 pause/resume、API 失敗計數、resume save/load 等可追蹤事件。
- **VII. Versioning and Release Evidence**: PASS  
  已規劃 release notes 任務，確保 CLI 契約與相容性變更可追溯。
- **Technology Stack Requirements**: PASS  
  延用 Go 與 `net` 掃描實作，不引入違反憲章的新技術。
- **Quality Gates**: PASS  
  本計畫輸出明確納入這些 gate。

**Gate Result**: PASS（無需違規豁免）

## Project Structure

### Documentation (this feature)

```text
specs/001-ip-aware-baseline-spec/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── verification/
├── contracts/
│   ├── cli-contract.md
│   ├── file-formats.md
│   └── pressure-api-contract.md
└── tasks.md                  # 將由 /speckit.tasks 產生
```

### Source Code (repository root)

```text
cmd/
└── port-scan/

internal/

pkg/
├── cli/
├── config/
├── input/
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
├── inputs/
├── mock-pressure-api/
├── mock-target-open/
├── mock-target-closed/
└── report/
```

**Structure Decision**: 採單一 Go 專案結構，維持 library-first + CLI-first。核心邏輯在 `pkg/*`，`cmd/port-scan` 僅做命令路由與 I/O glue，`tests/integration` 與 `e2e/` 分層驗證。

## Phase 0: Outline & Research

### Research Inputs

- **Unknowns from Technical Context**: 無未決事項。
- **Dependency best-practice tasks**:
  - Go CLI 旗標與錯誤碼契約一致性
  - CSV/JSON 輸出契約穩定性與可測試性
  - `golang.org/x/term` raw-mode 鍵盤控制安全收斂策略
- **Integration pattern tasks**:
  - 壓力 API 失敗重試與 fatal cutoff（第 3 次）
  - resume 檔讀寫路徑決策與中斷一致性
  - 時間戳批次輸出（同秒衝突）命名策略

### Phase 0 Output

- [research.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/research.md)

## Phase 1: Design & Contracts

### Design Artifacts

- [data-model.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/data-model.md)
- [quickstart.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/quickstart.md)
- Contracts:
  - [cli-contract.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/contracts/cli-contract.md)
  - [file-formats.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/contracts/file-formats.md)
  - [pressure-api-contract.md](/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/contracts/pressure-api-contract.md)

### Agent Context Update

- Run: `.specify/scripts/bash/update-agent-context.sh codex`
- Purpose: 同步本 feature 的技術與契約重點到 agent context，並保留手動區塊。

## Post-Design Constitution Check

- **I. Library-First Design**: PASS（資料模型與契約都對應 `pkg/*` 模組邊界）
- **II. CLI Contract-First**: PASS（CLI contract 已文檔化並定義 exit code/輸出格式）
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS（quickstart 與後續 tasks 將以測試先行）
- **IV. Integration Coverage for Contract Boundaries**: PASS（integration 驗證路徑已納入）
- **V. Isolated End-to-End Verification**: PASS（Docker e2e 場景含 normal/5xx/timeout，並檢查 `e2e/out/` 報告）
- **VI. Observability by Default**: PASS（記錄契約已覆蓋 pause/resume/API fail streak）
- **VII. Versioning and Release Evidence**: PASS（release notes 與遷移資訊流程已保留）
- **Technology Stack Requirements**: PASS（Go + net-based TCP scan 路徑不變）
- **Quality Gates**: PASS（coverage/integration/e2e gates 皆保留）

**Post-Design Gate Result**: PASS

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
