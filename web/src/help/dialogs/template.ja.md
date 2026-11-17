---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# テンプレートを編集 'Ansible Playbook'

タスクテンプレートは、Semaphore が何をどのように実行するかを示します。リポジトリとその中のファイル、インベントリと変数、そしてタスクの開始時にユーザーが変更できるオプションです。すべてのタスクはテンプレートから開始されます。ダイアログのフィールドはテンプレートのアプリ (Ansible、Terraform、OpenTofu、Terragrunt、Bash、Python、PowerShell など) によって異なります。一部のアプリにしかないフィールドには、そのことが書かれています。変更は、**保存** の後に開始されたタスクに反映されます。

## テンプレートのタイプ {#templateForm-typeTabs}

上部のタブで、テンプレートの種類を設定します。

- **タスク**: コードを実行します。ほとんどのテンプレートはこの種類です。
- **ビルド**: アーティファクトを作成します。各ビルドタスクは、**開始バージョン** から数えたバージョンを得ます。
- **デプロイ**: ビルドで作成されたアーティファクトをデプロイします。開始するときに、どのバージョンをデプロイするかをビルドタスクで選びます。

ビルドまたはデプロイのタスクは、環境に変数 `SEMAPHORE_TASK_TYPE`、`SEMAPHORE_TASK_TARGET_VERSION` (ビルド)、`SEMAPHORE_TASK_INCOMING_VERSION` (デプロイ) を受け取ります。Ansible は同じ値を追加の変数 `semaphore_vars.task_details` で受け取ります。

## 開始バージョン {#templateForm-startVersion}

タイプが **ビルド** の場合に表示されます。必須です。

最初のビルドタスクのバージョンです。例: `1.0.0` または `v0.1`。以降の各ビルドでは、前のバージョンの最後の数値が 1 増え、その前後のテキストとゼロ埋めは保持されます: `1.0.0`、`1.0.1`、`1.0.2`。先に進めるには、ここで数値を最後のビルドより大きくします。例: `2.0.0`。数値を含まないバージョンは変わりません。

## ビルドテンプレート {#templateForm-buildTemplateId}

タイプが **デプロイ** の場合に表示されます。必須です。

このテンプレートがデプロイするアーティファクトを、そのタスクで作成するテンプレートです。一覧には、プロジェクトのビルドテンプレートと、その下にデプロイテンプレートが表示されます。デプロイは別のデプロイの後に続けることができ、その場合はそのデプロイがビルドから得たバージョンを得ます。

## 自動実行 {#templateForm-autorun}

タイプが **デプロイ** の場合に表示されます。

**ビルドテンプレート** のタスクが成功するたびに、そのタスクが作成したバージョンで、このテンプレートのタスクを自動的に開始します。失敗したビルドでは何も開始されません。

## 名前 {#templateForm-name}

タスクテンプレート ページに表示されるテンプレートの名前です。このフィールドは必須です。

## リポジトリ {#templateForm-repositoryId}

実行するコードを保持するプロジェクトのリポジトリです。このフィールドは必須です。Semaphore はタスクのたびにリポジトリのコピーを更新し、そこからコードを実行します。

フィールドの右側のラベルには、タスクがチェックアウトするブランチが表示されます。クリックすると **ブランチ** フィールドが開きます。

## ブランチ {#templateForm-gitBranch}

**リポジトリ** の右側のブランチのラベルをクリックした後に表示されます。

リポジトリに設定されたブランチの代わりに、このテンプレートのタスクがチェックアウトするブランチまたはタグです。リポジトリのブランチを使う場合は、このフィールドを空のままにします。Semaphore がリポジトリのブランチを読み取れる場合、このフィールドにはそれらが一覧表示されます。

リポジトリがモノレポの場合、このフィールドは無効です。すべてのテンプレートが、リポジトリのブランチで 1 つのチェックアウトを共有します。

## プレイブックファイルへのパス {#templateForm-playbook}

