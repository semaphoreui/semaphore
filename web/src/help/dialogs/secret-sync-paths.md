---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

The paths of a secret storage that Semaphore imports secrets from, and how often it does so on
its own. Opened from a secret storage, the import fills the Key Store of the project; opened
from a variable group, it fills that group. What you change here is kept when you press
**Close** and saved with the storage or the group, by its **Save**.

## Auto-sync interval {#varGroupSync-syncInterval}

Shown when the dialog is opened from a variable group.

How often the group is imported on its own, in minutes. `0`, the default, turns the automatic
import off: the group is then imported only when you press its sync button on the Variable
Groups page.

The server checks once a minute and imports when the interval has passed since the last
attempt, successful or not; the first check after you set an interval imports at once. A failed
import is tried again after the next interval.

## Auto-sync interval {#secretStorageSync-syncInterval}

Shown when the dialog is opened from a secret storage.

How often the storage is imported on its own, in minutes. `0`, the default, turns the automatic
import off: the storage is then imported only when you press its sync button on the
**Storages** tab of the Key Store.

The server checks once a minute and imports when the interval has passed since the last
attempt, successful or not; the first check after you set an interval imports at once. A failed
import is tried again after the next interval.

## Path {#secretStorageSync-path}

Shown in every row of the list. Add a row with **Add path**.

A path in the storage to import the secrets from, written as the storage names it. Each row is
one path; the import reads all of them.

## separator {#secretStorageSync-separator}

Shown in every row of the list. Optional.

A separator that the import of this path uses. Leave it empty unless the names in your storage
need one.

## prefix {#secretStorageSync-prefix}

Shown in every row of the list. Optional.

A prefix that the import of this path uses. Leave it empty when the imported secrets need none.

## Remove path {#secretStorageSync-removePath}

Shown in every row of the list.

Removes the path from the list. The change takes effect when the storage or the group is saved.

## Add path {#secretStorageSync-addPath}

Adds an empty row to the list. Until there is at least one path, the dialog says that no sync
paths are defined, and the sync button of the storage or the group stays disabled.
