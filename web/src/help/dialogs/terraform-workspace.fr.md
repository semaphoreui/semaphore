---
form: src/components/TerraformInventoryForm.vue
---

# Nouveau Terraform Workspace

Un workspace est l'inventaire d'un modèle Terraform, OpenTofu ou Terragrunt : il indique le
workspace qu'une tâche sélectionne après `init`, et la clé avec laquelle `init` télécharge les
modules privés. Le workspace que vous créez ici est rattaché à ce modèle et apparaît au-dessus de
son état. Les tâches du modèle s'exécutent dans son workspace par défaut, celui marqué d'une coche ;
**Make default** le change. La même boîte de dialogue modifie un workspace lorsque vous l'ouvrez
avec le bouton crayon.

## Workspace name {#tfWorkspace-inventory}

Le nom du workspace Terraform, par exemple `production`. Le champ est obligatoire.

Avant un plan, Semaphore exécute `workspace select -or-create=true` avec ce nom, de sorte que le
workspace est créé s'il n'existe pas encore. Lorsque `workspace list` échoue, par exemple parce que
le backend ne prend pas en charge les workspaces, Semaphore ignore la sélection.

Semaphore enregistre le workspace sous la forme `TEMPLATE - NAME`, le nom du modèle suivi de ce
nom : c'est ce nom complet que vous voyez sur la page Inventaire et lorsque vous rattachez le
workspace à un modèle. Renommer un workspace ne le renomme pas dans Terraform : la tâche suivante
sélectionne, ou crée, un workspace portant le nouveau nom.

## SSH key for private modules {#tfWorkspace-sshKeyId}

La clé qu'utilise `init` pour télécharger des modules depuis des dépôts Git privés via SSH. Le
champ est obligatoire dans cette boîte de dialogue. La liste propose les clés de type **Clé SSH**,
**Connexion avec mot de passe** et **Aucun** du Magasin de clés.

Seule une **Clé SSH** a un effet : Semaphore la charge dans un agent SSH pendant `init`, et Git
l'utilise pour des sources de modules comme `git::ssh://git@example.com/modules/vpc.git`. Une clé
de type **Connexion avec mot de passe** ne transmet aucun identifiant à Git. Choisissez **Aucun**
lorsque les modules n'ont besoin d'aucune clé.
