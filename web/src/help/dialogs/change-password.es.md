---
form: src/components/ChangePasswordForm.vue
---

# Cambiar contraseña

Establece una nueva contraseña para iniciar sesión en Semaphore con un nombre de usuario y una
contraseña. La nueva contraseña se aplica cuando pulsa **Guardar**. Un usuario externo, que inicia
sesión mediante LDAP u OpenID Connect, no tiene aquí ninguna contraseña que cambiar.

## Current password {#changePassword-currentPassword}

Se muestra cuando cambia su propia contraseña.

Su contraseña actual. El campo es obligatorio: Semaphore la comprueba antes de establecer la nueva,
para que una sesión o un token de API robados no puedan convertirse en su contraseña. El servidor
rechaza el cambio cuando la contraseña no coincide.

Un administrador o un gestor de usuarios que restablece la contraseña de otro usuario no ve este
campo.

## Contraseña {#changePassword-password}

La nueva contraseña. El campo es obligatorio. Haga clic en el icono del ojo para mostrar lo que ha
escrito.

Semaphore guarda solo un hash de la contraseña.
