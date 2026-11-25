---
form: src/components/TerraformAliasForm.vue
---

# 編輯 Key

別名是 Semaphore 為此工作區提供的 HTTP 後端位址：指向它的 Terraform 或 OpenTofu 會在 Semaphore 中讀取、寫入、鎖定和解鎖該工作區的狀態。
Semaphore 會在您建立別名時產生此位址；在此對話方塊中，您只需選擇保護它的金鑰。別名是 Semaphore PRO 的功能。

## Auth key {#tfAlias-authKeyId}

保護此別名的金鑰。清單中只提供專案金鑰儲存中類型為**使用密碼登入**的金鑰；若清單是空的，請先在該處建立一把金鑰。

別名的用戶端會傳送此金鑰的登入名稱和密碼。儲存之後，請從別名清單中複製位址，並將它連同金鑰的登入名稱和密碼，填入您設定中的 `http` 後端：

```hcl
terraform {
  backend "http" {
    address        = "ALIAS_URL"
    lock_address   = "ALIAS_URL"
    unlock_address = "ALIAS_URL"
    username       = "LOGIN"
    password       = "PASSWORD"
  }
}
```

此位址會回應 `http` 後端的預設方法，因此一個位址即可服務狀態、鎖定和解鎖。為現有的別名選擇另一把金鑰時，其位址會保留不變：
此時用戶端需要使用新金鑰的登入名稱和密碼。
