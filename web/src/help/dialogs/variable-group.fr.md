---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Modifier le groupe de variables

Un groupe de variables contient les variables supplémentaires, les variables d'environnement et
les secrets que Semaphore transmet à chaque tâche d'un modèle de tâche qui utilise le groupe.
Ansible reçoit les variables supplémentaires sous forme de `--extra-vars`, Terraform et OpenTofu
sous forme de `-var` ; les variables d'environnement sont définies pour le processus de la tâche.
Un modèle peut utiliser plusieurs groupes : lorsque deux d'entre eux définissent le même nom, le
groupe qui vient en dernier dans le modèle l'emporte.

## Nom du groupe {#varGroup-name}

Le nom sous lequel le groupe apparaît sur la page Groupes de variables et dans le sélecteur de
groupes de variables d'un modèle de tâche. Le champ est obligatoire.

## Secret storage {#varGroup-secretStorageId}

Affiché lorsque le groupe est nouveau et que le serveur prend en charge les stockages de secrets.
Optionnel.

Le stockage de secrets dans lequel les secrets de ce groupe sont conservés, par exemple un
HashiCorp Vault. Laissez le champ vide pour conserver les secrets dans la base de données de
Semaphore, chiffrés.

Le stockage est choisi une seule fois, à la création du groupe : un groupe existant conserve son
stockage, et le champ n'est plus affiché. Un stockage marqué **Read only** n'accepte pas de
nouveaux secrets : l'enregistrement d'un secret dans ce stockage échoue.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Affiché lorsque le groupe est nouveau et que le serveur prend en charge les stockages de secrets.
Disponible dès qu'un **Secret storage** est choisi.

Le début du chemin que reçoit chaque secret de ce groupe dans le stockage. Semaphore y ajoute 10
caractères aléatoires : avec le préfixe `semaphore/prod/`, un secret est enregistré sous
`semaphore/prod/` suivi de ces caractères. Le préfixe est pris tel quel ; terminez-le donc par `/`
lorsque les secrets doivent se trouver dessous comme dans un dossier.

Le préfixe est défini une seule fois, à la création du groupe. L'onglet **Secrets** l'affiche comme
modèle du chemin source.

## Table, JSON ou YAML {#varGroup-extraVarsMode}

La façon dont vous modifiez les variables supplémentaires :

- **Table** : une ligne par variable, chacune avec un type. Ajoutez une ligne avec le bouton **+**
  à côté du sélecteur.
- **JSON** : toutes les variables sous la forme d'un seul objet JSON.
- **YAML** : toutes les variables sous la forme d'un seul mappage YAML.

La boîte de dialogue s'ouvre en mode **Table**. Changer de mode convertit ce que vous avez saisi ;
lorsque le contenu ne peut pas être lu, la boîte de dialogue reste dans le mode où vous étiez et
affiche l'erreur. Quel que soit le mode, les variables sont enregistrées sous la forme d'un seul
objet de noms et de valeurs.

## Variables supplémentaires en JSON {#varGroup-extraVarsJson}

Affiché lorsque les variables supplémentaires sont modifiées en **JSON**.

