---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Modifier le modèle 'Ansible Playbook'

Un modèle de tâche indique ce que Semaphore exécute et de quelle manière : le dépôt et le fichier
qu'il contient, l'inventaire et les variables, ainsi que les options qu'un utilisateur peut
modifier au lancement d'une tâche. Chaque tâche est lancée à partir d'un modèle. Les champs de la
boîte de dialogue dépendent de l'application du modèle (Ansible, Terraform, OpenTofu, Terragrunt,
Bash, Python, PowerShell et d'autres) : un champ que seules certaines applications possèdent le
précise. Les modifications s'appliquent aux tâches lancées après **Enregistrer**.

## Type de modèle {#templateForm-typeTabs}

Les onglets en haut définissent le genre de modèle :

- **Tâche** : exécute le code. La plupart des modèles sont de ce genre.
- **Construire** : produit un artefact. Chaque tâche de construction reçoit une version, comptée à
  partir de **Version de départ**.
- **Déployer** : déploie un artefact produit par une construction. Au lancement, vous choisissez
  la tâche de construction dont la version est déployée.

Une tâche de construction ou de déploiement reçoit dans son environnement les variables
`SEMAPHORE_TASK_TYPE`, `SEMAPHORE_TASK_TARGET_VERSION` (construction) et
`SEMAPHORE_TASK_INCOMING_VERSION` (déploiement). Ansible reçoit les mêmes valeurs dans la variable
supplémentaire `semaphore_vars.task_details`.

## Version de départ {#templateForm-startVersion}

Affiché lorsque le type est **Construire**. Obligatoire.

La version de la première tâche de construction, par exemple `1.0.0` ou `v0.1`. Chaque
construction suivante augmente de un le dernier nombre de la version précédente et conserve le
texte qui l'entoure ainsi que ses zéros de remplissage : `1.0.0`, `1.0.1`, `1.0.2`. Indiquez ici un
nombre supérieur à celui de la dernière construction pour faire un saut, par exemple à `2.0.0` ;
une version qui ne contient aucun nombre ne change jamais.

## Modèle de construction {#templateForm-buildTemplateId}

Affiché lorsque le type est **Déployer**. Obligatoire.

Le modèle dont les tâches produisent les artefacts que ce modèle déploie. La liste contient les
modèles de construction du projet et, en dessous, les modèles de déploiement : un déploiement peut
suivre un autre déploiement, et il reçoit alors la version que ce déploiement a reçue de sa
construction.

## Exécution automatique {#templateForm-autorun}

Affiché lorsque le type est **Déployer**.

Lance de lui-même une tâche de ce modèle chaque fois qu'une tâche du **Modèle de construction**
réussit, avec la version que cette tâche a produite. Une construction échouée ne lance rien.

## Nom {#templateForm-name}

Le nom sous lequel le modèle apparaît sur la page Modèles de tâches. Le champ est obligatoire.

## Dépôt {#templateForm-repositoryId}

Le dépôt du projet qui contient le code à exécuter. Le champ est obligatoire. Avant chaque tâche,
Semaphore met à jour sa copie du dépôt et exécute le code depuis celle-ci.

L'étiquette à droite du champ affiche la branche que les tâches extraient. Cliquez dessus pour
ouvrir le champ **Branche**.

## Branche {#templateForm-gitBranch}

Affiché après un clic sur l'étiquette de branche à droite de **Dépôt**.

La branche ou le tag que les tâches de ce modèle extraient, à la place de la branche définie sur
le dépôt. Laissez le champ vide pour utiliser la branche du dépôt. Lorsque Semaphore peut lire les
branches du dépôt, le champ les propose.

Le champ est désactivé lorsque le dépôt est un monorepo : tous les modèles partagent une seule
copie de celui-ci, sur la branche du dépôt.

## Chemin vers le fichier playbook {#templateForm-playbook}

Ce qu'il faut exécuter, sous forme de chemin depuis la racine du dépôt, par exemple
`deploy/site.yml`. Le chemin ne peut pas être absolu ni sortir du dépôt. La liste propose les
fichiers du dépôt ; vous pouvez aussi saisir un chemin absent de la liste. Le bouton à droite du
champ relit la liste.

