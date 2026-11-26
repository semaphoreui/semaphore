---
form: src/components/HostConfigForm.vue
---

# Edit mapping

對應會告訴 Semaphore，針對某個 SSH 主機或 Git URL 要使用金鑰儲存中的哪個憑證。儲存庫只帶有一把金鑰，但任務也可能連線到其他伺服器：
子模組、`requirements.yml` 中的角色和集合、Terraform 模組、清單儲存庫。專案的對應適用於其任務的每個 Git 和 SSH 連線，
以及 Semaphore 為該專案執行的 Git 命令，例如檢查儲存庫是否有新提交。您儲存庫中的檔案都不會改變。

## 類型 {#hostConfigForm-type}

對應所比對的對象：

- **Host**：SSH 主機名稱。任務對該主機的每個 SSH 連線都會使用此對應的金鑰：透過 SSH 的 Git，以及 Ansible
  連線到同名的清單主機。對於同一個主機，此對應優先於伺服器的 SSH 設定。
- **URL**：Git URL 或其開頭部分。Git 會使用此對應的憑證，連線到該 URL 之下的每個儲存庫。

切換為 **Host** 時，會清除主機對應無法使用的 **Credential**。

## 主機或 URL {#hostConfigForm-host}

此欄位為必填。每個主機或 URL 在每種類型中只能有一個對應。

對於 **Host**：主機名稱，須與 Git 或 Ansible 連線時所用的完全相同，例如 `github.com` 或 `10.0.0.21`。
只能使用字母、數字、句點、連字號和底線：不含使用者、不含連接埠、不含萬用字元。

對於 **URL**：以 `https://` 或 `http://` 開頭的位址，其中不含登入資訊，也不含空格、引號或 `=`。
可以提供完整的儲存庫，或以 `/` 結尾的前綴以涵蓋一組儲存庫：

```
https://github.com/acme/
https://github.com/acme/infra.git
```

當多個對應都符合時，最長的那個勝出：上面的儲存庫對應會針對該儲存庫覆寫 `acme` 的對應。

使用 SSH 金鑰時，URL 對應會讓 Git 透過 SSH 連線到該 URL 的主機的連接埠 `22`，無論 URL 指定的是哪個連接埠。

## Credential {#hostConfigForm-sshKeyId}

金鑰儲存中的金鑰，用於此對應所比對的連線。此欄位為必填。

- **Host** 只接受 **SSH 金鑰**。若金鑰設有登入名稱，SSH 會以該使用者身分連線。
- **URL** 接受 **SSH 金鑰**或**使用密碼登入**。SSH 金鑰會以其登入名稱連線；若沒有登入名稱，則以 `git` 連線。
  使用密碼登入只適用於 `https://` URL；若要使用權杖，請將權杖填入密碼並將登入名稱留空。

使用密碼登入的憑證會透過 Git 的環境傳給它，而不是透過其命令列，任務日誌中也只會顯示原始的 URL。
