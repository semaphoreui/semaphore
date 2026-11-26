---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Ścieżki magazynu sekretów, z których Semaphore importuje sekrety, oraz to, jak często robi to
samoczynnie. Otwarty z magazynu sekretów import wypełnia Magazyn kluczy projektu; otwarty z grupy
zmiennych wypełnia tę grupę. To, co tu zmienisz, jest zachowywane po naciśnięciu **Zamknij** i
zapisywane z magazynem lub grupą przyciskiem **Zapisz** ich okna dialogowego.

## Auto-sync interval {#varGroupSync-syncInterval}

Wyświetlane, gdy okno dialogowe otwarto z grupy zmiennych.

Jak często grupa jest importowana samoczynnie, w minutach. `0`, wartość domyślna, wyłącza
automatyczny import: grupa jest wtedy importowana tylko po naciśnięciu jej przycisku synchronizacji
na stronie Grupy zmiennych.

Serwer sprawdza raz na minutę i importuje, gdy od ostatniej próby, udanej lub nie, upłynął zadany
odstęp; pierwsze sprawdzenie po ustawieniu odstępu importuje od razu. Nieudany import jest
ponawiany po następnym odstępie.

## Auto-sync interval {#secretStorageSync-syncInterval}

Wyświetlane, gdy okno dialogowe otwarto z magazynu sekretów.

Jak często magazyn jest importowany samoczynnie, w minutach. `0`, wartość domyślna, wyłącza
automatyczny import: magazyn jest wtedy importowany tylko po naciśnięciu jego przycisku
synchronizacji na karcie **Storages** Magazynu kluczy.

Serwer sprawdza raz na minutę i importuje, gdy od ostatniej próby, udanej lub nie, upłynął zadany
odstęp; pierwsze sprawdzenie po ustawieniu odstępu importuje od razu. Nieudany import jest
ponawiany po następnym odstępie.

## Ścieżka {#secretStorageSync-path}

Wyświetlane w każdym wierszu listy. Dodaj wiersz przyciskiem **Add path**.

Ścieżka w magazynie, z której importowane są sekrety, zapisana tak, jak nazywa ją magazyn. Każdy
wiersz to jedna ścieżka; import odczytuje wszystkie.

## separator {#secretStorageSync-separator}

Wyświetlane w każdym wierszu listy. Opcjonalne.

Separator używany przy imporcie tej ścieżki. Pozostaw pole puste, chyba że nazwy w twoim magazynie
go wymagają.

## prefix {#secretStorageSync-prefix}

Wyświetlane w każdym wierszu listy. Opcjonalne.

Przedrostek używany przy imporcie tej ścieżki. Pozostaw pole puste, gdy importowane sekrety go nie
potrzebują.

## Usuń ścieżkę {#secretStorageSync-removePath}

Wyświetlane w każdym wierszu listy.

Usuwa ścieżkę z listy. Zmiana zaczyna obowiązywać po zapisaniu magazynu lub grupy.

## Add path {#secretStorageSync-addPath}

Dodaje pusty wiersz do listy. Dopóki nie ma co najmniej jednej ścieżki, okno dialogowe informuje,
że nie zdefiniowano ścieżek synchronizacji, a przycisk synchronizacji magazynu lub grupy pozostaje
wyłączony.
