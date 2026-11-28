---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# 編輯整合

整合是一組從同一個端點啟動的執行。當送往該端點某個 URL 的請求通過授權時，整合中比對器接受該請求的每個執行都會啟動。

## 名稱 {#integrationGroupForm-name}

整合在「整合」頁面中顯示的名稱。此欄位為必填。

## 端點 {#integrationGroupForm-endpointId}

其 URL 會啟動此整合的執行的端點。多個整合可以使用同一個端點。此欄位為必填。

## 覆寫授權 {#integrationGroupForm-authOverride}

保持關閉時，請求會依端點的方式驗證。開啟後，會使用在下方出現的欄位中設定的獨立方式與密鑰來驗證此整合的請求；此時端點的設定不再套用於此整合。

## 驗證方式 {#integrationAuthForm-authMethod}

Semaphore 檢查請求是否來自您所預期之服務的方式。未通過檢查的請求不會啟動任何東西，而 Webhook 對它的回應與對其他任何請求的回應相同。

- **無**：不檢查。任何知道此 URL 的人都可以啟動任務。
- **GitHub Webhooks**：`X-Hub-Signature-256` 標頭必須包含以密鑰對主體產生的 HMAC-SHA256 簽章，與 GitHub 傳送的格式相同。
- **Bitbucket Webhooks**：對 `X-Hub-Signature` 標頭進行相同的檢查，與 Bitbucket 傳送的格式相同。
- **Token**：**驗證標頭** 中指定的標頭必須包含密鑰本身。
- **HMAC** 和 **HMAC (SHA-512)**：**驗證標頭** 中指定的標頭必須包含以密鑰對主體產生的 HMAC-SHA256 或
  HMAC-SHA512，以小寫十六進位表示，不含 `sha256=` 之類的前綴。
- **BasicAuth**：請求必須在其 `Authorization` 標頭中帶有**驗證密鑰**所指定之金鑰的登入名稱和密碼。

## 驗證標頭 {#integrationAuthForm-authHeader}

當 **驗證方式** 為 **Token**、**HMAC** 或 **HMAC (SHA-512)** 時顯示。

攜帶權杖或簽章的請求標頭名稱，例如 `X-Webhook-Token`。名稱的大小寫無關緊要。沒有該標頭的請求會被拒絕。

## 驗證密鑰 {#integrationAuthForm-authSecretId}

當 **驗證方式** 不是 **無** 時顯示。

金鑰儲存中用來檢查請求的金鑰。對於 **BasicAuth**，只有類型為**使用密碼登入**的金鑰適用。對於其他方法，類型為**字串**的金鑰也適用：
檢查時會使用其值，或使用**使用密碼登入**金鑰的密碼。

除了 **無** 之外，每種方法都需要此欄位：沒有金鑰或密鑰為空時，每個請求都會被拒絕。
