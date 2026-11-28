---
form: src/views/Tokens.vue
---

# New Token

API 令牌让脚本或其他服务能够以您的身份调用 Semaphore API，并拥有与您的账户相同的权限。请在请求的
`Authorization` 标头中发送它：`Bearer` 加上令牌。令牌只在创建后立即显示一次：请在那时复制它。

## Token Name {#token-name}

可选。用于在令牌列表中区分此令牌的名称，例如使用它的脚本的名称。名称不是令牌的一部分，也不会改变令牌的权限。

## Expires {#token-expires}

令牌何时停止工作。可以选择从现在开始计算的一段时间，选择 **Never** 表示令牌一直有效直到被删除，或者选择
**Custom** 自行选择时间点。

永不过期的令牌会一直有效，直到您从列表中删除它。对于交给脚本或其他服务使用的令牌，建议选择一段有效期。

## Expiration date {#token-expiresAt}

当 **Expires** 为 **Custom** 时显示。

令牌停止工作的日期和时间，采用您浏览器的时区。该时间点必须在将来：服务器会拒绝创建一个已经过期的令牌。
