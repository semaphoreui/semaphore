---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# 編輯範本 'Ansible Playbook'

任務範本說明 Semaphore 要執行什麼以及如何執行：儲存庫及其中的檔案、清單和變數，以及使用者在啟動任務時可以變更哪些選項。每個任務都是從範本啟動的。
對話方塊的欄位取決於範本的應用程式（Ansible、Terraform、OpenTofu、Terragrunt、Bash、Python、PowerShell 及其他）：
只有部分應用程式才有的欄位會加以說明。變更會套用於按下**儲存**之後啟動的任務。

## 範本類型 {#templateForm-typeTabs}

頂端的分頁設定範本的種類：

- **任務**：執行程式碼。大多數範本都屬於這一種。
- **建置**：產生產出物。每個建置任務都會取得一個版本，從**起始版本**開始計算。
- **部署**：部署由建置所產生的產出物。啟動時，您要選擇所要部署之版本的建置任務。

建置或部署任務會在其環境中取得變數 `SEMAPHORE_TASK_TYPE`、`SEMAPHORE_TASK_TARGET_VERSION`（建置）和
`SEMAPHORE_TASK_INCOMING_VERSION`（部署）。Ansible 會在額外變數 `semaphore_vars.task_details` 中取得相同的值。

## 起始版本 {#templateForm-startVersion}

當類型為**建置**時顯示。必填。

第一個建置任務的版本，例如 `1.0.0` 或 `v0.1`。之後的每次建置，都會將前一個版本的最後一個數字加一，並保留其周圍的文字及補零：
`1.0.0`、`1.0.1`、`1.0.2`。將此處的數字調高到超過最後一次建置的版本，即可跳號，例如跳到 `2.0.0`；不含數字的版本永遠不會改變。

## 建置範本 {#templateForm-buildTemplateId}

當類型為**部署**時顯示。必填。

其任務會產生此範本所部署之產出物的範本。清單中列出專案的建置範本，其下則是部署範本：部署可以接在另一個部署之後，此時它會取得該部署從其建置所取得的版本。

## 自動執行 {#templateForm-autorun}

當類型為**部署**時顯示。

每當**建置範本**的任務成功時，自動以該任務所產生的版本啟動此範本的一個任務。失敗的建置不會啟動任何東西。

## 名稱 {#templateForm-name}

範本在「任務範本」頁面上列出時所用的名稱。此欄位為必填。

## 儲存庫 {#templateForm-repositoryId}

專案中保存要執行之程式碼的儲存庫。此欄位為必填。每個任務開始前，Semaphore 都會更新其儲存庫副本，並從該處執行程式碼。

欄位右側的標籤顯示任務要簽出的分支。點選它即可開啟**分支**欄位。

## 分支 {#templateForm-gitBranch}

在點選**儲存庫**右側的分支標籤後顯示。

此範本的任務要簽出的分支或標籤，用來取代儲存庫上設定的分支。將此欄位留空即使用儲存庫的分支。當 Semaphore 能讀取儲存庫的分支時，此欄位會列出它們。

當儲存庫是 monorepo 時，此欄位會停用：所有範本共用該儲存庫的同一份副本，位於儲存庫的分支上。

## Playbook 檔案路徑 {#templateForm-playbook}

要執行的內容，以從儲存庫根目錄起算的路徑表示，例如 `deploy/site.yml`。此路徑不能是絕對路徑，也不能通往儲存庫之外。
清單中提供儲存庫中的檔案；您也可以輸入不在清單中的路徑。欄位右側的按鈕會重新讀取清單。

標籤和意義會隨應用程式而改變：

- **Ansible**：Playbook。此欄位為必填；清單中提供 `.yml` 和 `.yaml` 檔案。
- **Terraform**、**OpenTofu**、**Terragrunt**：**Subdirectory path**，即程式碼所在的目錄。將其留空即在儲存庫的根目錄中執行程式碼。
  輸入 `/` 可列出某個目錄的內容。
- **Bash**、**Python**、**PowerShell** 及其他應用程式：**Script Filename**，即要執行的腳本。此欄位為必填。腳本會在儲存庫的根目錄中執行。

## Different working directory {#templateForm-workingDirectoryToggle}

在 Ansible 範本中顯示。

在儲存庫根目錄以外的目錄中執行 `ansible-playbook`。勾選它即可顯示 **Working directory** 欄位。取消勾選即可再次在根目錄中執行。

