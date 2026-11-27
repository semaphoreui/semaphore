---
form: src/components/RepositoryForm.vue
---

# 編輯 Repository

儲存庫是任務範本的 Playbook、Terraform 程式碼或腳本的來源：一個 Git 儲存庫，或執行任務之機器上的一個目錄。每個任務開始前，
Semaphore 會複製該 Git 儲存庫，或更新先前複製的副本，簽出分支，然後在該副本中執行任務。

## 名稱 {#repository-name}

儲存庫在「儲存庫」頁面，以及任務範本和清單的儲存庫選擇器中顯示的名稱。此欄位為必填。

## URL 或路徑 {#repository-gitUrl}

程式碼所在的位置。此欄位為必填。欄位下方的標籤會顯示您所輸入位址的種類，點選某個標籤會將位址改寫成該種類。

- **ssh**：`git@github.com:org/repo.git` 或 `ssh://git@github.com/org/repo.git`。Git 會以位址中的使用者，
  以及在**存取金鑰**中選擇的 **SSH 金鑰**登入。
- **https** 和 **http**：`https://github.com/org/repo.git`。私人儲存庫需要**使用密碼登入**金鑰；公開儲存庫需要**無**。
  請避免使用 `http`：此時登入名稱和密碼會以未加密的方式傳輸。
- **git**：`git://host/org/repo.git`，唯讀的 Git 協定，不需要登入。
- **file**：`file:///srv/git/repo.git`，位於執行任務之機器磁碟上的 Git 儲存庫，會像遠端儲存庫一樣被複製。
- **絕對路徑**：絕對路徑，例如 `/srv/playbooks` 或 `D:\playbooks`。Semaphore 不會複製它：任務會直接在該目錄中執行，
  該目錄必須存在於執行任務的機器上（Semaphore 伺服器或執行器）。分支不適用於它。

變更位址會移除 Semaphore 伺服器從舊位址複製的副本。

## 分支 {#repository-gitBranch}

Semaphore 複製並保持更新的分支，例如 `main`。此欄位為必填，但絕對路徑除外，此時此欄位會被鎖定。

任務範本可以指定自己的分支，其優先於此處的分支；範本也可以讓其任務在啟動時選擇分支。標籤也可以使用，但標籤無法被
pull，因此 Semaphore 會在每個任務開始前重新複製儲存庫。

## Single checkout for all templates {#repository-monorepo}

預設為關閉：每個任務範本都會取得自己的儲存庫副本。勾選此方塊即可只複製儲存庫一次，並讓所有使用它的範本共用該副本，
適用於大到無法為每個範本各保存一份副本的儲存庫。

共用的副本會停留在此儲存庫的分支上：使用它的範本不能有自己的分支，其任務也不能選擇其他分支或提交。絕對路徑時此選項會被鎖定。

## 存取金鑰 {#repository-sshKeyId}

金鑰儲存中 Git 用來登入儲存庫的金鑰。此欄位為必填。它提供三種類型的金鑰，應選擇哪一種取決於 URL：

- **SSH 金鑰**用於 `ssh` URL。在複製或更新期間，Semaphore 會將私鑰載入 SSH 代理程式。使用者是 URL 中的使用者，例如
  `git@github.com:org/repo.git` 中的 `git`。
- **使用密碼登入**用於 `https` URL：登入名稱和密碼，或 Git 伺服器的存取權杖。登入名稱為空時，只會傳送密碼並將其當作使用者名稱，
  這正是某些伺服器對權杖的要求。相同的憑證也會提供給 `ansible-galaxy`，用於同一伺服器上的角色和集合。
- **無**用於公開儲存庫、`git` 或 `file` URL，或絕對路徑。

與 URL 不符的金鑰不會為 Git 提供任何憑證，此時私人儲存庫將無法複製。
