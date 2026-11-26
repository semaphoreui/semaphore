---
form: src/components/TerraformInventoryForm.vue
---

# 新しい Terraform Workspace

ワークスペースは、Terraform、OpenTofu、Terragrunt のテンプレートのインベントリです。タスクが `init` の後に選択するワークスペースと、`init` が非公開のモジュールのダウンロードに使うキーを指定します。ここで作成したワークスペースはこのテンプレートに接続され、その状態の上に一覧表示されます。テンプレートのタスクは、チェックマークの付いたデフォルトのワークスペースで実行されます。**Make default** でそれを変更します。鉛筆のボタンで開くと、同じダイアログでワークスペースを編集できます。

## Workspace name {#tfWorkspace-inventory}

Terraform のワークスペースの名前です。例: `production`。このフィールドは必須です。

プランの前に、Semaphore はこの名前で `workspace select -or-create=true` を実行するため、ワークスペースがまだ存在しない場合は作成されます。バックエンドがワークスペースをサポートしていないなどの理由で `workspace list` が失敗した場合、Semaphore は選択をスキップします。

Semaphore はワークスペースを `TEMPLATE - NAME`、つまりテンプレートの名前とこの名前で保存します。この完全な名前は、インベントリ ページと、ワークスペースをテンプレートに接続するときに表示されます。ワークスペースの名前を変更しても、Terraform 内の名前は変わりません。次のタスクは新しい名前のワークスペースを選択するか、作成します。

## SSH key for private modules {#tfWorkspace-sshKeyId}

`init` が非公開の Git リポジトリから SSH でモジュールをダウンロードするときに使うキーです。このダイアログではこのフィールドは必須です。一覧には、キーのストアにある **SSHキー**、**パスワードでログイン**、**なし** タイプのキーが表示されます。

効果があるのは **SSHキー** だけです。Semaphore は `init` の間それを SSH エージェントに読み込み、Git は `git::ssh://git@example.com/modules/vpc.git` のようなモジュールのソースにそれを使います。**パスワードでログイン** タイプのキーは、Git に資格情報を渡しません。モジュールにキーが不要な場合は **なし** を選びます。
