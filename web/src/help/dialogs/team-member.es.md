---
form: src/components/TeamMemberForm.vue
---

# Nuevo Miembro del Equipo

Añade un usuario al proyecto con un rol. El rol decide lo que el miembro puede ver y hacer en el
proyecto; nada más de su cuenta cambia. Cuando el administrador ha activado las invitaciones, el
usuario recibe en su lugar una invitación y se convierte en miembro una vez que la acepta.

## Email o Username {#teamMember-inviteType}

Se muestra cuando se puede invitar a los usuarios tanto por correo electrónico como por nombre de
usuario.

Cómo elige a la persona: **Email** para invitar a una dirección de correo electrónico, **Username**
para elegir a un usuario que ya tiene una cuenta en este servidor de Semaphore.

## Correo Electrónico {#teamMember-email}

Se muestra cuando se invita a la persona por correo electrónico.

La dirección de correo electrónico de la persona a la que invitar. Se convierte en miembro del
proyecto una vez que acepta la invitación.

## Usuario {#teamMember-userId}

Se muestra cuando se elige a la persona por nombre de usuario.

El usuario que se va a añadir, listado con el nombre de usuario y el nombre completo. El campo es
obligatorio. La lista contiene a todos los usuarios del servidor de Semaphore que todavía no son
miembros del proyecto: una persona sin cuenta tiene que obtener primero una de un administrador.

## Rol {#teamMember-role}

Lo que el miembro puede hacer en el proyecto. El campo es obligatorio. Los roles integrados:

- **Owner**: todo, incluida la configuración del proyecto, su equipo y su eliminación.
- **Manager**: ejecuta tareas y workflows, aprueba workflows, y crea, cambia y elimina plantillas
  de tareas, inventarios, claves, repositorios, grupos de variables, horarios, integraciones y
  workflows. Un manager no puede cambiar la configuración del proyecto ni su equipo.
- **Task Runner**: ve el proyecto, ejecuta tareas y workflows y aprueba workflows. Un task runner
  no puede cambiar nada en el proyecto.
- **Guest**: ve el proyecto, sus tareas y sus registros, y no cambia nada.

La lista también contiene los roles creados para este proyecto (**Roles de proyecto**) y los roles
compartidos por todos los proyectos (**Shared roles**). Un rol así concede los permisos definidos
para él en la página Roles. Una invitación solo puede llevar un rol integrado.

Un miembro que puede gestionar el equipo puede dar cualquier rol, **Owner** incluido. El diálogo da
un rol; añada más después en la columna **Rol** de la lista del equipo.
