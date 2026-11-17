---
form: src/components/KeyForm.vue
---

# 編輯 Key

金鑰儲存中的金鑰保存專案使用但不顯示的憑證：Ansible 登入主機時使用的 SSH 金鑰或密碼、Git 儲存庫的憑證、Ansible vault
的密碼、整合的密鑰。清單、儲存庫、任務範本和整合會依名稱參照金鑰。儲存之後，機密值永遠不會再傳回瀏覽器。

## 金鑰名稱 {#key-name}

金鑰在金鑰儲存中，以及在清單、儲存庫、任務範本和整合的金鑰選擇器中顯示的名稱。此欄位為必填。請選擇一個能說明此金鑰用途的名稱，例如
`Deploy key of the web servers`。

## 來源 {#key-source}

金鑰的機密值保存的位置。各分頁：

- **Local**：機密值在此對話方塊中輸入，並儲存在 Semaphore 的資料庫中；若管理員已為存取金鑰設定加密金鑰，則會加密儲存。這是預設值。
- **Storage**：機密值保存在專案的外部密鑰儲存區中，於下方的 **Storage** 中選擇。此分頁只在 Pro 版本中提供。
  密鑰儲存區本身是在金鑰儲存的 **Storages** 分頁中設定的。
- **Env**：機密值從 Semaphore 伺服器程序的環境變數讀取。請在分頁下方的欄位中輸入變數名稱。
- **File**：機密值從 Semaphore 伺服器上的檔案讀取。請在分頁下方的欄位中輸入檔案的絕對路徑。該檔案必須位於伺服器的機密目錄中
  （其設定中的 `dirs.secrets`，預設為 `/tmp/semaphore`）；位於該目錄之外的路徑，或含有 `..` 的路徑，在使用金鑰時會失敗。
  檔案結尾的換行會被捨棄。

使用 **Env** 和 **File** 時，Semaphore 永遠不會寫入機密值：每次任務使用該金鑰時它都會讀取其值，因此請在機密值所在之處變更它。
該值必須符合金鑰的**類型**。**字串**金鑰會原樣採用該值。**SSH 金鑰**或**使用密碼登入**金鑰則需要含有該類型欄位的 JSON，例如：

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

編輯金鑰時，在勾選**覆寫**之前，各分頁都會被鎖定。從密鑰儲存區同步而來的金鑰會保留其來源，其名稱和類型也無法變更。

## Storage {#key-sourceStorageId}

在選擇 **Storage** 分頁時顯示。

保存機密值的專案密鑰儲存區：HashiCorp Vault、OpenBao、Devolutions Server、AWS Secrets Manager 或
Azure Key Vault，依金鑰儲存的 **Storages** 分頁中的設定。

- 非唯讀的儲存區會接收您在此對話方塊中輸入的機密值：Semaphore 會在您儲存時將其寫入該處。刪除金鑰也會從儲存區中刪除該機密值。
  保存在這類儲存區中的金鑰，之後無法移到其他儲存區。
- 唯讀的儲存區只會被讀取。此時對話方塊不會顯示機密值的欄位，而 Semaphore 會在每次任務使用該金鑰時從儲存區讀取它。

## Source Key {#key-sourceStorageKey}

在選取了 **Storage** 時顯示。

機密值在該儲存區中的位置，例如 HashiCorp Vault 中該機密值的路徑。

## 類型 {#key-type}

金鑰所保存的憑證種類。此欄位為必填，其下方的欄位會隨類型而改變。

- **SSH 金鑰**：私鑰，可附帶登入名稱和密碼短語。Ansible 用它連線到主機，Git 用它複製 `ssh` URL 的儲存庫。
- **使用密碼登入**：登入名稱和密碼。Ansible 用它們連線到主機，或用它們切換成其他使用者（清單的 **Sudo 憑證**）；Git
  用它們複製 `https` 儲存庫；密碼可用來開啟任務範本的 Ansible vault。
- **字串**：不帶登入名稱的單一機密值：權杖、API 金鑰、密碼。請將它用於整合的密鑰。它不能用作 SSH 金鑰、sudo 憑證、vault
  的密碼或儲存庫的存取金鑰。
- **無**：沒有憑證。請將它用於不需要登入的儲存庫（公開的 `https` 儲存庫、`file` URL 或本機路徑），以及其主機不需要憑證或自行定義連線方式的清單。

