---
form: src/components/InventorySourceForm.vue
---

# 編輯來源

來源從 Git 儲存庫中的檔案匯入受管清單的主機和群組。每次按下 Sync 都會讀取該檔案，檔案中描述的主機和群組會與它保持一致。伺服器需要 Ansible 並能存取該儲存庫。

## 名稱 {#inventorySourceForm-name}

來源的名稱，顯示在它的卡片上以及它帶來的主機旁。此欄位必填。

## 儲存庫 {#inventorySourceForm-repository}

讀取檔案所用的專案儲存庫。此欄位必填。使用它的分支。

## 路徑 {#inventorySourceForm-path}

清單檔案在儲存庫中的路徑，例如 `inventory/prod.yml`，也可以是存放清單檔案的目錄。此欄位必填。

檔案以靜態 YAML 或 INI 讀取，指令碼和外掛設定檔會被拒絕。在目錄中，Ansible 不略過的每個檔案都必須是有效的清單，否則同步失敗，什麼也不會改變。Ansible 在目錄中會略過一些副檔名，例如 `.cfg`、`.md` 和 `.txt`。若儲存庫中有指向儲存庫外部的連結，則會被拒絕。
