---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# 編輯排程

排程會在無人按下**執行**的情況下啟動任務範本的任務：依 cron 表示式反覆執行，或在設定的時間點執行一次。任務會取得此對話方塊中儲存的參數。
所有時間都採用 Semaphore 伺服器的時區，顯示在時間欄位旁：除非管理員設定了 `schedule.timezone` 或
`SEMAPHORE_SCHEDULE_TIMEZONE`，否則為 `UTC`。

## Name {#scheduleForm-name}

排程在「排程」頁面上列出時所用的名稱。此欄位為必填。它不會改變排程的行為。

## Template {#scheduleForm-templateId}

排程所啟動的任務範本。此欄位為必填。任務的執行方式就如同從該範本啟動一樣，使用範本本身的清單、儲存庫、變數群組及其他設定，
並在範本允許任務變更的地方使用下方的參數。

## 任務參數 {#scheduleForm-taskParams}

在選取範本後顯示。

排程所啟動之每個任務的參數。顯示哪些欄位取決於範本：其調查變數，以及它允許任務變更的選項，例如分支、清單、**限制**、
**標籤**、**跳過標籤**或 **CLI args**，以及 Ansible 或 Terraform、OpenTofu 和 Terragrunt 的選項。

每次排程觸發時，伺服器都會檢查範本：一旦範本不再允許任務變更某個值，此處儲存的該值就會被忽略。

## 訊息 {#taskParams-message}

選填。與排程所啟動之每個任務一起儲存並顯示在任務歷史中的備註，例如 `Nightly deploy`。它不會改變任務的執行方式。

## 調查變數 {#taskParams-survey}

在範本有調查變數時顯示。

排程任務所用的範本調查變數值。它們傳給任務的方式，與有人手動填寫調查時相同。必填的變數必須有值，才能儲存排程。

開啟對話方塊時，範本的預設值會被複製進來，並隨排程一起儲存。之後在範本中變更的預設值，不會影響先前儲存的排程。

類型為 secret 的變數的值不會隨排程儲存：排程任務會在沒有該值的情況下執行。

## 分支 {#taskParams-gitBranch}

在範本允許任務覆寫分支時顯示。

排程任務要簽出的儲存庫分支或標籤。此欄位的初始值為範本的分支。只有在範本允許覆寫時，伺服器才會使用此值。

## 清單 {#taskParams-inventoryId}

在範本允許任務選擇清單時顯示。

排程任務執行時所針對的清單，用來取代範本的清單。對於 Terraform 或 OpenTofu 範本，此欄位稱為 **Workspace**，但伺服器只會將此選擇套用於
Ansible 範本，其他範本則保留範本的工作區。

## 除錯 {#taskParams-debug}

在允許除錯的 Ansible 範本中顯示。

以更多的輸出執行 `ansible-playbook`：`-v` 會依**除錯等級**所設定的次數重複。

## 除錯等級 {#taskParams-debugLevel}

在允許除錯的 Ansible 範本中顯示；勾選**除錯**時才會啟用。

輸出的詳細程度，從 `1`（`-v`）到 `6`（`-vvvvvv`）。預設值為 `4`。

## 模擬執行 {#taskParams-dryRun}

在 Ansible 範本中顯示，除非範本將其隱藏。

以 `--check` 執行 Playbook：Ansible 會回報它將變更的內容，而不做任何變更。

## 差異 {#taskParams-diff}

在 Ansible 範本中顯示，除非範本將其隱藏。

以 `--diff` 執行 Playbook：Ansible 會顯示它對檔案和範本所做的變更。

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

在允許任務變更此選項的 Ansible 範本中顯示。

在 Playbook 執行之前，略過安裝 `requirements.yml` 檔案中的角色和集合。當需求已經安裝好，且任務不應連線到 Ansible Galaxy
時，請勾選此選項。

## Plan {#taskParams-plan}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

只執行 `plan`。任務在 plan 之後結束，不套用任何變更。

## Destroy {#taskParams-destroy}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

在 plan 和 apply 中加上 `-destroy`：任務會銷毀工作區的資源，而不是建立或更新它們。

## Auto Approve {#taskParams-autoApprove}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

不等待確認即套用 plan。只有在範本允許自動核准時才有效；設定為核准每個任務的範本，無論如何都會套用 plan。

若未使用自動核准，排程任務會在 plan 之後停止，並等候有人在任務頁面上確認或拒絕它。

## Upgrade {#taskParams-upgrade}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

