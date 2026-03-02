# Phase 0 Research: IP-Aware Baseline Specification

## Scope

本研究針對三類任務整合決策：
- Technical Context 未知項釐清（本次無未決項目）
- 依賴選型最佳實務
- 外部整合（CLI、壓力 API、檔案契約）模式

## Decisions

### 1) Core runtime stack

- **Decision**: 維持 Go 1.24 + standard library 為主，僅保留既有 `golang.org/x/term`、`golang.org/x/sys`。
- **Rationale**: 符合 Constitution 的 Go 技術棧要求，且目前功能（TCP scan、CSV/JSON、signal/control）已足夠，不需新增框架。
- **Alternatives considered**: 引入第三方 CLI/CSV framework；因增加依賴面、測試矩陣與維運成本而不採用。

### 2) CLI contract stability

- **Decision**: 保持 `port-scan validate` 與 `port-scan scan` 雙命令模式，維持 human/json 輸出與既有 exit code 語意。
- **Rationale**: 既有測試與使用方式已綁定該契約，穩定介面可降低升級風險。
- **Alternatives considered**: 合併為單命令多模式；會提高使用錯誤率並破壞現有腳本相容性。

### 3) Header mapping strictness

- **Decision**: `-cidr-ip-col`、`-cidr-ip-cidr-col` 採區分大小寫精確匹配，重複同名欄位直接 fail-fast。
- **Rationale**: 可避免隱式欄位選擇與歧義，錯誤可在啟動前被明確攔截。
- **Alternatives considered**: 不分大小寫匹配或「取第一欄」策略；會造成不可預期資料映射。

### 4) Resume path semantics

- **Decision**: 有 `-resume` 時，該路徑同時用於讀取與後續寫回；未提供時預設 `<output-dir>/resume_state.json`。
- **Rationale**: 讀寫路徑一致可避免 state 檔散落與排障困難，符合可追蹤性要求。
- **Alternatives considered**: 固定單一路徑、只讀不寫 `-resume`；均會增加理解與維護負擔。

### 5) Output file versioning

- **Decision**: 每次執行建立 UTC 時間戳批次檔：`YYYYMMDDTHHMMSSZ`；同秒衝突使用遞增序號 `-<n>`。
- **Rationale**: 兼具可追溯性、時區無歧義與衝突可預測解法，不覆蓋舊檔。
- **Alternatives considered**: 覆寫舊檔、append、UUID；覆寫/append 會污染結果，UUID 可讀性較差。

### 6) Pressure API failure handling

- **Decision**: API 連續失敗 1~2 次記錄錯誤並沿用上次狀態，第 3 次連續失敗轉為 fatal 並保存 resume。
- **Rationale**: 在容錯與安全停機間取得平衡，並與既有設計文件一致。
- **Alternatives considered**: 無限重試或立即失敗；前者可能長期不確定，後者過度敏感。

### 7) Test layering

- **Decision**: 以 unit + integration + docker e2e 三層驗證，coverage gate >85%。
- **Rationale**: 直接對應 Constitution Quality Gates，且能覆蓋邏輯、流程與環境互動。
- **Alternatives considered**: 僅 unit 或僅 e2e；會留下契約縫隙或回歸定位困難。

## Resolved Clarifications

- `Technical Context` 中所有項目已明確，無殘留未決項目。
