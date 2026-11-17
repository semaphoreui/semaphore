---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# 編集 Secret Storage

シークレットストレージは、Semaphore がシークレットを保管し、そこから読み込む外部のシークレットマネージャーです。HashiCorp Vault、OpenBao、Devolutions Server、AWS Secrets Manager、Azure Key Vault があります。ストレージを指定して作成した変数グループは、シークレットをそこに保管します。同期を有効にしたストレージは、そのシークレットをプロジェクトのキーのストアにインポートします。ストレージの種類は **New Storage** メニューで選択し、このダイアログでは変更できません。下のフィールドは種類によって異なります。

## 名前 {#secretStorage-name}

キーのストアの **Storages** タブと、新しい変数グループの **Secret storage** 選択欄に表示されるストレージの名前です。このフィールドは必須です。

## Server URL {#secretStorage-vaultURL}

ストレージが HashiCorp Vault、OpenBao、Devolutions Server の場合に表示されます。

スキームを含むサーバーのアドレスです。例: `https://vault.example.com:8200`。このフィールドは必須です。

## Mount {#secretStorage-vaultMount}

ストレージが HashiCorp Vault または OpenBao の場合に表示されます。任意です。

シークレットを保持するシークレットエンジンのマウントパスです。デフォルトの `secret` を使う場合は、このフィールドを空のままにします。

## Namespace {#secretStorage-vaultNamespace}

ストレージが HashiCorp Vault または OpenBao の場合に表示されます。任意です。

シークレットが存在する名前空間です。名前空間は Vault Enterprise と HCP Vault Dedicated、およびバージョン 2.3 以降の OpenBao にあります。名前空間のないサーバーでは、このフィールドを空のままにします。

## シークレットのソース {#secretSource-options}

すべての種類のストレージで、Semaphore がサインインに使うシークレットの上に表示されます。Vault または OpenBao の **Token**、Devolutions Server のアプリのシークレット、AWS のシークレットキー、Azure のクライアントシークレットです。

Semaphore がそのシークレットを取得する場所です。

- **Store in DB**: シークレットをこのダイアログに入力します。Semaphore はそれを暗号化してデータベースに保管します。
- **From ENV**: Semaphore サーバーの環境変数の名前を入力します。ストレージを使うたびに、シークレットはその変数から読み込まれます。
- **From File**: Semaphore サーバー上のファイルの絶対パスを入力します。シークレットは毎回そのファイルから、末尾の改行を除いて読み込まれます。ファイルはサーバーのシークレットディレクトリ (設定の `dirs.secrets`、デフォルトは `/tmp/semaphore`) の中にある必要があり、`..` を含むパスは拒否されます。

ソースを切り替えると、その下のフィールドはクリアされます。

## Token {#secretStorage-vaultToken}

ストレージが HashiCorp Vault または OpenBao で、シークレットのソースが **Store in DB** の場合に表示されます。

Semaphore がサーバーへの認証に使うトークンです。シークレットの読み取り、およびストレージが **Read only** でない限りその書き込みと削除が許可されている必要があります。

新しいストレージではこのフィールドは必須です。既存のストレージを開くと空になっています。保存されたトークンを保持するには空のままにし、置き換えるには新しいトークンを入力します。

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

ストレージが HashiCorp Vault または OpenBao で、シークレットのソースが **From ENV** または **From File** の場合に表示されます。

トークンを保持する環境変数の名前、またはファイルのパスです。新しいストレージではこのフィールドは必須です。

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

ストレージが Devolutions Server の場合に表示されます。

Semaphore は TLS 証明書を確認せずにサーバーに接続します。その場合、Semaphore とサーバーの間にいる誰でもシークレットを読み取れます。オフのままにし、代わりに Semaphore が信頼する証明書をサーバーに設定してください。

## Vault ID {#secretStorage-dvlsKey}

ストレージが Devolutions Server の場合に表示されます。

シークレットを保持する Devolutions Server のボールトの ID です。新しいストレージではこのフィールドは必須です。

## App Key {#secretStorage-dvlsAppKey}

ストレージが Devolutions Server の場合に表示されます。

Semaphore がサインインに使うアプリケーション ID のアプリケーションキーで、そのシークレットを **Secret** に入力します。新しいストレージではこのフィールドは必須です。

## Secret {#secretStorage-dvlsSecret}

ストレージが Devolutions Server で、シークレットのソースが **Store in DB** の場合に表示されます。

**App Key** のアプリケーション ID のアプリのシークレットです。新しいストレージではこのフィールドは必須です。既存のストレージを開くと空になっています。保存されたシークレットを保持するには空のままにし、置き換えるには新しいシークレットを入力します。

## Env var name / Path to the file {#secretStorage-dvlsEnv}

ストレージが Devolutions Server で、シークレットのソースが **From ENV** または **From File** の場合に表示されます。

アプリのシークレットを保持する環境変数の名前、またはファイルのパスです。新しいストレージではこのフィールドは必須です。

## Region {#secretStorage-awsRegion}

