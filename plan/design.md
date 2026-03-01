# Port Scan Tool 設計文檔

## 1. 專案概述

- **專案名稱**: TCP Port Scanner
- **專案類型**: Golang CLI 工具
- **核心功能**: 對指定 CIDR 範圍和 Port 進行 TCP 連接掃描，支援速率限制和全局流量控制
- **目標用戶**: 網路管理員、安全工程師

---

## 2. 命令行參數

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-cidr-file` | CIDR CSV 檔案路徑 (含 fab_name,cidr,cidr_name) | 必填 |
| `-port-file` | Port CSV 檔案路徑 (無表頭，格式 80/tcp) | 必填 |
| `-output` | 輸出 CSV 檔案路徑 | scan_results.csv |
| `-timeout` | TCP 連接超時時間 (如 2s, 100ms) | 100ms |
| `-delay` | 每次掃描間隔時間 | 10ms |
| `-bucket-rate` | Leaky Bucket 速率 (每秒目標數) | 100 |
| `-bucket-capacity` | Leaky Bucket 容量 | 100 |
| `-workers` | Worker 數量 | 10 |
| `-pressure-api` | 路由器壓力 API URL | http://localhost:8080/api/pressure |
| `-pressure-interval` | 獲取壓力間隔 (秒) | 5 |
| `-disable-api` | 關閉 API 壓力檢測模式 (僅依賴手動暫停) | false |
| `-resume` | 從上次 SIGINT 停止程式後的斷點繼續進行掃描 | (無) |
| `-log-level` | 日誌輸出層級 (debug, info, error) | info |

---

## 3. 輸入解析與防呆校驗 (Fail-Fast Validation)

在程式啟動並讀取完 CSV 檔案後，必須執行嚴格的輸入防呆校驗。若校驗失敗，程式應立即中止 (Abort) 並提示使用者修正，絕不隱式忽略錯誤。

### 3.1 CIDR 重疊檢查

- 掃描任務開始前，將所有輸入的 CIDR 轉換為 `net.IPNet` 進行兩兩比對。
- **判斷邏輯**：若 `Network_A` 包含 `Network_B` 的 Base IP，或反之，則視為重疊。
- **處理方式**：一旦發現重疊（例如 `10.0.0.0/8` 與 `10.1.1.0/24`），立刻觸發 Fatal Error 終止程式，並於終端機印出衝突的 CIDR 組合與名稱，要求使用者修正設定檔。

### 3.2 輸入檔案格式

#### 輸入 - CIDR CSV (有表頭)

```csv
fab_name,cidr,cidr_name
fab1,192.168.1.0/24,office-network
fab2,10.0.0.0/8,datacenter

```

#### 輸入 - Port CSV (無表頭)

```csv
80/tcp
443/tcp
22/tcp
8080/tcp

```

#### 輸出 - 結果 CSV

```csv
ip,port,status,response_time_ms,fab_name,cidr,cidr_name
192.168.1.1,80,open,12,fab1,192.168.1.0/24,office-network
192.168.1.1,443,close(timeout),0,fab1,192.168.1.0/24,office-network

```

**Status 欄位值**:

* `open`: 連接成功
* `close`: 連接被拒絕
* `close(timeout)`: 連接超時

### 3.4 輸出 - 斷點狀態檔案 (resume_state.json)

用於 `-resume` 參數，記錄每個 CIDR Chunk 的掃描進度。

```json
[
  {
    "cidr": "192.168.1.0/24",
    "cidr_name": "office-network",
    "ports": ["80/tcp", "443/tcp"],
    "next_index": 512,
    "scanned_count": 512,
    "total_count": 512,
    "status": "completed"
  },
  {
    "cidr": "10.0.0.0/8",
    "cidr_name": "datacenter",
    "ports": ["80/tcp", "443/tcp"],
    "next_index": 1500,
    "scanned_count": 1500,
    "total_count": 33554432,
    "status": "scanning"
  }
]

```

---

## 4. 架構設計

### 4.1 整體架構

```
┌───────────────────────────────────────────────────────────────┐
│                       Main Controller                         │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐  │
│  │ CSV Reader   │  │ Rate Limiter │  │ Global Speed Ctrl   │  │
│  │ (CIDR+Port)  │──│ (Leaky Bucket│──│ (Router Pressure &   │  │
│  └──────────────┘  │ per CIDR)    │  │  Broadcast Gate)    │  │
│                    └──────────────┘  └─────────────────────┘  │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ Task Generator (Handles Chunking & Waits for Speed Ctrl)│  │
│  └─────────────────────────────────────────────────────────┘  │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │ State Manager (Graceful Shutdown & JSON Chunk Save)     │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
                              │ (Global Task Queue)
              ┌───────────────┴───────────────┐
              ▼                               ▼
    ┌──────────────────┐            ┌──────────────────┐
    │     Worker 1     │    ...     │     Worker N     │
    └──────────────────┘            └──────────────────┘
              │                               │
              ▼                               ▼
    ┌──────────────────┐            ┌──────────────────┐
    │ TCP Scanner      │            │ TCP Scanner      │
    │ (net.DialTimeout)│            │ (net.DialTimeout)│
    └──────────────────┘            └──────────────────┘
              │                               │
              ▼ (Result Queue)                ▼
    ┌─────────────────────────────────────────────────────────┐
    │ Result Writer (Single Goroutine to CSV)                 │
    └─────────────────────────────────────────────────────────┘

