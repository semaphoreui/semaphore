---
form: src/components/RoleMappingForm.vue
---

# New mapping

對應會將專案中的某個角色授予每一位透過 LDAP 或 OpenID Connect 提供者登入、且屬於該提供者中某個群組的使用者。每次使用者透過該提供者登入時，
Semaphore 都會讀取其群組，並據此授予或撤銷所對應的角色，因此離開群組的使用者會在下次登入時失去該角色。透過對應取得的角色會讓使用者成為該專案的成員。
對應無法編輯：請刪除它並新增一個新的對應。

## Provider {#roleMapping-provider}

群組所來自的 LDAP 目錄或 OpenID Connect 提供者。清單中列出伺服器設定中已設定的提供者。此欄位為必填。

只有在提供者已為此進行設定時，才會讀取群組：LDAP 提供者需要群組搜尋基礎（`group_search_dn`），OpenID Connect
提供者需要指定包含群組的宣告（`groups_claim`）。若提供者之後從設定中移除，其對應仍會留在清單中，並標示為
**Provider not configured**。

## Group {#roleMapping-group}

提供者所給出的群組名稱。此欄位為必填。

- LDAP：群組項目的群組名稱屬性值，預設為 `cn`，例如 `devops`。完整的 DN，例如
  `cn=devops,ou=groups,dc=example,dc=com`，會被簡化為其 `cn`。
- OpenID Connect：群組宣告的其中一個值，與提供者傳送的內容相同。

## Project {#roleMapping-projectId}

授予角色所在的專案。此欄位為必填。管理員可以看到所有專案；其他使用者只能看到自己是成員的專案。

## 角色 {#roleMapping-role}

群組成員在專案中取得的角色。此欄位為必填，選取專案後即可使用。清單中列出該專案的角色、共用角色和內建角色。
當您選擇其他專案時，該專案中不存在的角色會被清除。

透過對應授予的 **Owner** 角色，在 Semaphore 防止您移除專案最後一位擁有者的角色時並不計入，因為它可能在下次登入時消失。
請至少保留一位手動新增的擁有者。
