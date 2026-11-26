---
form: src/components/KeyForm.vue
---

# 编辑 Key

密钥存储中的密钥保存项目在不显示其内容的情况下使用的凭据：Ansible 登录主机时使用的 SSH 密钥或密码、Git
仓库的凭据、Ansible vault 的密码、集成的秘密值。库存、仓库、任务模板和集成通过名称引用密钥。保存之后，秘密值永远不会再发送回浏览器。

## 密钥名称 {#key-name}

密钥在密钥存储以及库存、仓库、任务模板和集成的密钥选择器中显示的名称。此字段为必填项。请选择一个能说明密钥用途的名称，例如
`Deploy key of the web servers`。

## 来源 {#key-source}

密钥的秘密值保存在哪里。标签页：

- **Local**：秘密值在此对话框中输入，并保存在 Semaphore 的数据库中；如果管理员为访问密钥设置了加密密钥，则会加密保存。
  这是默认选项。
- **Storage**：秘密值保存在项目的外部秘密存储中，在下方的 **Storage** 中选择。此标签页仅在 Pro 版本中提供。
  存储本身在密钥存储的 **Storages** 标签页中设置。
- **Env**：从 Semaphore 服务器进程的环境变量中读取秘密值。请在标签页下方的字段中输入变量名称。
- **File**：从 Semaphore 服务器上的文件中读取秘密值。请在标签页下方的字段中输入文件的绝对路径。该文件必须位于服务器的秘密目录中
  （其配置中的 `dirs.secrets`，默认为 `/tmp/semaphore`）；位于该目录之外或包含 `..` 的路径，会在使用密钥时失败。
  文件末尾的换行符会被去除。

使用 **Env** 和 **File** 时，Semaphore 从不写入秘密值：每当任务使用该密钥时，它都会读取该值，因此请在秘密值所在之处更改它。
该值必须与密钥的**类型**相符。**字符串**密钥按原样使用该值。**SSH 密钥**或**使用密码登录**密钥需要包含该类型字段的
JSON，例如：

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

编辑密钥时，在勾选**覆盖**之前，标签页处于锁定状态。从秘密存储同步而来的密钥会保留其来源，其名称和类型无法更改。

## Storage {#key-sourceStorageId}

当选择 **Storage** 标签页时显示。

保存秘密值的项目秘密存储：HashiCorp Vault、OpenBao、Devolutions Server、AWS Secrets Manager 或 Azure Key
Vault，在密钥存储的 **Storages** 标签页中设置。

- 非只读的存储会接收您在此对话框中输入的秘密值：Semaphore 会在您保存时将其写入该存储。删除密钥也会从存储中删除该秘密值。
  保存在这种存储中的密钥以后无法移到其他存储。
- 只读存储只会被读取。此时对话框不显示秘密值的字段，每当任务使用该密钥时，Semaphore 都会从存储中读取它。

## Source Key {#key-sourceStorageKey}

当选择了 **Storage** 时显示。

秘密值在该存储中的位置，例如 HashiCorp Vault 中秘密值的路径。

## 类型 {#key-type}

密钥保存的凭据类型。此字段为必填项，其下方的字段会随类型而变化。

- **SSH 密钥**：私钥，可以带有登录名和密码短语。Ansible 使用它连接主机，Git 使用它克隆带有 `ssh` URL 的仓库。
- **使用密码登录**：登录名和密码。Ansible 使用它们连接主机，或者切换为另一个用户（库存的 **Sudo 凭据**）；Git
  使用它们克隆 `https` 仓库；其密码可以打开任务模板的 Ansible vault。
- **字符串**：不带登录名的单个秘密值：令牌、API 密钥或密码。可用作集成的秘密值。它不能用作 SSH 密钥、sudo 凭据、vault
  的密码或仓库的访问密钥。
- **无**：没有凭据。用于不需要登录的仓库（公共 `https` 仓库、`file` URL 或本地路径），以及其主机不需要凭据或自行定义连接方式的库存。

