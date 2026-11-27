---
form: src/components/AppForm.vue
---

# Edit App

應用程式是 Semaphore 用來執行任務的工具：Ansible、Terraform、OpenTofu、Terragrunt、Bash、
Python、PowerShell，或您自己的命令。每個任務範本都是為某一個應用程式建立的，此對話方塊設定伺服器如何啟動該應用程式，
以及該應用程式在介面中如何顯示。

這些設定儲存在資料庫中，並在儲存它們的伺服器上立即生效；叢集的其他節點會在重新啟動時讀取它們。被清空的欄位會保留舊值，
直到伺服器重新啟動。

## ID {#appForm-id}

必填。應用程式儲存時所用的鍵；任務範本透過此鍵參照其應用程式。內建的應用程式為 `ansible`、`terraform`、`tofu`、
`terragrunt`、`pulumi`、`bash`、`python` 和 `powershell`。任何其他 ID 都會新增一個您自己的應用程式：
Semaphore 會像執行腳本一樣執行它，使用 **Path** 中指定的命令；若 **Path** 為空，則使用與 ID 同名的命令。

請使用不含句點的 ID。變更現有應用程式的 ID 並不會將其重新命名：設定會以新 ID 儲存為一個新的應用程式，而舊的應用程式仍會保留。

## Icon {#appForm-icon}

選填。Material Design Icons 圖示集中某個圖示的名稱，不含 `mdi-` 前綴，例如 `language-python`。
將此欄位留空，即保留應用程式的內建圖示；此時您自己的應用程式會顯示一個問號。

## 名稱 {#appForm-title}

選填。應用程式在應用程式清單以及**新增範本**選單中顯示的名稱。將此欄位留空，即保留內建名稱，例如
`Ansible Playbook`；此時您自己的應用程式沒有名稱，因此請為它命名。

## Path {#appForm-path}

選填。Semaphore 用來取代預設命令執行的命令，可以是在伺服器的 `PATH` 中查找的名稱，也可以是完整路徑，例如
`/usr/local/bin/terraform-1.9`。預設命令為 `terraform`、`tofu`、`terragrunt`、`bash`、`python3` 和
`powershell`；您自己的應用程式會執行與其 ID 同名的命令。

Ansible 會忽略此欄位：Semaphore 一律執行 `ansible-playbook` 和 `ansible-galaxy`。

## Priority {#appForm-priority}

應用程式在應用程式清單以及**新增範本**選單中的位置：數字較大的應用程式排在前面。內建應用程式從 Ansible 的 1000
到 Python 的 500。優先順序為 `0` 或更小的內建應用程式，會在伺服器啟動時恢復其內建的優先順序。

## Args {#appForm-args}

選填。緊接在命令之後放在命令列上的參數，每個標籤一個。點選 **+** 新增一個參數，點選標籤以修改它，拖曳標籤以變更其順序。

- Terraform、OpenTofu 和 Terragrunt：這些參數位於 Semaphore 為每個命令加上的參數之前，例如 `init` 和
  `apply`。
- 腳本類應用程式：這些參數位於範本的腳本之前，並取代預設參數：PowerShell 預設以 `-File` 執行，因此請將 `-File`
  放在清單的最後。
- Ansible 會忽略此欄位。

## Active {#appForm-active}

啟用中的應用程式會出現在**新增範本**選單中。取消勾選此方塊，即可在該選單中隱藏此應用程式：已使用它建立的範本會保留其應用程式，
並且仍可執行。

新的應用程式預設不勾選此方塊：勾選它，應用程式才可供使用。
