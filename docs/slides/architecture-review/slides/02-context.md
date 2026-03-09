---
layout: default
---

<div class="eyebrow">Review Scope</div>

# 我們在審什麼

<div class="review-grid">
  <div class="review-card">
    <h3>目標能力</h3>
    <ul class="tight">
      <li>TCP port scan 的 deterministic pipeline</li>
      <li>fail-fast input validation</li>
      <li>pressure-aware pacing 與 pause / resume</li>
      <li>可追蹤輸出與 resumable state</li>
    </ul>
  </div>
  <div class="review-card">
    <h3>不是這次的目標</h3>
    <ul class="tight">
      <li>不是 distributed scanning 平台</li>
      <li>不是 raw packet / SYN scanner</li>
      <li>不是為了極致吞吐而犧牲可審查性</li>
      <li>不是把所有邏輯都塞進單一 CLI entrypoint</li>
    </ul>
  </div>
</div>

<div class="signal-list">
  <div class="signal"><strong>Requirement</strong> 輸入錯誤要盡早失敗</div>
  <div class="signal"><strong>Requirement</strong> 執行中斷要可恢復</div>
  <div class="signal"><strong>Requirement</strong> 掃描節奏要能被外部壓力調節</div>
  <div class="signal"><strong>Requirement</strong> 結果與 runtime 行為都要可驗證</div>
</div>

<!--
這一頁先把 review 邊界說清楚。
如果需求本身不需要分散式、raw packet 或極端效能優化，那架構就不應該提前為那些能力付出複雜度成本。
-->
