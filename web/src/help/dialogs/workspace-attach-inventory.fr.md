---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Rattache à ce modèle un espace de travail qui existe déjà dans le projet, afin qu'il apparaisse
parmi les espaces de travail du modèle avec son état et ses alias. Un espace de travail appartient
à un seul modèle à la fois. Le rattachement ne fait pas de l'espace de travail celui par défaut :
les tâches continuent de s'exécuter dans l'espace de travail par défaut jusqu'à ce que vous
appuyiez sur **Make default**.

## Workspace {#workspaceSelect-inventoryId}

L'espace de travail à rattacher. Le champ est obligatoire. La liste affiche les espaces de travail
du projet créés pour l'application de ce modèle (Terraform, OpenTofu ou Terragrunt) et rattachés à
aucun modèle ; un espace de travail détaché de son modèle avec **Detach** réapparaît ici.

L'espace de travail conserve son nom, sa clé et son état. Le serveur refuse un espace de travail
qu'un autre modèle a pris entre-temps.
