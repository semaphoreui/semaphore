---
form: src/components/ProjectForm.vue
---

# 新增專案

專案保存一項工作的任務範本、清單、儲存庫、變數群組和金鑰，以及參與這項工作的人員。您會成為您所建立之專案的擁有者。
只有管理員可以建立專案，除非伺服器的管理員允許所有使用者建立專案。下列所有設定之後都可以在專案的「設定」頁面中變更。

## 專案名稱 {#newProject-name}

專案在側邊欄頂端的專案選單中顯示的名稱。此欄位為必填。

## 最大平行任務數 {#newProject-maxParallelTasks}

此專案可以同時執行的任務數量。超過上限後啟動的任務會在佇列中等候，直到某個執行中的任務結束。

將此欄位留空或設為 `0`，表示專案本身不設上限。無論哪種情況，由管理員設定的整個伺服器的上限都會適用。

## Telegram 聊天 ID {#newProject-tg}

選填。接收此專案警報的 Telegram 聊天室。若此欄位為空，警報會傳送到為整個伺服器設定的聊天室。

只有在管理員已為伺服器啟用 Telegram 警報，且此專案允許警報時，此欄位才會生效。

## 啟用此專案的警報 {#newProject-alert}

為此專案的任務啟用警報。警報的傳送管道由管理員為整個伺服器設定：電子郵件、Telegram、Slack、Rocket.Chat、
Microsoft Teams、DingTalk 或 Gotify。

- 任務失敗時會寄送電子郵件，收件者是能夠檢視該任務、且在自己的帳號中啟用了警報的專案成員。
- 任務成功、失敗或等待確認時，會向通訊軟體傳送訊息。

## Demo {#newProject-demo}

在新專案中填入範例，供您試用 Semaphore：公開儲存庫 `https://github.com/semaphoreui/semaphore-demo.git`、
清單、一把 vault 密碼金鑰，以及適用於 Ansible、OpenTofu、Terragrunt、Bash 和 PowerShell 的任務範本，分別歸入
**Build**、**Deploy** 和 **Tools** 檢視。

若要建立不含範例的專案，請保持此開關關閉：專案只會取得金鑰 **None** 和檢視 **All**。這些範例都是專案的一般項目：
您可以隨意編輯或刪除它們。
