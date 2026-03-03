# Data Model: IP-Aware Baseline Specification

## 1. Entities

### CIDR Input Row

| Field | Type | Required | Description |
|---|---|---|---|
| `row_number` | integer | Yes | 原始 CSV 資料列號（1-based, 含表頭偏移） |
| `ip_selector` | string | Yes | `ip` 欄位內容；可為單一 IPv4 或 IPv4 CIDR |
| `ip_cidr` | string | Yes | 該列目標必須落入的邊界 CIDR |
| `ip_col_name` | string | Yes | 使用者指定的目標欄位名 |
| `ip_cidr_col_name` | string | Yes | 使用者指定的邊界欄位名 |
| `metadata` | map<string,string> | No | 其餘業務欄位（如 fab/cidr_name） |

Validation rules:
- 欄位名必須區分大小寫精確匹配。
- 同名欄位重複屬 fatal。
- `ip_selector` 與 `ip_cidr` 必須可解析；不可解析屬 fatal。

### Target Group

| Field | Type | Required | Description |
|---|---|---|---|
| `group_key` | string | Yes | `ip_cidr`，作為分組唯一鍵 |
| `target_ips` | string[] | Yes | 由該組 `ip_selector` 展開且去重後的 IPv4 集合 |
| `ports` | string[] | Yes | 由 port CSV 載入的 TCP port 規格 |
| `next_index` | integer | Yes | 下次待派發任務索引 |
| `scanned_count` | integer | Yes | 已完成任務數 |
| `total_count` | integer | Yes | 總任務數 = `len(target_ips) * len(ports)` |
| `status` | enum | Yes | `pending` / `scanning` / `completed` |

Validation rules:
- 不同 `group_key` 之間不得重疊。
- 同一 `group_key` 內不同 selector 展開結果不得互相重疊。

### Scan Task

| Field | Type | Required | Description |
|---|---|---|---|
| `chunk_cidr` | string | Yes | 所屬 `Target Group` |
| `ip` | string | Yes | 單一目標 IPv4 |
| `port` | integer | Yes | 單一 TCP port |
| `index` | integer | Yes | group-local 線性索引 |

### Scan Result Record

| Field | Type | Required | Description |
|---|---|---|---|
| `ip` | string | Yes | 掃描目標 IPv4 |
| `ip_cidr` | string | Yes | 所屬目標邊界 |
| `port` | integer | Yes | 掃描 port |
| `status` | enum | Yes | `open` / `close` / `close(timeout)` |
| `response_time_ms` | integer | Yes | 響應時間（毫秒） |
| `fab_name` | string | No | 業務欄位 |
| `cidr_name` | string | No | 業務欄位 |

### Output Batch

| Field | Type | Required | Description |
|---|---|---|---|
| `batch_ts_utc` | string | Yes | `YYYYMMDDTHHMMSSZ` |
| `batch_seq` | integer | Yes | 同秒衝突序號；無衝突時可視為 0 |
| `scan_results_path` | string | Yes | `scan_results-<ts>[-n].csv` |
| `opened_results_path` | string | Yes | `opened_results-<ts>[-n].csv` |

### Resume State Entry

| Field | Type | Required | Description |
|---|---|---|---|
| `cidr` | string | Yes | 分組鍵（對應 `ip_cidr`） |
| `cidr_name` | string | No | 顯示用途 |
| `ports` | string[] | Yes | 該組 ports |
| `next_index` | integer | Yes | 待續跑索引 |
| `scanned_count` | integer | Yes | 已掃描數 |
| `total_count` | integer | Yes | 總任務數 |
| `status` | enum | Yes | `pending` / `scanning` / `completed` |

## 2. Relationships

- `CIDR Input Row` N:1 `Target Group`（以 `ip_cidr` 分組）。
- `Target Group` 1:N `Scan Task`。
- `Scan Task` 1:1 `Scan Result Record`。
- `Target Group` 1:1 `Resume State Entry`（每次保存時更新）。
- `Output Batch` 1:N `Scan Result Record`（按檔案落地批次聚合）。

## 3. State Transitions

### Target Group status

- `pending` -> `scanning`: 開始派發該組任務。
- `scanning` -> `completed`: `next_index == total_count` 且 in-flight 任務清空。
- `scanning` -> `pending`: 中斷保存後待下次續跑。

### Scan lifecycle

- `initialized` -> `running`: 通過 fail-fast 驗證。
- `running` -> `paused`: `api_paused || manual_paused`。
- `paused` -> `running`: `api_paused == false && manual_paused == false`。
- `running|paused` -> `stopped`: SIGINT 或 API 連續第 3 次失敗。
- `stopped` -> `running`: 帶 `-resume` 重新啟動。
