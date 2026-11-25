---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

將專案中已存在的工作區附加到此範本，使其連同狀態和別名出現在範本的工作區之中。一個工作區同一時間只屬於一個範本。附加並不會讓該工作區成為預設工作區：
在您按下 **Make default** 之前，任務會繼續在預設工作區中執行。

## Workspace {#workspaceSelect-inventoryId}

要附加的工作區。此欄位為必填。清單中顯示專案中為此範本之應用程式（Terraform、OpenTofu 或 Terragrunt）建立、且未附加到任何範本的工作區；
透過 **Detach** 從其範本分離的工作區會再次出現在此處。

工作區會保留其名稱、金鑰和狀態。若工作區在此期間已被其他範本取用，伺服器會拒絕它。
