---
form: src/components/InventoryForm.vue
---

# Modifier Ansible Inventory

Un inventaire indique à Ansible sur quels hôtes exécuter un playbook et contient les identifiants
que Semaphore utilise pour s'y connecter. Un modèle de tâche fait référence à un inventaire, et
chaque tâche de ce modèle s'exécute sur cet inventaire. Le format d'un inventaire est décrit dans le
[guide des inventaires Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Nom {#inventoryForm-name}

Le nom sous lequel l'inventaire apparaît sur la page Inventaire et dans le sélecteur d'inventaire
d'un modèle de tâche. Choisissez un nom qui distingue les hôtes, par exemple `Production web servers`.

## Étiquette de l'exécuteur {#inventoryForm-runnerTag}

Affiché lorsque des exécuteurs de projet sont disponibles.

Les tâches qui utilisent cet inventaire s'exécutent sur un exécuteur portant cette étiquette. Une
étiquette d'exécuteur définie sur le modèle de tâche l'emporte sur celle de l'inventaire. Laissez le
champ vide pour que n'importe quel exécuteur puisse prendre les tâches.

## Identifiants de l'utilisateur {#inventoryForm-sshKeyId}

La clé du Magasin de clés avec laquelle Ansible se connecte aux hôtes. Le champ est obligatoire et
accepte trois types de clés :

- **Clé SSH** : Ansible se connecte avec la clé privée et, si la clé en possède un, avec son
  identifiant.
- **Connexion avec mot de passe** : Ansible se connecte avec l'identifiant et le mot de passe.
- **Aucun** : Semaphore ne transmet aucun identifiant. Utilisez-le pour les hôtes qui n'en ont pas
  besoin, comme `localhost`, ou lorsque l'inventaire lui-même définit la manière de se connecter.

## Identifiants sudo {#inventoryForm-becomeKeyId}

Optionnel. Une clé de type **Connexion avec mot de passe** du Magasin de clés, utilisée lorsqu'un
play élève ses privilèges avec `become` : son identifiant est l'utilisateur qu'Ansible devient, et
son mot de passe est celui qu'Ansible fournit à l'invite.

Laissez le champ vide lorsque les plays n'élèvent pas leurs privilèges, ou lorsque l'utilisateur
ci-dessus n'a pas besoin de mot de passe pour le faire.

## Type {#inventoryForm-type}

D'où vient l'inventaire :

- **Static** : l'inventaire est saisi dans cette boîte de dialogue au format INI.
- **Static YAML** : l'inventaire est saisi dans cette boîte de dialogue au format YAML. Utilisez-le
  également pour la configuration d'un plugin d'inventaire.
- **File** : l'inventaire est un fichier dans un dépôt Git.
- **Managed** : les hôtes et leurs variables sont conservés sous forme d'enregistrements. Ouvrez
  l'inventaire depuis la liste pour les modifier.

Les champs situés sous le type changent en fonction de celui-ci.

## Chemin vers le fichier d'inventaire {#inventoryForm-inventoryPath}

Affiché lorsque le type est **File**.

Le chemin du fichier d'inventaire depuis la racine de son dépôt, par exemple
`inventory/production.yml`. Le dépôt est celui sélectionné dans le champ Dépôt ou, si ce champ est
vide, le dépôt du modèle de tâche qui utilise l'inventaire.

## Dépôt {#inventoryForm-repositoryId}

Affiché lorsque le type est **File**. Optionnel.

Le dépôt qui contient le fichier d'inventaire. Semaphore met à jour sa copie du dépôt avant chaque
tâche. Laissez le champ vide lorsque le fichier se trouve dans le même dépôt que le playbook : le
chemin est alors recherché dans le dépôt du modèle de tâche.

## Inventaire {#inventoryForm-inventoryCode}

Affiché lorsque le type est **Static** ou **Static YAML**.

Le texte de l'inventaire. Semaphore l'enregistre et l'écrit dans un fichier temporaire pour chaque
tâche. Un inventaire statique au format INI :

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Les mêmes hôtes en **Static YAML** :

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
