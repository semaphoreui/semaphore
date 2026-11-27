---
form: src/components/InventorySourceForm.vue
---

# Modifier la source

Une source importe les hôtes et les groupes d'un inventaire géré depuis un fichier d'un dépôt Git. Le fichier est lu à chaque clic sur Sync, et les hôtes et groupes qu'il décrit restent alignés sur lui. Le serveur a besoin d'Ansible et d'un accès au dépôt.

## Nom {#inventorySourceForm-name}

Le nom de la source, affiché sur sa carte et à côté des hôtes qu'elle apporte. Le champ est obligatoire.

## Dépôt {#inventorySourceForm-repository}

Le dépôt du projet dans lequel le fichier est lu. Le champ est obligatoire. Sa branche est utilisée.

## Chemin {#inventorySourceForm-path}

Le chemin d'un fichier d'inventaire dans le dépôt, par exemple `inventory/prod.yml`, ou d'un répertoire de fichiers d'inventaire. Le champ est obligatoire.

Les fichiers sont lus comme du YAML ou de l'INI statique, les scripts et les fichiers de configuration de plugins sont refusés. Dans un répertoire, chaque fichier qu'Ansible n'ignore pas doit être un inventaire valide, sinon la synchronisation échoue et rien ne change. Ansible ignore certaines extensions dans un répertoire, par exemple `.cfg`, `.md` et `.txt`. Un dépôt contenant un lien qui mène hors de lui est refusé.
