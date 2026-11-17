---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# 编辑 Secret Storage

秘密存储是一个外部秘密管理器，Semaphore 将秘密值保存在其中并从中读取：HashiCorp Vault、OpenBao、Devolutions
Server、AWS Secrets Manager 或 Azure Key Vault。使用存储创建的变量组会将其秘密值保存在那里，开启了同步的存储会将其秘密值导入项目的密钥存储。
存储的种类在 **New Storage** 菜单中选择，无法在此对话框中更改；下面的字段取决于存储的种类。

## 名称 {#secretStorage-name}

存储在密钥存储的 **Storages** 标签页以及新建变量组的 **Secret storage** 选择器中显示的名称。此字段为必填项。

## Server URL {#secretStorage-vaultURL}

当存储为 HashiCorp Vault、OpenBao 或 Devolutions Server 时显示。

服务器的地址，包括协议，例如 `https://vault.example.com:8200`。此字段为必填项。

## Mount {#secretStorage-vaultMount}

当存储为 HashiCorp Vault 或 OpenBao 时显示。可选。

保存秘密值的秘密引擎的挂载路径。将此字段留空则使用默认值 `secret`。

## Namespace {#secretStorage-vaultNamespace}

当存储为 HashiCorp Vault 或 OpenBao 时显示。可选。

秘密值所在的命名空间。命名空间存在于 Vault Enterprise 和 HCP Vault Dedicated 中，以及 2.3 及以上版本的 OpenBao
中。对于没有命名空间的服务器，请将此字段留空。

## 秘密值来源 {#secretSource-options}

对于每种存储都会显示，位于 Semaphore 登录所用的秘密值上方：Vault 或 OpenBao 的 **Token**、Devolutions Server
的应用秘密、AWS 的秘密密钥、Azure 的客户端密码。

Semaphore 从何处获取该秘密值：

- **Store in DB**：您在此对话框中输入秘密值；Semaphore 将其加密并保存在数据库中。
- **From ENV**：您输入 Semaphore 服务器的一个环境变量的名称；每次使用该存储时都会从中读取秘密值。
- **From File**：您输入 Semaphore 服务器上一个文件的绝对路径；每次都会从中读取秘密值，不包括末尾的换行符。该文件必须位于服务器的秘密目录中
  （配置中的 `dirs.secrets`，默认为 `/tmp/semaphore`），包含 `..` 的路径会被拒绝。

切换来源会清除其下方的字段。

## Token {#secretStorage-vaultToken}

当存储为 HashiCorp Vault 或 OpenBao、且秘密值来源为 **Store in DB** 时显示。

Semaphore 向服务器进行身份验证时使用的令牌。它必须有权读取秘密值，并且除非存储为 **Read only**，还必须有权写入和删除秘密值。

对于新建的存储，此字段为必填项。打开已有的存储时，此字段为空：留空则保留已保存的令牌，输入新令牌则替换它。

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

当存储为 HashiCorp Vault 或 OpenBao、且秘密值来源为 **From ENV** 或 **From File** 时显示。

保存令牌的环境变量的名称或文件的路径。对于新建的存储，此字段为必填项。

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

当存储为 Devolutions Server 时显示。

Semaphore 连接服务器时不检查其 TLS 证书。此时，Semaphore 与服务器之间的任何人都可以读取秘密值。请保持关闭，并改为给服务器配置一个
Semaphore 信任的证书。

## Vault ID {#secretStorage-dvlsKey}

当存储为 Devolutions Server 时显示。

Devolutions Server 中保存秘密值的保管库的 ID。对于新建的存储，此字段为必填项。

## App Key {#secretStorage-dvlsAppKey}

当存储为 Devolutions Server 时显示。

Semaphore 登录时所用的应用程序身份的应用程序密钥，其秘密填入 **Secret**。对于新建的存储，此字段为必填项。

## Secret {#secretStorage-dvlsSecret}

当存储为 Devolutions Server、且秘密值来源为 **Store in DB** 时显示。

**App Key** 所对应的应用程序身份的应用秘密。对于新建的存储，此字段为必填项。打开已有的存储时，此字段为空：留空则保留已保存的秘密，
输入新秘密则替换它。

## Env var name / Path to the file {#secretStorage-dvlsEnv}

当存储为 Devolutions Server、且秘密值来源为 **From ENV** 或 **From File** 时显示。

保存应用秘密的环境变量的名称或文件的路径。对于新建的存储，此字段为必填项。

## Region {#secretStorage-awsRegion}