実行するもので、リポジトリのルートからのパスで指定します。例: `deploy/site.yml`。パスは絶対パスにできず、リポジトリの外に出ることもできません。一覧にはリポジトリのファイルが表示されますが、一覧にないパスを入力することもできます。フィールドの右側のボタンで一覧を再度読み込みます。

ラベルと意味はアプリによって変わります。

- **Ansible**: プレイブックです。必須です。一覧には `.yml` と `.yaml` のファイルが表示されます。
- **Terraform**、**OpenTofu**、**Terragrunt**: **Subdirectory path**、つまりコードのあるディレクトリです。リポジトリのルートでコードを実行する場合は空のままにします。ディレクトリの内容を一覧表示するには `/` を入力します。
- **Bash**、**Python**、**PowerShell**、その他のアプリ: **Script Filename**、つまり実行するスクリプトです。必須です。スクリプトはリポジトリのルートで実行されます。

## Different working directory {#templateForm-workingDirectoryToggle}

Ansible のテンプレートで表示されます。

`ansible-playbook` をリポジトリのルート以外のディレクトリで実行します。チェックを入れると **Working directory** フィールドが表示されます。再びルートで実行するには、チェックを外します。

## Working directory {#templateForm-workingDirectory}

**Different working directory** にチェックが入っている場合に表示されます。必須です。

`ansible-playbook` と `ansible-galaxy` が実行されるディレクトリで、リポジトリのルートからのパスで指定します。例: `ansible`。Ansible はそのファイルを現在のディレクトリから読み込むため、プロジェクトの `ansible.cfg` がサブディレクトリにある場合に使います。プレイブックのパスは、引き続きリポジトリのルートからの相対パスです。パスは絶対パスにできず、リポジトリの外に出ることもできません。

## インベントリ {#templateForm-inventoryId}

Ansible、Terraform、OpenTofu、Terragrunt のテンプレートで表示されます。

- **Ansible**: プレイブックが対象とするインベントリで、ホストの資格情報を含みます。このフィールドは必須です。
- **Terraform**、**OpenTofu**、**Terragrunt**: **Workspace**、つまりコードが実行されるワークスペースです。Semaphore はプランの前にそれを選択し、存在しない場合は作成します。ワークスペースなしで保存された新しいテンプレートには、`default` という名前のワークスペースが作成されます。

一覧には、テンプレートのアプリ用に作られたプロジェクトのインベントリが表示されます。

## 変数グループ {#templateForm-environmentIds}

任意です。このテンプレートのすべてのタスクが、その変数、環境変数、シークレットを受け取る変数グループです。複数選択できます。フィールドの順序で統合され、両方が同じ名前を設定している場合は後のグループが前のグループより優先されます。

変数がコードに渡される方法はアプリによって異なります。Ansible では追加の変数、Terraform と OpenTofu では `-var`、スクリプトでは `name=value` 引数です。環境変数はすべてのアプリで設定されます。

## ビュー {#templateForm-viewId}

任意です。テンプレートが一覧表示されるビューです。ビューはタスクテンプレート ページのタブで、プロジェクトのテンプレートをグループにまとめます。テンプレートをどのビューにも入れない場合は、このフィールドを空のままにします。

## ランナータグ {#templateForm-runnerTag}

プロジェクトのランナーが利用できる場合に表示されます。

このテンプレートのタスクは、サーバー自身で実行できる場合でも、このタグを持つランナーでのみ実行されます。このタグはインベントリのランナータグより優先されます。インベントリのタグを使う場合、またはタグを使わない場合は、このフィールドを空のままにします。

## Executor image {#templateForm-executorImage}

このテンプレートのタスクのためにランナーが起動するコンテナーイメージです。例: `semaphoreui/job:latest`。ランナーの設定のイメージを使う場合は、このフィールドを空のままにします。

タスクをコンテナーで実行するのは Docker と Kubernetes のエグゼキューターだけなので、このフィールドはそのいずれかが利用できる場合にのみ有効です。コンテナーなしで実行されるタスクは、これを無視します。

