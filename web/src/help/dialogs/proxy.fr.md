---
form: src/components/ProxyForm.vue
---

# Modifier Proxy

Un proxy est un hôte par lequel passent les tâches du projet pour atteindre des serveurs auxquels
elles ne peuvent pas se connecter directement. Choisissez-le dans le champ **Proxy** d'un dépôt,
pour récupérer le dépôt à travers lui, ou d'un inventaire, pour y connecter Ansible aux hôtes.

## Nom {#proxyForm-name}

Obligatoire et unique dans le projet. Les dépôts et les inventaires affichent le proxy sous ce nom.

## Type {#proxyForm-type}

- **SSH (jump host)** : Semaphore se connecte à l'hôte en SSH et ouvre de là la connexion vers le
  serveur. Seul ce type peut requérir un autre proxy.
- **SOCKS5**, **HTTP**, **HTTPS** : un serveur proxy qui transporte la connexion. Git l'utilise
  lui-même pour un dépôt `https` ; Semaphore fait passer une connexion SSH par lui, aucun autre
  outil n'est donc à installer.

## Host {#proxyForm-host}

Obligatoire. Un nom d'hôte ou une adresse IP, sans utilisateur, port ni schéma. Un nom d'hôte ne
peut contenir que des lettres, des chiffres, des points, des tirets et des tirets bas.

## Port {#proxyForm-port}

Facultatif, de 1 à 65535. Vide signifie `22` pour un jump host SSH. Indiquez le port pour les autres
types : sans lui, Git et les connexions SSH supposent des ports différents.

## User {#proxyForm-user}

SSH uniquement : l'identifiant sur le jump host. Vide signifie l'utilisateur sous lequel tourne
Semaphore. Lettres, chiffres, points, tirets et tirets bas, sans tiret au début. Un proxy SOCKS5
ou HTTP prend son identifiant dans la **Clé d'accès**.

## Clé d'accès {#proxyForm-sshKeyId}

Facultatif. La clé du magasin de clés avec laquelle on se connecte au proxy lui-même, en général
pas la clé des serveurs derrière lui :

- une **Clé SSH** pour un jump host SSH ;
- une **Connexion avec mot de passe** pour SOCKS5, HTTP et HTTPS. Le mot de passe est transmis
  dans l'environnement, jamais sur une ligne de commande.

## Requires Proxy {#proxyForm-requiresProxyId}

SSH uniquement, facultatif. Un autre proxy SSH du projet à traverser d'abord, pour un jump host
joignable seulement depuis un autre. Une chaîne compte au plus 5 proxys et ne peut pas revenir à
un proxy qui en fait déjà partie.
