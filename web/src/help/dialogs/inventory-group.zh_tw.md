---
form: src/components/InventoryGroupForm.vue
---

# 編輯群組

受管清單中的一組主機。Playbook 可以透過群組名稱在部分主機上執行，群組的變數只需設定一次即可作用於群組內所有主機。在 Hosts 分頁、主機對話框中或透過對所選主機的操作，可以把主機放入群組中。

## 名稱 {#inventoryGroupForm-name}

Ansible 用來識別群組的名稱，例如 Playbook 中的 `hosts: web`。此欄位為必填。名稱在清單內唯一，並且區分大小寫：`web` 和 `Web` 是兩個群組。名稱不能包含空格、逗號、方括號或冒號，長度最多為 250 個字元，並且不能是 `all` 或 `ungrouped`。對於包含字母、數字和底線以外字元的名稱（例如 `web-eu`），Ansible 會發出警告，但名稱仍然可用。

## 描述 {#inventoryGroupForm-description}

選填。給人看的備註，顯示在群組列表中。Ansible 不會收到它。

## Child groups {#inventoryGroupForm-children}

選填。位於此群組內的群組。子群組的主機也屬於此群組。群組不能成為它自己的祖先。

## 變數 {#inventoryGroupForm-variables}

選填。此群組的變數，以 YAML 或 JSON 對應的形式書寫：

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

群組內的每台主機都會得到這些變數。主機的變數優先於群組中同名的變數，群組的變數優先於清單的變數。不接受 YAML 的錨點、別名和合併鍵。
