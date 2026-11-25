---
form: src/components/TerraformAliasForm.vue
---

# 编辑 Key

别名是 Semaphore 为此工作区提供的 HTTP 后端的地址：指向它的 Terraform 或 OpenTofu 会在 Semaphore
中读取、写入、锁定和解锁工作区的状态。Semaphore 在您创建别名时生成该地址；在此对话框中您只需选择保护它的密钥。
别名是 Semaphore PRO 的功能。

## Auth key {#tfAlias-authKeyId}

保护别名的密钥。列表中只提供项目密钥存储中类型为**使用密码登录**的密钥；如果列表为空，请先在那里创建一个。

别名的客户端会发送此密钥的登录名和密码。保存之后，从别名列表中复制地址，并将其连同密钥的登录名和密码一起填入配置的
`http` 后端：

```hcl
terraform {
  backend "http" {
    address        = "ALIAS_URL"
    lock_address   = "ALIAS_URL"
    unlock_address = "ALIAS_URL"
    username       = "LOGIN"
    password       = "PASSWORD"
  }
}
```

该地址响应 `http` 后端的默认方法，因此一个地址即可用于状态、锁定和解锁。为已有的别名选择另一个密钥会保留其地址：
此后客户端需要使用新密钥的登录名和密码。
