---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

I percorsi di un archivio di segreti da cui Semaphore importa i segreti, e con quale frequenza lo fa
da solo. Aperta da un archivio di segreti, l'importazione riempie il Negozio chiavi del progetto;
aperta da un gruppo di variabili, riempie quel gruppo. Ciò che modifichi qui viene mantenuto quando
premi **Chiudi** e salvato con l'archivio o con il gruppo, dal suo **Salva**.

## Auto-sync interval {#varGroupSync-syncInterval}

Visibile quando la finestra di dialogo è aperta da un gruppo di variabili.

Ogni quanto il gruppo viene importato da solo, in minuti. `0`, il valore predefinito, disattiva
l'importazione automatica: il gruppo viene allora importato solo quando premi il suo pulsante di
sincronizzazione nella pagina Gruppi di variabili.

Il server controlla una volta al minuto e importa quando l'intervallo è trascorso dall'ultimo
tentativo, riuscito o meno; il primo controllo dopo che hai impostato un intervallo importa subito.
Un'importazione fallita viene ritentata dopo l'intervallo successivo.

## Auto-sync interval {#secretStorageSync-syncInterval}

Visibile quando la finestra di dialogo è aperta da un archivio di segreti.

Ogni quanto l'archivio viene importato da solo, in minuti. `0`, il valore predefinito, disattiva
l'importazione automatica: l'archivio viene allora importato solo quando premi il suo pulsante di
sincronizzazione nella scheda **Storages** del Negozio chiavi.

Il server controlla una volta al minuto e importa quando l'intervallo è trascorso dall'ultimo
tentativo, riuscito o meno; il primo controllo dopo che hai impostato un intervallo importa subito.
Un'importazione fallita viene ritentata dopo l'intervallo successivo.

## Percorso {#secretStorageSync-path}

Visibile in ogni riga dell'elenco. Aggiungi una riga con **Add path**.

Un percorso nell'archivio da cui importare i segreti, scritto come lo indica l'archivio. Ogni riga è
un percorso; l'importazione li legge tutti.

## separator {#secretStorageSync-separator}

Visibile in ogni riga dell'elenco. Opzionale.

Un separatore che l'importazione di questo percorso usa. Lascialo vuoto a meno che i nomi nel tuo
archivio non ne abbiano bisogno.

## prefix {#secretStorageSync-prefix}

Visibile in ogni riga dell'elenco. Opzionale.

Un prefisso che l'importazione di questo percorso usa. Lascialo vuoto quando i segreti importati non
ne hanno bisogno.

## Remove path {#secretStorageSync-removePath}

Visibile in ogni riga dell'elenco.

Rimuove il percorso dall'elenco. La modifica ha effetto quando l'archivio o il gruppo viene salvato.

## Add path {#secretStorageSync-addPath}

Aggiunge una riga vuota all'elenco. Finché non c'è almeno un percorso, la finestra di dialogo dice
che non sono definiti percorsi di sincronizzazione, e il pulsante di sincronizzazione dell'archivio
o del gruppo resta disattivato.
