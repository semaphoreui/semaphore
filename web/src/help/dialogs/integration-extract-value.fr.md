---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Une valeur extraite prélève une partie de la requête entrante, par exemple la branche d'un
événement push, et la transmet à la tâche que l'intégration lance. Chaque valeur de l'intégration
est extraite de chaque requête qui lance une tâche.

## Name {#integrationValueForm-name}

Le nom sous lequel la valeur apparaît sur la page de l'intégration. Il ne change ni ce qui est
extrait ni sa destination. Le champ est obligatoire.

## Source of the Value {#integrationValueForm-valueSource}

La partie de la requête dont la valeur est extraite :

- **Header** : la valeur d'un en-tête de la requête.
- **Body** : le corps de la requête.

## Data Type of Body {#integrationValueForm-bodyDataType}

Affiché lorsque **Source of the Value** vaut **Body**.

La manière dont le corps est lu :

- **JSON** : le corps est analysé comme du JSON, et **Key** est le chemin vers une valeur qu'il
  contient. La valeur garde son type JSON : un nombre, une liste ou un objet parviennent tels quels
  à la tâche.
- **String** : le corps entier est pris comme texte, et **Key** n'est pas utilisé.

## Key {#integrationValueForm-key}

Ce qu'il faut extraire de la requête. Le champ est obligatoire.

- Pour **Header** : le nom de l'en-tête, par exemple `X-GitHub-Event`. La casse n'a pas
  d'importance. Un en-tête absent de la requête donne une valeur vide.
- Pour **Body** en **JSON** : le chemin vers la valeur, ses parties reliées par des points ; un
  élément d'un tableau s'écrit `.[0]`, par exemple `repository.full_name` ou `commits.[0].id`. Un
  chemin absent du corps ne donne aucune valeur, et la tâche garde celle définie dans
  l'intégration.
- Pour **Body** en **String** : non utilisé, mais le formulaire exige une valeur ; saisissez
  n'importe quoi.

## Variable Usage {#integrationValueForm-variableType}

L'endroit où la valeur est placée dans une tâche d'un modèle :

- **Variables** : une variable supplémentaire de la tâche, transmise comme la réponse à une
  variable d'enquête. Elle remplace une réponse ou une variable du groupe de variables portant le
  même nom.
- **Task Params** : un paramètre de la tâche, stocké comme les options de la boîte de dialogue de
  la tâche. Le nom `git_branch` définit la branche de la tâche lorsque le modèle permet de remplacer
  la branche ; un préfixe `refs/heads/` est retiré.

Une intégration qui exécute un workflow ignore l'usage : chaque valeur devient la réponse à la
variable d'enquête du workflow nommée dans **Variable**.

## Variable {#integrationValueForm-variable}

Le nom sous lequel la valeur est transmise, par exemple `branch` pour une variable supplémentaire
que le playbook lit sous la forme `{{ branch }}`. Le champ est obligatoire.
