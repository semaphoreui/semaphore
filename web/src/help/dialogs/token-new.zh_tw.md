---
form: src/views/Tokens.vue
---

# New Token

API 權杖讓腳本或其他服務能以您的身分呼叫 Semaphore API，擁有與您帳號相同的權限。請在請求的 `Authorization` 標頭中傳送它：
`Bearer` 加上權杖。權杖只會在建立後立即顯示一次：請在那時複製它。

## Token Name {#token-name}

選填。用來在權杖清單中區分此權杖的名稱，例如使用它的腳本名稱。它不是權杖的一部分，也不會改變權杖可以執行的操作。

## Expires {#token-expires}

權杖停止運作的時間。請選擇從現在起算的一段期間；選擇 **Never** 表示權杖在被刪除之前都有效；選擇 **Custom** 則可自行選擇時間點。

永不過期的權杖在您從清單中將其刪除之前都有效。對於提供給腳本或其他服務的權杖，建議設定期間。

## Expiration date {#token-expiresAt}

當 **Expires** 為 **Custom** 時顯示。

權杖停止運作的日期和時間，採用您瀏覽器的時區。此時間點必須在未來：伺服器會拒絕已經過期的權杖。
