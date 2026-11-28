---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# 編輯端點

端點是 webhook 的入口：其他服務呼叫的 URL，以及驗證這些請求的方式。整合會連結到端點，並在端點接受的請求上啟動自己的執行。端點的 URL（即別名）在「端點」分頁中新增。

## 名稱 {#endpointForm-name}

端點在「端點」分頁與整合表單中顯示的名稱。此欄位為必填。

## 描述 {#endpointForm-description}

選填。關於端點用途的備註，顯示在其名稱旁邊。

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
