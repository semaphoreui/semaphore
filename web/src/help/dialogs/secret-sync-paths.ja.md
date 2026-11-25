---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Semaphore がシークレットをインポートするシークレットストレージのパスと、自動でインポートする頻度です。シークレットストレージから開いた場合、インポートはプロジェクトのキーのストアに追加されます。変数グループから開いた場合は、そのグループに追加されます。ここで変更した内容は **閉じる** を押しても保持され、ストレージまたはグループの **保存** によってそれとともに保存されます。

## Auto-sync interval {#varGroupSync-syncInterval}

ダイアログを変数グループから開いた場合に表示されます。

グループを自動でインポートする頻度で、分単位です。デフォルトの `0` は自動インポートをオフにします。その場合、グループは 変数グループ ページでその同期ボタンを押したときにのみインポートされます。

サーバーは 1 分に 1 回確認し、成功したかどうかに関係なく、前回の試行から間隔が経過していればインポートします。間隔を設定した後の最初の確認では、すぐにインポートします。失敗したインポートは、次の間隔の後に再試行されます。

## Auto-sync interval {#secretStorageSync-syncInterval}

ダイアログをシークレットストレージから開いた場合に表示されます。

ストレージを自動でインポートする頻度で、分単位です。デフォルトの `0` は自動インポートをオフにします。その場合、ストレージはキーのストアの **Storages** タブでその同期ボタンを押したときにのみインポートされます。

サーバーは 1 分に 1 回確認し、成功したかどうかに関係なく、前回の試行から間隔が経過していればインポートします。間隔を設定した後の最初の確認では、すぐにインポートします。失敗したインポートは、次の間隔の後に再試行されます。

## パス {#secretStorageSync-path}

一覧のすべての行に表示されます。**Add path** で行を追加します。

シークレットをインポートするストレージ内のパスで、ストレージでの名前のとおりに書きます。各行が 1 つのパスで、インポートはそのすべてを読み込みます。

## separator {#secretStorageSync-separator}

一覧のすべての行に表示されます。任意です。

このパスのインポートで使われる区切り文字です。ストレージ内の名前に必要でない限り、空のままにします。

## prefix {#secretStorageSync-prefix}

一覧のすべての行に表示されます。任意です。

このパスのインポートで使われる接頭辞です。インポートされるシークレットに必要ない場合は、空のままにします。

## パスを削除 {#secretStorageSync-removePath}

一覧のすべての行に表示されます。

パスを一覧から削除します。変更は、ストレージまたはグループを保存したときに反映されます。

## Add path {#secretStorageSync-addPath}

一覧に空の行を追加します。少なくとも 1 つのパスがあるまで、ダイアログには同期パスが定義されていないと表示され、ストレージまたはグループの同期ボタンは無効のままです。