ストレージが AWS Secrets Manager の場合に表示されます。

シークレットがある AWS のリージョンです。例: `us-east-1`。このフィールドは必須です。

## Endpoint URL {#secretStorage-awsEndpointURL}

ストレージが AWS Secrets Manager の場合に表示されます。任意です。

リージョンのデフォルトのエンドポイントの代わりに呼び出す Secrets Manager API のアドレスです。例: VPC エンドポイント。デフォルトの AWS エンドポイントを使う場合は、このフィールドを空のままにします。

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

ストレージが AWS Secrets Manager の場合に表示されます。

オンの場合、Semaphore はアクセスキーを保存せず、サーバーが動作している環境 (インスタンスの IAM ロールなど) から AWS の資格情報を取得します。その場合、下のアクセスキーのフィールドは無効になります。

オンにすると、保存時に保存済みのアクセスキーが削除されます。再びオフにするには、新しい **Secret Access Key** が必要です。

## Access Key ID {#secretStorage-awsAccessKeyId}

ストレージが AWS Secrets Manager の場合に表示されます。**Use IAM Role / Instance Profile** がオンの場合は無効です。

Semaphore がサインインに使うアクセスキーの ID です。例: `AKIAIOSFODNN7EXAMPLE`。IAM ロールがオフの場合、このフィールドは必須です。ID はシークレットではなく、そのまま保存され、表示されます。

## Secret Access Key {#secretStorage-awsSecretKey}

ストレージが AWS Secrets Manager で、シークレットのソースが **Store in DB** の場合に表示されます。**Use IAM Role / Instance Profile** がオンの場合は無効です。

アクセスキーのシークレットです。新しいストレージ、および IAM ロールをオフにした後は、このフィールドは必須です。それ以外の場合、ストレージを開くと空になっています。保存されたキーを保持するには空のままにし、置き換えるには新しいキーを入力します。

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

ストレージが AWS Secrets Manager で、シークレットのソースが **From ENV** または **From File** の場合に表示されます。**Use IAM Role / Instance Profile** がオンの場合は無効です。

シークレットアクセスキーを保持する環境変数の名前、またはファイルのパスです。このフィールドは **Secret Access Key** と同じ場合に必須です。

## Vault URL {#secretStorage-azureVaultURL}

ストレージが Azure Key Vault の場合に表示されます。

キーコンテナーのアドレスです。例: `https://my-vault.vault.azure.net`。このフィールドは必須です。

## Tenant ID {#secretStorage-azureTenantId}

ストレージが Azure Key Vault の場合に表示されます。

Semaphore がサインインに使うアプリケーションの Microsoft Entra テナントの ID です。このフィールドは必須です。

## Client ID {#secretStorage-azureClientId}

ストレージが Azure Key Vault の場合に表示されます。

そのアプリケーションのアプリケーション (クライアント) ID です。キーコンテナーのシークレットへのアクセス権が必要です。このフィールドは必須です。

## Client Secret {#secretStorage-azureClientSecret}

ストレージが Azure Key Vault で、シークレットのソースが **Store in DB** の場合に表示されます。

アプリケーションのクライアントシークレットです。新しいストレージではこのフィールドは必須です。既存のストレージを開くと空になっています。保存されたシークレットを保持するには空のままにし、置き換えるには新しいシークレットを入力します。

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

ストレージが Azure Key Vault で、シークレットのソースが **From ENV** または **From File** の場合に表示されます。

クライアントシークレットを保持する環境変数の名前、またはファイルのパスです。新しいストレージではこのフィールドは必須です。

## Read only {#secretStorage-readonly}

オンの場合、Semaphore はストレージから読み取るだけです。変数グループのシークレットをそこで作成または変更することはできず、保存は失敗します。Semaphore でシークレット、キー、変数グループを削除しても、そのシークレットはストレージに残ります。

オフの場合、Semaphore はストレージを使う変数グループのシークレットをそこに書き込み、グループとともにそこから削除します。

## Sync keys enabled {#secretStorage-syncEnabled}

オンにすると、**Sync paths** で設定したパスから、ストレージのシークレットをプロジェクトのキーのストアにインポートします。インポートは、**Storages** タブでストレージの同期ボタンを押したとき、および **Sync paths** で設定した間隔で自動的に実行されます。

インポートされたキーは同期済みとしてマークされます。その名前とタイプは変更できず、Semaphore で削除してもシークレットはストレージに残ります。同期がオンのままストレージを削除すると、インポートされたキーも削除されます。

## Sync paths {#secretStorage-syncPaths}

**Sync keys enabled** がオンの場合に表示されます。

インポート元となるストレージのパスと、自動インポートの間隔を設定するダイアログを開きます。ボタンの数字はパスの数です。少なくとも 1 つのパスがあるまで、ストレージの同期ボタンは無効のままです。

そのダイアログには **閉じる** しかありません。そこで変更した内容は、このダイアログの **保存** (新しいストレージでは **作成**) によってストレージとともに保存されます。
