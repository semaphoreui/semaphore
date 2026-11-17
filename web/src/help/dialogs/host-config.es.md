---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Una asignación indica a Semaphore qué credencial del Almacén de Claves debe usar para un host SSH o
una URL de Git. Un repositorio lleva una clave, pero una tarea también puede llegar a otros
servidores: submódulos, roles y colecciones de `requirements.yml`, módulos de Terraform,
repositorios de inventario. Las asignaciones del proyecto se aplican a todas las conexiones Git y
SSH de sus tareas, y a los comandos Git que Semaphore ejecuta para el proyecto, como la comprobación
de nuevos commits en un repositorio. No cambia ningún archivo de su repositorio.

## Tipo {#hostConfigForm-type}

Lo que coincide con la asignación:

- **Host**: un nombre de host SSH. Cada conexión SSH de una tarea a ese host usa la clave de la
  asignación: Git sobre SSH y Ansible al conectarse a un host del inventario con ese nombre. La
  asignación tiene prioridad sobre la configuración SSH del servidor para el mismo host.
- **URL**: una URL de Git o el comienzo de una. Git se conecta a cada repositorio bajo esa URL con
  la credencial de la asignación.

Cambiar a **Host** vacía una **Credential** que una asignación de host no puede usar.

## Host o URL {#hostConfigForm-host}

El campo es obligatorio. Puede existir una asignación de un tipo para cada host o URL.

Para **Host**: el nombre de host exactamente como Git o Ansible se conectan a él, por ejemplo
`github.com` o `10.0.0.21`. Solo letras, dígitos, puntos, guiones y guiones bajos: sin usuario, sin
puerto, sin comodines.

Para **URL**: una dirección que empieza por `https://` o `http://`, sin credenciales de inicio de
sesión y sin espacios, comillas ni `=`. Indique un repositorio completo, o un prefijo que termine
en `/` para abarcar un grupo de repositorios:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Cuando coinciden varias asignaciones, gana la más larga: la asignación del repositorio anterior
tiene prioridad sobre la asignación de `acme` para ese repositorio.

Con una clave SSH, una asignación de URL hace que Git se conecte por SSH al host de la URL en el
puerto `22`, sea cual sea el puerto que indica la URL.

## Credential {#hostConfigForm-sshKeyId}

La clave del Almacén de Claves que se usa para las conexiones que coinciden con la asignación. El
campo es obligatorio.

- **Host** acepta solo una clave del tipo **Clave SSH**. Cuando la clave tiene un usuario, SSH se
  conecta como ese usuario.
- **URL** acepta una clave del tipo **Clave SSH** o **Iniciar sesión con contraseña**. Una clave SSH se conecta
  con su usuario, o como `git` cuando no tiene ninguno. Una clave de inicio de sesión con contraseña
  solo funciona con una URL `https://`; para un token, ponga el token en la contraseña y deje vacío
  el usuario.

Una clave de inicio de sesión con contraseña llega a Git a través de su entorno, no de su línea de
comandos, y el registro de la tarea muestra solo la URL original.