Le libellé et la signification changent selon l'application :

- **Ansible** : le playbook. Il est obligatoire ; la liste propose les fichiers `.yml` et `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt** : **Subdirectory path**, le répertoire qui contient
  le code. Laissez-le vide pour exécuter le code à la racine du dépôt. Saisissez `/` pour lister le
  contenu d'un répertoire.
- **Bash**, **Python**, **PowerShell** et les autres applications : **Script Filename**, le script
  à exécuter. Il est obligatoire. Le script s'exécute à la racine du dépôt.

## Different working directory {#templateForm-workingDirectoryToggle}

Affiché pour les modèles Ansible.

Exécute `ansible-playbook` dans un répertoire autre que la racine du dépôt. Cochez la case pour
afficher le champ **Working directory**. Décochez-la pour revenir à la racine.

## Working directory {#templateForm-workingDirectory}

Affiché lorsque **Different working directory** est coché. Obligatoire.

Le répertoire dans lequel `ansible-playbook` et `ansible-galaxy` s'exécutent, sous forme de chemin
depuis la racine du dépôt, par exemple `ansible`. Utilisez-le lorsque le fichier `ansible.cfg` du
projet se trouve dans un sous-répertoire, car Ansible lit ce fichier dans le répertoire courant. Le
chemin du playbook reste relatif à la racine du dépôt. Le chemin ne peut pas être absolu ni sortir
du dépôt.

## Inventaire {#templateForm-inventoryId}

Affiché pour les modèles Ansible, Terraform, OpenTofu et Terragrunt.

- **Ansible** : l'inventaire sur lequel le playbook s'exécute, avec les identifiants des hôtes. Le
  champ est obligatoire.
- **Terraform**, **OpenTofu**, **Terragrunt** : **Workspace**, l'espace de travail dans lequel le
  code s'exécute. Semaphore le sélectionne avant le plan et le crée s'il n'existe pas. Un nouveau
  modèle enregistré sans espace de travail en reçoit un nommé `default`.

La liste contient les inventaires du projet créés pour l'application du modèle.

## Groupes de variables {#templateForm-environmentIds}

Optionnel. Les groupes de variables dont chaque tâche de ce modèle reçoit les variables, les
variables d'environnement et les secrets. Vous pouvez en choisir plusieurs : ils sont fusionnés
dans l'ordre du champ, et un groupe placé plus loin l'emporte sur un groupe précédent lorsque les
deux définissent le même nom.

La manière dont les variables parviennent au code dépend de l'application : variables
supplémentaires pour Ansible, `-var` pour Terraform et OpenTofu, arguments `name=value` pour un
script. Les variables d'environnement sont définies pour toutes les applications.

## Vue {#templateForm-viewId}

Optionnel. La vue dans laquelle le modèle apparaît : les vues sont les onglets de la page Modèles
de tâches, qui regroupent les modèles d'un projet. Laissez le champ vide pour ne placer le modèle
dans aucune vue.

## Étiquette de l'exécuteur {#templateForm-runnerTag}

Affiché lorsque des exécuteurs de projet sont disponibles.

Les tâches de ce modèle ne s'exécutent que sur un exécuteur portant cette étiquette, même lorsque
le serveur pourrait les exécuter lui-même. L'étiquette l'emporte sur l'étiquette de l'exécuteur de
l'inventaire. Laissez le champ vide pour utiliser l'étiquette de l'inventaire, ou aucune.

## Executor image {#templateForm-executorImage}

L'image de conteneur que l'exécuteur démarre pour les tâches de ce modèle, par exemple
`semaphoreui/job:latest`. Laissez le champ vide pour utiliser l'image de la configuration de
l'exécuteur.

Seuls les exécuteurs Docker et Kubernetes exécutent les tâches dans un conteneur ; le champ n'est
donc activé que lorsque l'un d'eux est disponible. Une tâche qui s'exécute sans conteneur l'ignore.

## Variables d'enquête {#surveyVars}

Les valeurs demandées à l'utilisateur au lancement d'une tâche de ce modèle. Cliquez sur **+**
pour ajouter une variable, sur une puce pour en modifier une, sur la croix d'une puce pour la
supprimer ; faites glisser les puces pour définir l'ordre dans lequel les questions sont posées.

Les réponses parviennent au code comme les variables d'un groupe de variables, ou comme variables
d'environnement, selon ce que chaque variable définit. Elles l'emportent sur un groupe de
variables qui définit le même nom.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Permet aux tâches de ce modèle de s'exécuter en même temps. Sans cette option, une tâche du modèle
attend dans la file d'attente jusqu'à la fin de celle en cours. Les limites de tâches parallèles du
projet et du serveur s'appliquent dans tous les cas.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Affiché lorsque l'administrateur a activé l'émission de JWT sur le serveur.

Remet à chaque tâche de ce modèle un JSON Web Token signé par Semaphore, dans la variable
d'environnement `SEMAPHORE_JWT`. Le code peut présenter ce jeton à un autre service pour prouver de
quelle tâche il s'agit : le jeton nomme la tâche, le projet, le modèle et l'utilisateur qui l'a
lancée. Le service vérifie la signature avec les clés publiques disponibles à
`/.well-known/jwks.json` sur le serveur Semaphore.

## JWT audience {#templateForm-jwtAudience}

Affiché lorsque **Issue JWT to task runner** est coché. Optionnel.

Les services auxquels le jeton est destiné, inscrits dans sa revendication `aud`, par exemple
`https://vault.example.com`. Appuyez sur Entrée après chaque valeur. Jusqu'à 32 valeurs ; laissez
le champ vide pour un jeton sans audience.

