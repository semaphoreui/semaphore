---
form: src/components/TerraformInventoryForm.vue
---

# 新增 Terraform Workspace

工作區是 Terraform、OpenTofu 或 Terragrunt 範本的清單：它指定任務在 `init` 之後選取的工作區，以及 `init` 下載私人模組時所用的金鑰。
您在此建立的工作區會附加到此範本，並列在其狀態上方。範本的任務會在其預設工作區（以勾號標示的那一個）中執行；**Make default**
可變更預設工作區。以鉛筆按鈕開啟時，同一個對話方塊也用來編輯工作區。

## Workspace name {#tfWorkspace-inventory}

Terraform 工作區的名稱，例如 `production`。此欄位為必填。

在執行 plan 之前，Semaphore 會以此名稱執行 `workspace select -or-create=true`，因此若工作區尚不存在就會被建立。當
`workspace list` 失敗時（例如因為後端不支援工作區），Semaphore 會略過選取步驟。

Semaphore 會將工作區儲存為 `TEMPLATE - NAME`，即範本名稱加上此名稱：您會在「清單」頁面上，以及將工作區附加到範本時看到此完整名稱。
重新命名工作區並不會在 Terraform 中將其重新命名：下一個任務會選取（或建立）新名稱的工作區。

## SSH key for private modules {#tfWorkspace-sshKeyId}

`init` 透過 SSH 從私人 Git 儲存庫下載模組時所用的金鑰。在此對話方塊中此欄位為必填。清單中提供金鑰儲存中類型為 **SSH 金鑰**、
**使用密碼登入**和**無**的金鑰。

只有 **SSH 金鑰**會生效：在 `init` 期間，Semaphore 會將它載入 SSH 代理程式，Git 會將它用於
`git::ssh://git@example.com/modules/vpc.git` 之類的模組來源。類型為**使用密碼登入**的金鑰不會將任何憑證傳給 Git。
若模組不需要金鑰，請選擇**無**。