以 `-upgrade` 執行 `init`：提供者和模組會更新到設定所允許的最新版本。

## Reconfigure {#taskParams-reconfigure}

在 Terraform、OpenTofu 或 Terragrunt 範本中顯示。

以 `-reconfigure` 取代 `-migrate-state` 執行 `init`：後端會重新設定，現有的狀態不會被遷移。

## Run at {#scheduleForm-runAt}

在只執行一次的排程中顯示。

任務開始的日期和時間，採用欄位末端所顯示的伺服器時區。新排程的初始值為下一個整點。此時間點必須在未來：伺服器會拒絕已經過去的時間。

任務開始之後，排程會被關閉；若勾選了 **Delete after run**，則會被刪除。若 Semaphore 在該時間點沒有在執行，任務之後也不會再啟動：
下次啟動時，排程會在未執行的情況下被關閉或刪除。

## Delete after run {#scheduleForm-deleteAfterRun}

在只執行一次的排程中顯示。

在任務開始後刪除排程。保持不勾選即可保留已關閉的排程，以便您設定新的時間並再次開啟它。

## Show cron format {#scheduleForm-rawCron}

在 cron 排程中顯示。

在輸入 cron 表示式（開啟）與透過 **Timing** 和時間核取方塊建立表示式（關閉）之間切換。兩者編輯的是同一個表示式。瀏覽器會記住您的選擇。

當表示式無法以核取方塊表示時，此開關會被鎖定為開啟：例如表示式無效，或使用了 `@daily` 或 `CRON_TZ=` 之類的形式。

## Cron {#scheduleForm-cronFormat}

在 cron 排程中，當 **Show cron format** 開啟時顯示。

由五個欄位組成的 cron 表示式：分鐘、小時、日、月、星期（`0` 為星期日）。此欄位為必填，伺服器會在您輸入時檢查表示式。
新排程的初始值為 `* * * * *`，即每分鐘。

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

伺服器也接受 `@hourly`、`@daily`、`@weekly`、`@monthly`、`@yearly` 和 `@every 1h30m`，以及
`CRON_TZ=Europe/Berlin` 之類用於指定其他時區的前綴。下方的預覽無法顯示這些表示式的下一次執行時間。

## Timing {#scheduleForm-timing}

在 cron 排程中，當 **Show cron format** 關閉時顯示。

任務執行的頻率。下方的核取方塊會隨之改變：

- **Hourly**：**Minutes**。
- **Daily**：**Hours** 和 **Minutes**。
- **Weekly**：**Weekdays**、**Hours** 和 **Minutes**。
- **Monthly**：**Days**、**Hours** 和 **Minutes**。
- **Yearly**：**Months**、**Days**、**Hours** 和 **Minutes**。

## Months {#scheduleForm-months}

當 **Timing** 為 **Yearly** 時顯示。

任務執行的月份。勾選多個月份，即會在每個勾選的月份中執行。都不勾選時，任務會在每個月執行。

## Weekdays {#scheduleForm-weekdays}

當 **Timing** 為 **Weekly** 時顯示。

任務執行的星期幾。都不勾選時，任務會在每一天執行。

## Days {#scheduleForm-days}

當 **Timing** 為 **Monthly** 或 **Yearly** 時顯示。

任務執行的日期（每月的第幾天）。沒有該日期的月份會被略過：設定為 31 日的任務不會在四月執行。都不勾選時，任務會在每一天執行。

## Hours {#scheduleForm-hours}

當 **Timing** 為 **Daily**、**Weekly**、**Monthly** 或 **Yearly** 時顯示。

任務執行的小時，採用右側所顯示的時區。都不勾選時，任務會在每個小時執行。

## Minutes {#scheduleForm-minutes}

在 cron 排程中，當 **Show cron format** 關閉時顯示。

任務在一小時中執行的分鐘，以五分鐘為間隔。都不勾選時，任務會每分鐘執行：若任務應每小時或每天執行一次，請至少勾選一個，例如 `:00`。

## Next run time {#scheduleForm-nextRun}

依表單中目前的排程，下一個任務將會開始的時間：分別以伺服器的時區和您瀏覽器的本地時間顯示。破折號表示表示式無效，或是預覽無法解讀的表示式，例如
`@daily`。

## 已啟用 {#scheduleForm-active}

排程是否會啟動任務。取消勾選即可保留排程而不執行它。排程清單中的開關也有相同作用。新排程預設為已啟用。
