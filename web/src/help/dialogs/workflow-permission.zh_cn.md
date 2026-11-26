---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

为一个角色授予对这一个工作流的额外权限。持有该角色的成员会在该角色在整个项目中拥有的权限之上获得这些权限：
此处的授权只会增加权限，绝不会收回权限。拥有多个角色的成员会获得授予其中任一角色的所有权限。

## 角色 {#templatePerm-roleSlug}

获得这些权限的角色。列表显示此项目的角色、共享角色和内置角色。此字段为必填项，并且一个角色在一个工作流上只能有一项授权：
服务器会拒绝第二项授权。

已保存的授权的角色无法更改。无论您在此处选择什么，服务器都会保留原来的角色，只保存权限。要将这些权限授予另一个角色，
请从列表中删除此授权并添加一项新的授权。

## Permissions {#templatePerm-permissionsTitle}

持有该角色的成员可以对此工作流执行的操作：

- **Can see this workflow**：工作流出现在工作流列表中，并且可以查看其运行。
- **Can run this workflow**：启动工作流的运行。
- **Can edit this workflow**：更改和删除工作流，以及管理这些权限。
- **Can approve this workflow**：批准或拒绝其运行中的审批步骤。

运行、编辑或批准都包含查看工作流的权限：勾选其中任一项时，**Can see this workflow** 也会被勾选且无法取消。
如果不勾选任何一项，保存的授权将不授予任何权限。
