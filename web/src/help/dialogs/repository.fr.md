---
form: src/components/RepositoryForm.vue
---

# Modifier Repository

Un dépôt est l'endroit d'où proviennent les playbooks, le code Terraform ou les scripts d'un modèle
de tâche : un dépôt Git, ou un répertoire sur la machine qui exécute les tâches. Avant chaque tâche,
Semaphore clone le dépôt Git, ou met à jour la copie qu'il a clonée auparavant, extrait la branche
et exécute la tâche dans cette copie.

## Nom {#repository-name}

Le nom sous lequel le dépôt apparaît sur la page Dépôts et dans le sélecteur de dépôt d'un modèle
de tâche et d'un inventaire. Le champ est obligatoire.

## URL ou chemin {#repository-gitUrl}

L'emplacement du code. Le champ est obligatoire. Les puces situées en dessous indiquent le genre
d'adresse que vous avez saisie, et un clic sur une puce réécrit l'adresse dans ce genre.

- **ssh** : `git@github.com:org/repo.git` ou `ssh://git@github.com/org/repo.git`. Git se connecte
  avec l'utilisateur figurant dans l'adresse et la **Clé SSH** choisie dans **Clé d'accès**.
- **https** et **http** : `https://github.com/org/repo.git`. Un dépôt privé nécessite une clé
  **Connexion avec mot de passe** ; un dépôt public nécessite **Aucun**. Évitez `http` :
  l'identifiant et le mot de passe circulent alors sans chiffrement.
- **git** : `git://host/org/repo.git`, le protocole Git en lecture seule, qui n'a pas de connexion.
- **file** : `file:///srv/git/repo.git`, un dépôt Git sur le disque de la machine qui exécute la
  tâche, cloné comme un dépôt distant.
- **chemin absolu** : un chemin absolu tel que `/srv/playbooks` ou `D:\playbooks`. Semaphore ne le
  clone pas : la tâche s'exécute directement dans ce répertoire, qui doit exister sur la machine qui
  exécute la tâche (le serveur Semaphore ou un exécuteur). La branche ne s'y applique pas.

Modifier l'adresse supprime les copies que le serveur Semaphore a clonées depuis l'ancienne.

## Branche {#repository-gitBranch}

La branche que Semaphore clone et maintient à jour, par exemple `main`. Le champ est obligatoire,
sauf pour un chemin absolu, pour lequel il est verrouillé.

Un modèle de tâche peut indiquer sa propre branche, qui l'emporte sur celle-ci, et un modèle peut
laisser ses tâches choisir une branche au démarrage. Une étiquette Git fonctionne aussi, mais une
étiquette ne peut pas être récupérée par un pull : Semaphore clone donc à nouveau le dépôt avant
chaque tâche.

## Single checkout for all templates {#repository-monorepo}

Désactivé par défaut : chaque modèle de tâche reçoit sa propre copie du dépôt. Cochez la case pour
cloner le dépôt une seule fois et partager cette copie entre tous les modèles qui l'utilisent, pour
un dépôt trop volumineux pour garder une copie par modèle.

La copie partagée reste sur la branche de ce dépôt : les modèles qui l'utilisent ne peuvent pas
avoir leur propre branche, et leurs tâches ne peuvent pas choisir une autre branche ni un autre
commit. Verrouillé pour un chemin absolu.

## Clé d'accès {#repository-sshKeyId}

La clé du Magasin de clés avec laquelle Git se connecte au dépôt. Le champ est obligatoire. Il
propose des clés de trois types ; celui à choisir dépend de l'URL :

- **Clé SSH** pour une URL `ssh`. Semaphore charge la clé privée dans un agent SSH pendant la durée
  du clonage ou de la mise à jour. L'utilisateur est celui de l'URL, comme `git` dans
  `git@github.com:org/repo.git`.
- **Connexion avec mot de passe** pour une URL `https` : l'identifiant et le mot de passe, ou le
  jeton d'accès du serveur Git. Avec un identifiant vide, le mot de passe est envoyé seul, comme
  nom d'utilisateur, ce que certains serveurs attendent d'un jeton. Les mêmes identifiants sont
  transmis à `ansible-galaxy` pour les rôles et les collections situés sur le même serveur.
- **Aucun** pour un dépôt public, une URL `git` ou `file`, ou un chemin absolu.

Une clé qui ne correspond pas à l'URL ne donne aucun identifiant à Git, et le clonage d'un dépôt
privé échoue alors.
