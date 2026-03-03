# Tasks: IP-Aware Baseline Specification

**Input**: Design documents from `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: 本功能在 spec/plan/constitution 已明確要求測試先行與品質 gate，因此每個 User Story 都包含先寫失敗測試再實作的任務。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Constitution Alignment Check

- **I. Library-First Design**: PASS（核心行為實作任務集中在 `pkg/*`，CLI glue 任務分離於 `cmd/`。）
- **II. CLI Contract-First**: PASS（CLI 契約測試/實作任務：T012、T013、T019、T029、T040。）
- **III. Test-First Delivery (NON-NEGOTIABLE)**: PASS（每個 US 先測試任務再實作任務，且含獨立驗證任務。）
- **IV. Integration Coverage for Contract Boundaries**: PASS（整合測試任務：T018、T025、T039、T050。）
- **V. Isolated End-to-End Verification**: PASS（e2e 任務：T035、T041、T057，另加 `e2e/out/` 產物驗證 T058。）
- **VI. Observability by Default**: PASS（觀測性契約/實作任務：T042、T043、T047、T048。）
- **VII. Versioning and Release Evidence**: PASS（release notes 任務：T054。）
- **Technology Stack Requirements**: PASS（任務均對齊現有 Go 專案結構與模組。）
- **Quality Gates**: PASS（全量測試/coverage/e2e gate：T055、T056、T057。）

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- 所有任務描述都包含精確檔案路徑

## Path Conventions

- 單一 Go 專案，程式碼在 `cmd/`, `pkg/`, `internal/`, `tests/`, `e2e/`
- 本文件採用絕對路徑，避免執行歧義

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 為本 feature 建立共用測試資料與輔助工具

- [X] T001 建立驗證輸出目錄結構於 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/`（含 `.gitkeep`）
- [X] T002 建立 US1/US2/US3 測試輸入樣本於 `/Users/xuxiping/tsmc/port-scan-mk3/tests/integration/testdata/ip_aware/` 與 `/Users/xuxiping/tsmc/port-scan-mk3/e2e/inputs/`
- [X] T003 [P] 新增共用檔名斷言 helper 於 `/Users/xuxiping/tsmc/port-scan-mk3/internal/testkit/output_batch_assert.go`
- [X] T004 [P] 新增共用 resume 狀態比對 helper 於 `/Users/xuxiping/tsmc/port-scan-mk3/internal/testkit/resume_assert.go`
- [X] T005 更新或新增 testkit 測試於 `/Users/xuxiping/tsmc/port-scan-mk3/internal/testkit/output_batch_assert_test.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/internal/testkit/resume_assert_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 建立所有 user story 共享的契約骨幹（旗標、路徑、批次命名、CLI 基礎）

**⚠️ CRITICAL**: User Story implementation 必須在本階段完成後才開始

- [X] T006 擴充 config 契約測試（`-cidr-ip-col`/`-cidr-ip-cidr-col`/`-resume`/`-output`）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/config/config_test.go`
- [X] T007 實作或調整 config 解析與預設值於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/config/config.go`
- [X] T008 [P] 新增批次輸出命名單元測試（UTC `YYYYMMDDTHHMMSSZ` + `-n`）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/batch_output_test.go`
- [X] T009 實作批次命名與同秒衝突遞增序號邏輯於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/batch_output.go`
- [X] T010 [P] 新增 resume 路徑解析單元測試於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/resume_path_test.go`
- [X] T011 實作 resume 讀寫路徑解析 helper（含 `<output-dir>/resume_state.json` fallback）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/resume_path.go`
- [X] T012 更新 CLI 使用說明與旗標顯示契約測試於 `/Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main_extra_test.go`
- [X] T013 調整 CLI usage 與解析 glue code 於 `/Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main.go`
- [X] T014 執行 foundational 驗證：`go test ./pkg/config ./pkg/scanapp ./cmd/port-scan`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/foundational-test.log`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - 依欄位名稱定義掃描目標 (Priority: P1) 🎯 MVP

**Goal**: 以 case-sensitive 欄位映射與 fail-fast 驗證，正確建立 `ip/ip_cidr` 掃描目標模型

**Independent Test**: 提供含額外欄位與自訂欄位名的 CIDR CSV，執行 validate/scan，確認只使用指定欄位、違規資料在派發前終止。

### Tests for User Story 1 ⚠️

- [X] T015 [P] [US1] 新增欄位名精確匹配/重複欄位 fatal 測試於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/input/cidr_columns_test.go`
- [X] T016 [P] [US1] 新增 fail-fast 驗證矩陣測試（containment/overlap/duplicate）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/input/validate_ip_rules_test.go`
- [X] T017 [P] [US1] 新增 selector 展開與去重行為測試於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/task/selector_expand_test.go`
- [X] T018 [P] [US1] 新增任務數公式斷言測試（`unique(ip-expand) * port_count == task_count`，對應 SC-003）於 `/Users/xuxiping/tsmc/port-scan-mk3/tests/integration/task_count_formula_test.go`
- [X] T019 [P] [US1] 新增 validate 子命令整合測試（自訂欄位名 + 錯誤訊息）於 `/Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main_test.go`

### Implementation for User Story 1

- [X] T020 [US1] 實作欄位名稱 case-sensitive 精確匹配與重複欄位檢查於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/input/cidr.go`
- [X] T021 [US1] 實作完整 fail-fast 驗證矩陣（格式、包含關係、跨組重疊、組內重疊）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/input/validate.go`
- [X] T022 [US1] 調整 CIDR row/domain 結構承載 `ip/ip_cidr` 與 metadata 於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/input/types.go`
- [X] T023 [US1] 實作僅由 `ip` selector 展開目標的任務生成邏輯於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/task/selector_expand.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/task/ipv4.go`
- [X] T024 [US1] 串接 validate/scan 啟動流程，保證驗證在任務派發前完成於 `/Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go`
- [X] T025 [US1] 執行 US1 測試：`go test ./pkg/input ./pkg/task ./tests/integration ./cmd/port-scan -run 'CIDR|Validate|selector|task|validate'`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/us1-test.log`

**Checkpoint**: User Story 1 可獨立完成並驗證（MVP）

---

## Phase 4: User Story 2 - 取得雙輸出結果檔 (Priority: P2)

**Goal**: 每次掃描輸出同批次主檔與 open-only 檔，採 UTC 時間戳命名並處理同秒衝突

**Independent Test**: 執行一次含 open/close 目標掃描，確認輸出目錄產生同批次 `scan_results-*` 與 `opened_results-*`，且 open-only 僅含 open。

### Tests for User Story 2 ⚠️

- [X] T026 [P] [US2] 新增 CSV header/row 契約測試（含 `ip_cidr` 欄位）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/csv_writer_test.go`
- [X] T027 [P] [US2] 新增 open-only 行為測試（無 open 時僅表頭）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/open_writer_test.go`
- [X] T028 [P] [US2] 新增 scanapp 批次輸出檔名與共用序號測試於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_test.go`
- [X] T029 [P] [US2] 新增 CLI 掃描輸出契約測試（檔名 pattern 與雙檔存在）於 `/Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main_scan_test.go`
- [X] T030 [P] [US2] 新增固定欄位順序與缺值 metadata 空字串輸出測試（對應 FR-011）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/csv_writer_contract_test.go`

### Implementation for User Story 2

- [X] T031 [US2] 實作 scan 結果檔批次命名套用（`scan_results-YYYYMMDDTHHMMSSZ[-n].csv`）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/batch_output.go`
- [X] T032 [US2] 實作 open-only 同批次檔名套用（`opened_results-YYYYMMDDTHHMMSSZ[-n].csv`）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go`
- [X] T033 [US2] 保證 open-only writer 只落 open 且空集合仍寫表頭於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/open_writer.go`
- [X] T034 [US2] 對齊 writer record 輸出欄位順序，並將缺值 `fab_name`/`cidr_name` 正規化為空字串於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/csv_writer.go`
- [X] T035 [US2] 更新 e2e 正常情境輸出斷言於 `/Users/xuxiping/tsmc/port-scan-mk3/e2e/run_e2e.sh` 與 `/Users/xuxiping/tsmc/port-scan-mk3/e2e/report/generate_report.go`
- [X] T036 [US2] 執行 US2 測試：`go test ./pkg/writer ./pkg/scanapp ./cmd/port-scan -run 'open|output|batch|scan|header|metadata'`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/us2-test.log`

**Checkpoint**: User Story 1 + 2 各自可獨立驗證

---

## Phase 5: User Story 3 - 在中斷與壓力異常下可恢復 (Priority: P3)

**Goal**: 完整落實 OR-gate 暫停、API 連續三次失敗 fatal、resume 路徑讀寫一致與續跑無重複/遺漏

**Independent Test**: 模擬 manual pause、API 5xx/timeout、SIGINT，確認第 3 次 API 失敗終止並保存 state，續跑後結果與基準一致。

### Tests for User Story 3 ⚠️

- [X] T037 [P] [US3] 新增 resume 路徑語意測試（顯式 `-resume` 與 default fallback）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_test.go`
- [X] T038 [P] [US3] 新增壓力 API 連續失敗 1/2/3 次測試於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_test.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_helpers_test.go`
- [X] T039 [P] [US3] 新增 resume 續跑無重複無遺漏整合測試於 `/Users/xuxiping/tsmc/port-scan-mk3/tests/integration/resume_flow_test.go`
- [X] T040 [P] [US3] 新增 CLI 掃描中斷/恢復測試於 `/Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main_scan_test.go`
- [X] T041 [P] [US3] 新增 e2e API 5xx/timeout/conn-fail 斷言（non-zero + resume 檔）於 `/Users/xuxiping/tsmc/port-scan-mk3/e2e/run_e2e.sh`
- [X] T042 [P] [US3] 新增結構化事件欄位契約測試（`target`/`port`/`state_transition`/`error_cause`）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/logx/logx_contract_test.go`
- [X] T043 [P] [US3] 新增長掃描 progress 與 completion summary 事件測試於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_observability_test.go`

### Implementation for User Story 3

- [X] T044 [US3] 實作 `-resume` 同路徑讀寫與預設 fallback 套用於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/config/config.go`
- [X] T045 [US3] 強化中斷保存流程，確保最新 `NextIndex` 與狀態一致性於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/state/state.go`
- [X] T046 [US3] 落實 API 失敗計數升級策略（第 3 次 fatal）與可追蹤事件輸出於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/logx/logx.go`
- [X] T047 [US3] 落實執行事件欄位正規化輸出（含無錯誤時 `error_cause=none`）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/logx/logx.go`
- [X] T048 [US3] 落實長掃描 progress 與 completion summary 事件發送於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go`
- [X] T049 [US3] 驗證/調整 OR-gate 暫停恢復邏輯於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/speedctrl/controller.go` 與 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan.go`
- [X] T050 [US3] 執行 US3 測試：`go test ./pkg/scanapp ./pkg/logx ./pkg/state ./tests/integration ./cmd/port-scan -run 'resume|pressure|pause|cancel|observability|progress|summary'`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/us3-test.log`

**Checkpoint**: 三個 user stories 均可獨立測試且功能完整

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 收斂跨故事品質、文件與最終驗證

- [X] T051 [P] 盤點本 feature 變更涉及的 public package API doc comments 缺口，輸出清單到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/public-api-doc-audit.md`
- [X] T052 補齊缺漏 Go doc comments（輸入、輸出、失敗模式）於 `/Users/xuxiping/tsmc/port-scan-mk3/pkg/input/`、`/Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/`、`/Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/`、`/Users/xuxiping/tsmc/port-scan-mk3/pkg/state/` 公開 API
- [X] T053 [P] 更新使用文件與範例命令（時間戳輸出與 resume 規則）於 `/Users/xuxiping/tsmc/port-scan-mk3/README.md` 與 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/quickstart.md`
- [X] T054 [P] 更新版本化 release notes（新契約/相容性/遷移提醒）於 `/Users/xuxiping/tsmc/port-scan-mk3/docs/release-notes/<version>.md`（`<version>` 採 `MAJOR.MINOR.PATCH`）
- [X] T055 執行完整單元與整合測試：`cd /Users/xuxiping/tsmc/port-scan-mk3 && go test ./...`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/full-test.log`
- [X] T056 執行 coverage gate：`cd /Users/xuxiping/tsmc/port-scan-mk3 && bash scripts/coverage_gate.sh`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/coverage.log`
- [X] T057 執行 e2e gate：`cd /Users/xuxiping/tsmc/port-scan-mk3 && bash e2e/run_e2e.sh`，並輸出結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/e2e.log`
- [X] T058 驗證 e2e 報告產物已輸出至 `/Users/xuxiping/tsmc/port-scan-mk3/e2e/out/`（檔案存在與最小摘要欄位），並輸出檢查結果到 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/verification/e2e-artifacts.log`
- [X] T059 彙整最終驗證與輸出樣本（含 batch 檔名、resume 狀態、observability 事件）於 `/Users/xuxiping/tsmc/port-scan-mk3/specs/001-ip-aware-baseline-spec/`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: 無依賴，可立即開始
- **Phase 2 (Foundational)**: 依賴 Phase 1，且阻擋所有 user story
- **Phase 3~5 (US1~US3)**: 依賴 Phase 2
- **Phase 6 (Polish)**: 依賴已選定 user stories 完成

### User Story Dependencies

- **US1 (P1)**: 可在 Foundational 完成後立即開始，為 MVP 最小可交付範圍
- **US2 (P2)**: 依賴 Foundational；建議在 US1 後執行（共享 scan/writer 路徑，避免衝突）
- **US3 (P3)**: 依賴 Foundational；建議在 US1 後執行（共享 scan/resume/logx 路徑，避免衝突）

### Within Each User Story

- 測試任務必須先寫、先失敗，再進入實作任務
- 先完成資料/契約層，再完成流程整合
- 每個故事完成後先做獨立驗證，再進入下一故事

### Parallel Opportunities

- Setup：T003 與 T004 可平行
- Foundational：T008 與 T010 可平行
- US1：T015/T016/T017/T018/T019 可平行（測試）
- US2：T026/T027/T028/T029/T030 可平行（測試）
- US3：T037/T038/T039/T040/T041/T042/T043 可平行（測試）
- Polish：T053 與 T054 可平行

---

## Parallel Example: User Story 1

```bash
Task: "T015 [US1] 新增欄位名精確匹配測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/input/cidr_columns_test.go"
Task: "T016 [US1] 新增 fail-fast 驗證矩陣測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/input/validate_ip_rules_test.go"
Task: "T017 [US1] 新增 selector 展開測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/task/selector_expand_test.go"
Task: "T018 [US1] 新增任務數公式斷言測試於 /Users/xuxiping/tsmc/port-scan-mk3/tests/integration/task_count_formula_test.go"
Task: "T019 [US1] 新增 validate CLI 整合測試於 /Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main_test.go"
```

## Parallel Example: User Story 2

```bash
Task: "T026 [US2] CSV 欄位契約測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/csv_writer_test.go"
Task: "T027 [US2] open-only 測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/open_writer_test.go"
Task: "T028 [US2] batch 命名測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_test.go"
Task: "T029 [US2] scan CLI 輸出契約測試於 /Users/xuxiping/tsmc/port-scan-mk3/cmd/port-scan/main_scan_test.go"
Task: "T030 [US2] 欄位順序/缺值測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/writer/csv_writer_contract_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T037 [US3] resume 路徑語意測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_test.go"
Task: "T038 [US3] pressure API 升級策略測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_helpers_test.go"
Task: "T039 [US3] 續跑一致性整合測試於 /Users/xuxiping/tsmc/port-scan-mk3/tests/integration/resume_flow_test.go"
Task: "T042 [US3] 觀測欄位契約測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/logx/logx_contract_test.go"
Task: "T043 [US3] progress/completion 事件測試於 /Users/xuxiping/tsmc/port-scan-mk3/pkg/scanapp/scan_observability_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1 + Phase 2
2. 完成 US1 (Phase 3)
3. 執行 T025 做獨立驗證
4. 若通過即可先交付 MVP

### Incremental Delivery

1. Foundation 完成後先交付 US1
2. 再加入 US2（輸出契約）
3. 最後加入 US3（韌性與恢復）
4. 每個增量都維持可獨立驗證與可示範

### Parallel Team Strategy

1. 共同完成 Phase 1/2
2. 分流執行：
   - 開發者 A：US1
   - 開發者 B：US2
   - 開發者 C：US3
3. 以 Phase 6 收斂整合驗證

---

## Notes

- `[P]` 任務代表無檔案衝突且依賴可分離
- `[USx]` 標籤確保任務可追溯到 user story
- 所有任務皆可直接指派給 LLM 執行，不需額外上下文
- 每個故事都包含獨立測試準則，避免跨故事耦合
