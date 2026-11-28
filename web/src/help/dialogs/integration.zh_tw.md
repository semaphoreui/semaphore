---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# 編輯執行

執行是當其他服務呼叫 Semaphore 的 webhook URL 時（例如 GitHub 的 push 事件）整合所啟動的內容：任務範本的一個任務，或一個 workflow。URL
與請求的驗證屬於整合的端點，而不屬於執行。執行的頁面包含它的比對器以及它從請求中擷取的值。

## 名稱 {#integrationForm-name}

整合在「整合」頁面上列出時所用的名稱。由此整合啟動的任務，在稽核日誌中會歸屬於它。此欄位為必填。

## 整合 {#integrationForm-groupId}

執行所屬的整合。到達該整合的端點並通過其授權的請求會啟動此執行。此欄位為必填。

## 目標 {#integrationForm-targetToggle}

在伺服器支援工作流程時顯示。

整合所啟動的對象：**Template** 會啟動任務範本的一個任務，**Workflow** 會啟動工作流程的一次執行。切換為
**Workflow** 會捨棄下方的任務參數。

## Workflow to run {#integrationForm-workflowId}

當 **目標** 為 **Workflow** 時顯示。

請求所啟動的工作流程。整合從請求中取得的每個值，都會作為工作流程中同名調查變數的答案傳給工作流程。工作流程整合需要
Pro 版本：沒有 Pro 版本時，此欄位會停用，伺服器也會拒絕此變更。

## 要執行的任務範本 {#integrationForm-templateId}

當 **目標** 為 **Template** 時顯示。

請求要為其啟動任務的任務範本。此欄位為必填。選取範本之後，其任務的參數會出現在下方。

## 任務參數 {#integrationForm-taskParams}

在選取任務範本後顯示。

此整合所啟動的每個任務都會取得的參數，就像您自己從範本啟動任務一樣。這裡只會出現範本允許任務覆寫的選項。從請求中取得的值會覆寫這些參數：
類型為 **Variables** 的值會覆寫同名的答案，類型為 **Task Params** 的值會覆寫同名的參數。

當範本允許在任務中使用 CLI 參數時，**CLI args** 會保存任務命令列的額外參數。

## 訊息 {#taskParams-message}

選填。此整合所啟動的每個任務的訊息，顯示在任務歷史中。

## 調查變數 {#taskParams-survey}

在範本有調查變數時顯示。

範本調查的答案，每個任務都相同。從請求中取得、類型為 **Variables** 的值會取代同名的答案。

類型為 secret 的變數的答案不會隨整合儲存：請改透過範本的變數群組傳遞機密值。

## 分支 {#taskParams-gitBranch}

在範本允許於任務中覆寫分支時顯示。

任務要簽出的儲存庫分支或標籤。保持原樣即使用範本的分支。從請求中取得、名為 `git_branch` 且類型為
**Task Params** 的值會覆寫它；該值開頭的 `refs/heads/` 會被移除。

## 清單 {#taskParams-inventoryId}

在範本允許於任務中覆寫清單時顯示。

任務執行時所針對的清單（對 Terraform 而言是工作區），用來取代範本的清單。

## Ansible 選項 {#taskParams-ansible}

在範本執行 Ansible Playbook 時顯示。

每個任務的 `ansible-playbook` 選項，每一項都在範本允許時才顯示：**限制**、**標籤**、**跳過標籤**、**除錯**、
**模擬執行**（`--check`）、**差異**和 **Skip Galaxy install**。

## Terraform 選項 {#taskParams-terraform}

在範本執行 Terraform、OpenTofu 或 Terragrunt 時顯示。

- **Plan**：任務執行 plan 後即停止，不會套用任何變更。
- **Destroy**：plan 和 apply 會銷毀資源。
- **Auto Approve**：當範本允許自動核准時，plan 會在未經確認的情況下套用。若未勾選，任務會等候，直到有人在
  Semaphore 中確認它。
- **Upgrade** 和 **Reconfigure**：`init` 會以 `-upgrade` 和 `-reconfigure` 執行。
