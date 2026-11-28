---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Шляхи сховища секретів, з яких Semaphore імпортує секрети, і як часто він робить це
автоматично. Якщо діалог відкрито зі сховища секретів, імпорт заповнює «Сховище ключів»
проєкту; якщо з групи змінних — цю групу. Те, що ви тут змінюєте, зберігається, коли ви
натискаєте **Закрити**, і записується разом зі сховищем або групою кнопкою **Зберегти**.

## Auto-sync interval {#varGroupSync-syncInterval}

Відображається, коли діалог відкрито з групи змінних.

Як часто група імпортується автоматично, у хвилинах. `0`, значення за замовчуванням, вимикає
автоматичний імпорт: тоді група імпортується лише тоді, коли ви натискаєте її кнопку
синхронізації на сторінці «середовище».

Сервер перевіряє раз на хвилину й імпортує, коли з останньої спроби, успішної чи ні, минув
інтервал; перша перевірка після того, як ви задали інтервал, імпортує одразу. Невдалий імпорт
повторюється після наступного інтервалу.

## Auto-sync interval {#secretStorageSync-syncInterval}

Відображається, коли діалог відкрито зі сховища секретів.

Як часто сховище імпортується автоматично, у хвилинах. `0`, значення за замовчуванням, вимикає
автоматичний імпорт: тоді сховище імпортується лише тоді, коли ви натискаєте його кнопку
синхронізації на вкладці **Storages** «Сховища ключів».

Сервер перевіряє раз на хвилину й імпортує, коли з останньої спроби, успішної чи ні, минув
інтервал; перша перевірка після того, як ви задали інтервал, імпортує одразу. Невдалий імпорт
повторюється після наступного інтервалу.

## Шлях {#secretStorageSync-path}

Відображається в кожному рядку списку. Рядок додається кнопкою **Add path**.

Шлях у сховищі, з якого імпортуються секрети, записаний так, як його називає сховище. Кожен рядок
— один шлях; імпорт зчитує їх усі.

## separator {#secretStorageSync-separator}

Відображається в кожному рядку списку. Необов’язкове поле.

Роздільник, який використовує імпорт цього шляху. Залиште поле порожнім, якщо назвам у вашому
сховищі він не потрібен.

## prefix {#secretStorageSync-prefix}

Відображається в кожному рядку списку. Необов’язкове поле.

Префікс, який використовує імпорт цього шляху. Залиште поле порожнім, якщо імпортованим секретам
він не потрібен.

## Видалити шлях {#secretStorageSync-removePath}

Відображається в кожному рядку списку.

Видаляє шлях зі списку. Зміна набуває чинності, коли сховище або групу збережено.

## Add path {#secretStorageSync-addPath}

Додає до списку порожній рядок. Доки немає принаймні одного шляху, діалог повідомляє, що шляхи
синхронізації не задано, а кнопка синхронізації сховища або групи залишається неактивною.