当存储为 AWS Secrets Manager 时显示。

秘密值所在的 AWS 区域，例如 `us-east-1`。此字段为必填项。

## Endpoint URL {#secretStorage-awsEndpointURL}

当存储为 AWS Secrets Manager 时显示。可选。

用来替代该区域默认端点进行调用的 Secrets Manager API 地址，例如 VPC 端点。将此字段留空则使用默认的 AWS 端点。

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

当存储为 AWS Secrets Manager 时显示。

开启时，Semaphore 不保存访问密钥，而是从服务器运行的环境中获取 AWS 凭据，例如其实例的 IAM 角色。此时下方的访问密钥字段处于禁用状态。

开启后，保存时会删除已保存的访问密钥。再次关闭时需要一个新的 **Secret Access Key**。

## Access Key ID {#secretStorage-awsAccessKeyId}

当存储为 AWS Secrets Manager 时显示。当 **Use IAM Role / Instance Profile** 开启时处于禁用状态。

Semaphore 登录时使用的访问密钥的 ID，例如 `AKIAIOSFODNN7EXAMPLE`。当 IAM 角色关闭时，此字段为必填项。ID
不是秘密值：它按原样保存和显示。

## Secret Access Key {#secretStorage-awsSecretKey}

当存储为 AWS Secrets Manager、且秘密值来源为 **Store in DB** 时显示。当 **Use IAM Role / Instance Profile**
开启时处于禁用状态。

访问密钥的秘密部分。对于新建的存储以及关闭 IAM 角色之后，此字段为必填项。其他情况下，打开存储时此字段为空：留空则保留已保存的密钥，
输入新密钥则替换它。

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

当存储为 AWS Secrets Manager、且秘密值来源为 **From ENV** 或 **From File** 时显示。当
**Use IAM Role / Instance Profile** 开启时处于禁用状态。

保存秘密访问密钥的环境变量的名称或文件的路径。在与 **Secret Access Key** 相同的情况下，此字段为必填项。

## Vault URL {#secretStorage-azureVaultURL}

当存储为 Azure Key Vault 时显示。

密钥保管库的地址，例如 `https://my-vault.vault.azure.net`。此字段为必填项。

## Tenant ID {#secretStorage-azureTenantId}

当存储为 Azure Key Vault 时显示。

Semaphore 登录时所用的应用程序所在的 Microsoft Entra 租户的 ID。此字段为必填项。

## Client ID {#secretStorage-azureClientId}

当存储为 Azure Key Vault 时显示。

该应用程序的应用程序（客户端）ID。它需要有权访问密钥保管库中的秘密值。此字段为必填项。

## Client Secret {#secretStorage-azureClientSecret}

当存储为 Azure Key Vault、且秘密值来源为 **Store in DB** 时显示。

应用程序的客户端密码。对于新建的存储，此字段为必填项。打开已有的存储时，此字段为空：留空则保留已保存的秘密，输入新秘密则替换它。

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

当存储为 Azure Key Vault、且秘密值来源为 **From ENV** 或 **From File** 时显示。

保存客户端密码的环境变量的名称或文件的路径。对于新建的存储，此字段为必填项。

## Read only {#secretStorage-readonly}

开启时，Semaphore 只从存储中读取。无法在其中创建或更改变量组的秘密值：保存会失败。在 Semaphore 中删除秘密值、密钥或变量组时，
其秘密值会保留在存储中。

关闭时，Semaphore 会将使用该存储的变量组的秘密值写入其中，并在删除变量组时从中删除这些秘密值。

## Sync keys enabled {#secretStorage-syncEnabled}

开启后，会从 **Sync paths** 下设置的路径，将秘密值从存储导入项目的密钥存储。当您在 **Storages** 标签页中点击存储的同步按钮时，
以及按照 **Sync paths** 中设置的间隔自动进行时，都会运行导入。

导入的密钥会被标记为已同步：其名称和类型无法更改，在 Semaphore 中删除它时，秘密值会保留在存储中。在同步开启时删除存储，
也会删除它所导入的密钥。

## Sync paths {#secretStorage-syncPaths}

当 **Sync keys enabled** 开启时显示。

打开一个对话框，其中包含要从中导入的存储路径以及自动导入的间隔。按钮上的数字是路径的数量。在存储至少有一个路径之前，
其同步按钮保持禁用状态。

该对话框只有**关闭**按钮：您在其中所做的更改会随存储一起，通过此对话框的**保存**（对于新建的存储为**创建**）进行保存。
