---
form: src/components/InventoryHostForm.vue
---

# 編輯主機

受管清單中的一台主機。使用該清單的任務會在主機列表中已啟用的主機上執行。

## 名稱 {#inventoryHostForm-name}

Ansible 用來識別主機的名稱；如果變數中沒有設定 `ansible_host`，它也是連線的位址。此欄位為必填。名稱在清單內唯一，並且區分大小寫：`web1` 和 `Web1` 是兩台主機。名稱不能包含空格、逗號、方括號或冒號，長度最多為 250 個字元。

## 描述 {#inventoryHostForm-description}

選填。給人看的備註，顯示在主機列表中。Ansible 不會收到它。

## 變數 {#inventoryHostForm-variables}

選填。此主機的變數，以 YAML 或 JSON 對應的形式書寫：

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

主機的變數優先於清單中同名的變數。不接受 YAML 的錨點、別名和合併鍵。