## JWT TTL {#templateForm-jwtTtl}

Affiché lorsque **Issue JWT to task runner** est coché. Optionnel.

La durée de validité du jeton à partir du début de la tâche, sous forme de durée telle que `30s`,
`15m` ou `1h`. Laissez le champ vide pour utiliser la valeur par défaut du serveur, une heure sauf
si l'administrateur en a défini une autre. La valeur ne peut pas dépasser le maximum du serveur,
24 heures sauf réglage contraire.

## Je veux exécuter une tâche par cron uniquement pour les nouveaux commits d'un dépôt {#templateForm-cronVisible}

Lance une tâche de ce modèle lorsqu'un nouveau commit apparaît dans un dépôt. Semaphore demande au
dépôt son dernier commit à l'intervalle choisi et ne lance une tâche que si le commit a changé
depuis la dernière vérification.

Cocher la case et enregistrer crée le calendrier de la vérification ; la décocher et enregistrer
supprime ce calendrier. Les autres calendriers du modèle sont conservés.

## Dépôt {#templateForm-cronRepositoryId}

Affiché lorsque **Je veux exécuter une tâche par cron uniquement pour les nouveaux commits d'un
dépôt** est coché. Obligatoire.

Le dépôt à surveiller pour de nouveaux commits. Semaphore surveille la branche définie sur ce
dépôt, et non la **Branche** du modèle.

## intervalleDeVérification {#templateForm-cronFormat}

Affiché lorsque **Je veux exécuter une tâche par cron uniquement pour les nouveaux commits d'un
dépôt** est coché.

La fréquence à laquelle Semaphore vérifie si le dépôt contient un nouveau commit : de toutes les
minutes à une fois par jour. La valeur par défaut est toutes les minutes.

## Supprimer les alertes de succès {#templateForm-suppressSuccessAlerts}

Empêche les tâches réussies de ce modèle d'envoyer des alertes. Les alertes des tâches échouées
sont toujours envoyées. L'option n'a d'effet que si les alertes sont activées pour le projet.

## Supprimer les alertes d'erreur {#templateForm-suppressErrorAlerts}

Empêche les tâches échouées de ce modèle d'envoyer des alertes. Les alertes des tâches réussies
sont toujours envoyées. L'option n'a d'effet que si les alertes sont activées pour le projet.

