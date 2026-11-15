---
form: src/components/AppForm.vue
---

# Edit App

Une application est un outil avec lequel Semaphore exécute des tâches : Ansible, Terraform,
OpenTofu, Terragrunt, Bash, Python, PowerShell ou une commande de votre choix. Chaque modèle de
tâche est conçu pour une application, et cette boîte de dialogue définit la manière dont le serveur
la démarre et dont l'application est affichée dans l'interface.

Les paramètres sont enregistrés dans la base de données et prennent effet immédiatement sur le
serveur qui les enregistre ; les autres nœuds d'un cluster les lisent à leur redémarrage. Un champ
que vous videz conserve son ancienne valeur jusqu'au redémarrage du serveur.

## ID {#appForm-id}

Obligatoire. La clé sous laquelle l'application est enregistrée ; un modèle de tâche fait référence
à son application par cette clé. Les applications intégrées sont `ansible`, `terraform`, `tofu`,
`terragrunt`, `pulumi`, `bash`, `python` et `powershell`. Tout autre ID ajoute une application de
votre choix : Semaphore l'exécute comme un script, avec la commande indiquée dans **Path** ou,
lorsque **Path** est vide, la commande nommée comme l'ID.

Utilisez un ID sans points. Modifier l'ID d'une application existante ne la renomme pas : les
paramètres sont enregistrés comme une nouvelle application sous le nouvel ID, et l'ancienne
application reste.

## Icon {#appForm-icon}

Optionnel. Le nom d'une icône du jeu Material Design Icons sans le préfixe `mdi-`, par exemple
`language-python`. Laissez le champ vide pour conserver l'icône intégrée de l'application ;
une application de votre choix affiche alors un point d'interrogation.

## Nom {#appForm-title}

Optionnel. Le nom de l'application dans la liste des applications et dans le menu
**Nouveau modèle**. Laissez le champ vide pour conserver le nom intégré, par exemple
`Ansible Playbook` ; une application de votre choix n'a alors pas de nom, donnez-lui-en donc un.

## Path {#appForm-path}

Optionnel. La commande que Semaphore exécute à la place de celle par défaut, soit un nom recherché
dans le `PATH` du serveur, soit un chemin complet, par exemple `/usr/local/bin/terraform-1.9`. Les
commandes par défaut sont `terraform`, `tofu`, `terragrunt`, `bash`, `python3` et `powershell` ;
une application de votre choix exécute la commande nommée comme son ID.

Ansible ignore ce champ : Semaphore exécute toujours `ansible-playbook` et `ansible-galaxy`.

## Priority {#appForm-priority}

La place de l'application dans la liste des applications et dans le menu **Nouveau modèle** : une
application avec un nombre plus élevé apparaît en premier. Les applications intégrées vont de 1000
pour Ansible à 500 pour Python. Une application intégrée dont la priorité vaut `0` ou moins
retrouve sa priorité intégrée au démarrage du serveur.

## Arguments {#appForm-args}

Optionnel. Des arguments placés sur la ligne de commande juste après la commande, un par puce.
Cliquez sur **+** pour en ajouter un, cliquez sur une puce pour la modifier, faites glisser les
puces pour changer leur ordre.

- Terraform, OpenTofu et Terragrunt : les arguments précèdent ceux que Semaphore ajoute à chaque
  commande, comme `init` et `apply`.
- Applications de script : les arguments précèdent le script du modèle et remplacent ceux par
  défaut : PowerShell s'exécute avec `-File` par défaut, placez donc `-File` en dernier dans votre
  liste.
- Ansible ignore ce champ.

## Active {#appForm-active}

Une application active est proposée dans le menu **Nouveau modèle**. Décochez la case pour masquer
l'application de ce menu : les modèles déjà créés avec elle conservent leur application et
s'exécutent toujours.

La case est décochée pour une nouvelle application : cochez-la pour rendre l'application
disponible.
