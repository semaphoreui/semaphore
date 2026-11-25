---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Modifier Secret Storage

Un stockage de secrets est un gestionnaire de secrets externe dans lequel Semaphore conserve des
secrets et depuis lequel il les lit : HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets
Manager ou Azure Key Vault. Un groupe de variables créé avec un stockage y conserve ses secrets, et
un stockage dont la synchronisation est activée importe ses secrets dans le Magasin de clés du
projet. Le type du stockage se choisit dans le menu **New Storage** et ne peut pas être modifié
dans cette boîte de dialogue ; les champs ci-dessous en dépendent.

## Nom {#secretStorage-name}

Le nom sous lequel le stockage apparaît dans l'onglet **Storages** du Magasin de clés et dans le
sélecteur **Secret storage** d'un nouveau groupe de variables. Le champ est obligatoire.

## Server URL {#secretStorage-vaultURL}

Affiché lorsque le stockage est un HashiCorp Vault, un OpenBao ou un Devolutions Server.

L'adresse du serveur, avec son schéma, par exemple `https://vault.example.com:8200`. Le champ est
obligatoire.

## Mount {#secretStorage-vaultMount}

Affiché lorsque le stockage est un HashiCorp Vault ou un OpenBao. Optionnel.

Le chemin de montage du moteur de secrets qui contient les secrets. Laissez le champ vide pour
utiliser `secret`, la valeur par défaut.

## Namespace {#secretStorage-vaultNamespace}

Affiché lorsque le stockage est un HashiCorp Vault ou un OpenBao. Optionnel.

L'espace de noms dans lequel se trouvent les secrets. Les espaces de noms existent dans Vault
Enterprise et HCP Vault Dedicated, ainsi que dans OpenBao à partir de la version 2.3. Laissez le
champ vide pour un serveur sans espaces de noms.

## Source du secret {#secretSource-options}

Affiché pour tous les types de stockage, au-dessus du secret avec lequel Semaphore se connecte : le
**Token** d'un Vault ou d'un OpenBao, le secret d'application d'un Devolutions Server, la clé
secrète d'AWS, le secret client d'Azure.

L'endroit où Semaphore prend ce secret :

- **Store in DB** : vous saisissez le secret dans cette boîte de dialogue ; Semaphore le chiffre et
  le conserve dans sa base de données.
- **From ENV** : vous saisissez le nom d'une variable d'environnement du serveur Semaphore ; le
  secret y est lu à chaque utilisation du stockage.
- **From File** : vous saisissez le chemin absolu d'un fichier sur le serveur Semaphore ; le secret
  y est lu à chaque fois, sans le saut de ligne final. Le fichier doit se trouver dans le
  répertoire des secrets du serveur (`dirs.secrets` dans la configuration, `/tmp/semaphore` par
  défaut), et un chemin contenant `..` est refusé.

Changer de source efface le champ situé en dessous.

## Token {#secretStorage-vaultToken}

Affiché lorsque le stockage est un HashiCorp Vault ou un OpenBao et que la source du secret est
**Store in DB**.

Le jeton avec lequel Semaphore s'authentifie auprès du serveur. Il doit être autorisé à lire les
secrets et, sauf si le stockage est **Read only**, à les écrire et à les supprimer.

Le champ est obligatoire pour un nouveau stockage. Lorsque vous ouvrez un stockage existant, il est
vide : laissez-le vide pour conserver le jeton enregistré, saisissez-en un nouveau pour le
remplacer.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Affiché lorsque le stockage est un HashiCorp Vault ou un OpenBao et que la source du secret est
**From ENV** ou **From File**.

Le nom de la variable d'environnement, ou le chemin du fichier, qui contient le jeton. Le champ est
obligatoire pour un nouveau stockage.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Affiché lorsque le stockage est un Devolutions Server.

Semaphore se connecte au serveur sans vérifier son certificat TLS. Quiconque se trouve entre
Semaphore et le serveur peut alors lire les secrets. Laissez l'option désactivée et donnez plutôt
au serveur un certificat auquel Semaphore fait confiance.

## Vault ID {#secretStorage-dvlsKey}

Affiché lorsque le stockage est un Devolutions Server.

L'ID du coffre de Devolutions Server qui contient les secrets. Le champ est obligatoire pour un
nouveau stockage.

## App Key {#secretStorage-dvlsAppKey}

Affiché lorsque le stockage est un Devolutions Server.

La clé d'application de l'identité d'application avec laquelle Semaphore se connecte, celle dont
le secret va dans **Secret**. Le champ est obligatoire pour un nouveau stockage.

## Secret {#secretStorage-dvlsSecret}

Affiché lorsque le stockage est un Devolutions Server et que la source du secret est
**Store in DB**.

Le secret d'application de l'identité d'application de **App Key**. Le champ est obligatoire pour
un nouveau stockage. Lorsque vous ouvrez un stockage existant, il est vide : laissez-le vide pour
conserver le secret enregistré, saisissez-en un nouveau pour le remplacer.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Affiché lorsque le stockage est un Devolutions Server et que la source du secret est **From ENV**
ou **From File**.

Le nom de la variable d'environnement, ou le chemin du fichier, qui contient le secret
d'application. Le champ est obligatoire pour un nouveau stockage.

## Region {#secretStorage-awsRegion}