## Working directory {#templateForm-workingDirectory}

在勾選 **Different working directory** 時顯示。必填。

`ansible-playbook` 和 `ansible-galaxy` 執行時所在的目錄，以從儲存庫根目錄起算的路徑表示，例如 `ansible`。當專案的
`ansible.cfg` 位於子目錄中時請使用它，因為 Ansible 會從目前目錄讀取該檔案。Playbook 路徑仍然相對於儲存庫的根目錄。
此路徑不能是絕對路徑，也不能通往儲存庫之外。

## 清單 {#templateForm-inventoryId}

在 Ansible、Terraform、OpenTofu 和 Terragrunt 範本中顯示。

- **Ansible**：Playbook 執行時所針對的清單，連同主機的憑證。此欄位為必填。
- **Terraform**、**OpenTofu**、**Terragrunt**：**Workspace**，即程式碼執行時所在的工作區。Semaphore 會在 plan
  之前選取它，若它不存在則會建立它。未設定工作區即儲存的新範本，會取得一個名為 `default` 的工作區。

清單中列出專案中為該範本之應用程式建立的清單。

## 變數群組 {#templateForm-environmentIds}

選填。此範本的每個任務都會取得其變數、環境變數和機密值的變數群組。您可以選擇多個：它們會依欄位中的順序合併，當兩個群組設定了相同的名稱時，後面的群組優先於前面的群組。

變數傳給程式碼的方式取決於應用程式：對 Ansible 是額外變數，對 Terraform 和 OpenTofu 是 `-var`，對腳本是 `name=value`
參數。環境變數則會為每個應用程式設定。

## 檢視 {#templateForm-viewId}

選填。列出此範本的檢視：檢視是「任務範本」頁面上的分頁，用來將專案的範本分組。將此欄位留空即不將範本放入任何檢視。

## 執行器標籤 {#templateForm-runnerTag}

在有專案執行器可用時顯示。

此範本的任務只會在帶有此標籤的執行器上執行，即使伺服器本身可以執行它們。此標籤會覆寫清單的執行器標籤。將此欄位留空，即使用清單的標籤，或不使用標籤。

## Executor image {#templateForm-executorImage}

執行器為此範本的任務所啟動的容器映像檔，例如 `semaphoreui/job:latest`。將此欄位留空即使用執行器設定中的映像檔。

只有 Docker 和 Kubernetes 執行環境會在容器中執行任務，因此只有在其中之一可用時，此欄位才會啟用。不使用容器執行的任務會忽略它。

## 調查變數 {#surveyVars}

此範本的任務啟動時要求使用者提供的值。點選 **+** 新增變數，點選標籤以修改變數，點選標籤上的叉號以移除變數；拖曳標籤即可設定提問的順序。

答案會依各變數的設定，以與變數群組中之變數相同的方式傳給程式碼，或作為環境變數傳遞。它們優先於設定了相同名稱的變數群組。

## 允許平行任務 {#templateForm-allowParallelTasks}

讓此範本的任務可以同時執行。若未勾選，範本的任務會在佇列中等候，直到執行中的任務結束。無論哪種情況，專案和伺服器的平行任務上限都會適用。

## Issue JWT to task runner {#templateForm-jwtEnabled}

在管理員已為伺服器開啟 JWT 核發時顯示。

為此範本的每個任務提供一個由 Semaphore 簽署的 JSON Web Token，放在環境變數 `SEMAPHORE_JWT` 中。程式碼可以向其他服務出示此權杖，
以證明自己是哪個任務：權杖中指明了任務、專案、範本以及啟動它的使用者。該服務會使用 Semaphore 伺服器 `/.well-known/jwks.json`
上的公鑰來檢查簽章。

## JWT audience {#templateForm-jwtAudience}

在勾選 **Issue JWT to task runner** 時顯示。選填。

此權杖的目標服務，寫入其 `aud` 宣告中，例如 `https://vault.example.com`。每輸入一個值後請按 Enter。最多 32 個值；
若權杖不需要受眾，請將此欄位留空。

## JWT TTL {#templateForm-jwtTtl}

在勾選 **Issue JWT to task runner** 時顯示。選填。

權杖從任務開始起保持有效的時間長度，以 `30s`、`15m` 或 `1h` 之類的時間長度表示。將此欄位留空即使用伺服器的預設值，除非管理員另有設定，否則為一小時。
此值不能超過伺服器的上限，除非另有設定，否則為 24 小時。

