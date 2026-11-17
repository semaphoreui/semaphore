---
form: src/components/EditRoleForm.vue
---

# Edit Role

角色是项目内一组具名的权限。您将角色授予项目的成员，成员可以执行其任一角色所允许的操作。从项目的 Roles
页面打开时，此对话框只编辑该项目的角色；从项目之外的 Roles 页面打开时，它编辑一个所有项目都可以使用的共享角色。
内置角色 Owner、Manager、Task Runner 和 Guest 无法更改。

## 名称 {#role-name}

角色在 Roles 页面以及任何选择角色的地方（例如向项目添加成员时）显示的名称。此字段为必填项。您可以随时重命名角色。

## Slug {#role-slug}

角色的标识符。在您输入名称时，Semaphore 会根据名称自动填写它，直到您亲自编辑 slug 为止。它由 1 到 100
个字母、数字、下划线和连字符组成，例如 `release-managers`。

角色保存之后就无法更改 slug：编辑角色时此字段处于锁定状态。

- slug 必须唯一，检查时不区分字母大小写。共享角色的 slug 必须与其他所有角色都不同；项目角色的 slug
  必须与共享角色以及其所在项目的其他角色都不同。
- `owner`、`manager`、`task_runner`、`guest` 和 `all` 是保留的，无论大小写。

## Permissions {#role-permissionsTitle}

持有该角色的成员可以在项目中执行的操作。新角色不勾选任何权限。成员所持有的所有角色的权限会累加；无论角色如何，
Semaphore 管理员都可以在每个项目中执行所有操作。

- **Can read project**：查看项目：其任务模板、任务及任务日志、库存、密钥、仓库、变量组、计划和成员。
  项目中的每一次读取都需要此权限，而且其他权限都不包含它，因此几乎每个角色都应勾选它。没有它，成员只能看到那些自身向该角色授予了访问权限的任务模板和工作流，
  以及它们的任务。
- **Can run project tasks**：从任务模板启动任务、停止任务，以及确认或拒绝等待确认的任务。
- **Can update project**：更改项目的设置并删除项目。
- **Can manage project resources**：创建、更改和删除任务模板、库存、密钥、仓库、变量组、计划、视图、集成、
  秘密存储、项目运行器以及项目的角色；删除任务；设置哪些角色可以使用某个任务模板。
- **Can manage project users**：向项目添加成员、更改其角色以及移除成员。
- **Can run workflows**：启动工作流并停止其运行。
- **Can edit workflows**：创建、更改和删除工作流，并设置哪些角色可以使用某个工作流。
- **Can approve workflows**：批准或拒绝工作流运行中等待审批的步骤。
