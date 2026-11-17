---
form: src/components/RoleMappingForm.vue
---

# New mapping

Una asignación da un rol en un proyecto a cada usuario que inicia sesión mediante un proveedor LDAP
u OpenID Connect y pertenece allí a un grupo. Semaphore lee los grupos del usuario en cada inicio de
sesión a través del proveedor y concede o retira los roles asignados en consecuencia, de modo que un
usuario que abandona el grupo pierde el rol en su siguiente inicio de sesión. Un rol asignado
convierte al usuario en miembro del proyecto. Una asignación no se puede editar: elimínela y añada
una nueva.

## Provider {#roleMapping-provider}

El directorio LDAP o el proveedor OpenID Connect del que procede el grupo. La lista contiene los
proveedores configurados en la configuración del servidor. El campo es obligatorio.

Los grupos solo se leen cuando el proveedor está configurado para ello: un proveedor LDAP necesita
una base de búsqueda de grupos (`group_search_dn`), un proveedor OpenID Connect indica el claim con
los grupos (`groups_claim`). Una asignación de un proveedor que más tarde se elimina de la
configuración permanece en la lista, marcada como **Provider not configured**.

## Group {#roleMapping-group}

El nombre del grupo tal como lo da el proveedor. El campo es obligatorio.

- LDAP: el valor del atributo de nombre de grupo de la entrada del grupo, `cn` por defecto, por
  ejemplo `devops`. Un DN completo como `cn=devops,ou=groups,dc=example,dc=com` se reduce a su
  `cn`.
- OpenID Connect: uno de los valores del claim de grupos, tal como lo envía el proveedor.

## Project {#roleMapping-projectId}

El proyecto en el que se da el rol. El campo es obligatorio. Un administrador ve todos los
proyectos; cualquier otro usuario ve solo los proyectos de los que es miembro.

## Rol {#roleMapping-role}

El rol que reciben los miembros del grupo en el proyecto. El campo es obligatorio y está disponible
una vez seleccionado un proyecto. La lista contiene los roles de ese proyecto, los roles compartidos
y los roles integrados. Cuando elige otro proyecto, se vacía un rol que no existe allí.

Un rol **Owner** dado por una asignación no cuenta cuando Semaphore le impide quitar el rol al
último propietario de un proyecto, porque puede desaparecer en el siguiente inicio de sesión.
Mantenga al menos un propietario añadido a mano.
