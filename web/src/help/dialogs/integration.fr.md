---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Modifier l’exécution

Une exécution est ce qu’une intégration lance lorsqu’un autre service appelle une URL de webhook de
Semaphore, par exemple lors d’un événement push de GitHub : une tâche d’un modèle de tâche ou un
workflow. L’URL et la vérification de la requête relèvent du point de terminaison de l’intégration,
et non de l’exécution. La page d’une exécution contient ses matchers et les valeurs qu’elle extrait
de la requête.

## Nom {#integrationForm-name}

Le nom sous lequel l'intégration apparaît sur la page Intégrations. Les tâches que l'intégration
lance lui sont attribuées dans le journal d'audit. Le champ est obligatoire.

## Intégration {#integrationForm-groupId}

L’intégration à laquelle appartient l’exécution. L’exécution se lance sur les requêtes qui
atteignent le point de terminaison de cette intégration et passent son autorisation. Le champ est
obligatoire.

## Cible {#integrationForm-targetToggle}

Affiché lorsque le serveur prend en charge les workflows.

Ce que l'intégration lance : **Template** lance une tâche d'un modèle de tâche, **Workflow** lance
une exécution d'un workflow. Passer à **Workflow** supprime les paramètres de la tâche ci-dessous.

## Workflow to run {#integrationForm-workflowId}

Affiché lorsque **Cible** vaut **Workflow**.

Le workflow qu'une requête lance. Chaque valeur que l'intégration extrait de la requête est
transmise au workflow comme réponse à sa variable d'enquête du même nom. Les intégrations de
workflow nécessitent l'édition Pro : sans elle, le champ est désactivé et le serveur refuse la
modification.

## Modèle de tâche à exécuter {#integrationForm-templateId}

Affiché lorsque **Cible** vaut **Template**.

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