## 調査変数 {#surveyVars}

このテンプレートのタスクの開始時にユーザーに尋ねる値です。**+** をクリックすると変数を追加し、チップをクリックすると変更し、チップのバツ印で削除します。チップをドラッグすると、質問の順序を設定できます。

回答は、各変数の設定に応じて、変数グループの変数と同じように、または環境変数としてコードに渡されます。回答は、同じ名前を設定している変数グループより優先されます。

## Allow parallel tasks {#templateForm-allowParallelTasks}

このテンプレートのタスクを同時に実行できるようにします。これがない場合、テンプレートのタスクは実行中のタスクが終了するまでキューで待機します。いずれの場合も、プロジェクトとサーバーの並列タスク数の上限が適用されます。

## Issue JWT to task runner {#templateForm-jwtEnabled}

管理者がサーバーで JWT の発行を有効にしている場合に表示されます。

このテンプレートのすべてのタスクに、Semaphore が署名した JSON Web Token を環境変数 `SEMAPHORE_JWT` で渡します。コードはこのトークンを別のサービスに提示して、どのタスクであるかを証明できます。トークンには、タスク、プロジェクト、テンプレート、タスクを開始したユーザーが含まれます。サービスは、Semaphore サーバーの `/.well-known/jwks.json` にある公開鍵で署名を確認します。

## JWT audience {#templateForm-jwtAudience}

**Issue JWT to task runner** にチェックが入っている場合に表示されます。任意です。

トークンの対象となるサービスで、トークンの `aud` クレームに書き込まれます。例: `https://vault.example.com`。値を入力するたびに Enter キーを押します。最大 32 個までです。オーディエンスのないトークンにする場合は、このフィールドを空のままにします。

## JWT TTL {#templateForm-jwtTtl}

**Issue JWT to task runner** にチェックが入っている場合に表示されます。任意です。

タスクの開始からトークンが有効な期間で、`30s`、`15m`、`1h` のような期間で指定します。サーバーのデフォルトを使う場合は、このフィールドを空のままにします。管理者が別の値を設定していない限り、1 時間です。値はサーバーの最大値を超えることはできません。別に設定されていない限り、最大値は 24 時間です。

## 新しいコミットのためだけにCronでタスクを実行したい {#templateForm-cronVisible}

リポジトリに新しいコミットが現れたときに、このテンプレートのタスクを開始します。Semaphore は選択した間隔でリポジトリに最新のコミットを問い合わせ、前回の確認以降にコミットが変わった場合にのみタスクを開始します。

チェックを入れて保存すると確認のスケジュールが作成され、チェックを外して保存するとそのスケジュールが削除されます。テンプレートのその他のスケジュールは保持されます。

## リポジトリ {#templateForm-cronRepositoryId}

**新しいコミットのためだけにCronでタスクを実行したい** にチェックが入っている場合に表示されます。必須です。

新しいコミットを監視するリポジトリです。Semaphore は、テンプレートの **ブランチ** ではなく、そのリポジトリに設定されたブランチを監視します。

## チェック間隔 {#templateForm-cronFormat}

**新しいコミットのためだけにCronでタスクを実行したい** にチェックが入っている場合に表示されます。

Semaphore がリポジトリに新しいコミットがあるかを確認する頻度で、毎分から 1 日 1 回までです。デフォルトは毎分です。

## 成功アラートを抑制 {#templateForm-suppressSuccessAlerts}

このテンプレートの成功したタスクがアラートを送信しないようにします。失敗したタスクのアラートは引き続き送信されます。プロジェクトでアラートが有効な場合にのみ効果があります。

## エラーアラートを抑制 {#templateForm-suppressErrorAlerts}

このテンプレートの失敗したタスクがアラートを送信しないようにします。成功したタスクのアラートは引き続き送信されます。プロジェクトでアラートが有効な場合にのみ効果があります。

