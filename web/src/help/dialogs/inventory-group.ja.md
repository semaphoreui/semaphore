---
form: src/components/InventoryGroupForm.vue
---

# グループを編集

管理対象インベントリのホストのグループです。Playbook はグループ名でホストの一部に対して実行でき、グループの変数はそのすべてのホストに一度で設定できます。ホストは Hosts タブ、ホストのダイアログ、または選択したホストへの操作でグループに入れます。

## 名前 {#inventoryGroupForm-name}

Ansible がグループを識別する名前で、例えば Playbook の `hosts: web` で使います。このフィールドは必須です。名前はインベントリ内で一意で、大文字と小文字は区別されます。`web` と `Web` は別のグループです。空白、カンマ、角括弧、コロンは使えず、長さは 250 文字までで、`all` と `ungrouped` は使えません。文字、数字、アンダースコア以外の文字を含む名前 (`web-eu` など) には Ansible が警告を出しますが、名前はそのまま使えます。

## 説明 {#inventoryGroupForm-description}

任意です。人のためのメモで、グループ一覧に表示されます。Ansible には渡されません。

## Child groups {#inventoryGroupForm-children}

任意です。このグループの中にあるグループです。子グループのホストはこのグループにも属します。グループを自分自身の祖先にすることはできません。

## 変数 {#inventoryGroupForm-variables}

任意です。このグループの変数を YAML または JSON のマッピングとして記述します。

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

グループのすべてのホストがこれを受け取ります。ホストの変数は同じ名前のグループの変数より優先され、グループの変数はインベントリの変数より優先されます。YAML のアンカー、エイリアス、マージキーは使えません。
