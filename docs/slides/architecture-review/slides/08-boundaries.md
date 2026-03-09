---
layout: two-cols
---

<div class="eyebrow">Boundary Rationale</div>

# 為什麼這樣切 package boundary

::left::

### 保護的邊界

- `cmd/port-scan`
  - 只做 command routing、參數解析、使用者 I/O
- `pkg/input` / `pkg/task`
  - 把資料品質與任務建模切乾淨
- `pkg/scanapp`
  - 只做 orchestration，不再回長成 god file
- `pkg/writer` / `pkg/state`
  - 把 artifact contract 與 resume contract 顯式化

::right::

### 這樣切的理由

- boundary 對應 reviewer 可檢查的責任
- CLI contract 與 runtime implementation 解耦
- 測試可以對準 collaborator，而不是全都靠 e2e
- 後續調整 pacing、resume、writer 時，影響面更可控

<div class="rationale">
  架構切分不是為了抽象，而是為了讓「改哪裡、風險在哪裡、誰擁有責任」可被 reviewer 一眼判斷。
</div>

<!--
這頁可以連回最近的 SOLID refactor 文件。
核心論點是：邊界命名要對應風險與責任，不是對應程式碼行數。
-->
