---
form: src/components/RepositoryForm.vue
---

# 编辑 Repository

仓库是任务模板的剧本、Terraform 代码或脚本的来源：一个 Git 仓库，或者运行任务的机器上的一个目录。在每个任务开始之前，
Semaphore 会克隆 Git 仓库，或者更新之前克隆的副本，检出分支，然后在该副本中运行任务。

## 名称 {#repository-name}

仓库在仓库页面以及任务模板和库存的仓库选择器中显示的名称。此字段为必填项。

## URL 或路径 {#repository-gitUrl}

代码所在的位置。此字段为必填项。字段下方的标签显示您输入的地址类型，点击某个标签会将地址改写为该类型。

- **ssh**：`git@github.com:org/repo.git` 或 `ssh://git@github.com/org/repo.git`。Git 使用地址中的用户以及在**访问密钥**中选择的
  **SSH 密钥**登录。
- **https** 和 **http**：`https://github.com/org/repo.git`。私有仓库需要**使用密码登录**类型的密钥；公共仓库需要**无**。
  请避免使用 `http`：此时登录名和密码会以未加密的方式传输。
- **git**：`git://host/org/repo.git`，只读的 Git 协议，没有登录。
- **file**：`file:///srv/git/repo.git`，位于运行任务的机器磁盘上的 Git 仓库，像远程仓库一样被克隆。
- **绝对路径**：例如 `/srv/playbooks` 或 `D:\playbooks` 这样的绝对路径。Semaphore 不会克隆它：任务直接在该目录中运行，
  该目录必须存在于运行任务的机器（Semaphore 服务器或运行器）上。分支对它不适用。

更改地址会删除 Semaphore 服务器从旧地址克隆的副本。

## 分支 {#repository-gitBranch}

Semaphore 克隆并保持最新的分支，例如 `main`。此字段为必填项，但绝对路径除外，此时此字段处于锁定状态。

任务模板可以指定自己的分支，该分支优先于此处的分支；模板也可以允许其任务在启动时选择分支。也可以使用标签，但标签无法被拉取，
因此 Semaphore 会在每个任务开始之前重新克隆仓库。

## Single checkout for all templates {#repository-monorepo}

默认关闭：每个任务模板都会获得自己的仓库副本。勾选此框后，仓库只克隆一次，并在所有使用它的模板之间共享该副本，
适用于太大而无法为每个模板保留一份副本的仓库。

共享副本始终停留在此仓库的分支上：使用它的模板不能有自己的分支，其任务也不能选择其他分支或提交。对于绝对路径，此选项处于锁定状态。

## 访问密钥 {#repository-sshKeyId}

Git 登录仓库时使用的、来自密钥存储的密钥。此字段为必填项。它提供三种类型的密钥，应选择哪一种取决于 URL：

- **SSH 密钥**用于 `ssh` URL。Semaphore 会在克隆或更新期间将私钥加载到 SSH 代理中。用户是 URL 中的用户，例如
  `git@github.com:org/repo.git` 中的 `git`。
- **使用密码登录**用于 `https` URL：登录名和密码，或者 Git 服务器的访问令牌。如果登录名为空，则只发送密码，并将其作为用户名，
  这正是某些服务器对令牌的要求。同样的凭据也会提供给 `ansible-galaxy`，用于同一服务器上的角色和集合。
- **无**用于公共仓库、`git` 或 `file` URL，或者绝对路径。

与 URL 不匹配的密钥不会向 Git 提供任何凭据，此时私有仓库将无法克隆。
