---
form: src/components/RoleMappingForm.vue
---

# New mapping

映射为每个通过 LDAP 或 OpenID Connect 提供方登录、并且在该提供方中属于某个组的用户授予项目中的一个角色。
Semaphore 会在用户每次通过该提供方登录时读取其所属的组，并相应地授予或收回映射的角色，因此离开该组的用户会在下次登录时失去该角色。
映射的角色会使用户成为项目的成员。映射无法编辑：请删除它并添加一个新的映射。

## Provider {#roleMapping-provider}

组所在的 LDAP 目录或 OpenID Connect 提供方。列表包含服务器配置中设置的提供方。此字段为必填项。

只有当提供方为此进行了设置时才会读取组：LDAP 提供方需要组搜索基准（`group_search_dn`），OpenID Connect
提供方需要指定包含组的声明（`groups_claim`）。如果某个提供方之后从配置中移除，其映射仍会保留在列表中，并标记为
**Provider not configured**。

## Group {#roleMapping-group}

提供方给出的组名称。此字段为必填项。

- LDAP：组条目的组名属性的值，默认为 `cn`，例如 `devops`。像
  `cn=devops,ou=groups,dc=example,dc=com` 这样的完整 DN 会被简化为其 `cn`。
- OpenID Connect：组声明的值之一，与提供方发送的一致。

## Project {#roleMapping-projectId}

授予角色所在的项目。此字段为必填项。管理员可以看到所有项目；其他用户只能看到自己是其成员的项目。

## 角色 {#roleMapping-role}

组成员在项目中获得的角色。此字段为必填项，在选择项目之后才可用。列表包含该项目的角色、共享角色和内置角色。
当您选择另一个项目时，在该项目中不存在的角色会被清除。

当 Semaphore 阻止您收回项目最后一位所有者的角色时，由映射授予的 **Owner** 角色不计算在内，因为它可能在下次登录时消失。
请至少保留一位手动添加的所有者。