## 使用者名稱 {#key-loginPasswordLogin}

當類型為**使用密碼登入**時顯示。選填。

憑證的使用者名稱。

- Ansible 會以此使用者身分連線到主機（`--user`）；作為清單的 **Sudo 憑證**時，它是 Ansible 切換成的使用者
  （`--become-user`）。當此欄位為空時，Ansible 會使用清單或其本身設定所指定的使用者。
- 對於 `https` 儲存庫，Git 會以此使用者和密碼登入。若 Git 伺服器只接受權杖本身，請將此欄位留空：此時密碼會被當作使用者名稱傳送。

## 密碼 {#key-loginPasswordPassword}

當類型為**使用密碼登入**時顯示。必填。

憑證的密碼。Semaphore 會用它回應 Ansible 的密碼提示：作為使用者憑證時是 SSH 密碼（`--ask-pass`），作為 sudo 憑證時是
`become` 密碼（`--ask-become-pass`），用於 Ansible vault 時是 vault 密碼。對於 `https` 儲存庫，它是 Git
伺服器的密碼或存取權杖。

## 值 {#key-stringValue}

當類型為**字串**時顯示。必填。

機密值本身，例如整合用來檢查其所收到之請求的權杖。它會依輸入的內容原樣儲存。

## 使用者名稱 {#key-sshLogin}

當類型為 **SSH 金鑰**時顯示。選填。

Ansible 連線到主機時使用的使用者（`--user`）。當此欄位為空時，Ansible 會使用清單或其本身設定所指定的使用者。

對於 Git 儲存庫，請改將使用者寫在儲存庫 URL 中，例如 `git@github.com:org/repo.git`。

## Passphrase {#key-sshPassphrase}

當類型為 **SSH 金鑰**時顯示。選填。

用來加密私鑰的密碼短語。若金鑰沒有密碼短語，請將此欄位留空。錯誤的密碼短語會讓每個使用此金鑰的任務在載入金鑰時失敗。
由 Semaphore 產生的金鑰沒有密碼短語：此時此欄位中的內容會被捨棄。

## Generate SSH Key {#key-generateSshKey}

當類型為 **SSH 金鑰**時顯示。

Semaphore 會在您儲存時建立一組新的金鑰對：一把 2048 位元的 RSA 私鑰，保存在此金鑰中；以及其公鑰，儲存後會在對話方塊中顯示一次，
之後每次編輯此金鑰時都會顯示在 **Public key** 之下。請將公鑰加入主機上的 `~/.ssh/authorized_keys`，或作為 Git
儲存庫的部署金鑰。

勾選此方塊時，私鑰欄位會被鎖定。對於現有的金鑰，請先勾選**覆寫**：舊的金鑰對會被取代。機密值從 **Env**、**File**
或唯讀儲存區讀取的金鑰無法產生，因為 Semaphore 沒有地方可以寫入。

## 私鑰 {#key-sshPrivateKey}

當類型為 **SSH 金鑰**時顯示。除非勾選了 **Generate SSH Key**，否則為必填。

私鑰，採用 OpenSSH 或 PEM 格式，包含其 `BEGIN` 和 `END` 行。對於任務，Semaphore 會將它載入自己的 SSH
代理程式，Ansible 和 Git 會在任務期間使用它。

## Public key {#key-publicKey}

在您編輯由 Semaphore 產生的 SSH 金鑰，且未勾選**覆寫**時顯示。

所產生金鑰對的公開部分。**Show** 會將它展開，複製按鈕會將它放到剪貼簿。請將它加入主機上的
`~/.ssh/authorized_keys`，或作為 Git 儲存庫的部署金鑰。

這是 Semaphore 最後一次產生的金鑰對的公鑰。當您透過**覆寫**手動取代私鑰之後，這裡顯示的金鑰就不再與它相符。

## 覆寫 {#key-overrideSecret}

在您編輯現有的金鑰時顯示。

Semaphore 永遠不會再次顯示已儲存的機密值，因此機密值的欄位都是空的並且被鎖定。勾選此方塊即可取代機密值：來源、類型和機密值的欄位會開放編輯，
其中的所有內容都會重新儲存，因此請重新輸入完整的機密值，而不只是要變更的部分。若未勾選，儲存時只會變更金鑰的名稱。

儲存金鑰也會移除 Semaphore 伺服器為使用此金鑰的儲存庫所保存的副本：下一個任務會重新複製它們。
