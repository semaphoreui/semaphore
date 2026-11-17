---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

マッチャーは、受信したリクエストに対する条件です。マッチャーは、プロジェクトのエイリアスへのリクエストがどの統合を開始するかを決めるため、1 つの URL で複数の統合を扱えます。統合は、そのすべてのマッチャーが成立した場合にのみ実行されます。マッチャーは、統合で **プロジェクトエイリアスを使用** がオンの場合にのみ使われます。オンの場合、統合にはプロジェクトのエイリアスを通じてのみアクセスできます。

プロジェクトのエイリアスで見つかった統合にマッチャーがない場合、その統合は実行されません。少なくとも 1 つ追加してください。

## Name {#integrationMatcherForm-name}

統合のページに表示されるマッチャーの名前です。マッチャーが確認する内容には影響しません。このフィールドは必須です。

## Match on {#integrationMatcherForm-matchType}

マッチャーが読み取るリクエストの部分です。

- **Header**: リクエストヘッダーの値です。
- **Body**: リクエストのボディです。

## Body Data Format {#integrationMatcherForm-bodyDataType}

**Match on** が **Body** の場合に表示されます。

ボディの読み取り方です。

- **JSON**: ボディは JSON として解析され、**Key** はその中の 1 つの値へのパスです。
- **String**: ボディ全体がテキストとして比較され、**Key** は使われません。

## Key {#integrationMatcherForm-key}

リクエストから読み取るものです。このフィールドは必須です。

- **Header** の場合: ヘッダーの名前です。例: `X-GitHub-Event`。大文字と小文字は区別されません。
- **JSON** の **Body** の場合: 値へのパスで、各部分をドットでつなぎます。配列の要素は `.[0]` と書きます。例: `repository.full_name` または `commits.[0].id`。
- **String** の **Body** の場合: 使われませんが、フォームには値が必要です。何か入力してください。

## Comparison Method {#integrationMatcherForm-method}

リクエストから読み取った値を **Value** と比較する方法です。

- **==**: 両者が等しい。
- **!=**: 両者が異なる。リクエストにないヘッダーやキーは、異なるものとみなされます。
- **Contains**: リクエストから読み取った値が **Value** を含む。

## Value {#integrationMatcherForm-value}

比較するテキストです。例: `push` または `refs/heads/main`。比較は完全一致で、大文字と小文字が区別されます。JSON ボディの数値や真偽値は、テキストの形で比較されます: `42`、`true`。このフィールドは必須です。
