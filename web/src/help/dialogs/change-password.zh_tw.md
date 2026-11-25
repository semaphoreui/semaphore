---
form: src/components/ChangePasswordForm.vue
---

# 更改密碼

設定以使用者名稱和密碼登入 Semaphore 時所用的新密碼。新密碼在按下**儲存**後生效。外部使用者透過 LDAP 或
OpenID Connect 登入，在此沒有可更改的密碼。

## Current password {#changePassword-currentPassword}

在更改您自己的密碼時顯示。

您目前的密碼。此欄位為必填：Semaphore 會在設定新密碼之前先檢查它，以免遭竊的工作階段或 API 權杖被用來改掉您的密碼。
若密碼不符，伺服器會拒絕此變更。

為其他使用者重設密碼的管理員或使用者管理者不會看到此欄位。

## 密碼 {#changePassword-password}

新密碼。此欄位為必填。點選眼睛圖示可顯示您輸入的內容。

Semaphore 只儲存密碼的雜湊值。
