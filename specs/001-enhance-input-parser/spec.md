# Feature Specification: Enhanced Input Field Parsing

**Feature Branch**: `001-enhance-input-parser`  
**Created**: 2026-03-02  
**Status**: Draft  
**Input**: User description: "input IP CIDR file 的欄位眾多，修改input parser來從這些欄位中解析重要欄位並拿來使用。以下是重要欄位: src_ip, src_network_segment, dst_ip, dst_network_segment, service_label, protocol, port, decision, policy_id, reason"

## Clarifications

### Session 2026-03-02

- Q: `decision` 欄位如何影響可執行目標？ → A: `accept` 與 `deny` 都建立可執行掃描目標，`decision` 作為背景欄位保留。
- Q: 哪些欄位是「列資料可用」的必要欄位？ → A: 10 個關鍵欄位全部必填，缺任何一個即拒絕該列。
- Q: 當多列資料對應到同一個 `dst_ip + port + protocol` 時，執行層要怎麼處理？ → A: 執行層按 `dst_ip + port + protocol` 去重（只掃一次），但保留所有原始列與其背景欄位對應。
- Q: `src_ip` / `src_network_segment` 與 `dst_ip` / `dst_network_segment` 的一致性要怎麼驗證？ → A: 同時驗證來源與目標 IP 都必須落在各自 network segment 中，任一不符即拒絕該列。

### Session 2026-03-03

- Q: 欄位名稱匹配規則要多寬鬆？ → A: 以指定欄位名為唯一標準，但匹配時忽略大小寫與前後空白。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parse rich input rows into usable scan targets (Priority: P1)

作為網路掃描操作者，我希望系統可直接從欄位很多的輸入檔讀取關鍵欄位並建立可執行目標，避免手動整理欄位後才能執行流程。

**Why this priority**: 這是核心價值，若無法從來源檔直接取得 `dst_ip`、`port` 與 `protocol`，整個後續流程無法啟動。

**Independent Test**: 只提供包含必要欄位與額外雜項欄位的輸入檔，系統仍可完成資料解析並產出可執行目標清單。

**Acceptance Scenarios**:

1. **Given** 輸入檔含有多餘欄位且包含 `dst_ip`、`port`、`protocol`，**When** 系統讀取檔案，**Then** 系統必須成功建立目標資料且不受無關欄位影響。
2. **Given** 同一檔案含多筆列資料，**When** 系統完成解析，**Then** 每列都必須對應到唯一且可追蹤的解析結果（成功或失敗）。
3. **Given** 某列 `decision=deny` 且其他必要欄位皆合法，**When** 系統完成解析，**Then** 該列仍必須建立可執行目標，並保留 `decision` 供後續判讀。
4. **Given** 欄位名稱存在大小寫差異或前後空白（如 ` SRC_IP `），**When** 系統匹配欄位，**Then** 仍必須正確識別為 canonical 欄位名稱。

---

### User Story 2 - Retain operational context from input fields (Priority: P2)

作為網路安全分析人員，我希望解析後仍保留來源與策略語意（如 `src_ip`、`src_network_segment`、`service_label`、`decision`、`policy_id`、`reason`），以便判讀掃描結果的風險與決策背景。

**Why this priority**: 即使掃描可執行，若缺少政策背景欄位，結果難以解讀與落地。

**Independent Test**: 提供含政策與原因欄位的輸入資料，系統需在解析結果中保留這些欄位值，並可被查閱。

**Acceptance Scenarios**:

1. **Given** 輸入列含完整背景欄位，**When** 系統完成解析，**Then** 解析結果必須保留 `src_ip`、`src_network_segment`、`dst_network_segment`、`service_label`、`decision`、`policy_id`、`reason`。
2. **Given** 兩列有相同 `dst_ip` 與 `port` 但不同 `policy_id`，**When** 系統完成解析，**Then** 兩列不得被錯誤合併或覆蓋背景資訊。
3. **Given** 多列共享相同 `dst_ip + port + protocol` 但背景欄位不同，**When** 系統進入執行層，**Then** 掃描執行只發生一次，且所有來源列仍可回溯至該次掃描結果。
4. **Given** 解析資料進入結果輸出流程，**When** 系統產生對外輸出，**Then** 輸出契約必須保留列級背景欄位（至少 `service_label`、`decision`、`policy_id`、`reason`）與可追溯關聯資訊，避免 parser 與 writer 契約語意斷裂。

