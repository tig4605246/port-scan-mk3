---
layout: default
---

<div class="eyebrow">Runtime Flow</div>

# Happy Path：資料流與控制流都必須可追蹤

<div class="figure-frame">
  <img src="/diagrams/happy-path.svg" alt="Happy path data flow" />
</div>

<div class="rationale">
  reviewer 要看到的是：從輸入到產物，每一段責任都有落點，沒有把 validation、dispatch、probe、artifact generation 混成一團。
</div>

<!--
這裡可以順著圖走一次 happy path。
也要特別指出 scanner 的邊界是 Go 的 TCP dial/close 模型，而不是 raw packet 模式。
-->
