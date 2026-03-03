# Phase 0 Research: Enhanced Input Field Parsing

## Decision 1: 欄位匹配採 canonical 名稱 + 大小寫/空白正規化

- **Decision**: 欄位識別僅接受 10 個 canonical 欄位名；比對時忽略大小寫與前後空白，不支援同義別名。
- **Rationale**: 在不犧牲契約穩定性的前提下，容忍常見資料清洗問題（大小寫與空白），同時避免別名造成誤映射與維護成本膨脹。
- **Alternatives considered**:
  - 嚴格逐字匹配：過於脆弱，容易因格式瑕疵失敗。
  - 支援多別名：彈性高但歧義與回歸成本高。
  - 自動語意猜測：不可預測且不利驗證。

## Decision 2: 10 個關鍵欄位全部必填，缺任一欄拒絕該列

- **Decision**: `src_ip`、`src_network_segment`、`dst_ip`、`dst_network_segment`、`service_label`、`protocol`、`port`、`decision`、`policy_id`、`reason` 任一缺失即列級失敗。
- **Rationale**: 規格已明確要求完整背景語意；部分欄位缺失會導致後續判讀與稽核斷裂。
- **Alternatives considered**:
  - 只保留執行必要欄位：會遺失策略背景。
  - 以預設值補空缺：會引入資料誤導。

## Decision 3: `decision` 為背景欄位，不作執行阻擋

- **Decision**: `decision` 僅接受 `accept|deny`，且兩者皆可產生可執行目標。
- **Rationale**: 需完整保留流量模擬觀點，避免把 `deny` 訊息排除於掃描語境外。
- **Alternatives considered**:
  - 只執行 `accept`：減少掃描量但丟失 `deny` 目標的驗證可見性。
  - `deny` 使整批失敗：過度保守、破壞可用性。

## Decision 4: 執行層按 `dst_ip + port + protocol` 去重，保留列級映射

- **Decision**: 掃描只對每個 execution key 執行一次，並維持「來源列 -> execution key -> 掃描結果」映射。
- **Rationale**: 平衡執行效率與審計可追蹤性，避免重複掃描同一目標。
- **Alternatives considered**:
  - 不去重：效率差且可能放大 timeout/壓力。
  - 只保留第一列：背景資訊遺失。

## Decision 5: IP 與 network segment 必做雙向一致性驗證

- **Decision**: 同時驗證 `src_ip ∈ src_network_segment` 與 `dst_ip ∈ dst_network_segment`。
- **Rationale**: 來源與目標語意都會影響策略判讀，不可只驗證其中一方。
- **Alternatives considered**:
  - 只驗證目標：來源語意可能失真。
  - 只記錄不阻擋：會把錯誤資料帶入執行。

## Decision 6: 高欄位/大檔解析採單次串流處理假設（100k rows scope）

- **Decision**: 以單次串流讀取與列級即時驗證為設計基線，目標支援 100k 列輸入並輸出完整摘要。
- **Rationale**: 對 CLI 掃描前置解析而言，串流可控、記憶體壓力低，且符合現有 Go 標準庫模式。
- **Alternatives considered**:
  - 全量載入記憶體後再驗證：簡單但在大檔下記憶體成本高。
  - 多階段離線預處理：流程複雜、操作成本上升。

## Decision 7: 測試與 gate 策略

- **Decision**: 以 `go test ./...` 與 coverage gate 為必跑；若 parser 變更影響 scan pipeline 分派行為，必跑 e2e gate。
- **Rationale**: 符合憲章品質門檻，並避免 parser 合約變更破壞下游流程。
- **Alternatives considered**:
  - 只跑單元測試：無法覆蓋邊界整合風險。
  - 每次都跑完整 e2e：成本高但可在任務層做條件化觸發。
