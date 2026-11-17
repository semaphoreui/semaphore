---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Die Pfade eines Secret-Speichers, aus denen Semaphore Secrets importiert, und wie oft es das von
selbst tut. Von einem Secret-Speicher aus geöffnet, füllt der Import den Schlüsselspeicher des
Projekts; von einer Variablengruppe aus geöffnet, füllt er diese Gruppe. Was Sie hier ändern,
bleibt erhalten, wenn Sie auf **Schließen** klicken, und wird mit dem Speicher oder der Gruppe
durch deren **Speichern** gespeichert.

## Auto-sync interval {#varGroupSync-syncInterval}

Wird angezeigt, wenn der Dialog von einer Variablengruppe aus geöffnet wird.

Wie oft die Gruppe von selbst importiert wird, in Minuten. `0`, die Voreinstellung, schaltet den
automatischen Import aus: Die Gruppe wird dann nur importiert, wenn Sie auf der Seite
Variablengruppen auf ihre Synchronisierungsschaltfläche klicken.

Der Server prüft einmal pro Minute und importiert, wenn das Intervall seit dem letzten Versuch
abgelaufen ist, ob erfolgreich oder nicht; die erste Prüfung, nachdem Sie ein Intervall
festgelegt haben, importiert sofort. Ein fehlgeschlagener Import wird nach dem nächsten Intervall
erneut versucht.

## Auto-sync interval {#secretStorageSync-syncInterval}

Wird angezeigt, wenn der Dialog von einem Secret-Speicher aus geöffnet wird.

Wie oft der Speicher von selbst importiert wird, in Minuten. `0`, die Voreinstellung, schaltet
den automatischen Import aus: Der Speicher wird dann nur importiert, wenn Sie auf dem Tab
**Storages** des Schlüsselspeichers auf seine Synchronisierungsschaltfläche klicken.

Der Server prüft einmal pro Minute und importiert, wenn das Intervall seit dem letzten Versuch
abgelaufen ist, ob erfolgreich oder nicht; die erste Prüfung, nachdem Sie ein Intervall
festgelegt haben, importiert sofort. Ein fehlgeschlagener Import wird nach dem nächsten Intervall
erneut versucht.

## Pfad {#secretStorageSync-path}

Wird in jeder Zeile der Liste angezeigt. Fügen Sie mit **Add path** eine Zeile hinzu.

Ein Pfad im Speicher, aus dem die Secrets importiert werden, so geschrieben, wie der Speicher ihn
nennt. Jede Zeile ist ein Pfad; der Import liest alle.

## separator {#secretStorageSync-separator}

Wird in jeder Zeile der Liste angezeigt. Optional.

Ein Trennzeichen, das der Import dieses Pfads verwendet. Lassen Sie das Feld leer, es sei denn,
die Namen in Ihrem Speicher brauchen eines.

## prefix {#secretStorageSync-prefix}

Wird in jeder Zeile der Liste angezeigt. Optional.

Ein Präfix, das der Import dieses Pfads verwendet. Lassen Sie das Feld leer, wenn die
importierten Secrets keines brauchen.

## Pfad entfernen {#secretStorageSync-removePath}

Wird in jeder Zeile der Liste angezeigt.

Entfernt den Pfad aus der Liste. Die Änderung wird wirksam, wenn der Speicher oder die Gruppe
gespeichert wird.

## Add path {#secretStorageSync-addPath}

Fügt der Liste eine leere Zeile hinzu. Solange es keinen Pfad gibt, meldet der Dialog, dass keine
Sync-Pfade festgelegt sind, und die Synchronisierungsschaltfläche des Speichers oder der Gruppe
bleibt deaktiviert.
