---
form: src/components/InventoryForm.vue
---

# 編輯 Ansible Inventory

清單告訴 Ansible 要對哪些主機執行 Playbook，並保存 Semaphore 連線到這些主機時使用的憑證。任務範本會參照一份清單，
該範本的每個任務都會針對這份清單執行。清單的格式請參閱
[Ansible 清單指南](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html)。

## 名稱 {#inventoryForm-name}

清單在清單頁面以及任務範本的清單選擇器中顯示的名稱。請選擇一個能區分這些主機的名稱，例如
`Production web servers`。

## 執行器標籤 {#inventoryForm-runnerTag}

在專案有可用的執行器時顯示。

使用此清單的任務會在帶有此標籤的執行器上執行。任務範本上設定的執行器標籤優先於清單的標籤。將此欄位留空，
則任何執行器都可以接手這些任務。

## 使用者憑證 {#inventoryForm-sshKeyId}

Ansible 登入主機時使用的金鑰，來自金鑰儲存。此欄位為必填，接受三種類型的金鑰：

- **SSH 金鑰**：Ansible 使用私鑰連線；若金鑰設有登入名稱，也會使用該登入名稱。
- **使用密碼登入**：Ansible 使用登入名稱和密碼連線。
- **無**：Semaphore 不傳遞任何憑證。適用於不需要憑證的主機（例如
  `localhost`），或清單本身已定義連線方式的情況。

## Sudo 憑證 {#inventoryForm-becomeKeyId}

選填。來自金鑰儲存、類型為**使用密碼登入**的金鑰，在 play 透過 `become` 提升權限時使用：其登入名稱是
Ansible 切換成的使用者，其密碼是 Ansible 回應密碼提示時所用的密碼。

若 play 不會提升權限，或上述使用者提升權限時不需要密碼，請將此欄位留空。

## 類型 {#inventoryForm-type}

清單的來源：

- **Static**：以 INI 格式在此對話方塊中輸入清單。
- **Static YAML**：以 YAML 格式在此對話方塊中輸入清單。也可以用來撰寫清單外掛程式的設定。
- **File**：清單是 Git 儲存庫中的一個檔案。
- **Managed**：主機及其變數以記錄的形式保存。若要編輯，請從列表中開啟該清單。

類型下方的欄位會隨類型而改變。

## 清單檔案路徑 {#inventoryForm-inventoryPath}

當類型為 **File** 時顯示。

清單檔案相對於其儲存庫根目錄的路徑，例如 `inventory/production.yml`。該儲存庫是在「儲存庫」欄位中選取的儲存庫；
若該欄位為空，則是使用此清單的任務範本的儲存庫。

## 儲存庫 {#inventoryForm-repositoryId}

當類型為 **File** 時顯示。選填。

存放清單檔案的儲存庫。Semaphore 會在每個任務開始前更新它保存的儲存庫副本。若該檔案與 Playbook 位於同一個儲存庫，
請將此欄位留空：此時會在任務範本的儲存庫中尋找該路徑。

## 清單 {#inventoryForm-inventoryCode}

當類型為 **Static** 或 **Static YAML** 時顯示。

清單的文字內容。Semaphore 會儲存它，並在每個任務中將其寫入一個暫存檔案。INI 格式的靜態清單：

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

以 **Static YAML** 表示的相同主機：

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
