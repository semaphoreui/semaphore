---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# 新增任務

啟動範本的一個任務：執行一次其 Playbook、Terraform 程式碼或腳本。範本決定此對話方塊顯示哪些欄位，下方每個段落都會說明其欄位何時出現。
底部的按鈕依範本的類型而定，為**執行**、**建置**或**部署**。

當範本有調查時，其問題會以欄位的形式出現在**訊息**之後。標有 `*` 的欄位為必填，Integer 欄位只接受數字，並且會填入預設答案。
Semaphore 會將答案以變數的形式傳給任務：對 Ansible 是額外變數，對 Terraform 是 `-var`，對腳本是 `name=value`
參數，對設定為環境變數的問題則是環境變數。答案會取代範本變數群組中的同名變數。機密答案在輸入時會被隱藏，且不會隨任務保存。

## 提交 {#newTask-commit}

在您再次執行先前的任務，且該任務是在已知的提交上執行時顯示。

任務會簽出儲存庫的這個提交，而不是分支的最新提交。關閉此方塊即可執行最新的提交。

只有在範本允許覆寫分支時才會使用此提交，而且永遠不會用於標示為 monorepo 的儲存庫。否則任務會執行最新的提交。

## 建置版本 {#newTask-buildTaskId}

當範本的類型為**部署**時顯示。

要部署的建置：建置範本最近 20 個成功任務中的一個，依版本和訊息列出。預設選取最新的一個。此欄位為必填。

建置的版本會傳給任務：在 Ansible 中為 `semaphore_vars.task_details.incoming_version`，在腳本中為環境變數
`SEMAPHORE_TASK_DETAILS_INCOMING_VERSION`。若您無權檢視建置範本，清單會是空的。

## 訊息 {#newTask-message}

選填。說明執行此任務原因的備註，例如 `Hotfix for the login page`。它會與任務一起顯示在任務歷史中，不會改變執行的任何內容。

Playbook 可以透過 `semaphore_vars.task_details.message` 讀取它，腳本則可透過環境變數
`SEMAPHORE_TASK_DETAILS_MESSAGE` 讀取。

## 分支 {#newTask-gitBranch}

在範本允許覆寫分支時顯示，除非儲存庫是 monorepo。

任務要簽出的儲存庫分支或標籤。此欄位的初始值為範本的分支。當此欄位為空時，任務會使用範本的分支；若範本沒有分支，則使用儲存庫的分支。
不是有效 Git 分支名稱的名稱會被拒絕。

## 清單 {#newTask-inventoryId}

在範本允許於任務中選擇清單，且您可以檢視專案時顯示。對於 Terraform、OpenTofu 或 Terragrunt 範本，此欄位的標籤為
**Workspace**，但伺服器只會將此選擇用於 Ansible 範本：Terraform、OpenTofu 或 Terragrunt 任務會在範本的工作區中執行，無論您在此處選擇什麼。

任務執行時所針對的清單。此欄位的初始值為範本的清單，並列出專案中為該範本之應用程式建立的清單。若所選的清單在任務開始時已不存在，
任務會針對範本的清單執行。

## 限制 {#taskParams-limit}

在允許覆寫限制的 Ansible 範本中顯示。

Playbook 要在其上執行的清單主機或群組，每個標籤一個模式，例如 `web` 或 `10.0.0.11`。Semaphore 會以逗號連接它們，作為
`--limit` 傳遞。清單的初始值為範本的限制。移除所有標籤即可在清單的所有主機上執行。

## 標籤 {#taskParams-tags}

在允許覆寫標籤的 Ansible 範本中顯示。

只有帶有其中某個標籤的 Playbook 任務和角色才會執行：Semaphore 會將它們作為 `--tags` 傳遞，每個標籤一個。清單的初始值為範本的標籤。
移除所有標籤即可執行整個 Playbook。

## 跳過標籤 {#taskParams-skipTags}

在允許覆寫跳過標籤的 Ansible 範本中顯示。

帶有其中某個標籤的 Playbook 任務和角色不會執行：Semaphore 會將它們作為 `--skip-tags` 傳遞，每個標籤一個。清單的初始值為範本的跳過標籤。

## 除錯 {#taskParams-debug}

在允許除錯的 Ansible 範本中顯示。

以詳細程度旗標執行 `ansible-playbook`，讓任務日誌顯示更多細節。標籤會顯示此旗標：`-v` 重複的次數依**除錯等級**而定。

## 除錯等級 {#taskParams-debugLevel}

與**除錯**一起顯示，並在勾選**除錯**時啟用。

日誌的詳細程度，從 1（`-v`）到 6（`-vvvvvv`）。預設值為 4（`-vvvv`），它也會顯示與主機連線的詳細資訊。

## 模擬執行 {#taskParams-dryRun}

在 Ansible 範本中顯示，除非範本將其隱藏。

以 `--check` 執行 Playbook：Ansible 會回報它將在主機上變更的內容，而不做任何變更。

## 差異 {#taskParams-diff}

在 Ansible 範本中顯示，除非範本將其隱藏。

以 `--diff` 執行 Playbook：日誌會顯示對主機上的檔案和範本所做的變更。同時勾選**模擬執行**，即可在不實際變更的情況下查看變更。

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

在允許覆寫 Galaxy 安裝的 Ansible 範本中顯示。

在 Playbook 執行之前，若自上次安裝以來有檔案變更，Semaphore 會安裝儲存庫中 `requirements.yml` 檔案所列的角色和集合。
勾選此方塊即可為此任務略過該步驟。

無論範本如何設定，此方塊一開始都不勾選，而不勾選的方塊會安裝需求。

## Plan {#taskParams-plan}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

執行 `plan` 並在此結束任務：不會套用任何變更。可用它來先檢閱變更。若未勾選，當 plan 有變更時，任務會繼續套用 plan。

## Destroy {#taskParams-destroy}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

在 `plan` 和 `apply` 中加上 `-destroy`：任務會規劃並接著套用移除工作區狀態中的每一項資源。同時勾選 **Plan**，即可查看將被銷毀的內容而不實際銷毀。

## Auto Approve {#taskParams-autoApprove}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

plan 產生後立即套用。若未勾選，任務會在 plan 之後停止，並等候有人在任務頁面上確認或拒絕它。

只有在範本允許自動核准時，此方塊才會生效。設定為一律核准的範本，無論此方塊如何設定，都會直接套用而不詢問。

## Upgrade {#taskParams-upgrade}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

以 `-upgrade` 執行 `init`：提供者和模組會升級到其版本限制所允許的最新版本。

## Reconfigure {#taskParams-reconfigure}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

以 `-reconfigure` 執行 `init`：後端會重新設定，並忽略已儲存的後端設定。若未勾選，Semaphore 會執行
`init -migrate-state`，將現有的狀態移到已變更的後端。

## CLI args {#newTask-args}

在範本允許覆寫 CLI 參數時顯示。

任務命令列的額外參數，每個標籤一個：`--forks=10` 是一個標籤，`--forks` 和 `10` 則是兩個。點選標籤可編輯它，拖曳可變更順序，
使用最後一個標籤可新增參數。

清單的初始值為範本的參數。Semaphore 一律會傳遞範本的參數，並將此清單加在其後，因此一旦您變更了清單，留在清單中的範本參數就會被傳遞兩次。
請移除它們，只保留您新增的參數。
