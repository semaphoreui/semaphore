---
form: src/components/ProjectForm.vue
---

# Nouveau projet

Un projet regroupe les modèles de tâches, inventaires, dépôts, groupes de variables et clés d'un
même travail, ainsi que les personnes qui y travaillent. Vous devenez propriétaire du projet que
vous créez. Seul un administrateur peut créer un projet, sauf si l'administrateur du serveur le
permet à tous les utilisateurs. Chaque paramètre ci-dessous peut être modifié plus tard sur la page
Paramètres du projet.

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

## Demo {#newProject-demo}

Remplit le nouveau projet d'exemples pour essayer Semaphore : le dépôt public
`https://github.com/semaphoreui/semaphore-demo.git`, des inventaires, une clé de mot de passe de
coffre et des modèles de tâches pour Ansible, OpenTofu, Terragrunt, Bash et PowerShell, répartis
dans les vues **Build**, **Deploy** et **Tools**.

Laissez l'interrupteur désactivé pour un projet sans exemples : il ne reçoit que la clé **None** et
la vue **All**. Les exemples sont des éléments ordinaires du projet : modifiez-les ou supprimez-les
à votre guise.
