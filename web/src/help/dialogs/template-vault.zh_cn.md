---
form: src/components/TemplateVaults.vue
---

# 保管库

保管库为 Ansible 提供剧本仓库中使用 Ansible Vault 加密的文件和变量的密码。模板的每个保管库都会成为
`ansible-playbook` 的一个 `--vault-id` 选项。保管库随模板一起保存，即在您保存模板对话框时保存。

## 保管库 ID {#templateVaultDialog-name}

可选。密码所对应的 vault ID，即加密内容时使用的标签（`ansible-vault encrypt --vault-id prod@prompt`）。
留空则表示 vault ID `default`。

ID 以小写形式保存。模板的每个保管库都需要有自己的 ID，并且只能有一个为空。

## 类型 {#templateVaultDialog-type}

密码的来源：

- **密码**：来自密钥存储的密钥。Semaphore 会在 Ansible 请求密码时输入它，因此密码永远不会写入文件。
- **客户端脚本**：输出密码的 vault 密码客户端脚本；由 Ansible 运行它。

## 保管库密码 {#templateVaultDialog-vaultKeyId}

当类型为**密码**时显示。

保存保管库密码的密钥。只列出类型为**使用密码登录**的密钥；Semaphore 使用其密码并忽略登录名。此字段为必填项。

## 脚本路径 {#templateVaultDialog-script}

当类型为**客户端脚本**时显示。

脚本的路径，例如 `scripts/vault-pass-client.py`。相对路径从 Ansible 运行所在的目录算起：即仓库的根目录，
或者模板设置了工作目录时的该工作目录。此字段为必填项。

文件名必须以 `-client` 结尾，可以带扩展名也可以不带：只有这样，Ansible 才会将该文件作为客户端脚本运行，并通过
`--vault-id` 向它传递 vault ID。脚本在运行任务的机器上，以运行 Semaphore 的用户身份运行。
