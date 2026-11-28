---
form: src/components/ProxyForm.vue
---

# 編輯 Proxy

代理是一台主機，專案的工作經過它連到無法直接連線的伺服器。在儲存庫的 **Proxy** 欄位中選擇它，
即可透過它取得儲存庫；在清單中選擇它，Ansible 即透過它連線主機。

## 名稱 {#proxyForm-name}

必填，且在專案中唯一。儲存庫與清單以此名稱顯示代理。

## 類型 {#proxyForm-type}

- **SSH (jump host)**：Semaphore 透過 SSH 連到該主機，再從那裡開啟到伺服器的連線。只有此類型
  可以依賴另一個代理。
- **SOCKS5**、**HTTP**、**HTTPS**：轉送連線的代理伺服器。對於 `https` 儲存庫，Git 會自行使用它；
  SSH 連線由 Semaphore 經由它轉送，因此不必安裝其他工具。

## Host {#proxyForm-host}

必填。主機名稱或 IP 位址，不含使用者、連接埠與協定。主機名稱只能包含字母、數字、點、連字號與底線。

## Port {#proxyForm-port}

選填，1 到 65535。留空時，SSH 跳板機為 `22`。其他類型請填寫連接埠：未填時，Git 與 SSH 連線會假設
不同的連接埠。

## User {#proxyForm-user}

僅限 SSH：跳板機上的登入名稱。留空表示執行 Semaphore 的使用者。可用字母、數字、點、連字號與底線，
不能以連字號開頭。SOCKS5 或 HTTP 代理改為從**存取金鑰**取得登入名稱。

## 存取金鑰 {#proxyForm-sshKeyId}

選填。金鑰儲存中用來登入代理本身的金鑰，通常不是其後伺服器的金鑰：

- SSH 跳板機使用 **SSH 金鑰**；
- SOCKS5、HTTP 與 HTTPS 使用**使用密碼登入**。密碼透過環境變數傳遞，絕不出現在命令列中。

## Requires Proxy {#proxyForm-requiresProxyId}

僅限 SSH，選填。需要先經過的專案中另一個 SSH 代理，用於只能從另一台跳板機連到的跳板機。一條鏈最多
包含 5 個代理，且不能回到鏈中已有的代理。
