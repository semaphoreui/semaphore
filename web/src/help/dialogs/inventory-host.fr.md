---
form: src/components/InventoryHostForm.vue
---

# Modifier l'hôte

Un hôte d'un inventaire géré. Une tâche qui utilise l'inventaire s'exécute sur les hôtes qui sont
activés dans la liste des hôtes.

Un hôte importé depuis une source d'inventaire est géré par cette source. Son nom et ses variables
sont en lecture seule et sont mis à jour à chaque synchronisation. La description, les groupes non
importés et l'interrupteur Enabled restent modifiables.

## Nom {#inventoryHostForm-name}

Le nom sous lequel Ansible connaît l'hôte, et l'adresse à laquelle il se connecte, sauf si les
variables définissent `ansible_host`. Le champ est obligatoire. Un nom est unique dans son
inventaire et la casse compte : `web1` et `Web1` sont deux hôtes. Il ne peut pas contenir d'espaces,
de virgules, de crochets ni de deux-points, et compte au plus 250 caractères.

## Description {#inventoryHostForm-description}

Facultatif. Une note pour les personnes, affichée dans la liste des hôtes. Ansible ne la reçoit pas.

## Groups {#inventoryHostForm-groups}

Facultatif. Les groupes dont cet hôte fait partie. Un hôte peut faire partie de plusieurs groupes.
Les groupes se créent dans l'onglet Groups.

## Variables {#inventoryHostForm-variables}

Facultatif. Les variables de cet hôte, écrites sous forme de mapping YAML ou JSON :

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Une variable de l'hôte est prioritaire sur une variable de l'inventaire portant le même nom. Les
ancres, alias et clés de fusion YAML ne sont pas acceptés.
