---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nouvelle tâche

Lance une tâche du modèle : une exécution de son playbook, de son code Terraform ou de son script.
Le modèle détermine les champs que cette boîte de dialogue affiche, et chaque section ci-dessous
indique quand son champ est présent. Le bouton en bas est **Exécuter**, **Construire** ou
**Déployer**, selon le type du modèle.

Lorsque le modèle comporte une enquête, ses questions sont des champs placés après **Message**. Un
champ marqué `*` est obligatoire, un champ Integer n'accepte que des chiffres, et les réponses par
défaut sont pré-remplies. Semaphore transmet les réponses à la tâche sous forme de variables :
variables supplémentaires pour Ansible, `-var` pour Terraform, arguments `name=value` pour un
script, ou variables d'environnement pour une question configurée ainsi. Une réponse remplace une
variable du même nom dans les groupes de variables du modèle. Une réponse secrète est masquée
pendant la saisie et n'est pas conservée avec la tâche.

## Commit {#newTask-commit}

Affiché lorsque vous relancez une tâche précédente et que cette tâche s'est exécutée sur un commit
connu.

La tâche extrait ce commit du dépôt au lieu du commit le plus récent de la branche. Fermez la boîte
pour exécuter le commit le plus récent.

Le commit n'est utilisé que si le modèle permet de remplacer la branche, et jamais pour un dépôt
marqué comme monorepo. Sinon, la tâche exécute le commit le plus récent.

## Version de construction {#newTask-buildTaskId}

Affiché lorsque le type du modèle est **Déployer**.

La construction à déployer : l'une des 20 dernières tâches réussies du modèle de construction,
listées par version et par message. La plus récente est sélectionnée. Le champ est obligatoire.

La version de la construction parvient à la tâche : sous la forme
`semaphore_vars.task_details.incoming_version` dans Ansible, et de la variable d'environnement
`SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` dans un script. La liste est vide lorsque vous n'avez pas
le droit de voir le modèle de construction.

## Message {#newTask-message}

Optionnel. Une note expliquant pourquoi la tâche s'exécute, par exemple `Hotfix for the login page`.
Elle est affichée avec la tâche dans l'historique des tâches et ne change rien à l'exécution.

Le playbook peut la lire dans `semaphore_vars.task_details.message`, et un script dans la variable
d'environnement `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Branche {#newTask-gitBranch}

Affiché lorsque le modèle permet de remplacer la branche, sauf si le dépôt est un monorepo.

La branche ou le tag du dépôt que la tâche extrait. Le champ contient au départ la branche du
modèle. Lorsqu'il est vide, la tâche utilise la branche du modèle, ou la branche du dépôt si le
modèle n'en a pas. Un nom qui n'est pas un nom de branche Git valide est refusé.

## Inventaire {#newTask-inventoryId}

Affiché lorsque le modèle permet de choisir l'inventaire dans une tâche et que vous avez le droit de
voir le projet. Pour un modèle Terraform, OpenTofu ou Terragrunt, le champ s'intitule **Workspace**,
mais le serveur n'utilise le choix que pour les modèles Ansible : une tâche Terraform, OpenTofu ou
Terragrunt s'exécute dans l'espace de travail du modèle, quoi que vous choisissiez ici.

L'inventaire sur lequel la tâche s'exécute. Le champ contient au départ l'inventaire du modèle et
liste les inventaires du projet créés pour l'application du modèle. Si l'inventaire choisi a
disparu au moment où la tâche démarre, la tâche s'exécute sur l'inventaire du modèle.

## Limite {#taskParams-limit}

Affiché pour un modèle Ansible qui permet de remplacer la limite.

Les hôtes ou groupes de l'inventaire sur lesquels le playbook s'exécute, un motif par puce, par
exemple `web` ou `10.0.0.11`. Semaphore les transmet avec `--limit`, séparés par des virgules. La
liste contient au départ la limite du modèle. Supprimez toutes les puces pour exécuter sur tous les
hôtes de l'inventaire.

## Étiquettes {#taskParams-tags}

Affiché pour un modèle Ansible qui permet de remplacer les étiquettes.

Seuls les tâches et les rôles du playbook portant l'une de ces étiquettes s'exécutent : Semaphore
les transmet avec `--tags`, une étiquette par puce. La liste contient au départ les étiquettes du
modèle. Supprimez toutes les puces pour exécuter le playbook entier.

