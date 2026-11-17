---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Paramètres du projet

Les paramètres du projet dans son ensemble : son nom, le nombre de ses tâches pouvant s'exécuter
simultanément et l'envoi ou non d'alertes. La Zone de danger sous le formulaire regroupe les actions
qui concernent tout le projet : sa sauvegarde, son cache et sa suppression. Les modifications du
formulaire prennent effet après **Enregistrer**.

## Nom du projet {#newProject-name}

Le nom sous lequel le projet apparaît dans le menu des projets en haut de la barre latérale. Le
champ est obligatoire.

## Nombre maximum de tâches parallèles {#newProject-maxParallelTasks}

Le nombre de tâches de ce projet pouvant s'exécuter en même temps. Une tâche lancée au-delà de la
limite attend dans la file d'attente jusqu'à ce que l'une des tâches en cours se termine.

Laissez le champ vide ou indiquez `0` pour ne pas fixer de limite propre au projet. La limite du
serveur entier, définie par son administrateur, s'applique dans tous les cas.

## ID de chat Telegram {#newProject-tg}

Optionnel. Le chat Telegram qui reçoit les alertes de ce projet. Lorsque le champ est vide, les
alertes sont envoyées au chat défini pour l'ensemble du serveur.

Le champ n'a d'effet que si l'administrateur a activé les alertes Telegram sur le serveur et que les
alertes sont autorisées pour ce projet.

## Autoriser les alertes pour ce projet {#newProject-alert}

Active les alertes pour les tâches de ce projet. La destination des alertes est définie pour
l'ensemble du serveur par son administrateur : e-mail, Telegram, Slack, Rocket.Chat, Microsoft
Teams, DingTalk ou Gotify.

- Un e-mail est envoyé lorsqu'une tâche échoue, aux membres du projet qui peuvent voir la tâche et
  qui ont activé les alertes dans leur compte.
- Un message est envoyé aux messageries lorsqu'une tâche réussit, échoue ou attend une confirmation.

Un modèle de tâche peut empêcher ses tâches réussies ou échouées d'envoyer des alertes.

## Test Alerts {#settings-testAlerts}

Envoie une alerte de test par chaque canal activé sur le serveur, afin que vous puissiez vérifier
que les alertes de ce projet arrivent bien.

Le bouton ne fonctionne que si **Autoriser les alertes pour ce projet** est coché et enregistré.
Cochez-le et appuyez d'abord sur **Enregistrer**.

## Sauvegarde du projet {#settings-exportProject}

Télécharge un fichier JSON contenant la configuration du projet : modèles de tâches, inventaires,
dépôts, groupes de variables, clés, calendriers, intégrations et les autres éléments qui y sont
configurés. Le fichier peut être retransformé en projet avec **Restaurer le projet** dans le menu
des projets en haut de la barre latérale.

Le fichier omet les secrets des clés et les secrets de l'onglet **Secrets** des groupes de
variables : ils doivent être saisis à nouveau après une restauration. Tout le reste est copié tel
quel, y compris un mot de passe saisi directement dans un autre champ : l'adresse d'un dépôt
contenant un mot de passe, le texte d'un inventaire statique, les valeurs des variables ordinaires
et des variables d'environnement. Vérifiez que le fichier ne contient pas de telles valeurs avant de
le partager. L'historique des tâches ne fait pas partie de la sauvegarde.

## Effacer le cache {#settings-clearCache}

Supprime les fichiers que Semaphore conserve sur le serveur pour ce projet, comme ses copies des
dépôts. Ils sont téléchargés à nouveau au démarrage de la tâche suivante.

Utilisez-le lorsqu'une tâche échoue à cause d'une copie de dépôt obsolète ou endommagée. L'action
est irréversible, et elle ne modifie pas les paramètres du projet.

## Supprimer le projet {#settings-deleteProject}

Supprime le projet avec tout son contenu : modèles de tâches, historique des tâches, inventaires,
dépôts, groupes de variables, clés et calendriers. Semaphore demande d'abord une confirmation.

Un projet supprimé ne peut pas être récupéré. Téléchargez une sauvegarde avant de supprimer un projet
dont vous pourriez avoir à nouveau besoin.
