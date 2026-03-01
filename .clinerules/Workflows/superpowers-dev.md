# Superpowers Development Workflow

這是一套基於 Superpowers 方法論的標準軟體開發流程。請嚴格按照以下步驟循序執行，並在每一步調用對應的 Skill：

1. **Step 1: 需求與架構釐清 (調用 Skill: `brainstorming`)**
   - 任務：在開始寫程式碼之前，先向使用者提問以完善想法。探索替代方案，並將設計分為小區塊呈現給使用者確認。
   - 檢查點：必須產出設計文件並獲得使用者核准，才能進入下一步。

2. **Step 2: 建立隔離環境 (調用 Skill: `using-git-worktrees`)**
   - 任務：設計確認後，建立新的 Git branch 與隔離的工作區，設定專案並確保初始測試基準為乾淨狀態。

3. **Step 3: 制定實作計畫 (調用 Skill: `writing-plans`)**
   - 任務：將工作拆解為微型任務（每個約 2-5 分鐘的開發量）。每個任務需包含確切的檔案路徑、程式碼構想與驗證步驟。

4. **Step 4: 驅動開發與測試 (調用 Skills: `subagent-driven-development` & `test-driven-development`)**
   - 任務：針對計畫中的每個子任務執行開發。嚴格遵守 RED-GREEN-REFACTOR 原則：先寫會失敗的測試，接著寫最少量的程式碼讓測試通過，然後重構。未寫測試前禁止撰寫業務邏輯。

5. **Step 5: 程式碼審查 (調用 Skill: `requesting-code-review`)**
   - 任務：在任務完成段落間，對照計畫進行自我審查，報告嚴重程度問題。阻斷性問題必須立刻修復。

6. **Step 6: 完成分支與合併 (調用 Skill: `finishing-a-development-branch`)**
   - 任務：所有任務完成後，驗證所有測試。向使用者提供後續選項（合併 / 發布 PR / 保留 / 捨棄），並清理工作區。