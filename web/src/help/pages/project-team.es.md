---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Equipo

Los miembros del proyecto y los roles que tienen. Un rol determina lo que un miembro puede hacer en
el proyecto: ejecutar tareas, cambiar plantillas y otros recursos, administrar el equipo. Un miembro
puede tener varios roles y obtiene todos los permisos que concede cualquiera de ellos.

- **Nuevo Miembro del Equipo** añade un usuario al proyecto con un rol. El botón lo ven los
  miembros que pueden administrar el equipo.
- La columna **Rol** cambia los roles de un miembro al instante, sin guardar. Un miembro que no
  puede administrar el equipo ve los roles, pero no puede cambiarlos.
- Un rol con un candado viene de un grupo del usuario en LDAP u OIDC que el administrador asignó a
  ese rol. No se puede quitar aquí: desaparece cuando el usuario sale del grupo o se elimina la
  asignación.
- El proyecto necesita al menos un Owner. Semaphore no permite quitar a su último Owner ni retirarle
  el rol Owner, salvo que lo haga un administrador de Semaphore.
- **Salir del Proyecto** lo saca a usted del proyecto. Solo aparece si el administrador permite a
  los miembros salir de los proyectos; el último Owner tampoco puede salir.
