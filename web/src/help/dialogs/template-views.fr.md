---
form: src/components/EditViewsForm.vue
---

# Modifier les vues

Les vues sont les onglets situés au-dessus de la liste des modèles de tâches. Une vue affiche les
modèles que vous y placez, ou tous les modèles du projet. La boîte de dialogue enregistre chaque
modification immédiatement : il n'y a pas de bouton Enregistrer pour l'ensemble de la boîte de
dialogue. Faites glisser une vue par sa poignée pour changer l'ordre des onglets.

## Titre {#viewsForm-title}

Le nom de l'onglet. Saisissez-le et confirmez-le avec la coche à côté du champ ; la croix rétablit
le titre enregistré. Une nouvelle vue est enregistrée dès qu'elle a un titre et que vous le
confirmez ou modifiez l'un de ses réglages. Fermer la boîte de dialogue n'enregistre pas un titre
que vous n'avez pas confirmé.

L'engrenage ouvre les réglages de la vue sous son titre.

## Supprimer {#viewsForm-viewDelete}

L'icône de corbeille supprime la vue immédiatement, sans confirmation. Les modèles de la vue ne
sont pas supprimés : ils restent dans le projet et ne figurent plus dans aucune vue.

## Type {#viewsForm-type}

Les modèles que l'onglet affiche :

- **Custom** : les modèles dont le champ **Vue**, dans la boîte de dialogue du modèle, désigne
  cette vue. C'est la valeur par défaut.
- **All** : tous les modèles du projet, quelle que soit leur vue.

Un membre ne voit que les modèles qu'il a le droit de voir, dans n'importe quelle vue.

## Sort by {#viewsForm-sortColumn}

La colonne selon laquelle les modèles de l'onglet sont triés ; **Name** est la seule. Un onglet
sans tri est également trié par nom ; le champ n'a donc d'intérêt qu'avec **Reverse**.

## Reverse {#viewsForm-sortReverse}

Trie l'onglet par ordre décroissant. Ne fonctionne que si **Sort by** est défini.

## Hidden {#viewsForm-hidden}

Masque l'onglet aux membres qui ne peuvent pas gérer les ressources du projet. Ceux qui le peuvent
le voient toujours, barré. Masquer un onglet ne masque pas ses modèles : une vue de type **All**
les liste toujours.

## Ajouter une vue {#viewsForm-add}

Ajoute une vue à la fin de la liste, avec un titre vide à saisir.
