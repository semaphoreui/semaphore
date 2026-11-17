---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# 編輯 Secret Storage

密鑰儲存區是外部的機密管理器，Semaphore 將機密值保存在其中並從中讀取：HashiCorp Vault、OpenBao、Devolutions
Server、AWS Secrets Manager 或 Azure Key Vault。以儲存區建立的變數群組會將其機密值保存在該處，而開啟同步的儲存區會將其機密值匯入專案的金鑰儲存。
儲存區的種類在 **New Storage** 選單中選擇，無法在此對話方塊中變更；下方的欄位取決於它。

## 名稱 {#secretStorage-name}

儲存區在金鑰儲存的 **Storages** 分頁，以及新變數群組的 **Secret storage** 選擇器中顯示的名稱。此欄位為必填。

## Server URL {#secretStorage-vaultURL}

當儲存區為 HashiCorp Vault、OpenBao 或 Devolutions Server 時顯示。

伺服器的位址，包含其協定，例如 `https://vault.example.com:8200`。此欄位為必填。

## Mount {#secretStorage-vaultMount}

當儲存區為 HashiCorp Vault 或 OpenBao 時顯示。選填。

保存機密值之機密引擎的掛載路徑。將此欄位留空即使用預設值 `secret`。

## Namespace {#secretStorage-vaultNamespace}

當儲存區為 HashiCorp Vault 或 OpenBao 時顯示。選填。

機密值所在的命名空間。命名空間存在於 Vault Enterprise 和 HCP Vault Dedicated 中，以及 2.3 版起的 OpenBao 中。
若伺服器沒有命名空間，請將此欄位留空。

## 機密來源 {#secretSource-options}

在每一種儲存區中都會顯示，位於 Semaphore 登入時所用的機密值上方：Vault 或 OpenBao 的 **Token**、Devolutions
Server 的應用程式密鑰、AWS 的秘密金鑰、Azure 的用戶端密碼。

Semaphore 取得該機密值的來源：

- **Store in DB**：您在此對話方塊中輸入機密值；Semaphore 會將其加密並保存在資料庫中。
- **From ENV**：您輸入 Semaphore 伺服器的環境變數名稱；每次使用儲存區時，都會從該變數讀取機密值。
- **From File**：您輸入 Semaphore 伺服器上某個檔案的絕對路徑；每次都會從該檔案讀取機密值，不含結尾的換行。該檔案必須位於伺服器的機密目錄中
  （設定中的 `dirs.secrets`，預設為 `/tmp/semaphore`），含有 `..` 的路徑會被拒絕。

切換來源會清除其下方的欄位。

## Token {#secretStorage-vaultToken}

當儲存區為 HashiCorp Vault 或 OpenBao，且機密來源為 **Store in DB** 時顯示。

Semaphore 用來向伺服器驗證身分的權杖。它必須被允許讀取機密值；除非儲存區為 **Read only**，還必須被允許寫入和刪除機密值。

新的儲存區必須填寫此欄位。開啟現有的儲存區時，此欄位是空的：保持空白即保留已儲存的權杖，輸入新的權杖則會取代它。

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

當儲存區為 HashiCorp Vault 或 OpenBao，且機密來源為 **From ENV** 或 **From File** 時顯示。

保存權杖的環境變數名稱或檔案路徑。新的儲存區必須填寫此欄位。

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

當儲存區為 Devolutions Server 時顯示。

Semaphore 連線到伺服器時不檢查其 TLS 憑證。此時位於 Semaphore 與伺服器之間的任何人都能讀取機密值。請保持關閉，並改為給伺服器一個
Semaphore 信任的憑證。

## Vault ID {#secretStorage-dvlsKey}

當儲存區為 Devolutions Server 時顯示。

Devolutions Server 中保存機密值之保險庫的 ID。新的儲存區必須填寫此欄位。

## App Key {#secretStorage-dvlsAppKey}

當儲存區為 Devolutions Server 時顯示。

Semaphore 登入時所用之應用程式身分的應用程式金鑰，其密鑰填入 **Secret**。新的儲存區必須填寫此欄位。

## Secret {#secretStorage-dvlsSecret}

當儲存區為 Devolutions Server，且機密來源為 **Store in DB** 時顯示。

**App Key** 之應用程式身分的應用程式密鑰。新的儲存區必須填寫此欄位。開啟現有的儲存區時，此欄位是空的：保持空白即保留已儲存的密鑰，
輸入新的密鑰則會取代它。

## Env var name / Path to the file {#secretStorage-dvlsEnv}

當儲存區為 Devolutions Server，且機密來源為 **From ENV** 或 **From File** 時顯示。

保存應用程式密鑰的環境變數名稱或檔案路徑。新的儲存區必須填寫此欄位。

## Region {#secretStorage-awsRegion}

