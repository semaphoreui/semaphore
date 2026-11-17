---
form: src/components/KeyForm.vue
---

# Editar Key

Una clave del Almacén de Claves contiene una credencial que el proyecto usa sin mostrarla: la clave
SSH o la contraseña con la que Ansible inicia sesión en los hosts, las credenciales de un
repositorio Git, la contraseña de un vault de Ansible, el secreto de una integración. Los
inventarios, repositorios, plantillas de tareas e integraciones hacen referencia a una clave por su
nombre. Una vez guardado, el secreto nunca se devuelve al navegador.

## Nombre de la Clave {#key-name}

El nombre con el que la clave aparece en el Almacén de Claves y en los selectores de claves de los
inventarios, repositorios, plantillas de tareas e integraciones. El campo es obligatorio. Elija un
nombre que diga qué abre la clave, por ejemplo `Deploy key of the web servers`.

## Origen {#key-source}

Dónde se guarda el secreto de la clave. Las pestañas:

- **Local**: el secreto se escribe en este diálogo y se almacena en la base de datos de Semaphore,
  cifrado cuando el administrador ha configurado una clave de cifrado para las claves de acceso. Es
  la opción predeterminada.
- **Storage**: el secreto se guarda en un almacén de secretos externo del proyecto, elegido en
  **Storage** más abajo. La pestaña solo existe en la edición Pro. Los almacenes se configuran en
  la pestaña **Storages** del Almacén de Claves.
- **Env**: el secreto se lee de una variable de entorno del proceso del servidor de Semaphore.
  Escriba el nombre de la variable en el campo debajo de las pestañas.
- **File**: el secreto se lee de un archivo en el servidor de Semaphore. Escriba la ruta absoluta
  del archivo en el campo debajo de las pestañas. El archivo debe estar dentro del directorio de
  secretos del servidor (`dirs.secrets` en su configuración, `/tmp/semaphore` por defecto); una
  ruta fuera de él, o una que contenga `..`, falla cuando se usa la clave. Se descarta un salto de
  línea final del archivo.

Con **Env** y **File**, Semaphore nunca escribe el secreto: lee el valor cada vez que una tarea usa
la clave, así que el secreto se cambia allí donde está. El valor debe corresponder al **Tipo** de
la clave. Una clave **Cadena** toma el valor tal cual. Una clave **Clave SSH** o
**Iniciar sesión con contraseña** espera un JSON con los campos de ese tipo, por ejemplo:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Al editar una clave, las pestañas están bloqueadas hasta que se marca **Sobrescribir**. Una clave
sincronizada desde un almacén de secretos conserva su origen, y su nombre y su tipo no se pueden
cambiar.

## Storage {#key-sourceStorageId}

Se muestra cuando se elige la pestaña **Storage**.

El almacén de secretos del proyecto que contiene el secreto: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager o Azure Key Vault, tal como se configuró en la pestaña **Storages** del
Almacén de Claves.

- Un almacén que no es de solo lectura recibe el secreto que escribe en este diálogo: Semaphore lo
  escribe allí al guardar. Al eliminar la clave también se elimina el secreto del almacén. Una
  clave guardada en un almacén así no se puede trasladar después a otro almacén.
- Un almacén de solo lectura solo se lee. El diálogo no muestra entonces campos para el secreto, y
  Semaphore lo lee del almacén cada vez que una tarea usa la clave.

## Source Key {#key-sourceStorageKey}

Se muestra cuando se selecciona un **Storage**.

Dónde está el secreto en ese almacén, por ejemplo la ruta al secreto en HashiCorp Vault.

## Tipo {#key-type}

Qué tipo de credencial contiene la clave. El campo es obligatorio, y los campos de debajo cambian
según el tipo.

- **Clave SSH**: una clave privada, con un usuario y una frase de contraseña opcionales. Ansible se
  conecta con ella a los hosts, y Git clona con ella un repositorio con una URL `ssh`.
- **Iniciar sesión con contraseña**: un usuario y una contraseña. Ansible se conecta con ellos a
  los hosts o se convierte con ellos en otro usuario (**Credenciales de Sudo** de un inventario);
  Git clona con ellos un repositorio `https`; la contraseña abre un vault de Ansible de una
  plantilla de tarea.
- **Cadena**: un único valor secreto sin usuario: un token, una clave de API, una contraseña. Úsela
  para el secreto de una integración. No puede ser una clave SSH, credenciales de sudo, la
  contraseña de un vault ni la clave de acceso de un repositorio.
- **Ninguno**: sin credencial. Úsela para un repositorio que no necesita inicio de sesión (un
  repositorio `https` público, una URL `file` o una ruta local) y para un inventario cuyos hosts no
  necesitan credenciales o definen ellos mismos la conexión.

