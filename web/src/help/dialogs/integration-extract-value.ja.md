---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

抽出値は、受信したリクエストの一部 (プッシュイベントのブランチなど) を取り出し、統合が開始するタスクに渡します。統合のすべての値は、タスクを開始する各リクエストから取り出されます。

## Name {#integrationValueForm-name}

統合のページに表示される値の名前です。何が取り出され、どこに渡されるかには影響しません。このフィールドは必須です。

## Source of the Value {#integrationValueForm-valueSource}

値を取り出すリクエストの部分です。

- **Header**: リクエストヘッダーの値です。
- **Body**: リクエストのボディです。

## Data Type of Body {#integrationValueForm-bodyDataType}

**Source of the Value** が **Body** の場合に表示されます。

ボディの読み取り方です。

- **JSON**: ボディは JSON として解析され、**Key** はその中の 1 つの値へのパスです。値は JSON の型を保持します。数値、リスト、オブジェクトはそのままタスクに渡されます。
- **String**: ボディ全体がテキストとして取り出され、**Key** は使われません。

## Key {#integrationValueForm-key}

リクエストから取り出すものです。このフィールドは必須です。

- **Header** の場合: ヘッダーの名前です。例: `X-GitHub-Event`。大文字と小文字は区別されません。リクエストにないヘッダーは空の値になります。
- **JSON** の **Body** の場合: 値へのパスで、各部分をドットでつなぎます。配列の要素は `.[0]` と書きます。例: `repository.full_name` または `commits.[0].id`。ボディにないパスは値を返さず、タスクは統合で設定された値を保持します。
- **String** の **Body** の場合: 使われませんが、フォームには値が必要です。何か入力してください。

## Variable Usage {#integrationValueForm-variableType}

テンプレートのタスクで値が渡される先です。

- **Variables**: タスクの追加の変数です。調査変数への回答と同じ方法で渡されます。同じ名前の回答や変数グループの変数を置き換えます。
- **Task Params**: タスクのパラメーターです。タスクのダイアログのオプションと同じ方法で保存されます。名前 `git_branch` は、テンプレートがブランチの上書きを許可している場合にタスクのブランチを設定します。先頭の `refs/heads/` は削除されます。

ワークフローを実行する統合では、この用途は無視されます。すべての値は、**Variable** に指定された名前のワークフローの調査変数への回答になります。

## Variable {#integrationValueForm-variable}

値が渡されるときの名前です。例: プレイブックが `{{ branch }}` として読む追加の変数には `branch`。このフィールドは必須です。
