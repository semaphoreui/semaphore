---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Une correspondance indique à Semaphore quel identifiant du Magasin de clés utiliser pour un hôte
SSH ou une URL Git. Un dépôt porte une seule clé, mais une tâche peut aussi atteindre d'autres
serveurs : sous-modules, rôles et collections de `requirements.yml`, modules Terraform, dépôts
d'inventaire. Les correspondances du projet s'appliquent à toutes les connexions Git et SSH de ses
tâches, ainsi qu'aux commandes Git que Semaphore exécute pour le projet, comme la recherche de
nouveaux commits dans un dépôt. Aucun fichier de votre dépôt n'est modifié.

## Type {#hostConfigForm-type}

Ce que la correspondance cible :

- **Host** : un nom d'hôte SSH. Chaque connexion SSH d'une tâche vers cet hôte utilise la clé de la
  correspondance : Git via SSH, et Ansible se connectant à un hôte d'inventaire de ce nom. La
  correspondance l'emporte sur la configuration SSH du serveur pour le même hôte.
- **URL** : une URL Git ou son début. Git se connecte à chaque dépôt situé sous cette URL avec
  l'identifiant de la correspondance.

Passer à **Host** efface un **Credential** qu'une correspondance d'hôte ne peut pas utiliser.

## Hôte ou URL {#hostConfigForm-host}

Le champ est obligatoire. Une seule correspondance d'un type peut exister pour chaque hôte ou URL.

Pour **Host** : le nom d'hôte exactement tel que Git ou Ansible s'y connectent, par exemple
`github.com` ou `10.0.0.21`. Lettres, chiffres, points, traits d'union et tirets bas uniquement :
pas d'utilisateur, pas de port, pas de caractères génériques.

Pour **URL** : une adresse qui commence par `https://` ou `http://`, sans identifiant et sans
espaces, guillemets ni `=`. Indiquez un dépôt entier, ou un préfixe qui se termine par `/` pour
couvrir un groupe de dépôts :

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Lorsque plusieurs correspondances s'appliquent, la plus longue l'emporte : la correspondance du
dépôt ci-dessus remplace celle d'`acme` pour ce seul dépôt.

Avec une clé SSH, une correspondance d'URL fait se connecter Git en SSH à l'hôte de l'URL sur le
port `22`, quel que soit le port indiqué dans l'URL.

## Credential {#hostConfigForm-sshKeyId}

La clé du Magasin de clés utilisée pour les connexions que la correspondance cible. Le champ est
obligatoire.

- **Host** n'accepte qu'une **Clé SSH**. Lorsque la clé possède un identifiant, SSH se connecte
  avec cet utilisateur.
- **URL** accepte une **Clé SSH** ou une **Connexion avec mot de passe**. Une clé SSH se connecte
  avec son identifiant, ou avec `git` lorsqu'elle n'en a pas. Une connexion avec mot de passe ne
  fonctionne qu'avec une URL `https://` ; pour un jeton, placez le jeton dans le mot de passe et
  laissez l'identifiant vide.

Une connexion avec mot de passe parvient à Git par son environnement, pas par sa ligne de commande,
et le journal de la tâche n'affiche que l'URL d'origine.
