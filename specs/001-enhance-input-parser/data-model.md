# Data Model: Enhanced Input Field Parsing

## Entity: InputTrafficRow

- **Description**: 原始輸入檔中的單一資料列。
- **Fields**:
  - `row_index` (int, required): 來源列號（1-based data row）
  - `src_ip` (string, required)
  - `src_network_segment` (string, required)
  - `dst_ip` (string, required)
  - `dst_network_segment` (string, required)
  - `service_label` (string, required)
  - `protocol` (string, required, allowed: `tcp`)
  - `port` (int, required, range: 1..65535)
  - `decision` (string, required, allowed: `accept|deny`)
  - `policy_id` (string, required)
  - `reason` (string, required)

## Entity: RowValidationResult

- **Description**: 每列解析與驗證結果。
- **Fields**:
  - `row_index` (int, required)
  - `status` (enum, required): `accepted` | `rejected`
  - `error_codes` ([]string, optional): 失敗時的標準化錯誤碼集合
  - `error_message` (string, optional): 人類可讀錯誤摘要
  - `normalized_headers` (map[string]string, required): canonical -> original header 映射

## Entity: ParsedTargetRecord

- **Description**: 由驗證成功列產生的可執行目標資料（含背景欄位）。
- **Fields**:
  - `row_index` (int, required)
  - `execution_key` (string, required): `{dst_ip}:{port}/{protocol}`
  - `dst_ip` (string, required)
  - `port` (int, required)
  - `protocol` (string, required)
  - `src_ip` (string, required)
  - `src_network_segment` (string, required)
  - `dst_network_segment` (string, required)
  - `service_label` (string, required)
  - `decision` (string, required)
  - `policy_id` (string, required)
  - `reason` (string, required)

## Entity: ExecutionTarget

- **Description**: 執行層唯一掃描目標（去重後）。
- **Fields**:
  - `execution_key` (string, required, unique)
  - `dst_ip` (string, required)
  - `port` (int, required)
  - `protocol` (string, required)
  - `source_rows` ([]int, required): 對應來源 row index 清單

## Entity: ParseSummary

- **Description**: 解析完成後摘要資訊。
- **Fields**:
  - `total_rows` (int, required)
  - `accepted_rows` (int, required)
  - `rejected_rows` (int, required)
  - `error_buckets` (map[string]int, required): 錯誤類型統計
  - `deduped_execution_targets` (int, required)

## Relationships

- `InputTrafficRow` 1 -> 1 `RowValidationResult`
- `InputTrafficRow` 1 -> 0..1 `ParsedTargetRecord`
- `ParsedTargetRecord` N -> 1 `ExecutionTarget`
- `ExecutionTarget` 1 -> N `InputTrafficRow` (透過 `source_rows`)

## Validation Rules

1. Header matching must use canonical names with case/trim normalization only.
2. All 10 key fields are mandatory; missing any field rejects row.
3. `protocol` must be `tcp`.
4. `decision` must be `accept` or `deny`.
5. `port` must be integer in [1, 65535].
6. `src_ip` must belong to `src_network_segment`.
7. `dst_ip` must belong to `dst_network_segment`.

## State Transitions

### Row lifecycle

`raw_row` -> `headers_normalized` -> `parsed` -> `validated`

- If validation passes: `validated` -> `accepted` -> `mapped_to_execution_target`
- If validation fails: `validated` -> `rejected`

### Execution target lifecycle

`collected_from_accepted_rows` -> `deduplicated_by_execution_key` -> `dispatched_once`