## 偵測到新的 Git 提交時自動執行任務 {#templateForm-cronVisible}

當儲存庫中出現新的提交時，啟動此範本的一個任務。Semaphore 會依所選的間隔向儲存庫詢問其最新的提交，並只在提交自上次檢查以來有變更時才啟動任務。

勾選此方塊並儲存，會建立檢查用的排程；取消勾選並儲存，則會刪除該排程。範本的其他排程會保留。

## 儲存庫 {#templateForm-cronRepositoryId}

在勾選**偵測到新的 Git 提交時自動執行任務**時顯示。必填。

要監看是否有新提交的儲存庫。Semaphore 會監看該儲存庫上設定的分支，而不是範本的**分支**。

## 檢查間隔 {#templateForm-cronFormat}

在勾選**偵測到新的 Git 提交時自動執行任務**時顯示。

Semaphore 檢查儲存庫是否有新提交的頻率：從每分鐘一次到每天一次。預設為每分鐘一次。

## 隱藏成功通知 {#templateForm-suppressSuccessAlerts}

讓此範本中成功的任務不發送警報。失敗任務的警報仍會發送。只有在專案開啟了警報時才會生效。

## 隱藏錯誤通知 {#templateForm-suppressErrorAlerts}

讓此範本中失敗的任務不發送警報。成功任務的警報仍會發送。只有在專案開啟了警報時才會生效。

## CLI args {#templateForm-args}

加到每個任務之命令中的參數。每個標籤是一個參數，原樣傳遞，不經過 shell：將 `--forks=10` 放在一個標籤中，或將 `--forks`
和 `10` 放在兩個標籤中。拖曳標籤即可變更其順序。清單右側的按鈕會以 JSON 陣列的形式開啟這些參數：

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**：加到 `ansible-playbook`，位於 Playbook 之前。
- **Terraform**、**OpenTofu**、**Terragrunt**：加到 `plan` 和 `apply`，不會加到 `init`。
- 腳本：加在腳本名稱之後。

## CLI 參數 {#templateForm-allowOverrideArgsInTask}

位於**提示**之下。讓使用者在啟動任務時加入自己的參數。這些參數會加在範本的參數之後。

## 分支 {#templateForm-allowOverrideBranchInTask}

位於**提示**之下。讓使用者在啟動任務時選擇另一個分支、標籤或提交，以取代範本的分支。若未勾選，任務一律簽出範本的分支。

當儲存庫是 monorepo 時，此方塊會停用。

## 清單 {#templateForm-allowOverrideInventory}

位於**提示**之下。在 Ansible、Terraform、OpenTofu 和 Terragrunt 範本中顯示。

讓使用者在啟動任務時選擇另一個清單。只有 Ansible 範本會使用在該處選擇的清單：即使勾選了此方塊，Terraform、OpenTofu 或
Terragrunt 任務仍會在範本的工作區中執行。

## 限制 {#templateForm-limit}

在 Ansible 範本中顯示。

要執行 Playbook 的主機或群組，取自清單，以 `--limit` 傳遞。每個標籤是一個模式，例如 `web` 或 `db01`；將清單留空即在每台主機上執行。

## 標籤 {#templateForm-tags}

在 Ansible 範本中顯示。

只執行帶有這些標籤的任務和角色，以 `--tags` 傳遞。將清單留空即執行所有內容。

## 跳過標籤 {#templateForm-skipTags}

在 Ansible 範本中顯示。

略過帶有這些標籤的任務和角色，以 `--skip-tags` 傳遞。

## 保險庫 {#templateVaults}

在 Ansible 範本中顯示。

Playbook 解密其檔案所需的 Ansible Vault 密碼。點選 **+** 新增一個，點選標籤以修改它，點選標籤上的叉號以移除它。每個保險庫都有一個名稱，
即其 vault ID，並從金鑰儲存中的金鑰或儲存庫中的客戶端腳本提供密碼。

## 自動核准 {#templateForm-autoApprove}

在 Terraform、OpenTofu 和 Terragrunt 範本中顯示。

不經詢問即套用 plan。若未勾選，任務會在 plan 之後停止，並等候使用者確認或拒絕它；沒有變更的 plan 會在沒有任何內容可套用的情況下結束任務。

## 覆寫後端設定 {#templateForm-overrideBackend}