## Nombre de usuario {#key-loginPasswordLogin}

Se muestra cuando el tipo es **Iniciar sesión con contraseña**. Opcional.

El nombre de usuario de la credencial.

- Ansible se conecta a los hosts como este usuario (`--user`); como **Credenciales de Sudo** de un
  inventario, es el usuario en el que se convierte Ansible (`--become-user`). Cuando el campo está
  vacío, Ansible usa el usuario que indica el inventario o su propia configuración.
- Para un repositorio `https`, Git inicia sesión con este usuario y la contraseña. Déjelo vacío
  para un token que el servidor Git acepta solo: la contraseña se envía entonces como nombre de
  usuario.

## Contraseña {#key-loginPasswordPassword}

Se muestra cuando el tipo es **Iniciar sesión con contraseña**. Obligatorio.

La contraseña de la credencial. Semaphore responde con ella a las solicitudes de contraseña de
Ansible: la contraseña SSH (`--ask-pass`) para las credenciales de usuario, la contraseña de
`become` (`--ask-become-pass`) para las credenciales de sudo y la contraseña del vault para un
vault de Ansible. Para un repositorio `https` es la contraseña o el token de acceso del servidor
Git.

## Valor {#key-stringValue}

Se muestra cuando el tipo es **Cadena**. Obligatorio.

El propio valor secreto, por ejemplo el token con el que una integración comprueba las solicitudes
que recibe. Se guarda tal como se escribe.

## Nombre de usuario {#key-sshLogin}

Se muestra cuando el tipo es **Clave SSH**. Opcional.

El usuario como el que Ansible se conecta a los hosts (`--user`). Cuando el campo está vacío,
Ansible usa el usuario que indica el inventario o su propia configuración.

Para un repositorio Git, ponga el usuario en la URL del repositorio, como en
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Se muestra cuando el tipo es **Clave SSH**. Opcional.

La frase de contraseña con la que está cifrada la clave privada. Déjela vacía para una clave sin
frase de contraseña. Una frase de contraseña incorrecta hace fallar, al cargar la clave, cada tarea
que usa la clave. Una clave generada por Semaphore no tiene frase de contraseña: lo que contenga
este campo se descarta en ese caso.

## Generate SSH Key {#key-generateSshKey}

Se muestra cuando el tipo es **Clave SSH**.

Semaphore crea un nuevo par de claves al guardar: una clave privada RSA de 2048 bits, que se queda
en la clave, y su clave pública, que se muestra una vez en un diálogo después de guardar y luego en
**Public key** cada vez que edita la clave. Añada la clave pública a `~/.ssh/authorized_keys` en
los hosts, o como clave de despliegue del repositorio Git.

El campo de la clave privada está bloqueado mientras la casilla está marcada. En una clave
existente, marque primero **Sobrescribir**: el par de claves anterior se sustituye. Una clave cuyo
secreto se lee de **Env**, **File** o de un almacén de solo lectura no se puede generar, ya que
Semaphore no tiene dónde escribirla.

## Clave Privada {#key-sshPrivateKey}

Se muestra cuando el tipo es **Clave SSH**. Obligatorio salvo que **Generate SSH Key** esté
marcado.

La clave privada, en formato OpenSSH o PEM, incluidas sus líneas `BEGIN` y `END`. Para una tarea,
Semaphore la carga en un agente SSH propio, que Ansible y Git usan mientras dura la tarea.

## Public key {#key-publicKey}

Se muestra cuando edita una clave SSH que generó Semaphore, mientras **Sobrescribir** no está
marcado.

La mitad pública del par de claves generado. **Show** la despliega, el botón de copiar la pone en
el portapapeles. Añádala a `~/.ssh/authorized_keys` en los hosts, o como clave de despliegue del
repositorio Git.

Es la clave pública del último par que generó Semaphore. Después de sustituir a mano la clave
privada con **Sobrescribir**, la clave que se muestra aquí ya no le corresponde.

## Sobrescribir {#key-overrideSecret}

Se muestra cuando edita una clave existente.

Semaphore nunca vuelve a mostrar un secreto guardado, por lo que los campos del secreto están
vacíos y bloqueados. Marque la casilla para sustituir el secreto: el origen, el tipo y los campos
del secreto se abren, y todo lo que contienen se guarda de nuevo, así que vuelva a introducir el
secreto completo, no solo la parte que cambia. Sin ella, al guardar solo cambia el nombre de la
clave.

Al guardar una clave también se eliminan las copias que el servidor de Semaphore guarda de los
repositorios que la usan: la siguiente tarea los vuelve a clonar.