## Ignorer les étiquettes {#taskParams-skipTags}

Affiché pour un modèle Ansible qui permet de remplacer les étiquettes ignorées.

Les tâches et les rôles du playbook portant l'une de ces étiquettes ne s'exécutent pas : Semaphore
les transmet avec `--skip-tags`, une étiquette par puce. La liste contient au départ les
étiquettes ignorées du modèle.

## Déboguer {#taskParams-debug}

Affiché pour un modèle Ansible qui autorise le débogage.

Exécute `ansible-playbook` avec une option de verbosité, afin que le journal de la tâche affiche
plus de détails. Le libellé affiche l'option : `-v` autant de fois que l'indique **Niveau de
débogage**.

## Niveau de débogage {#taskParams-debugLevel}

Affiché avec **Déboguer**, et activé lorsque **Déboguer** est coché.

Le niveau de détail du journal, de 1 (`-v`) à 6 (`-vvvvvv`). La valeur par défaut est 4
(`-vvvv`), qui affiche aussi les détails des connexions aux hôtes.

## Exécution à blanc {#taskParams-dryRun}

Affiché pour un modèle Ansible, sauf si le modèle le masque.

Exécute le playbook avec `--check` : Ansible indique ce qu'il modifierait sur les hôtes et ne
modifie rien.

## Diff {#taskParams-diff}

Affiché pour un modèle Ansible, sauf si le modèle le masque.

Exécute le playbook avec `--diff` : le journal affiche les modifications apportées aux fichiers et
aux templates sur les hôtes. Cochez aussi **Exécution à blanc** pour voir les modifications sans
les effectuer.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Affiché pour un modèle Ansible qui permet de remplacer l'installation Galaxy.

Avant l'exécution du playbook, Semaphore installe les rôles et les collections listés dans les
fichiers `requirements.yml` du dépôt, lorsqu'un fichier a changé depuis la dernière installation.
Cochez la case pour ignorer cette étape pour cette tâche.

La case est décochée au départ quel que soit le réglage du modèle, et une case décochée installe
les dépendances.

## Plan {#taskParams-plan}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Exécute `plan` et termine la tâche à ce stade : rien n'est appliqué. Utilisez-le pour examiner
d'abord les modifications. Sans cette option, la tâche continue et applique le plan lorsqu'il
comporte des modifications.

## Destroy {#taskParams-destroy}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Ajoute `-destroy` à `plan` et `apply` : la tâche planifie, puis applique, la suppression de toutes
les ressources de l'état de l'espace de travail. Cochez aussi **Plan** pour voir ce qui serait
détruit sans le détruire.

## Auto Approve {#taskParams-autoApprove}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Applique le plan dès qu'il est établi. Sans cette option, la tâche s'arrête après le plan et
attend que quelqu'un le confirme ou le rejette sur la page de la tâche.

La case n'a d'effet que si le modèle autorise l'approbation automatique. Un modèle réglé sur
l'approbation automatique applique toujours le plan sans rien demander, quel que soit l'état de la
case.

## Upgrade {#taskParams-upgrade}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Exécute `init` avec `-upgrade` : les fournisseurs et les modules sont mis à niveau vers les
versions les plus récentes que permettent leurs contraintes de version.

## Reconfigure {#taskParams-reconfigure}

Affiché pour un modèle Terraform, OpenTofu ou Terragrunt.

Exécute `init` avec `-reconfigure` : le backend est configuré à nouveau, et la configuration
enregistrée du backend est ignorée. Sans cette option, Semaphore exécute `init -migrate-state`, qui
déplace l'état existant vers un backend modifié.

## CLI args {#newTask-args}

Affiché lorsque le modèle permet de remplacer les arguments CLI.

Des arguments supplémentaires pour la ligne de commande de la tâche, un par puce : `--forks=10` est
une puce, `--forks` et `10` en font deux. Cliquez sur une puce pour la modifier, faites-la glisser
pour changer l'ordre, et utilisez la dernière puce pour en ajouter une.

La liste contient au départ les arguments du modèle. Semaphore transmet toujours les arguments du
modèle et ajoute cette liste après eux ; dès que vous modifiez la liste, les arguments du modèle
qui y restent sont donc transmis deux fois. Supprimez-les et ne gardez que les arguments que vous
ajoutez.
