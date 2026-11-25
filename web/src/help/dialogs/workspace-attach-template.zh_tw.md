---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

將此工作區附加到某個範本，使其連同狀態和別名出現在該範本的工作區之中。一個工作區同一時間只屬於一個範本：附加之後，它會離開「清單」頁面，
改在範本中管理。附加並不會讓它成為範本的預設工作區：若要如此，請開啟範本並按下 **Make default**。

## Template {#inventoryAttach-templateId}

要附加工作區的範本。此欄位為必填。清單中顯示專案中為與該工作區相同之應用程式（Terraform 或 OpenTofu）建立的範本。

工作區會保留其名稱、金鑰和狀態。若之後要將它移到其他範本，請在其所附加的範本中按下 **Detach**：它會回到「清單」頁面。範本的預設工作區無法分離。
