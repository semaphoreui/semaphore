---
form: src/components/TerraformAliasForm.vue
---

# 編集 Key

エイリアスは、Semaphore がこのワークスペースのために提供する HTTP バックエンドのアドレスです。これを指定した Terraform または OpenTofu は、Semaphore 上のワークスペースの状態を読み取り、書き込み、ロック、ロック解除します。アドレスはエイリアスの作成時に Semaphore が作成します。このダイアログでは、それを保護するキーを選ぶだけです。エイリアスは Semaphore PRO の機能です。

## Auth key {#tfAlias-authKeyId}

エイリアスを保護するキーです。一覧には、プロジェクトのキーのストアにある **パスワードでログイン** タイプのキーだけが表示されます。一覧が空の場合は、先にそこでキーを作成してください。

エイリアスのクライアントは、このキーのログインとパスワードを送信します。保存した後、エイリアスの一覧からアドレスをコピーし、キーのログインとパスワードとともに、設定の `http` バックエンドに入力します。

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

アドレスは `http` バックエンドのデフォルトのメソッドに応答するため、1 つのアドレスで状態、ロック、ロック解除を扱えます。既存のエイリアスに別のキーを選んでもアドレスは保持されます。その場合、クライアントには新しいキーのログインとパスワードが必要です。
