---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Da a un rol permisos adicionales sobre esta plantilla concreta. Un miembro que tiene el rol los
recibe además de los permisos que el rol tiene en todo el proyecto: una concesión aquí solo añade,
nunca quita un permiso. Un miembro con varios roles recibe todos los permisos concedidos a
cualquiera de ellos.

## Rol {#templatePerm-roleSlug}

El rol que recibe los permisos. La lista muestra los roles de este proyecto, los roles compartidos y
los roles integrados. El campo es obligatorio, y un rol solo puede tener una concesión en una
plantilla: el servidor rechaza una segunda.

El rol de una concesión guardada no se puede cambiar. El servidor conserva el rol y guarda solo los
permisos, elija lo que elija aquí. Para dar los permisos a otro rol, quite esta concesión de la
lista y añada una nueva.

## Permisos {#templatePerm-permissionsTitle}

Lo que puede hacer con esta plantilla un miembro que tiene el rol:

- **Can see this template**: la plantilla aparece en la lista de plantillas, con sus tareas y sus
  registros.
- **Can run this template**: iniciar tareas desde la plantilla, y detener, confirmar o rechazar
  sus tareas.
- **Can edit this template**: cambiar y eliminar la plantilla, detener todas sus tareas y
  gestionar estos permisos.

Ejecutar o editar incluye ver la plantilla: cuando marca cualquiera de los dos, **Can see this
template** también se marca y no se puede desmarcar. Deje todas las casillas sin marcar para
guardar una concesión que no da nada.