```

### 4.2 核心組件

1. **CSV Reader**: 解析 CIDR 和 Port CSV 檔案。
2. **Task Generator**: 將輸入載入為 Chunk 結構，並將 IP × Port 展開為 Task 放入全局 Task Queue。負責監聽 Speed Controller 的暫停信號。
3. **Leaky Bucket Rate Limiter**: 每個 CIDR 獨立的速率限制器。
4. **Global Speed Controller**: 定時獲取路由器壓力，並透過**廣播閘門 (Broadcast Gate)** 動態控制 Task Generator 派發任務的速度。
5. **Worker Pool**: 並發執行 TCP 掃描，純粹的消費者，不負責暫停邏輯。
6. **TCP Scanner**: 使用 `net.DialTimeout` 發起 TCP 連接，並保證完整的 4-way handshake 關閉。
7. **Result Writer**: 獨立 Goroutine 接收掃描結果並寫入 CSV，避免併發寫入衝突。
8. **State Manager**: 處理 SIGINT 信號，協調優雅降落並儲存斷點 JSON。

### 4.3 斷點繼續 (Resume 機制)

掃描可能依需求中斷，當收到 `SIGINT` (Ctrl+C) 時，必須透過**優雅降落 (Graceful Shutdown)** 將當前進度存成人類可讀的 JSON 檔案 (`resume_state.json`)。下次繼續時使用 `-resume` flag 接續掃描。

* **Chunk 結構設計**：每次只拿出一個 CIDR 與其對應的 Ports 建成一個 Chunk struct。
* **斷點紀錄方式 (`NextIndex`)**：將該 Chunk 內的 IP × Port 組合視為一維陣列，記錄下一個準備派發的陣列索引 `NextIndex`。
* **SIGINT 中斷處理流程 (防亂序與防漏掃)**：
1. **停止派發**：攔截到 SIGINT 後，觸發 `context.Cancel()` 通知 Task Generator 停止產生新任務並退出迴圈。
2. **記錄游標**：Task Generator 退出時，將當下未派發的目標索引值更新至該 Chunk 的 `NextIndex`。
3. **等待消化 (Wait In-flight)**：呼叫 `sync.WaitGroup.Wait()`，等待 Worker 把已經從 Queue 拿出來的任務執行完畢。
4. **輸出 JSON**：所有 Worker 停工後，將所有 Chunk struct 陣列寫出為 `resume_state.json`。


* **接續掃描**：使用 `-resume` 時，直接讀取 JSON，找出狀態為 `scanning` 或 `pending` 的 Chunk，並從該 Chunk 的 `NextIndex` 繼續計算 IP 與 Port 並派發。嚴格保證不重複掃描且不遺漏。

### 4.4 進度條與當前資訊

掃描中必須要能顯示當下正在 scan 的 CIDR 與 port 資訊，以及全局速度控制器的狀態 (speed limit, 路由器 pressure 等)。

---

## 5. Leaky Bucket 速率限制器

* 每個 CIDR 分配獨立的 Leaky Bucket
* 參數：
* `rate`: 每秒處理數 (預設 100)
* `capacity`: 桶大小 (預設 100)


* 實現：使用 Go channel + time.Ticker

---

## 6. Worker Pattern

* 使用 Go channel 實現生產者-消費者模式，統一從全局 Task Queue 獲取任務。
* 任務池：緩衝 channel，容量建議設為 Worker 數量的 1~2 倍。
* Worker 數量：可通過 `-workers` 配置。

---

## 7. TCP 掃描邏輯與系統資源配置

1. 使用 `net.DialTimeout()` 發起 TCP 連接。
2. 記錄連接開始時間。
3. 如果連接成功：
* 記錄響應時間
* 呼叫 `conn.Close()` 觸發標準的 4-way handshake 關閉連接。
* 狀態記為 "open"


4. 如果連接超時：
* 狀態記為 "close(timeout)"


5. 如果連接被拒絕：
* 狀態記為 "close"



**系統資源配置要求 (應對 TIME_WAIT)**：
為避免高併發產生過多 `TIME_WAIT` 狀態耗盡本地 Port 資源，部署環境 (Linux) 建議配置以下 `sysctl` 參數：

* `sysctl -w net.ipv4.ip_local_port_range="1024 65535"`
* `sysctl -w net.ipv4.tcp_tw_reuse=1`
* `sysctl -w net.ipv4.tcp_fin_timeout=15`

---

## 8. 全局速度控制器 (優雅暫停與雙重控制)

系統支援「API 自動控制」與「手動鍵盤控制」兩種方式來觸發廣播閘門 (Broadcast Gate)，以實現優雅暫停 (Graceful Pause)。

### 8.1 雙重控制來源

1. **API 自動控制 (預設啟用)**：
   - 定時 (預設 5 秒) 從 API 獲取路由器壓力 (`GET {pressure-api}`)。
   - **關閉機制**：可透過 `-disable-api=true` 旗標關閉。關閉後，背景將不會啟動 API 輪詢 Goroutine。
2. **手動鍵盤控制 (永久啟用)**：
   - 背景啟動按鍵監聽 Goroutine (需將 Terminal 設為 Raw Mode 以攔截即時按鍵)。
   - 使用者可隨時按下 **空白鍵 (Space)** 切換「暫停/繼續」狀態。
   - 即使 API 控制被關閉，此功能依然完全生效。

### 8.2 狀態判定與阻塞邏輯 (OR-Gate Logic)

Speed Controller 內部維護兩個獨立的暫停旗標：`api_paused` 與 `manual_paused`。

- **暫停條件 (Pause)**：當 `api_paused == true` 或 `manual_paused == true` 任一條件成立時，重建阻塞 Channel，Task Generator 停止派發任務。
- **恢復條件 (Resume)**：必須滿足 `api_paused == false` 且 `manual_paused == false` 雙雙成立時，才會關閉阻塞 Channel (放行)。
- **保護情境**：
  - 若 API 偵測到設備壓力過大 (`api_paused=true`)，即使使用者狂按空白鍵 (`manual_paused=false`)，系統依然保持暫停，以保護實體網路設備不被掃描流量打掛。
  - 若網路正常 (`api_paused=false`)，使用者按空白鍵 (`manual_paused=true`) 即可隨時讓掃描停下。

### 8.3 優雅降落 (Graceful Pause) 行為

- 當進入暫停狀態時，**Task Generator 將被阻塞**，停止從 Chunk 中取出新任務。
- **Worker 不會被阻塞**，會繼續消化 Task Queue 中剩餘的任務，並**讓正在進行中 (In-flight) 的 TCP 握手自然完成（成功、超時或被拒絕）**，避免留下半開連接 (Half-open connections)。
- 狀態切換時，依據 `-log-level info` 輸出明確提示：
  - `[API] 路由器壓力過載，掃描已自動暫停`
  - `[Manual] 接收到按鍵指令，掃描已手動暫停`
- 結合 `context.Context`，若在暫停期間收到 SIGINT (Ctrl+C)，能強制打破阻塞並正確保存 `NextIndex` 斷點至 JSON，絕不遺漏任務。

---

## 9. 錯誤處理與日誌

系統將依據 `-log-level` 參數決定輸出詳細程度：

| 錯誤/事件類型 | 嚴重級別 | 處理方式 / 紀錄層級 |
|----------|----------|----------|
| CIDR 重疊衝突 | Fatal | 立即終止程式，輸出衝突的網段，要求使用者修正設定檔。 |
| CSV 檔案不存在 | Fatal | 輸出錯誤訊息並終止程式。 |
| File Descriptor 耗盡 | Fatal | 啟動前檢查 `ulimit -n`，若過低則提示使用者提升並終止。 |
| API 獲取失敗 | Error | 連續失敗 1~2 次記錄為 Error，使用上次成功值；連續失敗 3 次則觸發 Fatal 終止程式。 |
| 連接超時 / 拒絕 | Debug | 不印出 Error，僅在 Debug 模式下可查看底層 TCP 錯誤。結果交由 Result Writer 寫入 CSV。 |
| 路由器壓力過大暫停 | Info | 記錄 `Global Speed Controller: Paused scanning`，恢復時亦記錄 Info。 |
| 斷點儲存與接續 | Info | 記錄成功載入的斷點 Chunk 與 SIGINT 觸發時成功寫入的狀態。 |

---

## 10. 測試策略

1. **單元測試**: 測試 Leaky Bucket、CSV Parser、IP 展開與 Chunk 索引計算邏輯。
2. **整合測試**: 測試完整的掃描流程、Graceful Shutdown 行為與斷點 JSON 生成。
3. **Mock Server**: 模擬路由器壓力 API。
4. **e2e test**: 測試完整使用情境，環境以 docker compose 搭建，必須確保完全與真實網路隔離。測試必須產出一頁式 report (HTML 與 plain text 兩種輸出)。