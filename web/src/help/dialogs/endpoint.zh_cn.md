---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# 编辑端点

端点是 webhook 的入口：其他服务调用的 URL，以及校验这些请求的方式。集成关联到端点，并在端点接受的请求上启动自己的运行。端点的 URL（即别名）在“端点”选项卡中添加。

## 名称 {#endpointForm-name}

端点在“端点”选项卡和集成表单中显示的名称。此字段为必填项。

## 描述 {#endpointForm-description}

可选。关于端点用途的备注，显示在其名称旁边。

## 认证方式 {#integrationAuthForm-authMethod}

Semaphore 如何检查请求是否来自您所期望的服务。未通过检查的请求不会启动任何内容，webhook 对它的响应与对其他任何请求的响应相同。

- **无**：不检查。任何知道该 URL 的人都可以启动任务。
- **GitHub Webhooks**：`X-Hub-Signature-256` 标头必须包含使用秘密值生成的正文 HMAC-SHA256 签名，与 GitHub
  发送的一致。
- **Bitbucket Webhooks**：对 `X-Hub-Signature` 标头进行相同的检查，与 Bitbucket 发送的一致。
- **Token**：**认证头** 中指定的标头必须包含秘密值本身。
- **HMAC** 和 **HMAC (SHA-512)**：**认证头** 中指定的标头必须包含使用秘密值生成的正文 HMAC-SHA256 或
  HMAC-SHA512，采用小写十六进制，且不带 `sha256=` 之类的前缀。
- **BasicAuth**：请求必须在其 `Authorization` 标头中携带**认证密钥**中密钥的登录名和密码。

## 认证头 {#integrationAuthForm-authHeader}

当 **认证方式** 为 **Token**、**HMAC** 或 **HMAC (SHA-512)** 时显示。

携带令牌或签名的请求标头的名称，例如 `X-Webhook-Token`。名称不区分大小写。不带该标头的请求会被拒绝。

## 认证密钥 {#integrationAuthForm-authSecretId}

当 **认证方式** 不为 **无** 时显示。

用来检查请求的、来自密钥存储的密钥。对于 **BasicAuth**，只有类型为**使用密码登录**的密钥适用。对于其他方法，类型为**字符串**的密钥也适用：
检查使用其值，或者**使用密码登录**密钥的密码。

除 **无** 之外的每种方法都需要此字段：没有密钥或秘密值为空时，每个请求都会被拒绝。
