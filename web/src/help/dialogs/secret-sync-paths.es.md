---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Las rutas de un almacén de secretos desde las que Semaphore importa secretos, y con qué frecuencia
lo hace por sí solo. Abierta desde un almacén de secretos, la importación llena el Almacén de Claves
del proyecto; abierta desde un grupo de variables, llena ese grupo. Lo que cambie aquí se conserva
al pulsar **Cerrar** y se guarda con el almacén o el grupo, mediante su **Guardar**.

## Auto-sync interval {#varGroupSync-syncInterval}

Se muestra cuando el diálogo se abre desde un grupo de variables.

Con qué frecuencia se importa el grupo por sí solo, en minutos. `0`, el valor predeterminado,
desactiva la importación automática: el grupo solo se importa entonces cuando pulsa su botón de
sincronización en la página Grupos de Variables.

El servidor comprueba una vez por minuto e importa cuando ha pasado el intervalo desde el último
intento, haya tenido éxito o no; la primera comprobación después de establecer un intervalo importa
de inmediato. Una importación fallida se vuelve a intentar después del siguiente intervalo.

## Auto-sync interval {#secretStorageSync-syncInterval}

Se muestra cuando el diálogo se abre desde un almacén de secretos.

Con qué frecuencia se importa el almacén por sí solo, en minutos. `0`, el valor predeterminado,
desactiva la importación automática: el almacén solo se importa entonces cuando pulsa su botón de
sincronización en la pestaña **Storages** del Almacén de Claves.

El servidor comprueba una vez por minuto e importa cuando ha pasado el intervalo desde el último
intento, haya tenido éxito o no; la primera comprobación después de establecer un intervalo importa
de inmediato. Una importación fallida se vuelve a intentar después del siguiente intervalo.

## Ruta {#secretStorageSync-path}

Se muestra en cada fila de la lista. Añada una fila con **Add path**.

Una ruta del almacén desde la que importar los secretos, escrita tal como la nombra el almacén. Cada
fila es una ruta; la importación las lee todas.

## separator {#secretStorageSync-separator}

Se muestra en cada fila de la lista. Opcional.

Un separador que usa la importación de esta ruta. Déjelo vacío salvo que los nombres de su almacén
lo necesiten.

## prefix {#secretStorageSync-prefix}

Se muestra en cada fila de la lista. Opcional.

Un prefijo que usa la importación de esta ruta. Déjelo vacío cuando los secretos importados no lo
necesiten.

## Eliminar ruta {#secretStorageSync-removePath}

Se muestra en cada fila de la lista.

Elimina la ruta de la lista. El cambio se aplica cuando se guarda el almacén o el grupo.

## Add path {#secretStorageSync-addPath}

Añade una fila vacía a la lista. Mientras no haya al menos una ruta, el diálogo indica que no hay
rutas de sincronización definidas, y el botón de sincronización del almacén o del grupo permanece
deshabilitado.
