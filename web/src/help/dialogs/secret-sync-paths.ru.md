---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Пути хранилища секретов, из которых Semaphore импортирует секреты, и как часто он делает это
автоматически. Если диалог открыт из хранилища секретов, импорт заполняет раздел «Хранилище
ключей» проекта; если он открыт из группы переменных, импорт заполняет эту группу. То, что вы
здесь меняете, сохраняется при нажатии **Закрыть** и записывается вместе с хранилищем или группой
кнопкой **Сохранить** их диалога.

## Auto-sync interval {#varGroupSync-syncInterval}

Показывается, когда диалог открыт из группы переменных.

Как часто группа импортируется автоматически, в минутах. `0` — значение по умолчанию — выключает
автоматический импорт: тогда группа импортируется, только когда вы нажимаете ее кнопку
синхронизации на странице «Группы переменных».

Сервер проверяет раз в минуту и выполняет импорт, когда с последней попытки, успешной или нет,
прошел интервал; первая проверка после того, как вы задали интервал, сразу выполняет импорт.
Неудачный импорт повторяется через следующий интервал.

## Auto-sync interval {#secretStorageSync-syncInterval}

Показывается, когда диалог открыт из хранилища секретов.

Как часто хранилище импортируется автоматически, в минутах. `0` — значение по умолчанию —
выключает автоматический импорт: тогда хранилище импортируется, только когда вы нажимаете его
кнопку синхронизации на вкладке **Storages** раздела «Хранилище ключей».

Сервер проверяет раз в минуту и выполняет импорт, когда с последней попытки, успешной или нет,
прошел интервал; первая проверка после того, как вы задали интервал, сразу выполняет импорт.
Неудачный импорт повторяется через следующий интервал.

## Путь {#secretStorageSync-path}

Показывается в каждой строке списка. Строка добавляется кнопкой **Add path**.

Путь в хранилище, из которого импортируются секреты, записанный так, как его называет хранилище.
Каждая строка — один путь; импорт читает их все.

## separator {#secretStorageSync-separator}

Показывается в каждой строке списка. Необязательное поле.

Разделитель, который использует импорт этого пути. Оставьте поле пустым, если имена в вашем
хранилище не требуют разделителя.

## prefix {#secretStorageSync-prefix}

Показывается в каждой строке списка. Необязательное поле.

Префикс, который использует импорт этого пути. Оставьте поле пустым, если импортируемым секретам
префикс не нужен.

## Удалить путь {#secretStorageSync-removePath}

Показывается в каждой строке списка.

Удаляет путь из списка. Изменение вступает в силу, когда хранилище или группа сохранены.

## Add path {#secretStorageSync-addPath}

Добавляет в список пустую строку. Пока нет хотя бы одного пути, диалог сообщает, что пути
синхронизации не заданы, а кнопка синхронизации хранилища или группы остается недоступной.