## CLI args {#templateForm-args}

Des arguments ajoutés à la commande de chaque tâche. Chaque puce est un argument, transmis tel
quel, sans shell : mettez `--forks=10` dans une puce, ou `--forks` et `10` dans deux. Faites
glisser les puces pour changer leur ordre. Le bouton à droite de la liste ouvre les arguments sous
forme de tableau JSON :

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible** : ajoutés à `ansible-playbook`, avant le playbook.
- **Terraform**, **OpenTofu**, **Terragrunt** : ajoutés à `plan` et `apply`, pas à `init`.
- Un script : ajoutés après le nom du script.

## Arguments CLI {#templateForm-allowOverrideArgsInTask}

Sous **Invites**. Permet à l'utilisateur d'ajouter ses propres arguments au lancement d'une tâche.
Ils sont ajoutés après les arguments du modèle.

## Branche {#templateForm-allowOverrideBranchInTask}

Sous **Invites**. Permet à l'utilisateur de choisir une autre branche, un autre tag ou un autre
commit au lancement d'une tâche, à la place de la branche du modèle. Sans cette option, la tâche
extrait toujours la branche du modèle.

La case est désactivée lorsque le dépôt est un monorepo.

## Inventaire {#templateForm-allowOverrideInventory}

Sous **Invites**. Affiché pour les modèles Ansible, Terraform, OpenTofu et Terragrunt.

Permet à l'utilisateur de choisir un autre inventaire au lancement d'une tâche. Seuls les modèles
Ansible utilisent l'inventaire choisi à ce moment : une tâche Terraform, OpenTofu ou Terragrunt
s'exécute dans l'espace de travail du modèle même lorsque la case est cochée.

## Limite {#templateForm-limit}

Affiché pour les modèles Ansible.

Les hôtes ou groupes de l'inventaire sur lesquels exécuter le playbook, transmis avec `--limit`.
Chaque puce est un motif, par exemple `web` ou `db01` ; laissez la liste vide pour exécuter sur
tous les hôtes.

## Étiquettes {#templateForm-tags}

Affiché pour les modèles Ansible.

N'exécute que les tâches et les rôles portant ces étiquettes, transmises avec `--tags`. Laissez la
liste vide pour tout exécuter.

## Ignorer les étiquettes {#templateForm-skipTags}

Affiché pour les modèles Ansible.

Ignore les tâches et les rôles portant ces étiquettes, transmises avec `--skip-tags`.

## Coffres {#templateVaults}

Affiché pour les modèles Ansible.

Les mots de passe Ansible Vault dont le playbook a besoin pour déchiffrer ses fichiers. Cliquez sur
**+** pour en ajouter un, sur une puce pour le modifier, sur la croix d'une puce pour le supprimer.
Chaque coffre a un nom, son identifiant de coffre, et fournit le mot de passe à partir d'une clé du
Magasin de clés ou d'un script client du dépôt.

## Approbation automatique {#templateForm-autoApprove}

Affiché pour les modèles Terraform, OpenTofu et Terragrunt.

Applique le plan sans rien demander. Sans cette option, une tâche s'arrête après le plan et attend
qu'un utilisateur le confirme ou le rejette ; un plan sans modifications termine la tâche sans rien
appliquer.

## Override backend settings {#templateForm-overrideBackend}

Affiché pour les modèles Terraform, OpenTofu et Terragrunt lorsque Semaphore peut conserver l'état
Terraform.

Conserve l'état du code dans Semaphore. Avant `init`, Semaphore écrit dans le répertoire du code un
fichier qui définit le backend `http`, et le supprime après la tâche :

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Affiché avec **Override backend settings** ; activé lorsque cette case est cochée. Optionnel.

Le nom du fichier dans lequel Semaphore écrit le backend, par exemple `semaphore_backend.tf`. La
valeur par défaut est `backend.tf`. Le nom doit se terminer par `.tf` et ne contenir que des
lettres, des chiffres, `_`, `-` et `.`. Choisissez un nom qu'aucun fichier de votre code ne porte :
Semaphore écrase ce fichier et le supprime après la tâche.

