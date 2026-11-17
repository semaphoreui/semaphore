---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Editar Usuario

Un usuario es una cuenta que inicia sesión en Semaphore. La pestaña **Settings** contiene el perfil
de la cuenta y sus opciones; **Guardar** las escribe. La pestaña **Security** contiene la
contraseña, la autenticación de dos factores y las cuentas de inicio de sesión vinculadas, y la
pestaña **Roles** los roles globales del usuario: ambas actúan de inmediato, sin **Guardar**. Lo
que un usuario puede hacer dentro de un proyecto se define en la página **Equipo** de ese proyecto,
no aquí.

## Nombre {#user-name}

El nombre completo del usuario, que se muestra en la barra lateral y en las listas de usuarios y de
miembros de proyectos. El campo es obligatorio.

Para un usuario **Externo**, Semaphore toma el nombre del proveedor de identidad al iniciar sesión,
cuando el proveedor lo envía, así que un cambio hecho aquí puede no durar más allá del siguiente
inicio de sesión.

## Nombre de usuario {#user-username}

El login del usuario. El campo es obligatorio. El usuario inicia sesión con el nombre de usuario o
con el correo electrónico, y la contraseña.

El nombre de usuario de un usuario **Externo** no se puede cambiar: el servidor se niega a
guardarlo.

## Correo Electrónico {#user-email}

La dirección de correo electrónico del usuario. El campo es obligatorio. Semaphore envía a ella las
alertas de tareas del usuario, y una aplicación de autenticación la muestra como nombre de la
cuenta. Es privada: solo la ven el usuario y quienes pueden gestionar usuarios.

Para un usuario **Externo**, el correo electrónico procede del proveedor de identidad: el campo es
de solo lectura, y Semaphore actualiza la dirección en cada inicio de sesión.

## Contraseña {#user-password}

Se muestra cuando se crea un usuario nuevo.

La contraseña con la que el usuario inicia sesión. El campo es obligatorio, salvo para un usuario
**Externo**, que no tiene contraseña en Semaphore. Para cambiar la contraseña de un usuario
existente, use **Change password** en la pestaña **Security**.

## Enviar alertas {#user-alert}

Envía al usuario un correo electrónico cuando falla una tarea en un proyecto en el que el usuario
puede ver la tarea. El proyecto debe permitir alertas, y el administrador del servidor debe haber
activado las alertas por correo electrónico. Una plantilla de tarea puede impedir que sus tareas
fallidas envíen alertas.

La opción no afecta a las alertas de un proyecto enviadas a Telegram, Slack y los demás servicios de
mensajería.

## Usuario Administrador {#user-admin}

Se muestra cuando usted es administrador.

Un administrador puede hacer todo en el servidor: gestionar los usuarios, los roles globales y la
configuración del servidor, y abrir cualquier proyecto con todos los derechos.

No puede quitar la opción a su propia cuenta, y el servidor mantiene al menos un administrador: se
niega a quitar la opción al último.

## Pro user {#user-pro}

Se muestra en la edición Pro. Solo un administrador puede cambiar la opción.

Da al usuario una plaza de la suscripción Pro. La suscripción tiene un número fijo de usuarios Pro:
cuando se alcanza, el servidor se niega a guardar la opción. Sin una suscripción activa la opción
no tiene efecto.

## Externo {#user-external}

Se muestra cuando usted es administrador. Solo se puede establecer al crear el usuario.

Un usuario externo inicia sesión mediante LDAP u OpenID Connect, no con una contraseña de Semaphore:
el campo de contraseña está deshabilitado, y se rechaza un inicio de sesión con contraseña. El
nombre y el correo electrónico se toman del proveedor de identidad en cada inicio de sesión. Deje la
casilla vacía para un usuario que inicia sesión con una contraseña.

## Change password {#user-changePassword}

Se muestra en la pestaña **Security** de un usuario existente que no es **Externo**, cuando el
servidor permite iniciar sesión con contraseña.

Abre un diálogo que establece una nueva contraseña. Para cambiar su propia contraseña, tiene que
indicar la actual; para restablecer la contraseña de otro usuario, no.

## Time-based one-time password {#user-totp}

Se muestra en la pestaña **Security** cuando el administrador ha activado la autenticación de dos
factores en el servidor.

Activa la autenticación de dos factores para este usuario. Cuando activa el interruptor, Semaphore
crea un secreto de inmediato y lo muestra como un código QR: escanéelo con una aplicación de
autenticación. A partir de entonces, un inicio de sesión pide el código de la aplicación después de
la contraseña.

Cuando el servidor permite la recuperación, se muestra un código de recuperación debajo del código
QR. Solo se muestra una vez: cópielo y guárdelo en un lugar seguro. Al desactivar el interruptor, el
secreto se elimina de inmediato.

## Proveedor LDAP {#user-ldapProvider}

Se muestra en la pestaña **Security** de su propia cuenta cuando el servidor tiene un proveedor LDAP
al que su cuenta todavía no está vinculada.

El servidor LDAP al que vincular su cuenta, cuando hay varios.

## LDAP username {#user-ldapUsername}

Se muestra con **Proveedor LDAP**.

Su nombre de usuario en el servidor LDAP. La vinculación le permite iniciar sesión en esta cuenta
también mediante LDAP.

## LDAP password {#user-ldapPassword}

Se muestra con **Proveedor LDAP**.

Su contraseña en el servidor LDAP. Semaphore la usa una vez, para comprobar que la cuenta LDAP es
suya, y no la guarda.

## Link {#user-ldapLink}

Se muestra con **Proveedor LDAP**.

Vincula su cuenta a la cuenta LDAP indicada arriba. La vinculación se rechaza cuando las
credenciales son incorrectas, cuando la cuenta LDAP está vinculada a otro usuario o cuando su cuenta
ya está vinculada a ese proveedor. Una cuenta vinculada aparece en **Linked accounts**; cierre su
chip para desvincularla.

## Global role {#global-role-select}

Se muestra en la pestaña **Roles** de un usuario existente, en la edición Pro, cuando usted puede
gestionar los roles globales.

Un rol global que dar al usuario, de entre los roles que el usuario todavía no tiene. Un rol global
concede permisos sobre todo el servidor: gestionar usuarios, roles globales o la configuración del
servidor, o leer el registro de auditoría. Los propios roles globales se crean en la pestaña
**Roles globales** de la página Usuarios.

## Assign {#globalRoleAssign-assign}

Se muestra con **Global role**.

Da al usuario el rol seleccionado de inmediato. Los roles que tiene el usuario se listan debajo,
cada uno con un botón para quitarlo. **Effective global permissions** muestra lo que el usuario
puede hacer al final, y de dónde procede cada permiso.
