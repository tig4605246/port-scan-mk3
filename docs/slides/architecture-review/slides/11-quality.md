---
layout: two-cols
---

<div class="eyebrow">Evidence</div>

# 設計合理性的佐證，不只靠口頭說明

::left::

### 驗證鍊

- `go test ./...`
- `bash scripts/coverage_gate.sh`
- `bash e2e/run_e2e.sh`
- timestamped outputs 與 `resume_state` artifacts

::right::

### review 意義

- 確認 boundary 拆分後 contract 沒偷偷改
- 確認 happy path 與 failure path 都有 evidence
- 確認輸出檔與 runtime state 可被外部檢查
- 確認這不是只存在於設計圖上的 architecture

<div class="rationale">
  architecture review 最終仍要落到 evidence。這套設計之所以站得住腳，是因為關鍵控制點都有對應驗證與 artifact。
</div>

<!--
如果 reviewer 擔心這只是漂亮圖稿，這頁就是回應：這些 boundary 與流程都有可執行的驗證鍊支撐。
-->