Affiché lorsque le stockage est un AWS Secrets Manager.

La région AWS dans laquelle se trouvent les secrets, par exemple `us-east-1`. Le champ est
obligatoire.

## Endpoint URL {#secretStorage-awsEndpointURL}

Affiché lorsque le stockage est un AWS Secrets Manager. Optionnel.

L'adresse de l'API Secrets Manager à appeler à la place du point de terminaison par défaut de la
région, par exemple un point de terminaison VPC. Laissez le champ vide pour utiliser le point de
terminaison AWS par défaut.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Affiché lorsque le stockage est un AWS Secrets Manager.

Lorsque l'option est activée, Semaphore n'enregistre aucune clé d'accès et prend les identifiants
AWS dans l'environnement où le serveur s'exécute, comme le rôle IAM de son instance. Les champs de
la clé d'accès ci-dessous sont alors désactivés.

L'activer supprime la clé d'accès enregistrée lors de l'enregistrement. La désactiver à nouveau
exige une nouvelle **Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Affiché lorsque le stockage est un AWS Secrets Manager. Désactivé lorsque
**Use IAM Role / Instance Profile** est activé.

L'ID de la clé d'accès avec laquelle Semaphore se connecte, par exemple `AKIAIOSFODNN7EXAMPLE`. Le
champ est obligatoire lorsque le rôle IAM est désactivé. L'ID n'est pas un secret : il est
enregistré et affiché tel quel.

## Secret Access Key {#secretStorage-awsSecretKey}

Affiché lorsque le stockage est un AWS Secrets Manager et que la source du secret est
**Store in DB**. Désactivé lorsque **Use IAM Role / Instance Profile** est activé.

Le secret de la clé d'accès. Le champ est obligatoire pour un nouveau stockage et après la
désactivation du rôle IAM. Sinon, il est vide lorsque vous ouvrez le stockage : laissez-le vide
pour conserver la clé enregistrée, saisissez-en une nouvelle pour la remplacer.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Affiché lorsque le stockage est un AWS Secrets Manager et que la source du secret est **From ENV**
ou **From File**. Désactivé lorsque **Use IAM Role / Instance Profile** est activé.

Le nom de la variable d'environnement, ou le chemin du fichier, qui contient la clé d'accès
secrète. Le champ est obligatoire dans les mêmes cas que **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Affiché lorsque le stockage est un Azure Key Vault.

L'adresse du coffre de clés, par exemple `https://my-vault.vault.azure.net`. Le champ est
obligatoire.

## Tenant ID {#secretStorage-azureTenantId}

Affiché lorsque le stockage est un Azure Key Vault.

L'ID du locataire Microsoft Entra de l'application sous laquelle Semaphore se connecte. Le champ
est obligatoire.

## Client ID {#secretStorage-azureClientId}

Affiché lorsque le stockage est un Azure Key Vault.

L'ID d'application (client) de cette application. Elle doit avoir accès aux secrets du coffre de
clés. Le champ est obligatoire.

## Client Secret {#secretStorage-azureClientSecret}

Affiché lorsque le stockage est un Azure Key Vault et que la source du secret est **Store in DB**.

Le secret client de l'application. Le champ est obligatoire pour un nouveau stockage. Lorsque vous
ouvrez un stockage existant, il est vide : laissez-le vide pour conserver le secret enregistré,
saisissez-en un nouveau pour le remplacer.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Affiché lorsque le stockage est un Azure Key Vault et que la source du secret est **From ENV** ou
**From File**.

Le nom de la variable d'environnement, ou le chemin du fichier, qui contient le secret client. Le
champ est obligatoire pour un nouveau stockage.

## Read only {#secretStorage-readonly}

Lorsque l'option est activée, Semaphore ne fait que lire dans le stockage. Un secret d'un groupe de
variables ne peut pas y être créé ni modifié : son enregistrement échoue. Supprimer un secret, une
clé ou un groupe de variables dans Semaphore laisse son secret dans le stockage.

Lorsqu'elle est désactivée, Semaphore écrit dans le stockage les secrets des groupes de variables
qui l'utilisent et les en supprime avec les groupes.

## Sync keys enabled {#secretStorage-syncEnabled}

Activez-la pour importer des secrets du stockage dans le Magasin de clés du projet, depuis les
chemins définis sous **Sync paths**. L'import s'exécute lorsque vous appuyez sur le bouton de
synchronisation du stockage dans l'onglet **Storages**, et automatiquement à l'intervalle défini
dans **Sync paths**.

Une clé importée est marquée comme synchronisée : son nom et son type ne peuvent pas être modifiés,
et la supprimer dans Semaphore laisse le secret dans le stockage. Supprimer le stockage alors que
la synchronisation est activée supprime également les clés qu'il a importées.

## Sync paths {#secretStorage-syncPaths}

Affiché lorsque **Sync keys enabled** est activé.

Ouvre la boîte de dialogue contenant les chemins du stockage depuis lesquels importer et
l'intervalle de l'import automatique. Le nombre sur le bouton est le nombre de chemins. Le bouton
de synchronisation du stockage reste désactivé tant qu'il n'a pas au moins un chemin.

La boîte de dialogue n'a que **Fermer** : ce que vous y modifiez est enregistré avec le stockage,
par **Enregistrer** de cette boîte de dialogue (**Créer** pour un nouveau stockage).
