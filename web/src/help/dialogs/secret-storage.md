---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Edit Secret Storage

A secret storage is an external secret manager that Semaphore keeps secrets in and reads them
from: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager or Azure Key Vault. A
variable group created with a storage keeps its secrets there, and a storage with sync turned on
imports its secrets into the Key Store of the project. The kind of the storage is picked in the
**New Storage** menu and can not be changed in this dialog; the fields below depend on it.

## Name {#secretStorage-name}

The name the storage is listed under on the **Storages** tab of the Key Store and in the
**Secret storage** selector of a new variable group. The field is required.

## Server URL {#secretStorage-vaultURL}

Shown when the storage is a HashiCorp Vault, an OpenBao or a Devolutions Server.

The address of the server, with its scheme, for example `https://vault.example.com:8200`. The
field is required.

## Mount {#secretStorage-vaultMount}

Shown when the storage is a HashiCorp Vault or an OpenBao. Optional.

The mount path of the secrets engine that holds the secrets. Leave the field empty to use
`secret`, the default.

## Namespace {#secretStorage-vaultNamespace}

Shown when the storage is a HashiCorp Vault or an OpenBao. Optional.

The namespace the secrets live in. Namespaces exist in Vault Enterprise and HCP Vault Dedicated,
and in OpenBao from version 2.3. Leave the field empty for a server without namespaces.

## Secret source {#secretSource-options}

Shown for every kind of storage, above the secret Semaphore signs in with: the **Token** of a
Vault or an OpenBao, the app secret of a Devolutions Server, the secret key of AWS, the client
secret of Azure.

Where Semaphore takes that secret from:

- **Store in DB**: you type the secret into this dialog; Semaphore encrypts it and keeps it in
  its database.
- **From ENV**: you type the name of an environment variable of the Semaphore server; the secret
  is read from it every time the storage is used.
- **From File**: you type the absolute path of a file on the Semaphore server; the secret is read
  from it every time, without the final line break. The file must be inside the secrets
  directory of the server (`dirs.secrets` in the configuration, `/tmp/semaphore` by default), and
  a path with `..` is refused.

Switching the source clears the field below it.

## Token {#secretStorage-vaultToken}

Shown when the storage is a HashiCorp Vault or an OpenBao and the secret source is
**Store in DB**.

The token Semaphore authenticates to the server with. It must be allowed to read the secrets
and, unless the storage is **Read only**, to write and delete them.

The field is required for a new storage. When you open an existing storage, it is empty: leave
it empty to keep the stored token, type a new one to replace it.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Shown when the storage is a HashiCorp Vault or an OpenBao and the secret source is
**From ENV** or **From File**.

The name of the environment variable, or the path of the file, that holds the token. The field
is required for a new storage.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Shown when the storage is a Devolutions Server.

Semaphore connects to the server without checking its TLS certificate. Anyone between Semaphore
and the server can then read the secrets. Leave it off, and give the server a certificate that
Semaphore trusts instead.

## Vault ID {#secretStorage-dvlsKey}

Shown when the storage is a Devolutions Server.

The ID of the vault of Devolutions Server that holds the secrets. The field is required for a
new storage.

## App Key {#secretStorage-dvlsAppKey}

Shown when the storage is a Devolutions Server.

The application key of the application identity Semaphore signs in with, the one whose secret
goes into **Secret**. The field is required for a new storage.

## Secret {#secretStorage-dvlsSecret}

Shown when the storage is a Devolutions Server and the secret source is **Store in DB**.

The app secret of the application identity of **App Key**. The field is required for a new
storage. When you open an existing storage, it is empty: leave it empty to keep the stored
secret, type a new one to replace it.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Shown when the storage is a Devolutions Server and the secret source is **From ENV** or
**From File**.

The name of the environment variable, or the path of the file, that holds the app secret. The
field is required for a new storage.

## Region {#secretStorage-awsRegion}

Shown when the storage is an AWS Secrets Manager.

The AWS region the secrets are in, for example `us-east-1`. The field is required.

## Endpoint URL {#secretStorage-awsEndpointURL}

Shown when the storage is an AWS Secrets Manager. Optional.

The address of the Secrets Manager API to call instead of the default endpoint of the region, for
example a VPC endpoint. Leave the field empty to use the default AWS endpoint.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Shown when the storage is an AWS Secrets Manager.

When on, Semaphore stores no access key and takes the AWS credentials from the environment the
server runs in, such as the IAM role of its instance. The fields of the access key below are then
disabled.

Turning it on removes the stored access key when you save. Turning it off again requires a new
**Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Shown when the storage is an AWS Secrets Manager. Disabled when
**Use IAM Role / Instance Profile** is on.

The ID of the access key Semaphore signs in with, for example `AKIAIOSFODNN7EXAMPLE`. The field
is required when the IAM role is off. The ID is not a secret: it is stored and shown as it is.

## Secret Access Key {#secretStorage-awsSecretKey}

Shown when the storage is an AWS Secrets Manager and the secret source is **Store in DB**.
Disabled when **Use IAM Role / Instance Profile** is on.

The secret of the access key. The field is required for a new storage and after the IAM role is
turned off. Otherwise it is empty when you open the storage: leave it empty to keep the stored
key, type a new one to replace it.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Shown when the storage is an AWS Secrets Manager and the secret source is **From ENV** or
**From File**. Disabled when **Use IAM Role / Instance Profile** is on.

The name of the environment variable, or the path of the file, that holds the secret access key.
The field is required in the same cases as **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Shown when the storage is an Azure Key Vault.

The address of the key vault, for example `https://my-vault.vault.azure.net`. The field is
required.

## Tenant ID {#secretStorage-azureTenantId}

Shown when the storage is an Azure Key Vault.

The ID of the Microsoft Entra tenant of the application Semaphore signs in as. The field is
required.

## Client ID {#secretStorage-azureClientId}

Shown when the storage is an Azure Key Vault.

The application (client) ID of that application. It needs access to the secrets of the key
vault. The field is required.

## Client Secret {#secretStorage-azureClientSecret}

Shown when the storage is an Azure Key Vault and the secret source is **Store in DB**.

The client secret of the application. The field is required for a new storage. When you open an
existing storage, it is empty: leave it empty to keep the stored secret, type a new one to
replace it.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Shown when the storage is an Azure Key Vault and the secret source is **From ENV** or
**From File**.

The name of the environment variable, or the path of the file, that holds the client secret. The
field is required for a new storage.

## Read only {#secretStorage-readonly}

When on, Semaphore only reads from the storage. A secret of a variable group can not be created
or changed in it: saving it fails. Deleting a secret, a key or a variable group in Semaphore
leaves its secret in the storage.

When off, Semaphore writes the secrets of the variable groups that use the storage into it and
deletes them from it with the groups.

## Sync keys enabled {#secretStorage-syncEnabled}

Turn it on to import secrets from the storage into the Key Store of the project, from the paths
set under **Sync paths**. The import runs when you press the sync button of the storage on the
**Storages** tab, and on its own at the interval set in **Sync paths**.

An imported key is marked as synchronized: its name and type can not be changed, and deleting it
in Semaphore leaves the secret in the storage. Deleting the storage while sync is on deletes the
keys it imported as well.

## Sync paths {#secretStorage-syncPaths}

Shown when **Sync keys enabled** is on.

Opens the dialog with the paths of the storage to import from and the interval of the automatic
import. The number on the button is the number of paths. The sync button of the storage stays
disabled until it has at least one path.

The dialog has only **Close**: what you change in it is saved with the storage, by **Save** of
this dialog (**Create** for a new storage).
