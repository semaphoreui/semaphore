---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

Les chemins d'un stockage de secrets depuis lesquels Semaphore importe des secrets, et la fréquence
à laquelle il le fait automatiquement. Lorsque la boîte de dialogue est ouverte depuis un stockage
de secrets, l'import remplit le Magasin de clés du projet ; lorsqu'elle est ouverte depuis un
groupe de variables, il remplit ce groupe. Ce que
vous modifiez ici est conservé lorsque vous appuyez sur **Fermer** et enregistré avec le stockage
ou le groupe, par son **Enregistrer**.

## Auto-sync interval {#varGroupSync-syncInterval}

Affiché lorsque la boîte de dialogue est ouverte depuis un groupe de variables.

La fréquence, en minutes, à laquelle le groupe est importé automatiquement. `0`, la valeur par
défaut, désactive l'import automatique : le groupe n'est alors importé que lorsque vous appuyez sur
son bouton de synchronisation sur la page Groupes de variables.

Le serveur vérifie une fois par minute et importe lorsque l'intervalle s'est écoulé depuis la
dernière tentative, réussie ou non ; la première vérification après la définition d'un intervalle
importe immédiatement. Un import échoué est retenté après l'intervalle suivant.

## Auto-sync interval {#secretStorageSync-syncInterval}

Affiché lorsque la boîte de dialogue est ouverte depuis un stockage de secrets.

La fréquence, en minutes, à laquelle le stockage est importé automatiquement. `0`, la valeur par
défaut, désactive l'import automatique : le stockage n'est alors importé que lorsque vous appuyez
sur son bouton de synchronisation dans l'onglet **Storages** du Magasin de clés.

Le serveur vérifie une fois par minute et importe lorsque l'intervalle s'est écoulé depuis la
dernière tentative, réussie ou non ; la première vérification après la définition d'un intervalle
importe immédiatement. Un import échoué est retenté après l'intervalle suivant.

## Chemin {#secretStorageSync-path}

Affiché dans chaque ligne de la liste. Ajoutez une ligne avec **Add path**.

Un chemin dans le stockage depuis lequel importer les secrets, écrit tel que le stockage le nomme.
Chaque ligne est un chemin ; l'import les lit tous.

## separator {#secretStorageSync-separator}

Affiché dans chaque ligne de la liste. Optionnel.

Un séparateur qu'utilise l'import de ce chemin. Laissez-le vide, sauf si les noms de votre stockage
en ont besoin.

## prefix {#secretStorageSync-prefix}

Affiché dans chaque ligne de la liste. Optionnel.

Un préfixe qu'utilise l'import de ce chemin. Laissez-le vide lorsque les secrets importés n'en ont
pas besoin.

## Supprimer le chemin {#secretStorageSync-removePath}

Affiché dans chaque ligne de la liste.

Retire le chemin de la liste. La modification prend effet lorsque le stockage ou le groupe est
enregistré.

## Add path {#secretStorageSync-addPath}

Ajoute une ligne vide à la liste. Tant qu'il n'y a aucun chemin, la boîte de dialogue indique
qu'aucun chemin de synchronisation n'est défini, et le bouton de synchronisation du stockage ou du
groupe reste désactivé.
