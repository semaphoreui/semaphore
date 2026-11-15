---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

授予某個角色針對此工作流程的額外權限。擁有該角色的成員，會在該角色於整個專案中所擁有的權限之外，再取得這些權限：此處的授權只會增加權限，
永遠不會移除權限。擁有多個角色的成員，會取得其中任一角色所獲授予的所有權限。

## 角色 {#templatePerm-roleSlug}

取得這些權限的角色。清單中顯示此專案的角色、共用角色和內建角色。此欄位為必填，而且一個角色在一個工作流程上只能有一項授權：伺服器會拒絕第二項授權。

已儲存之授權的角色無法變更。無論您在此選擇什麼，伺服器都會保留原角色，只儲存權限。若要將權限授予其他角色，請從清單中移除此授權，並新增一項新的授權。

## 權限 {#templatePerm-permissionsTitle}

擁有該角色的成員可以對此工作流程執行的操作：

- **Can see this workflow**：工作流程會出現在工作流程清單中，連同其執行記錄。
- **Can run this workflow**：啟動工作流程的執行。
- **Can edit this workflow**：修改和刪除工作流程，以及管理這些權限。
- **Can approve this workflow**：核准或拒絕其執行中的核准步驟。

執行、編輯或核准都包含檢視工作流程：當您勾選其中任一項時，**Can see this workflow** 也會被勾選，且無法取消。若所有方塊都不勾選，儲存的授權將不授予任何權限。
