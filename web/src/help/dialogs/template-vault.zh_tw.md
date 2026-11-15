---
form: src/components/TemplateVaults.vue
---

# 保險庫

保險庫為 Ansible 提供 Playbook 儲存庫中以 Ansible Vault 加密之檔案和變數的密碼。範本的每個保險庫都會成為
`ansible-playbook` 的一個 `--vault-id` 選項。儲存範本對話方塊時，保險庫會隨範本一起儲存。

## 保險庫 ID {#templateVaultDialog-name}

選填。此密碼所對應的 vault ID，即內容加密時所用的標籤（`ansible-vault encrypt --vault-id prod@prompt`）。
將此欄位留空即為 vault ID `default`。

ID 會以小寫儲存。範本的每個保險庫都需要自己的 ID，且只能有一個是空的。

## 類型 {#templateVaultDialog-type}

密碼的來源：

- **密碼**：金鑰儲存中的金鑰。當 Ansible 要求密碼時，Semaphore 會輸入其密碼，因此密碼永遠不會被寫入檔案。
- **客戶端腳本**：會印出密碼的 vault 密碼客戶端腳本；由 Ansible 執行它。

## 保險庫密碼 {#templateVaultDialog-vaultKeyId}

當類型為**密碼**時顯示。

保存 vault 密碼的金鑰。只會列出類型為**使用密碼登入**的金鑰；Semaphore 會使用其密碼並忽略登入名稱。此欄位為必填。

## 腳本路徑 {#templateVaultDialog-script}

當類型為**客戶端腳本**時顯示。

腳本的路徑，例如 `scripts/vault-pass-client.py`。相對路徑以 Ansible 執行時所在的目錄為基準：儲存庫的根目錄，或範本的工作目錄（若有設定）。
此欄位為必填。

檔案名稱必須以 `-client` 結尾，可以有副檔名也可以沒有：只有這樣，Ansible 才會將該檔案當作客戶端腳本執行，並以 `--vault-id`
將 vault ID 傳給它。腳本會在執行任務的機器上，以 Semaphore 執行時所用的使用者身分執行。
