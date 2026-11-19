---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# 編輯整合

當其他服務呼叫 Semaphore 的 Webhook URL 時（例如 GitHub 的 push 事件），整合會啟動一個任務範本或一個工作流程。
此 URL 是一個別名：`/api/integrations/` 加上一段隨機代碼。專案的別名位於「整合」頁面，單一整合的別名則位於該整合自己的頁面。
該頁面也包含整合的比對器，以及整合從請求中取得的值。

## Name {#integrationForm-name}

整合在「整合」頁面上列出時所用的名稱。由此整合啟動的任務，在稽核日誌中會歸屬於它。此欄位為必填。

## Runs {#integrationForm-targetToggle}

在伺服器支援工作流程時顯示。

整合所啟動的對象：**Template** 會啟動任務範本的一個任務，**Workflow** 會啟動工作流程的一次執行。切換為
**Workflow** 會捨棄下方的任務參數。

## Workflow to run {#integrationForm-workflowId}

當 **Runs** 為 **Workflow** 時顯示。

請求所啟動的工作流程。整合從請求中取得的每個值，都會作為工作流程中同名調查變數的答案傳給工作流程。工作流程整合需要
Pro 版本：沒有 Pro 版本時，此欄位會停用，伺服器也會拒絕此變更。

## Task Template to run {#integrationForm-templateId}

當 **Runs** 為 **Template** 時顯示。

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

## Auth method {#integrationForm-authMethod}

Semaphore 檢查請求是否來自您所預期之服務的方式。未通過檢查的請求不會啟動任何東西，而 Webhook 對它的回應與對其他任何請求的回應相同。

- **None**：不檢查。任何知道此 URL 的人都可以啟動任務。
- **GitHub Webhooks**：`X-Hub-Signature-256` 標頭必須包含以密鑰對主體產生的 HMAC-SHA256 簽章，與 GitHub 傳送的格式相同。
- **Bitbucket Webhooks**：對 `X-Hub-Signature` 標頭進行相同的檢查，與 Bitbucket 傳送的格式相同。
- **Token**：**Auth header** 中指定的標頭必須包含密鑰本身。
- **HMAC** 和 **HMAC (SHA-512)**：**Auth header** 中指定的標頭必須包含以密鑰對主體產生的 HMAC-SHA256 或
  HMAC-SHA512，以小寫十六進位表示，不含 `sha256=` 之類的前綴。
- **BasicAuth**：請求必須在其 `Authorization` 標頭中帶有**驗證密鑰**所指定之金鑰的登入名稱和密碼。

## Auth header {#integrationForm-authHeader}

當 **Auth method** 為 **Token**、**HMAC** 或 **HMAC (SHA-512)** 時顯示。

攜帶權杖或簽章的請求標頭名稱，例如 `X-Webhook-Token`。名稱的大小寫無關緊要。沒有該標頭的請求會被拒絕。

## 驗證密鑰 {#integrationForm-authSecretId}

當 **Auth method** 不是 **None** 時顯示。

金鑰儲存中用來檢查請求的金鑰。對於 **BasicAuth**，只有類型為**使用密碼登入**的金鑰適用。對於其他方法，類型為**字串**的金鑰也適用：
檢查時會使用其值，或使用**使用密碼登入**金鑰的密碼。

除了 **None** 之外，每種方法都需要此欄位：沒有金鑰或密鑰為空時，每個請求都會被拒絕。
