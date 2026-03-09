---
layout: default
---

<div class="eyebrow">Architecture</div>

# 系統總覽：責任邊界先於實作細節

<div class="figure-frame">
  <img src="/diagrams/architecture-overview.svg" alt="Architecture overview" />
</div>

<div class="rationale">
  關鍵判斷：`cmd/port-scan` 只保留 CLI glue，真正可演化的邏輯移入 `pkg/`；這讓 contract 保持穩定，內部則能按責任拆分。
</div>

<!--
這頁是 review 的主架構圖。
重點不是 package 名稱，而是依賴方向：CLI 只做 composition root；可重用、可測試、可替換的邏輯都留在 pkg。
-->
