---
form: src/components/HostConfigForm.vue
---

# Edit mapping

映射告诉 Semaphore 对某个 SSH 主机或 Git URL 使用密钥存储中的哪个凭据。仓库只带有一个密钥，但任务还可能访问其他服务器：
子模块、`requirements.yml` 中的角色和集合、Terraform 模块、库存仓库。项目的映射适用于其任务的每一个 Git 和
SSH 连接，以及 Semaphore 为项目运行的 Git 命令，例如检查仓库是否有新提交。您仓库中的任何文件都不会改变。

## 类型 {#hostConfigForm-type}

映射匹配的内容：

- **Host**：SSH 主机名。任务到该主机的每一个 SSH 连接都使用映射的密钥：通过 SSH 的 Git，以及 Ansible
  连接到同名的库存主机。对于同一主机，映射优先于服务器的 SSH 配置。
- **URL**：Git URL 或其开头部分。Git 会使用映射的凭据连接到该 URL 下的每一个仓库。

切换到 **Host** 时，会清除主机映射无法使用的 **Credential**。

## Host 或 URL {#hostConfigForm-host}

此字段为必填项。对于每个主机或 URL，每种类型只能存在一个映射。

对于 **Host**：与 Git 或 Ansible 连接时完全一致的主机名，例如 `github.com` 或 `10.0.0.21`。
只能包含字母、数字、点号、连字符和下划线：不能包含用户、端口或通配符。

对于 **URL**：以 `https://` 或 `http://` 开头的地址，其中不包含登录名，也不包含空格、引号或 `=`。
可以指定一个完整的仓库，或者以 `/` 结尾的前缀以涵盖一组仓库：

```
https://github.com/acme/
https://github.com/acme/infra.git
```

当有多个映射匹配时，最长的那个优先：上面的仓库映射会针对该仓库覆盖 `acme` 的映射。

使用 SSH 密钥时，URL 映射会让 Git 通过 SSH 连接到 URL 中主机的 `22` 端口，无论 URL 中指定的是什么端口。

## Credential {#hostConfigForm-sshKeyId}

来自密钥存储、用于映射所匹配的连接的密钥。此字段为必填项。

- **Host** 只接受 **SSH 密钥**。如果密钥带有登录名，SSH 会以该用户身份连接。
- **URL** 接受 **SSH 密钥**或**使用密码登录**。SSH 密钥以其登录名连接，如果没有登录名则以 `git` 连接。
  使用密码登录只适用于 `https://` URL；对于令牌，请将令牌填入密码并将登录名留空。

使用密码登录时，凭据通过 Git 的环境而不是命令行传递给 Git，任务日志中只显示原始 URL。