---

### User Story 3 - Reject invalid rows with clear reasons (Priority: P3)

作為資料提供者，我希望當欄位值不合法時，系統可以清楚指出哪一列有問題與原因，讓我可快速修正來源資料。

**Why this priority**: 清楚的失敗訊息可降低排錯成本，避免整批資料難以落地。

**Independent Test**: 提供包含缺欄、格式錯誤、值域錯誤的混合資料，系統須明確區分可用列與不可用列並回報原因。

**Acceptance Scenarios**:

1. **Given** 某列缺少必要欄位值（如 `dst_ip` 或 `port`），**When** 系統解析該列，**Then** 系統必須標記該列不可用並回報具體欄位錯誤。
2. **Given** 某列 `protocol` 非 `tcp` 或 `decision` 非 `accept/deny`，**When** 系統解析該列，**Then** 系統必須拒絕該列並提供可讀錯誤原因。
3. **Given** 某列缺少 10 個關鍵欄位中的任一欄，**When** 系統解析該列，**Then** 系統必須拒絕該列並指出缺失欄位名稱。
4. **Given** 某列 `src_ip` 不屬於 `src_network_segment` 或 `dst_ip` 不屬於 `dst_network_segment`，**When** 系統解析該列，**Then** 系統必須拒絕該列並回報對應不一致欄位。

### Edge Cases

- 當輸入檔含有必要欄位名稱但大小寫或前後空白不一致時，系統仍需可正確識別欄位。
- 當輸入檔使用非 canonical 別名欄位（如 `source_ip`）而非指定欄位名時，該列需被視為缺少必要欄位。
- 當 `port` 超出有效範圍或非整數時，該列需被拒絕並回報無效埠號。
- 當 `src_network_segment` 與 `src_ip` 不一致時，該列需被標記為資料不一致。
- 當 `dst_network_segment` 與 `dst_ip` 不一致時，該列需被標記為資料不一致。
- 當整份檔案所有列都無效時，系統需回報無可用資料並停止後續流程。
- 當同一目標重複出現且背景欄位不同時，系統需保留列級語意，不得默默覆蓋來源資訊。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系統 MUST 從輸入資料列解析以下欄位：`src_ip`、`src_network_segment`、`dst_ip`、`dst_network_segment`、`service_label`、`protocol`、`port`、`decision`、`policy_id`、`reason`。
- **FR-002**: 系統 MUST 允許輸入檔包含其他非關鍵欄位，且不影響關鍵欄位解析。
- **FR-003**: 系統 MUST 將每一列解析結果維持為可追蹤的列級記錄，不得因目標相同而遺失背景欄位。
- **FR-004**: 系統 MUST 將 `dst_ip`、`port`、`protocol` 作為可執行目標資料的必要組成。
- **FR-005**: 系統 MUST 僅接受 `protocol=tcp` 的列資料；非 `tcp` 列資料必須標記為不可用。
- **FR-006**: 系統 MUST 僅接受 `decision` 為 `accept` 或 `deny` 的列資料；其他值必須標記為不可用。
- **FR-007**: 系統 MUST 在必要欄位缺失、格式錯誤或值域不合法時，為該列提供明確錯誤原因。
- **FR-008**: 系統 MUST 在輸入檔處理完成後，提供成功列數、失敗列數與失敗原因分類摘要。
- **FR-009**: 系統 MUST 在可用列存在時，讓後續流程可使用解析出的目標欄位與背景欄位。
- **FR-010**: 系統 MUST 在沒有任何可用列時，停止後續流程並回報「無可用輸入資料」。
- **FR-011**: 系統 MUST 將 `decision` 視為背景欄位而非執行門檻；`accept` 與 `deny` 兩種合法值在其他必要欄位有效時皆可建立可執行目標。
- **FR-012**: 系統 MUST 將 `src_ip`、`src_network_segment`、`dst_ip`、`dst_network_segment`、`service_label`、`protocol`、`port`、`decision`、`policy_id`、`reason` 全部視為必要欄位；缺任一欄位即拒絕該列。
- **FR-013**: 系統 MUST 在執行層以 `dst_ip + port + protocol` 去重，只執行一次掃描；同時 MUST 保留每一列原始背景欄位與該執行結果的對應關係。
- **FR-014**: 系統 MUST 驗證 `src_ip` 屬於 `src_network_segment` 且 `dst_ip` 屬於 `dst_network_segment`；任一不成立即拒絕該列。
- **FR-015**: 系統 MUST 以 canonical 欄位名作為唯一匹配標準，欄位匹配時僅允許忽略大小寫與前後空白；不支援同義別名或語意自動猜測。
- **FR-016**: 系統 MUST 維持 parser-output 與 writer-output 契約一致性：解析輸出需保留 `execution_key` 與列級對應資訊，結果輸出需保留 `service_label`、`decision`、`policy_id`、`reason` 等背景語意，不得因去重或寫出流程遺失可追溯性。

