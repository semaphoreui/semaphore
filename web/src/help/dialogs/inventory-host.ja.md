---
form: src/components/InventoryHostForm.vue
---

# ホストを編集

管理対象インベントリのホストです。このインベントリを使うタスクは、ホスト一覧で有効になっているホストで実行されます。

## 名前 {#inventoryHostForm-name}

Ansible がホストを識別する名前です。変数で `ansible_host` を設定しない場合は、接続先のアドレスにもなります。このフィールドは必須です。名前はインベントリ内で一意で、大文字と小文字は区別されます。`web1` と `Web1` は別のホストです。空白、カンマ、角括弧、コロンは使えず、長さは 250 文字までです。

## 説明 {#inventoryHostForm-description}

任意です。人のためのメモで、ホスト一覧に表示されます。Ansible には渡されません。

## 変数 {#inventoryHostForm-variables}

任意です。このホストの変数を YAML または JSON のマッピングとして記述します。

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

ホストの変数は、同じ名前のインベントリの変数より優先されます。YAML のアンカー、エイリアス、マージキーは使えません。
