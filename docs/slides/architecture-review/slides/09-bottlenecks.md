---
layout: default
---

<div class="eyebrow">Bottlenecks</div>

# 瓶頸不只在網路，也在控制與產物

<div class="review-grid">
  <div class="review-card">
    <h3>Probe throughput</h3>
    <p>`workers`、`timeout`、`delay` 直接決定 dial concurrency 與等待成本。</p>
  </div>
  <div class="review-card">
    <h3>Pressure gate</h3>
    <p>外部 pressure signal 讓吞吐量不能只靠固定 worker 數，必須有顯式 pause / resume 控制。</p>
  </div>
  <div class="review-card">
    <h3>Output I/O</h3>
    <p>結果寫檔與 open-only split 會影響 runtime 節奏，不能當成最後再 dump 的附加流程。</p>
  </div>
  <div class="review-card">
    <h3>Resume state</h3>
    <p>checkpoint 粒度太粗會浪費重跑成本，太細又增加同步負擔，所以要維持可推導的 state 邊界。</p>
  </div>
</div>

<div class="rationale">
  這也是為什麼 architecture 裡需要 `speedctrl`、`writer`、`state` 這些看似不是核心掃描的元件；它們其實在控制 throughput 與 operational safety。
</div>

<!--
reviewer 會看設計是否只顧主流程，不顧瓶頸與 side effect。
這頁要強調：瓶頸是整體 runtime system 的問題，不只是 scanner function 的問題。
-->
