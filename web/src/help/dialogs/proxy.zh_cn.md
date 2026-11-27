---
form: src/components/ProxyForm.vue
---

# 编辑 Proxy

代理是一台主机，项目的任务经过它访问无法直接连接的服务器。在存储库的 **Proxy** 字段中选择它，
即可通过它获取存储库；在清单中选择它，Ansible 即通过它连接主机。

## 名称 {#proxyForm-name}

必填，且在项目中唯一。存储库和清单以此名称显示代理。

## 类型 {#proxyForm-type}

- **SSH (jump host)**：Semaphore 通过 SSH 连接到该主机，再从那里打开到服务器的连接。只有此类型
  可以依赖另一个代理。
- **SOCKS5**、**HTTP**、**HTTPS**：转发连接的代理服务器。对于 `https` 存储库，Git 自己使用它；
  SSH 连接由 Semaphore 经由它转发，因此无需安装其他工具。

## Host {#proxyForm-host}

必填。主机名或 IP 地址，不含用户、端口和协议。主机名只能包含字母、数字、点、连字符和下划线。

## Port {#proxyForm-port}

可选，1 到 65535。留空时，SSH 跳板机为 `22`。其他类型请填写端口：不填时，Git 和 SSH 连接会假定
不同的端口。

## User {#proxyForm-user}

仅限 SSH：跳板机上的登录名。留空表示运行 Semaphore 的用户。可用字母、数字、点、连字符和下划线，
不能以连字符开头。SOCKS5 或 HTTP 代理改为从**访问密钥**获取登录名。

## 访问密钥 {#proxyForm-sshKeyId}

可选。密钥存储中用于登录代理本身的密钥，通常不是其后服务器的密钥：

- SSH 跳板机使用 **SSH 密钥**；
- SOCKS5、HTTP 和 HTTPS 使用**使用密码登录**。密码通过环境变量传递，绝不出现在命令行中。

## Requires Proxy {#proxyForm-requiresProxyId}

仅限 SSH，可选。需要先经过的项目中另一个 SSH 代理，用于只能从另一台跳板机访问的跳板机。一条链最多
包含 5 个代理，且不能回到链中已有的代理。
