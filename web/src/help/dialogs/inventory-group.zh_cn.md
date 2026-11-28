---
form: src/components/InventoryGroupForm.vue
---

# 编辑组

托管库存中的一组主机。Playbook 可以通过组名在部分主机上运行，组的变量只需设置一次即可作用于组内所有主机。在 Hosts 标签页、主机对话框中或通过对所选主机的操作，可以把主机放入组中。

从清单源导入的组由该源管理。其所有字段均为只读，每次同步时都会更新。

## 名称 {#inventoryGroupForm-name}

Ansible 用来识别组的名称，例如 Playbook 中的 `hosts: web`。此字段为必填项。名称在库存内唯一，并且区分大小写：`web` 和 `Web` 是两个组。名称不能包含空格、逗号、方括号或冒号，长度最多为 250 个字符，并且不能是 `all` 或 `ungrouped`。对于包含字母、数字和下划线以外字符的名称（例如 `web-eu`），Ansible 会发出警告，但名称仍然可用。也不能使用 `*`, `?`, `!`, `&`, `~` 和 `@`，因为 Ansible 会把它们当作模式。

## 描述 {#inventoryGroupForm-description}

可选。给人看的备注，显示在组列表中。Ansible 不会收到它。

## Child groups {#inventoryGroupForm-children}

可选。位于此组内的组。子组的主机也属于此组。组不能成为它自己的祖先。

## 变量 {#inventoryGroupForm-variables}

可选。此组的变量，以 YAML 或 JSON 映射的形式书写：

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

组内的每台主机都会得到这些变量。主机的变量优先于组中同名的变量，组的变量优先于库存的变量。不接受 YAML 的锚点、别名和合并键。
