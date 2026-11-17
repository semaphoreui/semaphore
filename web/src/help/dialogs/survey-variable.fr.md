---
form: src/components/SurveyVars.vue
---

# Variable d'enquête

Une variable d'enquête est une question que la boîte de dialogue de la tâche pose chaque fois que
vous exécutez le modèle. La réponse parvient à la tâche sous forme de variable : une variable
supplémentaire pour Ansible, un `-var` pour Terraform et OpenTofu, un argument pour un script, ou
une variable d'environnement. La variable est enregistrée avec le modèle, lorsque vous enregistrez
la boîte de dialogue du modèle. La même boîte de dialogue modifie l'enquête d'un workflow, où le
type **Secret** n'est pas proposé.

## Nom {#surveyVarDialog-name}

Le nom sous lequel la tâche reçoit la réponse : le nom de la variable supplémentaire, de la
variable Terraform, de l'argument `name=value` d'un script, ou de la variable d'environnement. Les
espaces aux extrémités sont supprimés. Le champ est obligatoire.

Utilisez un nom qu'aucune autre variable de l'enquête ne porte, et que l'outil accepte, par exemple
`app_version`. Une réponse remplace une variable du même nom provenant du groupe de variables du
modèle.

## Titre {#surveyVarDialog-title}

Le libellé du champ dans la boîte de dialogue de la tâche, par exemple `Version to deploy`. C'est
aussi le texte de la puce de la variable dans la boîte de dialogue du modèle. Le champ est
obligatoire.

## Description {#surveyVarDialog-description}

Optionnel. Une indication affichée sous le champ de la boîte de dialogue de la tâche lorsqu'il a le
focus : ce qu'il faut saisir, ou un exemple.

## Type {#surveyVarDialog-type}

Le genre de réponse que demande la boîte de dialogue de la tâche :

- **String** : une ligne de texte. C'est la valeur par défaut.
- **Integer** : des chiffres uniquement ; la boîte de dialogue de la tâche refuse tout le reste, y
  compris un signe moins. La réponse parvient tout de même à la tâche sous forme de texte :
  convertissez-la dans le playbook, par exemple avec `| int`.
- **Secret** : un texte que la boîte de dialogue de la tâche masque. La réponse n'est pas
  enregistrée avec la tâche : une nouvelle exécution la demande donc à nouveau, et Ansible la reçoit
  dans un fichier chiffré par vault. Un secret n'a pas de valeur par défaut.
- **Enum** : une valeur choisie dans une liste que vous définissez ci-dessous.
- **Text** : plusieurs lignes de texte.
- **Select** : plusieurs valeurs choisies dans une liste que vous définissez ci-dessous. La tâche
  les reçoit sous forme de liste ; dans un argument, un `-var` ou une variable d'environnement, la
  liste est écrite en JSON, par exemple `["a","b"]`.

Changer le type efface la valeur par défaut, et un type sans liste efface les valeurs.

## Pass variable as {#surveyVarDialog-target}

La manière dont la réponse parvient à la tâche :

- **Extra variable** : la manière habituelle de l'outil. Ansible la reçoit comme variable
  supplémentaire, Terraform et OpenTofu comme `-var name=value`, un script Bash, PowerShell ou
  Python comme argument `name=value` après le script.
- **Environment variable** : une variable d'environnement du processus, `name=value`, pour tous
  les genres de modèles. La variable n'est alors pas transmise de l'autre manière.

## Clé {#surveyVarDialog-valueName}

Affiché lorsque le type est **Enum** ou **Select**.

Le texte que la boîte de dialogue de la tâche affiche pour cette valeur dans la liste. Chaque valeur
a besoin d'une clé, et les clés doivent être différentes : sinon, la boîte de dialogue refuse
d'enregistrer la variable.

## Valeur {#surveyVarDialog-valueValue}

Affiché lorsque le type est **Enum** ou **Select**.

Ce que la tâche reçoit lorsque cette valeur est choisie. Elle peut différer de sa clé : la clé
`Production` peut transmettre `prod`.

## Add Value {#surveyVarDialog-addValue}

Affiché lorsque le type est **Enum** ou **Select**.

Ajoute une ligne vide à la liste des valeurs. Un **Enum** ou un **Select** a besoin d'au moins une
valeur. L'icône de corbeille d'une ligne la supprime.

## Default value {#surveyVarDialog-defaultValue}

Affiché pour tous les types sauf **Secret**.

La réponse que la boîte de dialogue de la tâche remplit à son ouverture ; vous pouvez la modifier
là. Pour un **Enum**, choisissez l'une des valeurs, pour un **Select**, autant que vous voulez.
Laissez le champ vide pour commencer avec une réponse vide.

La valeur par défaut n'est remplie que par la boîte de dialogue de la tâche. Une tâche lancée d'une
autre manière, par l'API par exemple, ne reçoit pas de valeur par défaut.

## Requis {#surveyVarDialog-required}

La boîte de dialogue de la tâche ne lance pas de tâche tant que la variable n'a pas de réponse.
Comme pour la valeur par défaut, la vérification est faite uniquement par la boîte de dialogue de
la tâche, pas par le serveur.
