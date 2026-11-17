---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Modifier l'intégration

Une intégration lance un modèle de tâche, ou un workflow, lorsqu'un autre service appelle une URL
de webhook de Semaphore, par exemple lors d'un événement push de GitHub. L'URL est un alias,
`/api/integrations/` suivi d'un code aléatoire : les alias du projet se trouvent sur la page
Intégrations, les alias d'une intégration sur sa propre page. Cette page contient aussi les
correspondants et les valeurs que l'intégration extrait de la requête.

## Name {#integrationForm-name}

Le nom sous lequel l'intégration apparaît sur la page Intégrations. Les tâches que l'intégration
lance lui sont attribuées dans le journal d'audit. Le champ est obligatoire.

## Runs {#integrationForm-targetToggle}

Affiché lorsque le serveur prend en charge les workflows.

Ce que l'intégration lance : **Template** lance une tâche d'un modèle de tâche, **Workflow** lance
une exécution d'un workflow. Passer à **Workflow** supprime les paramètres de la tâche ci-dessous.

## Workflow to run {#integrationForm-workflowId}

Affiché lorsque **Runs** vaut **Workflow**.

Le workflow qu'une requête lance. Chaque valeur que l'intégration extrait de la requête est
transmise au workflow comme réponse à sa variable d'enquête du même nom. Les intégrations de
workflow nécessitent l'édition Pro : sans elle, le champ est désactivé et le serveur refuse la
modification.

## Task Template to run {#integrationForm-templateId}

Affiché lorsque **Runs** vaut **Template**.

Le modèle de tâche dont une requête lance une tâche. Le champ est obligatoire. Une fois un modèle
choisi, les paramètres de ses tâches apparaissent en dessous.

## Paramètres de la tâche {#integrationForm-taskParams}

Affiché lorsqu'un modèle de tâche est choisi.

Les paramètres que reçoit chaque tâche lancée par cette intégration, comme si vous la lanciez
vous-même depuis le modèle. Seules les options que le modèle permet à une tâche de remplacer
figurent ici. Les valeurs extraites de la requête les remplacent : une valeur de type
**Variables** remplace une réponse du même nom, une valeur de type **Task Params** un paramètre du
même nom.

Lorsque le modèle autorise les arguments CLI dans une tâche, **CLI args** contient des arguments
supplémentaires pour la ligne de commande de la tâche.

## Message {#taskParams-message}

Optionnel. Le message de chaque tâche que l'intégration lance, affiché dans l'historique des
tâches.

## Variables d'enquête {#taskParams-survey}

Affiché lorsque le modèle possède des variables d'enquête.

Les réponses à l'enquête du modèle, identiques pour chaque tâche. Une valeur de type **Variables**
extraite de la requête remplace la réponse du même nom.

Les réponses aux variables de type secret ne sont pas enregistrées avec l'intégration : transmettez
plutôt un secret par le groupe de variables du modèle.

## Branche {#taskParams-gitBranch}

Affiché lorsque le modèle permet de remplacer la branche dans une tâche.

La branche ou l'étiquette Git du dépôt que les tâches extraient. Laissez la valeur telle quelle pour
utiliser la branche du modèle. Une valeur nommée `git_branch` de type **Task Params**, extraite de
la requête, la remplace ; un préfixe `refs/heads/` est retiré de cette valeur.

## Inventaire {#taskParams-inventoryId}

Affiché lorsque le modèle permet de remplacer l'inventaire dans une tâche.

L'inventaire, ou pour Terraform l'espace de travail, sur lequel les tâches s'exécutent à la place
de celui du modèle.

## Options Ansible {#taskParams-ansible}

Affiché lorsque le modèle exécute un playbook Ansible.

Les options d'`ansible-playbook` pour chaque tâche, chacune affichée lorsque le modèle l'autorise :
**Limite**, **Étiquettes**, **Ignorer les étiquettes**, **Déboguer**, **Exécution à blanc**
(`--check`), **Diff** et **Skip Galaxy install**.

## Options Terraform {#taskParams-terraform}

Affiché lorsque le modèle exécute Terraform, OpenTofu ou Terragrunt.

- **Plan** : la tâche exécute le plan et s'arrête, rien n'est appliqué.
- **Destroy** : le plan et l'application détruisent les ressources.
- **Auto Approve** : le plan est appliqué sans confirmation, lorsque le modèle autorise
  l'approbation automatique. Sans cette option, une tâche attend que quelqu'un la confirme dans
  Semaphore.
- **Upgrade** et **Reconfigure** : `init` s'exécute avec `-upgrade` et `-reconfigure`.

## Auth method {#integrationForm-authMethod}

La manière dont Semaphore vérifie qu'une requête provient du service attendu. Une requête qui
échoue à la vérification ne lance rien, et le webhook lui répond comme à n'importe quelle autre
requête.

- **None** : aucune vérification. Quiconque connaît l'URL peut lancer la tâche.
- **GitHub Webhooks** : l'en-tête `X-Hub-Signature-256` doit contenir la signature HMAC-SHA256 du
  corps calculée avec le secret, telle que GitHub l'envoie.
- **Bitbucket Webhooks** : la même vérification sur l'en-tête `X-Hub-Signature`, tel que Bitbucket
  l'envoie.
- **Token** : l'en-tête indiqué dans **Auth header** doit contenir le secret lui-même.
- **HMAC** et **HMAC (SHA-512)** : l'en-tête indiqué dans **Auth header** doit contenir le
  HMAC-SHA256 ou le HMAC-SHA512 du corps calculé avec le secret, en hexadécimal minuscule, sans
  préfixe tel que `sha256=`.
- **BasicAuth** : la requête doit porter l'identifiant et le mot de passe de la clé choisie dans
  **Secret d'authentification** dans son en-tête `Authorization`.

## Auth header {#integrationForm-authHeader}

Affiché lorsque **Auth method** vaut **Token**, **HMAC** ou **HMAC (SHA-512)**.

Le nom de l'en-tête de la requête qui porte le jeton ou la signature, par exemple
`X-Webhook-Token`. La casse du nom n'a pas d'importance. Une requête sans cet en-tête est refusée.

## Secret d'authentification {#integrationForm-authSecretId}

Affiché lorsque **Auth method** ne vaut pas **None**.

La clé du Magasin de clés à laquelle la requête est comparée. Pour **BasicAuth**, seule une clé de
type **Connexion avec mot de passe** convient. Pour les autres méthodes, une clé de type **Chaîne**
convient aussi : la vérification utilise sa valeur, ou le mot de passe d'une clé **Connexion avec
mot de passe**.

Le champ est nécessaire pour toutes les méthodes sauf **None** : sans clé, ou avec un secret vide,
toute requête est refusée.
