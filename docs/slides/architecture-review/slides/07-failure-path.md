---
layout: default
---

<div class="eyebrow">Failure Control</div>

# 失敗路徑：這個設計把 recovery 當成第一級責任

<div class="figure-frame">
  <img src="/diagrams/failure-recovery.svg" alt="Failure and recovery paths" />
</div>

<div class="rationale">
  設計合理性的核心在於：validation fail、pressure escalation、cancellation 這三種風險都有明確的停止與恢復語意，而不是由 operator 自行猜測狀態。
</div>

<!--
architecture review 不只看 happy path。
真正能拉開品質差距的是失敗時有沒有明確語意：何時停止、何時保存狀態、如何 rerun。
-->