### Key Entities *(include if feature involves data)*

- **Input Traffic Row**: 輸入檔中的單一列，包含掃描來源、目標、服務語意、策略判定與原因欄位。
- **Parsed Target Record**: 從單列成功解析出的可用目標資料，至少包含 `dst_ip`、`port`、`protocol`，並附帶背景欄位。
- **Row Validation Result**: 每列解析/驗證結果，包含是否可用、錯誤原因、原始列識別資訊。

## Assumptions

- 輸入檔是列式資料，且每列代表一筆獨立的流量模擬結果。
- 欄位名稱以英文識別，核心欄位語意固定為本需求定義的 10 欄。
- 欄位名稱匹配以 canonical 欄位名為準，僅容忍大小寫與前後空白差異，不納入同義別名。
- `protocol` 預期為 `tcp`；若來源資料含其他協定，本功能僅負責拒絕並回報。
- `decision` 為策略結果，僅接受 `accept` 與 `deny`。
- 本功能不新增身份驗證、授權或額外法規流程，沿用既有資料治理與稽核規範。

## Dependencies

- 依賴資料提供方穩定提供上述關鍵欄位名稱。
- 依賴既有掃描流程可接收並使用解析後的目標與背景資料。
- 依賴既有結果呈現機制可顯示列級錯誤與摘要資訊。

## Constitution Alignment *(mandatory)*

### Test Strategy

- **TS-001**: 先定義解析成功、欄位缺失、值域錯誤、列級摘要等失敗測試，再補齊功能行為。
- **TS-002**: 更新跨模組驗證，確認解析後資料可被下游流程使用，且列級背景欄位不遺失。
- **TS-003**: 若端到端流程已依賴舊輸入欄位假設，需新增或更新整體流程情境；若不需更新，需明確記錄理由。

### Observability Strategy

- **OS-001**: 定義並輸出可讀的解析統計：總列數、成功列數、失敗列數、主要失敗原因。
- **OS-002**: 提供可追蹤的列級錯誤資訊，讓操作人員可直接定位來源資料問題。

### Release Strategy

- **RS-001**: 預期產品版本為 `MINOR`，理由為新增對高欄位輸入資料的可用能力且維持既有核心流程定位。
- **RS-002**: 更新 `docs/release-notes/` 新增一條本功能的能力說明、限制條件與操作影響。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 對於包含必要欄位的輸入檔，100% 可用列可在單次執行中被成功解析為可執行目標。
- **SC-002**: 對於不合法資料列，100% 皆可提供對應欄位與原因的可讀錯誤訊息。
- **SC-003**: 以 `tests/integration/testdata/rich_input/invalid_rows.csv` 為固定資料集進行驗證時，操作人員自「取得解析摘要」起算 10 分鐘內，能完成至少 90% 原始錯誤列的修正並於重跑後轉為可用列；計算方式為 `修正成功列數 / 原始錯誤列數`，量測證據 MUST 記錄於 `specs/001-enhance-input-parser/quickstart.md`。
- **SC-004**: 以混合有效/無效固定資料集進行驗證時，主要流程任務完成率 MUST >= 95%，計算方式為 `實際成功執行的 execution_key 數 / 由有效列推導的預期 execution_key 數`；量測證據 MUST 記錄於 `specs/001-enhance-input-parser/quickstart.md`。
