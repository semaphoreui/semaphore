---
form: src/components/InventoryGroupForm.vue
---

# Editar grupo

Un grupo de hosts de un inventario gestionado. Un playbook se ejecuta en una parte de los hosts por
el nombre del grupo, y una variable del grupo se define una vez para todos sus hosts. Los hosts se
ponen en un grupo en la pestaña Hosts, en el diálogo del host o con las acciones sobre la selección.

Un grupo importado desde una fuente de inventario lo gestiona esa fuente. Todos sus campos son de solo
lectura y se actualizan en cada sincronización.

## Nombre {#inventoryGroupForm-name}

El nombre con el que Ansible conoce el grupo, por ejemplo en `hosts: web` de un playbook. El campo
es obligatorio. Un nombre es único dentro de su inventario y distingue mayúsculas de minúsculas:
`web` y `Web` son dos grupos. No puede contener espacios, comas, corchetes ni dos puntos, tiene como
máximo 250 caracteres y no puede ser `all` ni `ungrouped`. Ansible avisa de los nombres con
caracteres distintos de letras, dígitos y guiones bajos, como `web-eu`, pero el nombre funciona
igualmente. Tampoco se permiten los caracteres `*`, `?`, `!`, `&`, `~` y `@`, porque Ansible los lee
como un patrón.

## Descripción {#inventoryGroupForm-description}

Opcional. Una nota para las personas, que se muestra en la lista de grupos. Ansible no la recibe.

## Child groups {#inventoryGroupForm-children}

Opcional. Grupos dentro de este grupo. Los hosts de un grupo hijo también pertenecen a este grupo.
Un grupo no puede ser su propio ancestro.

## Variables {#inventoryGroupForm-variables}

Opcional. Variables de este grupo, escritas como un mapeo YAML o JSON:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Todos los hosts del grupo las reciben. Una variable de un host tiene prioridad sobre una variable
del grupo con el mismo nombre, y una variable del grupo tiene prioridad sobre una variable del
inventario. No se aceptan anclas, alias ni claves de fusión de YAML.