## 用户名 {#key-loginPasswordLogin}

当类型为**使用密码登录**时显示。可选。

凭据的用户名。

- Ansible 以此用户身份连接主机（`--user`）；作为库存的 **Sudo 凭据**时，它是 Ansible 切换到的用户（`--become-user`）。
  如果此字段为空，Ansible 使用库存或其自身配置中指定的用户。
- 对于 `https` 仓库，Git 使用此用户和密码登录。对于 Git 服务器单独接受的令牌，请将其留空：此时密码会作为用户名发送。

## 密码 {#key-loginPasswordPassword}

当类型为**使用密码登录**时显示。必填。

凭据的密码。Semaphore 用它回应 Ansible 的密码提示：对于用户凭据是 SSH 密码（`--ask-pass`），对于 sudo 凭据是
`become` 密码（`--ask-become-pass`），对于 Ansible vault 是 vault 密码。对于 `https` 仓库，它是 Git
服务器的密码或访问令牌。

## 值 {#key-stringValue}

当类型为**字符串**时显示。必填。

秘密值本身，例如集成用来核对其收到的请求的令牌。它按输入的原样保存。

## 用户名 {#key-sshLogin}

当类型为 **SSH 密钥**时显示。可选。

Ansible 连接主机时使用的用户（`--user`）。如果此字段为空，Ansible 使用库存或其自身配置中指定的用户。

对于 Git 仓库，请改为将用户写入仓库 URL，例如 `git@github.com:org/repo.git`。

## Passphrase {#key-sshPassphrase}

当类型为 **SSH 密钥**时显示。可选。

加密私钥时使用的密码短语。对于没有密码短语的密钥，请留空。错误的密码短语会使每个使用该密钥的任务在加载密钥时失败。
由 Semaphore 生成的密钥没有密码短语：此时此字段中的内容会被丢弃。

## Generate SSH Key {#key-generateSshKey}

当类型为 **SSH 密钥**时显示。

保存时，Semaphore 会创建一个新的密钥对：一个 2048 位的 RSA 私钥，保留在密钥中；以及其公钥，保存后会在对话框中显示一次，
之后每次编辑密钥时都会显示在 **Public key** 下。请将公钥添加到主机上的 `~/.ssh/authorized_keys`，或者作为 Git
仓库的部署密钥。

勾选此框时，私钥字段处于锁定状态。对于已有的密钥，请先勾选**覆盖**：旧的密钥对会被替换。秘密值从 **Env**、**File**
或只读存储读取的密钥无法生成，因为 Semaphore 无处写入它。

## 私钥 {#key-sshPrivateKey}

当类型为 **SSH 密钥**时显示。除非勾选了 **Generate SSH Key**，否则为必填。

OpenSSH 或 PEM 格式的私钥，包括其 `BEGIN` 和 `END` 行。对于任务，Semaphore 会将其加载到自己的 SSH 代理中，Ansible
和 Git 在任务期间使用该代理。

## Public key {#key-publicKey}

当您编辑由 Semaphore 生成的 SSH 密钥、且未勾选**覆盖**时显示。

所生成密钥对的公钥部分。**Show** 会展开它，复制按钮会将其放到剪贴板上。请将其添加到主机上的 `~/.ssh/authorized_keys`，
或者作为 Git 仓库的部署密钥。

它是 Semaphore 最近一次生成的密钥对的公钥。在您用**覆盖**手动替换私钥之后，此处显示的密钥就不再与之匹配。

## 覆盖 {#key-overrideSecret}

在编辑已有密钥时显示。

Semaphore 永远不会再次显示已保存的秘密值，因此秘密值的字段为空且处于锁定状态。勾选此框即可替换秘密值：来源、类型和秘密值的字段会被打开，
其中的所有内容都会重新保存，因此请重新输入完整的秘密值，而不仅是更改的部分。不勾选时，保存只会更改密钥的名称。

保存密钥还会删除 Semaphore 服务器为使用该密钥的仓库保存的副本：下一个任务会重新克隆它们。
