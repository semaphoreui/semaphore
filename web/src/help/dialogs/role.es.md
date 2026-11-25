---
form: src/components/EditRoleForm.vue
---

# Edit Role

Un rol es un conjunto de permisos con nombre dentro de un proyecto. Usted da roles a los miembros de
un proyecto, y un miembro puede hacer lo que permita cualquiera de sus roles. Abierto desde la
página Roles de un proyecto, el diálogo edita solo un rol de ese proyecto; abierto desde la página
Roles fuera de un proyecto, edita un rol compartido que puede usar cualquier proyecto. Los roles
integrados Owner, Manager, Task Runner y Guest no se pueden cambiar.

## Nombre {#role-name}

El nombre con el que el rol aparece en la página Roles y allí donde se elige un rol, por ejemplo al
añadir un miembro a un proyecto. El campo es obligatorio. Puede renombrar un rol en cualquier
momento.

## Slug {#role-slug}

El identificador del rol. Semaphore lo rellena a partir del nombre mientras escribe el nombre,
hasta que usted edite el slug. Admite de 1 a 100 letras, dígitos, guiones bajos y guiones, por
ejemplo `release-managers`.

El slug no se puede cambiar una vez guardado el rol: el campo está bloqueado cuando edita un rol.

- Un slug debe ser único, y la comprobación no distingue entre mayúsculas y minúsculas. El slug de
  un rol compartido debe ser distinto de todos los demás roles; el slug de un rol de proyecto debe
  ser distinto de los roles compartidos y de los demás roles de su proyecto.
- `owner`, `manager`, `task_runner`, `guest` y `all` están reservados, en mayúsculas o minúsculas.

## Permisos {#role-permissionsTitle}

Lo que pueden hacer en el proyecto los miembros que tienen el rol. Un rol nuevo no tiene ninguno
marcado. Los permisos de todos los roles de un miembro se suman; un administrador de Semaphore puede
hacer todo en cada proyecto, sean cuales sean los roles.

- **Can read project**: ver el proyecto: sus plantillas de tareas, las tareas y sus registros, los
  inventarios, claves, repositorios, grupos de variables, horarios y miembros. Toda lectura en el
  proyecto lo requiere, y ningún otro permiso lo incluye, así que márquelo para casi todos los
  roles. Sin él, un miembro ve solo las plantillas de tareas y los workflows que conceden acceso
  propio al rol, con sus tareas.
- **Can run project tasks**: iniciar tareas desde plantillas de tareas, detenerlas y confirmar o
  rechazar una tarea que espera una confirmación.
- **Can update project**: cambiar la configuración del proyecto y eliminarlo.
- **Can manage project resources**: crear, cambiar y eliminar plantillas de tareas, inventarios,
  claves, repositorios, grupos de variables, horarios, vistas, integraciones, almacenes de secretos,
  ejecutores del proyecto y los roles del proyecto; eliminar tareas; definir qué roles pueden usar
  una plantilla de tarea.
- **Can manage project users**: añadir miembros al proyecto, cambiar sus roles y quitarlos.
- **Can run workflows**: iniciar un workflow y detener sus ejecuciones.
- **Can edit workflows**: crear, cambiar y eliminar workflows, y definir qué roles pueden usar un
  workflow.
- **Can approve workflows**: aprobar o rechazar un paso de una ejecución de workflow que espera una
  aprobación.
