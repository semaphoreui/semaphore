---
form: src/components/RestoreProjectForm.vue
---

# Restaurer le projet...

Crée un nouveau projet à partir d'un fichier de sauvegarde réalisé avec **Sauvegarde du projet** sur
la page Paramètres d'un projet. Le nouveau projet reçoit les modèles de tâches, inventaires, dépôts,
groupes de variables, clés, calendriers, intégrations et les autres éléments configurés dans la
sauvegarde, et vous en devenez le propriétaire. Seul un administrateur peut restaurer un projet,
sauf si l'administrateur du serveur permet à tous les utilisateurs de créer des projets.

## Nom du projet {#restoreProject-name}

Le nom du nouveau projet. Le champ est obligatoire, et il remplace le nom inscrit dans le fichier de
sauvegarde.

Le nom doit être différent des noms de tous les projets du serveur : le serveur refuse de restaurer
un projet sous un nom déjà pris.

## Fichier de sauvegarde {#restoreProject-file}

Le fichier JSON de la sauvegarde. Choisissez le fichier avant d'appuyer sur **Restore**.

Une sauvegarde ne contient aucune valeur secrète : après la restauration, Semaphore vous indique
combien de clés sont vides et nécessitent que leurs secrets soient saisis à nouveau. Les secrets des
groupes de variables doivent également être saisis à nouveau. Les membres du projet et l'historique
de ses tâches ne font pas partie de la sauvegarde. Une sauvegarde créée par une version plus récente
de Semaphore est refusée.