## CLI args {#templateForm-args}

すべてのタスクのコマンドに追加される引数です。各チップが 1 つの引数で、シェルを介さずにそのまま渡されます。`--forks=10` を 1 つのチップに入れるか、`--forks` と `10` を 2 つのチップに入れます。チップをドラッグすると順序を変更できます。一覧の右側のボタンで、引数を JSON 配列として開きます。

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: `ansible-playbook` のプレイブックの前に追加されます。
- **Terraform**、**OpenTofu**、**Terragrunt**: `plan` と `apply` に追加され、`init` には追加されません。
- スクリプト: スクリプト名の後に追加されます。

## CLI引数 {#templateForm-allowOverrideArgsInTask}

**プロンプト** の下にあります。タスクの開始時に、ユーザーが独自の引数を追加できるようにします。それらはテンプレートの引数の後に追加されます。

## ブランチ {#templateForm-allowOverrideBranchInTask}

**プロンプト** の下にあります。タスクの開始時に、テンプレートのブランチの代わりに、ユーザーが別のブランチ、タグ、コミットを選択できるようにします。これがない場合、タスクは常にテンプレートのブランチをチェックアウトします。

リポジトリがモノレポの場合、このボックスは無効です。

## インベントリ {#templateForm-allowOverrideInventory}

**プロンプト** の下にあります。Ansible、Terraform、OpenTofu、Terragrunt のテンプレートで表示されます。

タスクの開始時に、ユーザーが別のインベントリを選択できるようにします。そこで選択したインベントリを使うのは Ansible のテンプレートだけです。Terraform、OpenTofu、Terragrunt のタスクは、ボックスにチェックが入っていても、テンプレートのワークスペースで実行されます。

## 制限 {#templateForm-limit}

Ansible のテンプレートで表示されます。

インベントリの中から、プレイブックを実行するホストまたはグループで、`--limit` として渡されます。各チップが 1 つのパターンです。例: `web` または `db01`。すべてのホストで実行するには、一覧を空のままにします。

## タグ {#templateForm-tags}

Ansible のテンプレートで表示されます。

これらのタグを持つタスクとロールだけを実行し、`--tags` として渡されます。すべてを実行するには、一覧を空のままにします。

## スキップタグ {#templateForm-skipTags}

Ansible のテンプレートで表示されます。

これらのタグを持つタスクとロールをスキップし、`--skip-tags` として渡されます。

## ボールト {#templateVaults}

Ansible のテンプレートで表示されます。

プレイブックがファイルを復号するために必要な Ansible Vault のパスワードです。**+** をクリックすると追加し、チップをクリックすると変更し、チップのバツ印で削除します。各ボールトには名前と Vault ID があり、キーのストアのキー、またはリポジトリのクライアントスクリプトからパスワードを渡します。

## 自動承認 {#templateForm-autoApprove}

Terraform、OpenTofu、Terragrunt のテンプレートで表示されます。

確認なしでプランを適用します。これがない場合、タスクはプランの後に停止し、ユーザーが確認または拒否するまで待機します。変更のないプランでは、適用するものがないままタスクが終了します。

## Override backend settings {#templateForm-overrideBackend}

Semaphore が Terraform の状態を保持できる場合に、Terraform、OpenTofu、Terragrunt のテンプレートで表示されます。

コードの状態を Semaphore に保持します。`init` の前に、Semaphore は `http` バックエンドを設定するファイルをコードのディレクトリに書き込み、タスクの後にそれを削除します。

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

**Override backend settings** とともに表示され、そのボックスにチェックが入っている場合に有効になります。任意です。

Semaphore がバックエンドを書き込むファイルの名前です。例: `semaphore_backend.tf`。デフォルトは `backend.tf` です。名前は `.tf` で終わり、英字、数字、`_`、`-`、`.` のみを含む必要があります。コードのファイルにない名前を選んでください。Semaphore はそのファイルを上書きし、タスクの後に削除します。

