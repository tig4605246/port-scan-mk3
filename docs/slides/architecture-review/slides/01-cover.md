---
layout: cover
---

<div class="eyebrow">Architecture Review</div>

# port-scan-mk3

## 從需求壓力到設計取捨

<div class="review-grid">
  <div class="review-card">
    <h3>這次要回答的問題</h3>
    <p>不是「功能做了什麼」，而是「為什麼這樣切分責任、控制風險、保護 operator contract」。</p>
  </div>
  <div class="review-card">
    <h3>Review 焦點</h3>
    <p>需求、架構、瓶頸、trade-offs，以及這些選擇是否與專案規模與風險相稱。</p>
  </div>
</div>

<div class="score-strip">
  <div class="score-box">
    <strong>Context</strong>
    輸入品質、掃描壓力、可恢復性
  </div>
  <div class="score-box">
    <strong>Architecture</strong>
    CLI glue、package boundary、runtime control
  </div>
  <div class="score-box">
    <strong>Decision</strong>
    合理性優先，不為複雜而複雜
  </div>
</div>

<!--
本 deck 用 architecture review 的角度來看 port-scan-mk3。
我要回答的是：在目前專案範圍下，這套設計是否足夠清楚、足夠可維護，並且能對應實際的 operational constraints。
-->
