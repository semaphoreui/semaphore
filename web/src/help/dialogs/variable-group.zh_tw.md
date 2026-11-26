---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# 編輯變數群組

變數群組保存額外變數、環境變數和機密值，Semaphore 會將它們傳給使用此群組之任務範本的每個任務。Ansible 以 `--extra-vars`
取得額外變數，Terraform 和 OpenTofu 則以 `-var` 取得；環境變數會設定在任務的程序中。一個範本可以使用多個群組：當其中兩個群組定義了相同的名稱時，
在範本中排在後面的群組勝出。

## 群組名稱 {#varGroup-name}

群組在「變數群組」頁面以及任務範本的變數群組選擇器中顯示的名稱。此欄位為必填。

## Secret storage {#varGroup-secretStorageId}

在群組為新建且伺服器支援密鑰儲存區時顯示。選填。

保存此群組機密值的密鑰儲存區，例如 HashiCorp Vault。將此欄位留空，即可將機密值加密保存在 Semaphore 的資料庫中。

儲存區只能在建立群組時選擇一次：現有的群組會保留其儲存區，此欄位也不會再顯示。標示為 **Read only** 的儲存區不接受新的機密值：
將機密值儲存到其中會失敗。

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

在群組為新建且伺服器支援密鑰儲存區時顯示。選擇了 **Secret storage** 之後才可使用。

此群組的每個機密值在儲存區中之路徑的開頭部分。Semaphore 會在其後加上 10 個隨機字元：使用前綴 `semaphore/prod/` 時，
機密值會儲存為 `semaphore/prod/` 加上這些字元。前綴會原樣採用，因此若機密值應像資料夾一樣位於其下，請以 `/` 結尾。

前綴只能在建立群組時設定一次。**Secrets** 分頁會將它顯示為來源路徑模式。

## Table、JSON 或 YAML {#varGroup-extraVarsMode}

編輯額外變數的方式：

- **Table**：每個變數一列，各有一個類型。使用切換按鈕旁的 **+** 按鈕新增一列。
- **JSON**：所有變數組成一個 JSON 物件。
- **YAML**：所有變數組成一個 YAML 對應。

對話方塊開啟時為 **Table**。切換時會轉換您輸入的內容；若內容無法讀取，對話方塊會停留在原本的模式並顯示錯誤。無論哪種模式，
變數都會儲存為一個由名稱和值組成的物件。

## JSON 格式的額外變數 {#varGroup-extraVarsJson}

在以 **JSON** 編輯額外變數時顯示。

以一個 JSON 物件表示的額外變數。名稱不能為空；值可以是字串、數字、清單或物件：

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## YAML 格式的額外變數 {#varGroup-extraVarsYaml}

在以 **YAML** 編輯額外變數時顯示。

以一個 YAML 對應表示的額外變數。文字會以 JSON 所知的類型讀取：未加引號的日期，例如 `2024-01-01`，仍是字串，而 `.inf`
或 `.nan` 會被拒絕。頂層為清單或純量時也會被拒絕。

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## 變數名稱 {#varGroup-varName}

在以 **Table** 編輯額外變數時，顯示於每一列中。

Playbook 或 Terraform 程式碼所看到的變數名稱，例如 `app_version`。

## 變數類型 {#varGroup-varType}

在以 **Table** 編輯額外變數時，顯示於每一列中。

該列之值的讀取方式：

- **String**：原樣的文字。
- **Number**：數字，例如 `42`。
- **List**：JSON 陣列，例如 `["a", "b"]`。
- **Dict**：JSON 物件，例如 `{"key": "value"}`。

與其類型不符的值會中止儲存，並顯示該變數的名稱。

## 變數值 {#varGroup-varValue}

在以 **Table** 編輯額外變數時，顯示於每一列中。

變數的值，採用其類型的格式。對於 **List** 或 **Dict**，欄位末端的按鈕會開啟一個較大的編輯器。

Semaphore 會將這些值作為資料傳給 Ansible：值中的 `{{ … }}` 表示式不會被求值，Playbook 會原樣取得該文字。

## 環境變數名稱 {#varGroup-envName}

位於 **Variables** 分頁的**環境變數**之下，每一列中都有。使用該區段的 **+** 按鈕新增一列。

在執行任務之程序的環境中設定的變數名稱，例如 `AWS_REGION`。

## 環境變數值 {#varGroup-envValue}

位於 **Variables** 分頁的**環境變數**之下，每一列中都有。

變數的值，以文字表示。它以純文字儲存，並會顯示給每一位開啟此群組的人：請改將密碼和權杖放在 **Secrets** 分頁中。

## Sync keys enabled {#varGroup-syncEnabled}

顯示於具有密鑰儲存區之群組的 **Secrets** 分頁上。

開啟此選項，即可從 **Sync paths** 中設定的路徑，將儲存區中的機密值匯入此群組。當您在「變數群組」頁面上按下群組的同步按鈕時，
以及依 **Sync paths** 中設定的間隔自動執行時，都會進行匯入。關閉時，不會匯入任何內容。

## Sync paths {#varGroup-syncPaths}

當 **Sync keys enabled** 開啟時顯示。

開啟一個對話方塊，其中包含要匯入的儲存區路徑以及自動匯入的間隔。按鈕上的數字是路徑的數量。在群組至少有一個路徑之前，「變數群組」頁面上該群組的同步按鈕會保持停用。

該對話方塊只有**關閉**按鈕：您在其中所做的變更，會透過此對話方塊的**儲存**隨群組一起儲存。

## 機密變數名稱 {#varGroup-secretVarName}

位於 **Secrets** 分頁的**額外變數**之下，每一列中都有。使用該區段的 **+** 按鈕新增一列。

值為機密的額外變數名稱。任務取得它的方式，與取得 **Variables** 分頁中的額外變數相同；機密值優先於同名的額外變數。Ansible
會從以 Ansible Vault 加密的檔案中取得機密值，Terraform 和 OpenTofu 則以 `-var` 取得。如對話方塊所警告的，以此方式傳遞的機密值仍可能以明文出現在
Ansible 的日誌中。

## 機密變數值 {#varGroup-secretVarValue}

位於 **Secrets** 分頁的**額外變數**之下，每一列中都有。

機密的值。Semaphore 會將其加密，或寫入群組的密鑰儲存區，並且永遠不會再次顯示：之後開啟群組時，此欄位是空的。保持空白即保留已儲存的值，
輸入新的值則會取代它。刪除該列會在儲存群組時刪除該機密值。

## 機密環境變數名稱 {#varGroup-secretEnvName}

位於 **Secrets** 分頁的**環境變數**之下，每一列中都有。使用該區段的 **+** 按鈕新增一列。

值為機密的環境變數名稱，例如 `DB_PASSWORD`。它會設定在執行任務之程序的環境中，與 **Variables** 分頁中的環境變數相同。

## 機密環境變數值 {#varGroup-secretEnvValue}

位於 **Secrets** 分頁的**環境變數**之下，每一列中都有。

機密的值。與機密額外變數一樣，它永遠不會再次顯示：保持空白即保留已儲存的值，輸入新的值則會取代它。
