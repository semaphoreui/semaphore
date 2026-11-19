---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Un correspondant est une condition portant sur la requête entrante. Les correspondants décident
quelle intégration lance une requête adressée à l'alias du projet, de sorte qu'une seule URL peut
servir plusieurs intégrations : une intégration ne s'exécute que lorsque tous ses correspondants
sont satisfaits. Ils ne sont utilisés que lorsque **Utiliser l'alias du projet** est activé pour
l'intégration ; dans ce cas, l'intégration n'est accessible que par l'alias du projet.

Une intégration trouvée par l'alias du projet sans aucun correspondant ne s'exécute jamais :
ajoutez-en au moins un.

## Name {#integrationMatcherForm-name}

Le nom sous lequel le correspondant apparaît sur la page de l'intégration. Il ne change pas ce que
le correspondant vérifie. Le champ est obligatoire.

## Match on {#integrationMatcherForm-matchType}

La partie de la requête que le correspondant lit :

- **Header** : la valeur d'un en-tête de la requête.
- **Body** : le corps de la requête.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Affiché lorsque **Match on** vaut **Body**.

La manière dont le corps est lu :

- **JSON** : le corps est analysé comme du JSON, et **Key** est le chemin vers une valeur qu'il
  contient.
- **String** : le corps entier est comparé comme texte, et **Key** n'est pas utilisé.

## Key {#integrationMatcherForm-key}

Ce qu'il faut lire dans la requête. Le champ est obligatoire.

- Pour **Header** : le nom de l'en-tête, par exemple `X-GitHub-Event`. La casse n'a pas
  d'importance.
- Pour **Body** en **JSON** : le chemin vers la valeur, ses parties reliées par des points ; un
  élément d'un tableau s'écrit `.[0]`, par exemple `repository.full_name` ou `commits.[0].id`.
- Pour **Body** en **String** : non utilisé, mais le formulaire exige une valeur ; saisissez
  n'importe quoi.

## Comparison Method {#integrationMatcherForm-method}

La manière dont la valeur lue dans la requête est comparée à **Value** :

- **==** : les deux sont égales.
- **!=** : les deux diffèrent. Un en-tête ou une clé absents de la requête comptent comme
  différents.
- **Contains** : la valeur lue dans la requête contient **Value**.

## Value {#integrationMatcherForm-value}

Le texte auquel comparer, par exemple `push` ou `refs/heads/main`. La comparaison est exacte et
sensible à la casse. Un nombre ou un booléen du corps JSON est comparé sous sa forme texte : `42`,
`true`. Le champ est obligatoire.
