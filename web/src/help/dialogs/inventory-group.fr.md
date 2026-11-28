---
form: src/components/InventoryGroupForm.vue
---

# Modifier le groupe

Un groupe d'hôtes d'un inventaire géré. Un playbook s'exécute sur une partie des hôtes par le nom du
groupe, et une variable du groupe est définie une seule fois pour tous ses hôtes. Les hôtes sont mis
dans un groupe dans l'onglet Hosts, dans la boîte de dialogue de l'hôte ou avec les actions sur la
sélection.

Un groupe importé depuis une source d'inventaire est géré par cette source. Tous ses champs sont en
lecture seule et sont mis à jour à chaque synchronisation.

## Nom {#inventoryGroupForm-name}

Le nom sous lequel Ansible connaît le groupe, par exemple dans `hosts: web` d'un playbook. Le champ
est obligatoire. Un nom est unique dans son inventaire et la casse compte : `web` et `Web` sont deux
groupes. Il ne peut pas contenir d'espaces, de virgules, de crochets ni de deux-points, compte au
plus 250 caractères et ne peut pas être `all` ni `ungrouped`. Ansible avertit pour les noms qui
contiennent d'autres caractères que des lettres, des chiffres et des tirets bas, comme `web-eu`,
mais le nom fonctionne quand même. Les caractères `*`, `?`, `!`, `&`, `~` et `@` ne sont pas
autorisés non plus, car Ansible les lit comme un motif.

## Description {#inventoryGroupForm-description}

Facultatif. Une note pour les personnes, affichée dans la liste des groupes. Ansible ne la reçoit
pas.

## Child groups {#inventoryGroupForm-children}

Facultatif. Les groupes à l'intérieur de ce groupe. Les hôtes d'un groupe enfant appartiennent aussi
à ce groupe. Un groupe ne peut pas être son propre ancêtre.

## Variables {#inventoryGroupForm-variables}

Facultatif. Les variables de ce groupe, écrites sous forme de mapping YAML ou JSON :

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Chaque hôte du groupe les reçoit. Une variable d'un hôte est prioritaire sur une variable du groupe
portant le même nom, et une variable du groupe est prioritaire sur une variable de l'inventaire. Les
ancres, alias et clés de fusion YAML ne sont pas acceptés.
