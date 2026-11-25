---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Cesty v úložišti tajných údajů, ze kterých Semaphore importuje tajné údaje, a jak často to dělá
sám. Když dialog otevřete z úložiště tajných údajů, import plní Úložiště klíčů projektu; když ho
otevřete ze skupiny proměnných, plní tuto skupinu. Co zde změníte, se zachová po stisknutí
**Zavřít** a uloží se s úložištěm nebo se skupinou jejich tlačítkem **Uložit**.

## Auto-sync interval {#varGroupSync-syncInterval}

Zobrazí se, když je dialog otevřen ze skupiny proměnných.

Jak často se skupina sama importuje, v minutách. `0`, výchozí hodnota, automatický import vypne:
skupina se pak importuje pouze tehdy, když stisknete její tlačítko synchronizace na stránce
Skupiny proměnných.

Server kontroluje jednou za minutu a importuje, když od posledního pokusu, úspěšného či ne,
uplynul interval; první kontrola po nastavení intervalu importuje okamžitě. Neúspěšný import se
zopakuje po dalším intervalu.

## Auto-sync interval {#secretStorageSync-syncInterval}

Zobrazí se, když je dialog otevřen z úložiště tajných údajů.

Jak často se úložiště samo importuje, v minutách. `0`, výchozí hodnota, automatický import
vypne: úložiště se pak importuje pouze tehdy, když stisknete jeho tlačítko synchronizace na kartě
**Storages** Úložiště klíčů.

Server kontroluje jednou za minutu a importuje, když od posledního pokusu, úspěšného či ne,
uplynul interval; první kontrola po nastavení intervalu importuje okamžitě. Neúspěšný import se
zopakuje po dalším intervalu.

## Cesta {#secretStorageSync-path}

Zobrazí se na každém řádku seznamu. Řádek přidáte tlačítkem **Add path**.

Cesta v úložišti, ze které se importují tajné údaje, zapsaná tak, jak ji úložiště pojmenovává.
Každý řádek je jedna cesta; import čte všechny.

## separator {#secretStorageSync-separator}

Zobrazí se na každém řádku seznamu. Volitelné.

Oddělovač, který používá import této cesty. Ponechte pole prázdné, pokud ho názvy ve vašem
úložišti nepotřebují.

## prefix {#secretStorageSync-prefix}

Zobrazí se na každém řádku seznamu. Volitelné.

Předpona, kterou používá import této cesty. Ponechte pole prázdné, když importované tajné údaje
žádnou nepotřebují.

## Odebrat cestu {#secretStorageSync-removePath}

Zobrazí se na každém řádku seznamu.

Odebere cestu ze seznamu. Změna se projeví po uložení úložiště nebo skupiny.

## Add path {#secretStorageSync-addPath}

Přidá do seznamu prázdný řádek. Dokud neexistuje alespoň jedna cesta, dialog hlásí, že nejsou
definované žádné cesty synchronizace, a tlačítko synchronizace úložiště nebo skupiny zůstane
neaktivní.
