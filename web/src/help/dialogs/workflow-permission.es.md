---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Da a un rol permisos adicionales sobre este workflow concreto. Un miembro que tiene el rol los
recibe además de los permisos que el rol tiene en todo el proyecto: una concesión aquí solo añade,
nunca quita un permiso. Un miembro con varios roles recibe todos los permisos concedidos a
cualquiera de ellos.

## Rol {#templatePerm-roleSlug}

El rol que recibe los permisos. La lista muestra los roles de este proyecto, los roles compartidos y
los roles integrados. El campo es obligatorio, y un rol solo puede tener una concesión en un
workflow: el servidor rechaza una segunda.

El rol de una concesión guardada no se puede cambiar. El servidor conserva el rol y guarda solo los
permisos, elija lo que elija aquí. Para dar los permisos a otro rol, quite esta concesión de la
lista y añada una nueva.

## Permisos {#templatePerm-permissionsTitle}

Lo que puede hacer con este workflow un miembro que tiene el rol:

- **Can see this workflow**: el workflow aparece en la lista de workflows, con sus ejecuciones.
- **Can run this workflow**: iniciar ejecuciones del workflow.
- **Can edit this workflow**: cambiar y eliminar el workflow, y gestionar estos permisos.
- **Can approve this workflow**: aprobar o rechazar los pasos de aprobación de sus ejecuciones.

Ejecutar, editar o aprobar incluye ver el workflow: cuando marca cualquiera de ellos, **Can
see this workflow** también se marca y no se puede desmarcar. Deje todas las casillas sin marcar
para guardar una concesión que no da nada.
