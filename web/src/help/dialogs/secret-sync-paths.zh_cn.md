---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Semaphore 从中导入秘密值的秘密存储路径，以及它自动导入的频率。从秘密存储打开时，导入会填充项目的密钥存储；
从变量组打开时，导入会填充该变量组。您在此处所做的更改会在点击**关闭**时保留，并在存储或变量组点击其**保存**时一并保存。

## Auto-sync interval {#varGroupSync-syncInterval}

从变量组打开此对话框时显示。

自动导入变量组的频率，以分钟为单位。默认值 `0` 表示关闭自动导入：此时只有在变量组页面中点击其同步按钮时，变量组才会被导入。

服务器每分钟检查一次，当距离上次尝试（无论成功与否）已经过了间隔时间时进行导入；设置间隔之后的第一次检查会立即导入。
导入失败后，会在下一个间隔之后重试。

## Auto-sync interval {#secretStorageSync-syncInterval}

从秘密存储打开此对话框时显示。

自动导入存储的频率，以分钟为单位。默认值 `0` 表示关闭自动导入：此时只有在密钥存储的 **Storages**
标签页中点击其同步按钮时，存储才会被导入。

服务器每分钟检查一次，当距离上次尝试（无论成功与否）已经过了间隔时间时进行导入；设置间隔之后的第一次检查会立即导入。
导入失败后，会在下一个间隔之后重试。

## 路径 {#secretStorageSync-path}

在列表的每一行中显示。使用 **Add path** 添加一行。

存储中要从中导入秘密值的路径，按存储命名它的方式书写。每一行是一个路径；导入会读取所有路径。

## separator {#secretStorageSync-separator}

在列表的每一行中显示。可选。

导入此路径时使用的分隔符。除非存储中的名称需要分隔符，否则请留空。

## prefix {#secretStorageSync-prefix}

在列表的每一行中显示。可选。

导入此路径时使用的前缀。如果导入的秘密值不需要前缀，请留空。

## 删除路径 {#secretStorageSync-removePath}

在列表的每一行中显示。

从列表中删除该路径。此更改在保存存储或变量组时生效。

## Add path {#secretStorageSync-addPath}

在列表中添加一个空行。在至少有一个路径之前，对话框会提示未定义同步路径，并且存储或变量组的同步按钮保持禁用状态。
