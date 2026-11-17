---
form: src/components/ChangePasswordForm.vue
---

# 更改密码

为使用用户名和密码登录 Semaphore 设置新密码。新密码在点击**保存**后生效。通过 LDAP 或 OpenID Connect
登录的外部用户在此没有可更改的密码。

## Current password {#changePassword-currentPassword}

在更改自己的密码时显示。

您当前的密码。此字段为必填项：Semaphore 会在设置新密码之前检查它，这样被盗的会话或 API 令牌就无法被用来更改您的密码。
如果密码不匹配，服务器会拒绝此更改。

管理员或用户管理员在重置其他用户的密码时看不到此字段。

## 密码 {#changePassword-password}

新密码。此字段为必填项。点击眼睛图标可以显示您输入的内容。

Semaphore 只保存密码的哈希值。