Les variables supplémentaires sous la forme d'un seul objet JSON. Un nom ne peut pas être vide ;
une valeur peut être une chaîne, un nombre, une liste ou un objet :

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Variables supplémentaires en YAML {#varGroup-extraVarsYaml}

Affiché lorsque les variables supplémentaires sont modifiées en **YAML**.

Les variables supplémentaires sous la forme d'un seul mappage YAML. Le texte est lu avec les types
que connaît JSON : une date sans guillemets comme `2024-01-01` reste une chaîne, et `.inf` ou
`.nan` sont refusés. Une liste ou un scalaire au premier niveau est également refusé.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Nom de la variable {#varGroup-varName}

Affiché lorsque les variables supplémentaires sont modifiées en **Table**, dans chaque ligne.

Le nom de la variable tel que le playbook ou le code Terraform le voit, par exemple
`app_version`.

## Type de la variable {#varGroup-varType}

Affiché lorsque les variables supplémentaires sont modifiées en **Table**, dans chaque ligne.

La façon dont la valeur de la ligne est lue :

- **String** : le texte tel quel.
- **Number** : un nombre, par exemple `42`.
- **List** : un tableau JSON, par exemple `["a", "b"]`.
- **Dict** : un objet JSON, par exemple `{"key": "value"}`.

Une valeur qui ne correspond pas à son type bloque l'enregistrement en indiquant le nom de la
variable.

## Valeur de la variable {#varGroup-varValue}

Affiché lorsque les variables supplémentaires sont modifiées en **Table**, dans chaque ligne.

La valeur de la variable, au format de son type. Pour une **List** ou un **Dict**, le bouton à la
fin du champ ouvre un éditeur plus grand.

Semaphore transmet les valeurs à Ansible en tant que données : une expression `{{ … }}` dans une
valeur n'est pas évaluée, le playbook reçoit le texte tel quel.

## Nom de la variable d'environnement {#varGroup-envName}

Dans l'onglet **Variables**, sous **Variables d'environnement**, dans chaque ligne. Ajoutez une
ligne avec le bouton **+** de la section.

Le nom d'une variable définie dans l'environnement du processus qui exécute la tâche, par exemple
`AWS_REGION`.

## Valeur de la variable d'environnement {#varGroup-envValue}

Dans l'onglet **Variables**, sous **Variables d'environnement**, dans chaque ligne.

La valeur de la variable, sous forme de texte. Elle est enregistrée en clair et affichée à toute
personne qui ouvre le groupe : placez plutôt les mots de passe et les jetons dans l'onglet
**Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Affiché dans l'onglet **Secrets** d'un groupe qui possède un stockage de secrets.

Activez-le pour importer des secrets du stockage dans ce groupe, depuis les chemins définis sous
**Sync paths**. L'import s'exécute lorsque vous appuyez sur le bouton de synchronisation du groupe
sur la page Groupes de variables, et automatiquement à l'intervalle défini dans **Sync paths**.
Lorsqu'il est désactivé, rien n'est importé.

## Sync paths {#varGroup-syncPaths}

Affiché lorsque **Sync keys enabled** est activé.

Ouvre la boîte de dialogue contenant les chemins du stockage depuis lesquels importer et
l'intervalle de l'import automatique. Le nombre sur le bouton est le nombre de chemins. Le bouton
de synchronisation du groupe sur la page Groupes de variables reste désactivé tant que le groupe
n'a pas au moins un chemin.

La boîte de dialogue n'a que **Fermer** : ce que vous y modifiez est enregistré avec le groupe, par
**Enregistrer** de cette boîte de dialogue.

## Nom de la variable secrète {#varGroup-secretVarName}

Dans l'onglet **Secrets**, sous **Variables supplémentaires**, dans chaque ligne. Ajoutez une ligne
avec le bouton **+** de la section.

Le nom d'une variable supplémentaire dont la valeur est un secret. La tâche la reçoit comme elle
reçoit les variables supplémentaires de l'onglet **Variables** ; un secret l'emporte sur une
variable supplémentaire du même nom. Ansible reçoit les secrets dans un fichier chiffré avec
Ansible Vault, Terraform et OpenTofu sous forme de `-var`. Comme l'avertit la boîte de dialogue, un
secret transmis de cette façon peut tout de même apparaître en clair dans le journal d'Ansible.

## Valeur de la variable secrète {#varGroup-secretVarValue}

Dans l'onglet **Secrets**, sous **Variables supplémentaires**, dans chaque ligne.

La valeur du secret. Semaphore la chiffre, ou l'écrit dans le stockage de secrets du groupe, et ne
l'affiche plus jamais : lorsque vous ouvrez le groupe plus tard, le champ est vide. Laissez-le vide
pour conserver la valeur enregistrée, saisissez-en une nouvelle pour la remplacer. Supprimer la
ligne supprime le secret lors de l'enregistrement du groupe.

## Nom de la variable d'environnement secrète {#varGroup-secretEnvName}

Dans l'onglet **Secrets**, sous **Variables d'environnement**, dans chaque ligne. Ajoutez une ligne
avec le bouton **+** de la section.

Le nom d'une variable d'environnement dont la valeur est un secret, par exemple `DB_PASSWORD`. Elle
est définie dans l'environnement du processus qui exécute la tâche, comme le sont les variables
d'environnement de l'onglet **Variables**.

## Valeur de la variable d'environnement secrète {#varGroup-secretEnvValue}

Dans l'onglet **Secrets**, sous **Variables d'environnement**, dans chaque ligne.

La valeur du secret. Comme pour une variable supplémentaire secrète, elle n'est plus jamais
affichée : laissez le champ vide pour conserver la valeur enregistrée, saisissez-en une nouvelle
pour la remplacer.
