---
form: src/components/InventoryHostForm.vue
---

# Editar host

Un host de un inventario gestionado. Una tarea que usa el inventario se ejecuta en los hosts que
están activados en la lista de hosts.

Un host importado desde una fuente de inventario lo gestiona esa fuente. Su nombre y sus variables son
de solo lectura y se actualizan en cada sincronización. La descripción, los grupos no importados y el
interruptor Enabled siguen siendo editables.

## Nombre {#inventoryHostForm-name}

El nombre con el que Ansible conoce el host y la dirección a la que se conecta, salvo que las
variables definan `ansible_host`. El campo es obligatorio. Un nombre es único dentro de su
inventario y distingue mayúsculas de minúsculas: `web1` y `Web1` son dos hosts. No puede contener
espacios, comas, corchetes ni dos puntos, y tiene como máximo 250 caracteres.

## Descripción {#inventoryHostForm-description}

Opcional. Una nota para las personas, que se muestra en la lista de hosts. Ansible no la recibe.

## Groups {#inventoryHostForm-groups}

Opcional. Los grupos en los que está este host. Un host puede estar en varios grupos. Los grupos se
crean en la pestaña Groups.

## Variables {#inventoryHostForm-variables}

Opcional. Variables de este host, escritas como un mapeo YAML o JSON:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Una variable del host tiene prioridad sobre una variable del inventario con el mismo nombre. No se
aceptan anclas, alias ni claves de fusión de YAML.
