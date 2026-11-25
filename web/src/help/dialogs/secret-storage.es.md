---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Editar Secret Storage

Un almacén de secretos es un gestor de secretos externo en el que Semaphore guarda secretos y del
que los lee: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager o Azure Key Vault.
Un grupo de variables creado con un almacén guarda allí sus secretos, y un almacén con la
sincronización activada importa sus secretos al Almacén de Claves del proyecto. El tipo de almacén
se elige en el menú **New Storage** y no se puede cambiar en este diálogo; los campos de abajo
dependen de él.

## Nombre {#secretStorage-name}

El nombre con el que el almacén aparece en la pestaña **Storages** del Almacén de Claves y en el
selector **Secret storage** de un nuevo grupo de variables. El campo es obligatorio.

## Server URL {#secretStorage-vaultURL}

Se muestra cuando el almacén es un HashiCorp Vault, un OpenBao o un Devolutions Server.

La dirección del servidor, con su esquema, por ejemplo `https://vault.example.com:8200`. El campo
es obligatorio.

## Mount {#secretStorage-vaultMount}

Se muestra cuando el almacén es un HashiCorp Vault o un OpenBao. Opcional.

La ruta de montaje del motor de secretos que contiene los secretos. Deje el campo vacío para usar
`secret`, el valor predeterminado.

## Namespace {#secretStorage-vaultNamespace}

Se muestra cuando el almacén es un HashiCorp Vault o un OpenBao. Opcional.

El espacio de nombres en el que están los secretos. Los espacios de nombres existen en Vault
Enterprise y HCP Vault Dedicated, y en OpenBao a partir de la versión 2.3. Deje el campo vacío para
un servidor sin espacios de nombres.

## Origen del secreto {#secretSource-options}

Se muestra para todos los tipos de almacén, encima del secreto con el que Semaphore inicia sesión:
el **Token** de un Vault o de un OpenBao, el secreto de aplicación de un Devolutions Server, la
clave secreta de AWS, el secreto de cliente de Azure.

De dónde toma Semaphore ese secreto:

- **Store in DB**: usted escribe el secreto en este diálogo; Semaphore lo cifra y lo guarda en su
  base de datos.
- **From ENV**: usted escribe el nombre de una variable de entorno del servidor de Semaphore; el
  secreto se lee de ella cada vez que se usa el almacén.
- **From File**: usted escribe la ruta absoluta de un archivo en el servidor de Semaphore; el
  secreto se lee de él cada vez, sin el salto de línea final. El archivo debe estar dentro del
  directorio de secretos del servidor (`dirs.secrets` en la configuración, `/tmp/semaphore` por
  defecto), y una ruta con `..` se rechaza.

Cambiar el origen vacía el campo de debajo.

## Token {#secretStorage-vaultToken}

Se muestra cuando el almacén es un HashiCorp Vault o un OpenBao y el origen del secreto es
**Store in DB**.

El token con el que Semaphore se autentica en el servidor. Debe tener permiso para leer los
secretos y, salvo que el almacén sea **Read only**, para escribirlos y eliminarlos.

El campo es obligatorio para un almacén nuevo. Cuando abre un almacén existente, está vacío: déjelo
vacío para conservar el token guardado, escriba uno nuevo para sustituirlo.

## Nombre de la variable de entorno / Ruta al archivo {#secretStorage-vaultTokenSource}

Se muestra cuando el almacén es un HashiCorp Vault o un OpenBao y el origen del secreto es
**From ENV** o **From File**.

El nombre de la variable de entorno, o la ruta del archivo, que contiene el token. El campo es
obligatorio para un almacén nuevo.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Se muestra cuando el almacén es un Devolutions Server.

Semaphore se conecta al servidor sin comprobar su certificado TLS. Cualquiera que esté entre
Semaphore y el servidor puede entonces leer los secretos. Déjelo desactivado y, en su lugar, dé al
servidor un certificado en el que Semaphore confíe.

## Vault ID {#secretStorage-dvlsKey}

Se muestra cuando el almacén es un Devolutions Server.

El ID del vault de Devolutions Server que contiene los secretos. El campo es obligatorio para un
almacén nuevo.

## App Key {#secretStorage-dvlsAppKey}

Se muestra cuando el almacén es un Devolutions Server.

La clave de aplicación de la identidad de aplicación con la que Semaphore inicia sesión, aquella
cuyo secreto va en **Secret**. El campo es obligatorio para un almacén nuevo.

## Secret {#secretStorage-dvlsSecret}

Se muestra cuando el almacén es un Devolutions Server y el origen del secreto es **Store in DB**.

El secreto de aplicación de la identidad de aplicación de **App Key**. El campo es obligatorio para
un almacén nuevo. Cuando abre un almacén existente, está vacío: déjelo vacío para conservar el
secreto guardado, escriba uno nuevo para sustituirlo.

## Nombre de la variable de entorno / Ruta al archivo {#secretStorage-dvlsEnv}

Se muestra cuando el almacén es un Devolutions Server y el origen del secreto es **From ENV** o
**From File**.

El nombre de la variable de entorno, o la ruta del archivo, que contiene el secreto de aplicación.
El campo es obligatorio para un almacén nuevo.

## Region {#secretStorage-awsRegion}

