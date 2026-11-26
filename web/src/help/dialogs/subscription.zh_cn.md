---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

订阅密钥会在此服务器上解锁 Semaphore PRO 和 Enterprise 的功能。此对话框用于激活密钥，并显示当前有效密钥的方案：
其持有者、状态、到期时间，以及它允许的和已在使用的 Pro 用户、运行器和其他项目的数量。管理员可以从 Pro 版本的用户菜单中打开它。
在订阅生效之前，此对话框的名称为 **Upgrade to Semaphore PRO**。

## Enter your PRO or EE key {#subscription-key}

订阅密钥，即您随 PRO 或 Enterprise 订阅获得的密钥。将它粘贴到这里，或者用上传按钮从文件中加载，然后点击
**Activate New Key**。激活时此字段为必填项。

当密钥在服务器的配置中设置时（`subscription.key` 或 `subscription.key_file`，或者变量
`SEMAPHORE_SUBSCRIPTION_KEY` 和 `SEMAPHORE_SUBSCRIPTION_KEY_FILE`），此字段为只读：此时只能在配置中更改密钥。

## 上传密钥文件 {#subscription-upload}

在尚未激活任何订阅时显示。

从您计算机上的文件（`.txt`、`.key`、`.pem` 或 `.lic`）中读取密钥，并填入上面的字段。在您点击
**Activate New Key** 之前，密钥不会被激活。

## 密钥操作 {#subscription-actionsMenu}

在订阅有效或已过期时显示。

针对服务器密钥的操作菜单：

- **Reload**：重新获取订阅的状态，例如在续订之后。
- **Upload**：从文件中读取新密钥并填入上面的字段，与上传按钮的作用相同。
- **Reset**：立即从服务器中删除密钥，不会请求确认。

## Activate New Key {#subscription-activate}

在此服务器上激活字段中的密钥。如果密钥有效，订阅的方案会显示在下方。如果您的账户还不是 Pro 用户，并且方案中有空闲的
Pro 用户席位，对话框会询问是否将您的账户设为 Pro 用户。

已经激活的密钥会以警告的形式报告。当密钥在服务器的配置中设置时，此按钮处于禁用状态。

## Buy Pro {#subscription-buy}

在浏览器的新标签页中打开 Semaphore 门户中购买 PRO 订阅的页面。

## Get 30-day free trial {#subscription-trial}

在没有有效订阅时显示。

在浏览器的新标签页中打开 Semaphore 门户中发放 30 天免费试用密钥的页面。将您获得的密钥粘贴到上面的字段中。
