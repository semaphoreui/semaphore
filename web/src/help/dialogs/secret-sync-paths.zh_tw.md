---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Semaphore 從密鑰儲存區匯入機密值的路徑，以及它自動匯入的頻率。從密鑰儲存區開啟時，匯入的內容會填入專案的金鑰儲存；從變數群組開啟時，
則會填入該群組。您在此處所做的變更，會在您按下**關閉**時保留，並透過儲存區或群組的**儲存**隨其一起儲存。

## Auto-sync interval {#varGroupSync-syncInterval}

從變數群組開啟此對話方塊時顯示。

群組自動匯入的頻率，以分鐘為單位。預設值 `0` 會關閉自動匯入：此時只有在您於「變數群組」頁面上按下群組的同步按鈕時，才會匯入該群組。

伺服器每分鐘檢查一次，當距離上次嘗試（無論成功與否）已經過了該間隔時就會匯入；您設定間隔後的第一次檢查會立即匯入。匯入失敗時，會在下一個間隔後重試。

## Auto-sync interval {#secretStorageSync-syncInterval}

從密鑰儲存區開啟此對話方塊時顯示。

儲存區自動匯入的頻率，以分鐘為單位。預設值 `0` 會關閉自動匯入：此時只有在您於金鑰儲存的 **Storages** 分頁中按下儲存區的同步按鈕時，
才會匯入該儲存區。

伺服器每分鐘檢查一次，當距離上次嘗試（無論成功與否）已經過了該間隔時就會匯入；您設定間隔後的第一次檢查會立即匯入。匯入失敗時，會在下一個間隔後重試。

## 路徑 {#secretStorageSync-path}

在清單的每一列中顯示。透過 **Add path** 新增一列。

儲存區中要匯入機密值的路徑，依儲存區本身的命名方式撰寫。每一列是一個路徑；匯入時會讀取所有路徑。

## separator {#secretStorageSync-separator}

在清單的每一列中顯示。選填。

匯入此路徑時所使用的分隔符號。除非您儲存區中的名稱需要分隔符號，否則請將此欄位留空。

## prefix {#secretStorageSync-prefix}

在清單的每一列中顯示。選填。

匯入此路徑時所使用的前綴。若匯入的機密值不需要前綴，請將此欄位留空。

## 移除路徑 {#secretStorageSync-removePath}

在清單的每一列中顯示。

從清單中移除此路徑。此變更會在儲存區或群組儲存時生效。

## Add path {#secretStorageSync-addPath}

在清單中新增一個空白列。在至少有一個路徑之前，對話方塊會顯示尚未定義同步路徑，而儲存區或群組的同步按鈕會保持停用。