## Limite {#templateForm-allowOverrideLimit}

Sous **Invites Ansible**. Permet à l'utilisateur de définir les hôtes sur lesquels exécuter le
playbook au lancement d'une tâche, à la place de la **Limite** du modèle.

## Étiquettes {#templateForm-allowOverrideTags}

Sous **Invites Ansible**. Permet à l'utilisateur de définir les étiquettes au lancement d'une
tâche, à la place des **Étiquettes** du modèle.

## Ignorer les étiquettes {#templateForm-allowOverrideSkipTags}

Sous **Invites Ansible**. Permet à l'utilisateur de définir les étiquettes à ignorer au lancement
d'une tâche, à la place de **Ignorer les étiquettes** du modèle.

## Déboguer {#templateForm-allowDebug}

Sous **Invites Ansible**. Permet à l'utilisateur d'activer une sortie détaillée au lancement d'une
tâche, de `-v` à `-vvvvvv`.

## Exécution à blanc {#templateForm-showDryRun}

Sous **Invites Ansible**. Cochée par défaut. Propose l'option **Exécution à blanc** au lancement
d'une tâche, qui exécute le playbook avec `--check` : Ansible indique ce qu'il modifierait et ne
modifie rien. Décochez la case pour retirer l'option.

## Diff {#templateForm-showDiff}

Sous **Invites Ansible**. Cochée par défaut. Propose l'option **Diff** au lancement d'une tâche,
qui exécute le playbook avec `--diff` pour afficher les modifications apportées aux fichiers.
Décochez la case pour retirer l'option.

## Approbation automatique {#templateForm-allowAutoApprove}

Sous les invites de Terraform, OpenTofu et Terragrunt. Permet à l'utilisateur de cocher **Auto
Approve** au lancement d'une tâche, pour appliquer le plan sans confirmation. La boîte de dialogue
de la tâche propose cette option dans tous les cas, mais sans cette case Semaphore l'ignore : la
tâche attend une confirmation après le plan, sauf si **Approbation automatique** du modèle est
cochée.

## Galaxy install options {#templateForm-galaxy}

Affiché pour les modèles Ansible. Cliquez dessus pour afficher les options de l'installation
Galaxy ; le libellé compte les options qui diffèrent des valeurs par défaut.

Avant l'exécution du playbook, Semaphore installe avec `ansible-galaxy` les rôles et les
collections que demande le dépôt. Il recherche `requirements.yml`, `roles/requirements.yml` et
`collections/requirements.yml` à côté du playbook et à la racine du dépôt, et n'exécute
l'installation que si un fichier a changé depuis la dernière tâche.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Sous **Galaxy install options**. Ignore l'installation Galaxy : le playbook s'exécute avec les
rôles et les collections installés auparavant. Utilisez-le lorsque les dépendances sont installées
d'une autre manière ou que l'installation est lente et inutile.

## Role install args {#templateForm-galaxyRoleArgs}

Sous **Galaxy install options**. Des arguments ajoutés à `ansible-galaxy role install`, un par
puce, par exemple `--ignore-certs` ou `--server=https://galaxy.example.com`. Semaphore n'accepte
que les options de l'installation qui peuvent être transmises sans risque : une option inconnue,
`--token`, `--api-key`, `-p` ou `--roles-path` empêche l'enregistrement du modèle.

## Collection install args {#templateForm-galaxyCollectionArgs}

Sous **Galaxy install options**. Des arguments ajoutés à `ansible-galaxy collection install`, un
par puce, par exemple `--pre` ou `--upgrade`. La même règle s'applique que pour **Role install
args** : `--token`, `--api-key`, `-p` et `--collections-path` ne sont pas acceptés.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Sous **Galaxy install options**, dans **Prompts**. Permet à l'utilisateur de choisir au lancement
d'une tâche s'il faut ignorer l'installation Galaxy. Le choix de la tâche l'emporte sur **Skip
Galaxy install** du modèle.
