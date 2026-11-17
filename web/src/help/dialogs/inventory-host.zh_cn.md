---
form: src/components/InventoryHostForm.vue
---

# 编辑主机

托管库存中的一台主机。使用该库存的任务会在主机列表中已启用的主机上运行。

## 名称 {#inventoryHostForm-name}

Ansible 用来识别主机的名称；如果变量中没有设置 `ansible_host`，它也是连接的地址。此字段为必填项。名称在库存内唯一，并且区分大小写：`web1` 和 `Web1` 是两台主机。名称不能包含空格、逗号、方括号或冒号，长度最多为 250 个字符。

## 描述 {#inventoryHostForm-description}

可选。给人看的备注，显示在主机列表中。Ansible 不会收到它。

## 变量 {#inventoryHostForm-variables}

可选。此主机的变量，以 YAML 或 JSON 映射的形式书写：

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

主机的变量优先于库存中同名的变量。不接受 YAML 的锚点、别名和合并键。
