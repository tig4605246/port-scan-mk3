---
layout: two-cols
---

<div class="eyebrow">Why This Design</div>

# 為什麼不是更簡單，或更複雜？

::left::

### 選擇這個設計，因為它剛好回應 4 類壓力

- `fail-fast validation` 避免把資料問題遞延到 runtime
- `task expansion + execution key` 讓工作單位可推導、可去重
- `pressure-aware control` 讓 pacing 成為顯式控制點
- `writer + state` 讓輸出與恢復成為架構責任，而不是附帶腳本

::right::

### 沒有選擇的方向

- 不把所有流程硬塞回 `main.go`
- 不把 orchestration 做成 god object
- 不把壓力控制外推到不透明 shell wrapper
- 不導入超過現階段所需的 queue / broker / distributed worker

<div class="rationale">
  這頁的結論是：目前的複雜度不是 accidental complexity，而是對應需求後的最小可接受結構。
</div>

<!--
review 常見問題是「這個專案有必要切這麼多 package 嗎」或「為什麼不再簡化」。
這頁直接回答：因為這些能力如果留在隱含流程裡，reviewer 就無法清楚判斷風險控制點。
-->
