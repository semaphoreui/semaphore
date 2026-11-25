---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

訂閱金鑰會在此伺服器上解鎖 Semaphore PRO 和 Enterprise 的功能。此對話方塊用來啟用金鑰，並顯示目前有效金鑰的方案：其持有者、
狀態、到期時間，以及它所允許和已使用的 Pro 使用者、執行器及其他項目的數量。它由管理員從 Pro 版本的使用者選單中開啟。在訂閱生效之前，
此對話方塊的名稱為 **Upgrade to Semaphore PRO**。

## Enter your PRO or EE key {#subscription-key}

訂閱金鑰，即您隨 PRO 或 Enterprise 訂閱取得的金鑰。請將它貼到此處，或使用上傳按鈕從檔案載入，然後按下
**Activate New Key**。啟用時此欄位為必填。

當金鑰設定在伺服器的設定中時（`subscription.key` 或 `subscription.key_file`，或變數
`SEMAPHORE_SUBSCRIPTION_KEY` 和 `SEMAPHORE_SUBSCRIPTION_KEY_FILE`），此欄位為唯讀：此時只能在該處變更金鑰。

## 上傳金鑰檔案 {#subscription-upload}

在尚未啟用任何訂閱時顯示。

從您電腦上的檔案（`.txt`、`.key`、`.pem` 或 `.lic`）將金鑰讀入上方的欄位。在您按下 **Activate New Key**
之前，金鑰不會被啟用。

## 金鑰操作 {#subscription-actionsMenu}

在訂閱有效或已過期時顯示。

針對伺服器金鑰的操作選單：

- **Reload**：重新取得訂閱的狀態，例如在續訂之後。
- **Upload**：從檔案將新的金鑰讀入上方的欄位，與上傳按鈕的作用相同。
- **Reset**：立即從伺服器移除金鑰，不經確認。

## Activate New Key {#subscription-activate}

在此伺服器上啟用欄位中的金鑰。當金鑰有效時，訂閱的方案會顯示在下方。若您的帳號還不是 Pro 使用者，且方案中有空出的 Pro
使用者席位，對話方塊會詢問是否將您的帳號設為 Pro 使用者。

已啟用過的金鑰會以警告的形式回報。當金鑰設定在伺服器的設定中時，此按鈕會停用。

## Buy Pro {#subscription-buy}

在瀏覽器的新分頁中，開啟 Semaphore 入口網站上購買 PRO 訂閱的頁面。

## Get 30-day free trial {#subscription-trial}

在沒有有效訂閱時顯示。

在瀏覽器的新分頁中，開啟 Semaphore 入口網站上核發 30 天免費試用金鑰的頁面。請將您取得的金鑰貼到上方的欄位。
