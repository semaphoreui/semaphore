---
form: src/components/InventorySourceForm.vue
---

# 编辑源

源从 Git 仓库中的文件导入托管清单的主机和组。每次点击 Sync 都会读取该文件，文件中描述的主机和组会与它保持一致。服务器需要 Ansible 并能访问该仓库。

## 名称 {#inventorySourceForm-name}

源的名称，显示在它的卡片上以及它带来的主机旁。此字段必填。

## 仓库 {#inventorySourceForm-repository}

读取文件所用的项目仓库。此字段必填。使用它的分支。

## 路径 {#inventorySourceForm-path}

清单文件在仓库中的路径，例如 `inventory/prod.yml`，也可以是存放清单文件的目录。此字段必填。

文件按静态 YAML 或 INI 读取，脚本和插件配置文件会被拒绝。在目录中，Ansible 不跳过的每个文件都必须是有效的清单，否则同步失败，什么也不会改变。Ansible 在目录中会跳过一些扩展名，例如 `.cfg`、`.md` 和 `.txt`。如果仓库中有指向仓库外部的链接，则会被拒绝。