Se muestra cuando el almacén es un AWS Secrets Manager.

La región de AWS en la que están los secretos, por ejemplo `us-east-1`. El campo es obligatorio.

## Endpoint URL {#secretStorage-awsEndpointURL}

Se muestra cuando el almacén es un AWS Secrets Manager. Opcional.

La dirección de la API de Secrets Manager a la que llamar en lugar del endpoint predeterminado de
la región, por ejemplo un endpoint de VPC. Deje el campo vacío para usar el endpoint predeterminado
de AWS.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Se muestra cuando el almacén es un AWS Secrets Manager.

Cuando está activado, Semaphore no guarda ninguna clave de acceso y toma las credenciales de AWS del
entorno en el que se ejecuta el servidor, como el rol de IAM de su instancia. Los campos de la clave
de acceso de abajo quedan entonces deshabilitados.

Activarlo elimina la clave de acceso guardada al guardar. Volver a desactivarlo requiere una nueva
**Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Se muestra cuando el almacén es un AWS Secrets Manager. Deshabilitado cuando
**Use IAM Role / Instance Profile** está activado.

El ID de la clave de acceso con la que Semaphore inicia sesión, por ejemplo `AKIAIOSFODNN7EXAMPLE`.
El campo es obligatorio cuando el rol de IAM está desactivado. El ID no es un secreto: se guarda y
se muestra tal cual.

## Secret Access Key {#secretStorage-awsSecretKey}

Se muestra cuando el almacén es un AWS Secrets Manager y el origen del secreto es **Store in DB**.
Deshabilitado cuando **Use IAM Role / Instance Profile** está activado.

El secreto de la clave de acceso. El campo es obligatorio para un almacén nuevo y después de
desactivar el rol de IAM. En los demás casos está vacío cuando abre el almacén: déjelo vacío para
conservar la clave guardada, escriba una nueva para sustituirla.

## Nombre de la variable de entorno / Ruta al archivo {#secretStorage-awsSecretKeySource}

Se muestra cuando el almacén es un AWS Secrets Manager y el origen del secreto es **From ENV** o
**From File**. Deshabilitado cuando **Use IAM Role / Instance Profile** está activado.

El nombre de la variable de entorno, o la ruta del archivo, que contiene la clave de acceso secreta.
El campo es obligatorio en los mismos casos que **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Se muestra cuando el almacén es un Azure Key Vault.

La dirección del key vault, por ejemplo `https://my-vault.vault.azure.net`. El campo es
obligatorio.

## Tenant ID {#secretStorage-azureTenantId}

Se muestra cuando el almacén es un Azure Key Vault.

El ID del tenant de Microsoft Entra de la aplicación con la que Semaphore inicia sesión. El campo es
obligatorio.

## Client ID {#secretStorage-azureClientId}

Se muestra cuando el almacén es un Azure Key Vault.

El ID de aplicación (cliente) de esa aplicación. Necesita acceso a los secretos del key vault. El
campo es obligatorio.

## Client Secret {#secretStorage-azureClientSecret}

Se muestra cuando el almacén es un Azure Key Vault y el origen del secreto es **Store in DB**.

El secreto de cliente de la aplicación. El campo es obligatorio para un almacén nuevo. Cuando abre
un almacén existente, está vacío: déjelo vacío para conservar el secreto guardado, escriba uno
nuevo para sustituirlo.

## Nombre de la variable de entorno / Ruta al archivo {#secretStorage-azureClientSecretSource}

Se muestra cuando el almacén es un Azure Key Vault y el origen del secreto es **From ENV** o
**From File**.

El nombre de la variable de entorno, o la ruta del archivo, que contiene el secreto de cliente. El
campo es obligatorio para un almacén nuevo.

## Read only {#secretStorage-readonly}

Cuando está activado, Semaphore solo lee del almacén. Un secreto de un grupo de variables no se
puede crear ni cambiar en él: guardarlo falla. Al eliminar un secreto, una clave o un grupo de
variables en Semaphore, su secreto se queda en el almacén.

Cuando está desactivado, Semaphore escribe en el almacén los secretos de los grupos de variables que
lo usan y los elimina de él junto con los grupos.

## Sync keys enabled {#secretStorage-syncEnabled}

Actívelo para importar secretos del almacén al Almacén de Claves del proyecto, desde las rutas
definidas en **Sync paths**. La importación se ejecuta cuando pulsa el botón de sincronización del
almacén en la pestaña **Storages**, y por sí sola con el intervalo definido en **Sync paths**.

Una clave importada se marca como sincronizada: su nombre y su tipo no se pueden cambiar, y al
eliminarla en Semaphore el secreto se queda en el almacén. Eliminar el almacén mientras la
sincronización está activada elimina también las claves que importó.

## Sync paths {#secretStorage-syncPaths}

Se muestra cuando **Sync keys enabled** está activado.

Abre el diálogo con las rutas del almacén desde las que importar y el intervalo de la importación
automática. El número del botón es el número de rutas. El botón de sincronización del almacén
permanece deshabilitado hasta que tenga al menos una ruta.

El diálogo solo tiene **Cerrar**: lo que cambie en él se guarda con el almacén, mediante
**Guardar** de este diálogo (**Crear** para un almacén nuevo).
