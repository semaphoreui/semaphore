---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Un rol global concede permisos sobre toda la instancia de Semaphore a los usuarios a los que se
asigna; se asigna en el diálogo de un usuario. No da acceso dentro de los proyectos: ese acceso
procede solo de los roles que un usuario tiene como miembro de un proyecto. Un administrador tiene
todos los permisos globales, sean cuales sean los roles asignados. Los roles globales son una
función Pro.

## Nombre {#globalRole-name}

El nombre con el que el rol aparece en la pestaña Roles globales de la página Usuarios y en el
diálogo de un usuario. El campo es obligatorio. Puede renombrar un rol en cualquier momento.

## Slug {#globalRole-slug}

El identificador del rol. Semaphore lo rellena a partir del nombre mientras escribe el nombre,
hasta que usted edite el slug. Admite de 1 a 100 letras, dígitos, guiones bajos y guiones, por
ejemplo `user-admins`.

El slug debe ser único entre los roles globales; un rol de proyecto puede tener el mismo slug. No se
puede cambiar una vez guardado el rol: el campo está bloqueado cuando edita un rol.

## Global permissions {#globalRole-permissionsTitle}

Lo que puede hacer en toda la instancia un usuario que tiene el rol. Un rol nuevo no tiene ninguno
marcado. Los permisos de todos los roles globales de un usuario se suman.

- **Can manage users**: crear usuarios, cambiarlos y eliminarlos, y ver sus datos en la lista de
  usuarios. Un usuario así no puede convertir a un usuario en administrador ni en usuario Pro, y no
  puede cambiar ni eliminar a un administrador ni a un usuario que tenga un permiso global que él
  mismo no tiene.
- **Can manage global roles**: crear, cambiar y eliminar los roles globales y los roles de proyecto
  compartidos, y decidir qué usuarios tienen qué rol global. Concédalo solo a personas a las que
  confiaría toda la instancia.
- **Can manage system settings**: cambiar la configuración de la instancia, las aplicaciones, los
  ejecutores globales y las asignaciones de grupos; ver el clúster y la información sobre el
  servidor; ver y eliminar las tareas de todos los proyectos; limpiar la caché del servidor.
- **Can read the audit log**: leer el registro de auditoría de la instancia y exportarlo.
