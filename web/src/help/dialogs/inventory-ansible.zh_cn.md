---
form: src/components/InventoryForm.vue
---

# 编辑 Ansible Inventory

库存告诉 Ansible 要对哪些主机运行剧本，并保存 Semaphore 连接这些主机时使用的凭据。任务模板引用一个库存，
该模板的每个任务都针对这个库存运行。库存的格式请参阅
[Ansible 库存指南](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html)。

## 名称 {#inventoryForm-name}

库存在库存页面以及任务模板的库存选择器中显示的名称。请选择一个能区分这些主机的名称，例如
`Production web servers`。

## 运行器标签 {#inventoryForm-runnerTag}

在项目有可用的运行器时显示。

使用此库存的任务会在带有此标签的运行器上运行。任务模板上设置的运行器标签优先于库存的标签。将此字段留空，
则任何运行器都可以领取这些任务。

## 用户凭据 {#inventoryForm-sshKeyId}

Ansible 登录主机时使用的密钥，来自密钥存储。此字段为必填项，接受三种类型的密钥：

- **SSH 密钥**：Ansible 使用私钥连接；如果密钥带有登录名，也会使用该登录名。
- **使用密码登录**：Ansible 使用登录名和密码连接。
- **无**：Semaphore 不传递任何凭据。适用于不需要凭据的主机（例如
  `localhost`），或者库存本身已经定义了连接方式的情况。

## Sudo 凭据 {#inventoryForm-becomeKeyId}

可选。来自密钥存储、类型为**使用密码登录**的密钥，在 play 通过 `become` 提升权限时使用：其登录名是
Ansible 切换到的用户，其密码是 Ansible 回应密码提示时使用的密码。

如果 play 不提升权限，或者上面的用户提升权限时不需要密码，请将此字段留空。

## 类型 {#inventoryForm-type}

库存的来源：

- **Static**：以 INI 格式在此对话框中输入库存。
- **Static YAML**：以 YAML 格式在此对话框中输入库存。也可以用它来编写库存插件的配置。
- **File**：库存是 Git 仓库中的一个文件。
- **Managed**：主机及其变量以记录的形式保存。要编辑它们，请从列表中打开该库存。

类型下方的字段会随类型而变化。

## 库存文件路径 {#inventoryForm-inventoryPath}

当类型为 **File** 时显示。

库存文件相对于其仓库根目录的路径，例如 `inventory/production.yml`。该仓库是在“仓库”字段中选择的仓库；
如果该字段为空，则是使用此库存的任务模板的仓库。

## 仓库 {#inventoryForm-repositoryId}

当类型为 **File** 时显示。可选。

存放库存文件的仓库。Semaphore 会在每个任务开始之前更新它保存的仓库副本。如果该文件与剧本位于同一个仓库中，
请将此字段留空：此时会在任务模板的仓库中查找该路径。

## 库存 {#inventoryForm-inventoryCode}

当类型为 **Static** 或 **Static YAML** 时显示。

库存的文本。Semaphore 会保存它，并在每个任务中将其写入一个临时文件。INI 格式的静态库存：

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

以 **Static YAML** 表示的相同主机：

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