## 制限 {#templateForm-allowOverrideLimit}

**Ansibleプロンプト** の下にあります。タスクの開始時に、テンプレートの **制限** の代わりに、ユーザーが実行するホストを設定できるようにします。

## タグ {#templateForm-allowOverrideTags}

**Ansibleプロンプト** の下にあります。タスクの開始時に、テンプレートの **タグ** の代わりに、ユーザーがタグを設定できるようにします。

## スキップタグ {#templateForm-allowOverrideSkipTags}

**Ansibleプロンプト** の下にあります。タスクの開始時に、テンプレートの **スキップタグ** の代わりに、ユーザーがスキップするタグを設定できるようにします。

## デバッグ {#templateForm-allowDebug}

**Ansibleプロンプト** の下にあります。タスクの開始時に、ユーザーが `-v` から `-vvvvvv` までの詳細な出力を有効にできるようにします。

## ドライラン {#templateForm-showDryRun}

**Ansibleプロンプト** の下にあります。デフォルトでチェックが入っています。タスクの開始時に **ドライラン** オプションを表示します。これはプレイブックを `--check` 付きで実行し、Ansible は何を変更するかを報告し、何も変更しません。このオプションをなくすには、チェックを外します。

## 差分 {#templateForm-showDiff}

**Ansibleプロンプト** の下にあります。デフォルトでチェックが入っています。タスクの開始時に **差分** オプションを表示します。これはプレイブックを `--diff` 付きで実行し、ファイルに加えられた変更を表示します。このオプションをなくすには、チェックを外します。

## 自動承認 {#templateForm-allowAutoApprove}

Terraform、OpenTofu、Terragrunt のプロンプトの下にあります。タスクの開始時に、ユーザーが **Auto Approve** にチェックを入れて、確認なしでプランを適用できるようにします。タスクのダイアログはいずれの場合もこのオプションを表示しますが、このボックスがない場合 Semaphore はそれを無視します。テンプレートの **自動承認** にチェックが入っていない限り、タスクはプランの後に確認を待ちます。

## Galaxy install options {#templateForm-galaxy}

Ansible のテンプレートで表示されます。クリックすると Galaxy のインストールのオプションが表示されます。ラベルには、デフォルトと異なるオプションの数が表示されます。

プレイブックの実行前に、Semaphore はリポジトリが求めるロールとコレクションを `ansible-galaxy` でインストールします。プレイブックの隣とリポジトリのルートで `requirements.yml`、`roles/requirements.yml`、`collections/requirements.yml` を探し、前回のタスク以降にファイルが変更された場合にのみインストールを実行します。

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

**Galaxy install options** の下にあります。Galaxy のインストールをスキップします。プレイブックは以前にインストールされたロールとコレクションで実行されます。要件が別の方法でインストールされている場合や、インストールが遅くて不要な場合に使います。

## Role install args {#templateForm-galaxyRoleArgs}

**Galaxy install options** の下にあります。`ansible-galaxy role install` に追加される引数で、チップ 1 つにつき 1 つです。例: `--ignore-certs` または `--server=https://galaxy.example.com`。Semaphore は、渡しても安全なインストールのフラグだけを受け付けます。不明なフラグ、`--token`、`--api-key`、`-p`、`--roles-path` があると、テンプレートの保存に失敗します。

## Collection install args {#templateForm-galaxyCollectionArgs}

**Galaxy install options** の下にあります。`ansible-galaxy collection install` に追加される引数で、チップ 1 つにつき 1 つです。例: `--pre` または `--upgrade`。**Role install args** と同じ規則が適用されます。`--token`、`--api-key`、`-p`、`--collections-path` は受け付けられません。

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

**Galaxy install options** の下の **Prompts** にあります。タスクの開始時に、Galaxy のインストールをスキップするかどうかをユーザーが選べるようにします。タスクでの選択は、テンプレートの **Skip Galaxy install** より優先されます。