在 Semaphore 可以保存 Terraform 狀態時，顯示於 Terraform、OpenTofu 和 Terragrunt 範本中。

將程式碼的狀態保存在 Semaphore 中。在 `init` 之前，Semaphore 會在程式碼的目錄中寫入一個設定 `http` 後端的檔案，並在任務結束後刪除它：

```hcl
terraform {
  backend "http" {
  }
}
```

## 後端檔案名稱 {#templateForm-backendFilename}

與**覆寫後端設定**一起顯示；勾選該方塊時才會啟用。選填。

Semaphore 寫入後端設定的檔案名稱，例如 `semaphore_backend.tf`。預設為 `backend.tf`。名稱必須以 `.tf` 結尾，且只能包含字母、
數字、`_`、`-` 和 `.`。請選擇您程式碼中沒有任何檔案使用的名稱：Semaphore 會覆寫該檔案，並在任務結束後刪除它。

## 限制 {#templateForm-allowOverrideLimit}

位於 **Ansible 提示**之下。讓使用者在啟動任務時設定要執行的主機，以取代範本的**限制**。

## 標籤 {#templateForm-allowOverrideTags}

位於 **Ansible 提示**之下。讓使用者在啟動任務時設定標籤，以取代範本的**標籤**。

## 跳過標籤 {#templateForm-allowOverrideSkipTags}

位於 **Ansible 提示**之下。讓使用者在啟動任務時設定要跳過的標籤，以取代範本的**跳過標籤**。

## 除錯 {#templateForm-allowDebug}

位於 **Ansible 提示**之下。讓使用者在啟動任務時開啟詳細輸出，從 `-v` 到 `-vvvvvv`。

## 模擬執行 {#templateForm-showDryRun}

位於 **Ansible 提示**之下。預設為勾選。在啟動任務時提供**模擬執行**選項，以 `--check` 執行 Playbook：Ansible
會回報它將變更的內容，而不做任何變更。取消勾選此方塊即可移除此選項。

## 差異 {#templateForm-showDiff}

位於 **Ansible 提示**之下。預設為勾選。在啟動任務時提供**差異**選項，以 `--diff` 執行 Playbook，顯示對檔案所做的變更。
取消勾選此方塊即可移除此選項。

## 自動核准 {#templateForm-allowAutoApprove}

位於 Terraform、OpenTofu 和 Terragrunt 的提示之下。讓使用者在啟動任務時勾選 **Auto Approve**，以不經確認即套用 plan。
任務對話方塊無論如何都會提供該選項，但若未勾選此方塊，Semaphore 會忽略它：任務會在 plan 之後等候確認，除非範本的**自動核准**已勾選。

## Galaxy install options {#templateForm-galaxy}

在 Ansible 範本中顯示。點選它即可顯示 Galaxy 安裝的選項；標籤會計算與預設值不同的選項數量。

在 Playbook 執行之前，Semaphore 會使用 `ansible-galaxy` 安裝儲存庫所需的角色和集合。它會在 Playbook 旁邊以及儲存庫的根目錄中尋找
`requirements.yml`、`roles/requirements.yml` 和 `collections/requirements.yml`，並且只在自上次任務以來有檔案變更時才執行安裝。

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

位於 **Galaxy install options** 之下。略過 Galaxy 安裝：Playbook 會使用先前安裝的角色和集合執行。當需求以其他方式安裝，
或安裝很慢且不需要時，請使用此選項。

## Role install args {#templateForm-galaxyRoleArgs}

位於 **Galaxy install options** 之下。加到 `ansible-galaxy role install` 的參數，每個標籤一個，例如
`--ignore-certs` 或 `--server=https://galaxy.example.com`。Semaphore 只接受可以安全傳遞的安裝旗標：未知的旗標、
`--token`、`--api-key`、`-p` 或 `--roles-path` 會讓範本無法儲存。

## Collection install args {#templateForm-galaxyCollectionArgs}

位於 **Galaxy install options** 之下。加到 `ansible-galaxy collection install` 的參數，每個標籤一個，例如
`--pre` 或 `--upgrade`。適用與 **Role install args** 相同的規則：不接受 `--token`、`--api-key`、`-p` 和
`--collections-path`。

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

位於**提示**中的 **Galaxy install options** 之下。讓使用者在啟動任務時選擇是否略過 Galaxy 安裝。任務的選擇優先於範本的
**Skip Galaxy install**。
