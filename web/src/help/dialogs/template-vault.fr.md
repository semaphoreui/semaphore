---
form: src/components/TemplateVaults.vue
---

# Coffre

Un coffre fournit à Ansible le mot de passe des fichiers et des variables chiffrés avec Ansible
Vault dans le dépôt du playbook. Chaque coffre du modèle devient une option `--vault-id` de
`ansible-playbook`. Le coffre est enregistré avec le modèle, lorsque vous enregistrez la boîte de
dialogue du modèle.

## ID de coffre {#templateVaultDialog-name}

Optionnel. L'identifiant de coffre auquel le mot de passe est destiné, le libellé avec lequel le
contenu a été chiffré (`ansible-vault encrypt --vault-id prod@prompt`). Laissez-le vide pour
l'identifiant de coffre `default`.

L'identifiant est enregistré en minuscules. Chaque coffre du modèle doit avoir son propre
identifiant, et un seul peut être vide.

## Type {#templateVaultDialog-type}

L'origine du mot de passe :

- **Mot de passe** : une clé du Magasin de clés. Semaphore saisit son mot de passe lorsque Ansible
  le demande, de sorte qu'il n'est jamais écrit dans un fichier.
- **Script client** : un script client de mot de passe de coffre qui affiche le mot de passe ;
  Ansible l'exécute.

## Mot de passe du coffre {#templateVaultDialog-vaultKeyId}

Affiché lorsque le type est **Mot de passe**.

La clé qui contient le mot de passe du coffre. Seules les clés de type **Connexion avec mot de
passe** sont listées ; Semaphore utilise leur mot de passe et ignore l'identifiant. Le champ est
obligatoire.

## Chemin du script {#templateVaultDialog-script}

Affiché lorsque le type est **Script client**.

Le chemin du script, par exemple `scripts/vault-pass-client.py`. Un chemin relatif part du
répertoire dans lequel Ansible s'exécute : la racine du dépôt, ou le répertoire de travail du
modèle s'il en a un. Le champ est obligatoire.

Le nom du fichier doit se terminer par `-client`, avec ou sans extension : ce n'est qu'à cette
condition qu'Ansible exécute le fichier comme un script client et lui transmet l'identifiant de
coffre avec `--vault-id`. Le script s'exécute sur la machine qui exécute la tâche, sous
l'utilisateur avec lequel Semaphore s'exécute.