當儲存區為 AWS Secrets Manager 時顯示。

機密值所在的 AWS 區域，例如 `us-east-1`。此欄位為必填。

## Endpoint URL {#secretStorage-awsEndpointURL}

當儲存區為 AWS Secrets Manager 時顯示。選填。

用來取代該區域預設端點而呼叫的 Secrets Manager API 位址，例如 VPC 端點。將此欄位留空即使用 AWS 的預設端點。

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

當儲存區為 AWS Secrets Manager 時顯示。

開啟時，Semaphore 不儲存任何存取金鑰，而是從伺服器執行所在的環境取得 AWS 憑證，例如其執行個體的 IAM 角色。此時下方存取金鑰的欄位會被停用。

開啟此選項後，儲存時會移除已儲存的存取金鑰。再次關閉它時，需要新的 **Secret Access Key**。

## Access Key ID {#secretStorage-awsAccessKeyId}

當儲存區為 AWS Secrets Manager 時顯示。當 **Use IAM Role / Instance Profile** 開啟時停用。

Semaphore 登入時所用之存取金鑰的 ID，例如 `AKIAIOSFODNN7EXAMPLE`。當 IAM 角色關閉時，此欄位為必填。此 ID 不是機密：它會原樣儲存和顯示。

## Secret Access Key {#secretStorage-awsSecretKey}

當儲存區為 AWS Secrets Manager，且機密來源為 **Store in DB** 時顯示。當
**Use IAM Role / Instance Profile** 開啟時停用。

存取金鑰的秘密。新的儲存區，以及關閉 IAM 角色之後，都必須填寫此欄位。其他情況下，開啟儲存區時此欄位是空的：保持空白即保留已儲存的金鑰，
輸入新的金鑰則會取代它。

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

當儲存區為 AWS Secrets Manager，且機密來源為 **From ENV** 或 **From File** 時顯示。當
**Use IAM Role / Instance Profile** 開啟時停用。

保存秘密存取金鑰的環境變數名稱或檔案路徑。此欄位在與 **Secret Access Key** 相同的情況下為必填。

## Vault URL {#secretStorage-azureVaultURL}

當儲存區為 Azure Key Vault 時顯示。

金鑰保存庫的位址，例如 `https://my-vault.vault.azure.net`。此欄位為必填。

## Tenant ID {#secretStorage-azureTenantId}

當儲存區為 Azure Key Vault 時顯示。

Semaphore 以其身分登入之應用程式所屬的 Microsoft Entra 租用戶 ID。此欄位為必填。

## Client ID {#secretStorage-azureClientId}

當儲存區為 Azure Key Vault 時顯示。

該應用程式的應用程式（用戶端）ID。它需要能夠存取金鑰保存庫的機密值。此欄位為必填。

## Client Secret {#secretStorage-azureClientSecret}

當儲存區為 Azure Key Vault，且機密來源為 **Store in DB** 時顯示。

應用程式的用戶端密碼。新的儲存區必須填寫此欄位。開啟現有的儲存區時，此欄位是空的：保持空白即保留已儲存的密碼，輸入新的密碼則會取代它。

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

當儲存區為 Azure Key Vault，且機密來源為 **From ENV** 或 **From File** 時顯示。

保存用戶端密碼的環境變數名稱或檔案路徑。新的儲存區必須填寫此欄位。

## Read only {#secretStorage-readonly}

開啟時，Semaphore 只會從儲存區讀取。無法在其中建立或變更變數群組的機密值：儲存時會失敗。在 Semaphore 中刪除機密值、金鑰或變數群組時，
其機密值會保留在儲存區中。

關閉時，Semaphore 會將使用此儲存區之變數群組的機密值寫入其中，並在刪除群組時一併從中刪除這些機密值。

## Sync keys enabled {#secretStorage-syncEnabled}

開啟此選項，即可從 **Sync paths** 中設定的路徑，將儲存區中的機密值匯入專案的金鑰儲存。當您在 **Storages** 分頁中按下儲存區的同步按鈕時，
以及依 **Sync paths** 中設定的間隔自動執行時，都會進行匯入。

匯入的金鑰會標示為已同步：其名稱和類型無法變更，在 Semaphore 中刪除它時，機密值會保留在儲存區中。在同步開啟時刪除儲存區，也會一併刪除它所匯入的金鑰。

## Sync paths {#secretStorage-syncPaths}

當 **Sync keys enabled** 開啟時顯示。

開啟一個對話方塊，其中包含要匯入的儲存區路徑以及自動匯入的間隔。按鈕上的數字是路徑的數量。在至少有一個路徑之前，儲存區的同步按鈕會保持停用。

該對話方塊只有**關閉**按鈕：您在其中所做的變更，會透過此對話方塊的**儲存**（新的儲存區則為**建立**）隨儲存區一起儲存。
