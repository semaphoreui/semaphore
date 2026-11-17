---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

为一个角色授予对这一个模板的额外权限。持有该角色的成员会在该角色在整个项目中拥有的权限之上获得这些权限：
此处的授权只会增加权限，绝不会收回权限。拥有多个角色的成员会获得授予其中任一角色的所有权限。

## 角色 {#templatePerm-roleSlug}

获得这些权限的角色。列表显示此项目的角色、共享角色和内置角色。此字段为必填项，并且一个角色在一个模板上只能有一项授权：
服务器会拒绝第二项授权。

已保存的授权的角色无法更改。无论您在此处选择什么，服务器都会保留原来的角色，只保存权限。要将这些权限授予另一个角色，
请从列表中删除此授权并添加一项新的授权。

## Permissions {#templatePerm-permissionsTitle}

持有该角色的成员可以对此模板执行的操作：

- **Can see this template**：模板出现在模板列表中，并且可以查看其任务及任务日志。
- **Can run this template**：从此模板启动任务，并停止、确认或拒绝其任务。
- **Can edit this template**：更改和删除模板、停止其所有任务，以及管理这些权限。

运行或编辑都包含查看模板的权限：勾选其中任一项时，**Can see this template** 也会被勾选且无法取消。
如果不勾选任何一项，保存的授权将不授予任何权限。
