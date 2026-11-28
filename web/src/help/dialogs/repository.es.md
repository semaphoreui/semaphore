---
form: src/components/RepositoryForm.vue
---

# Editar Repository

Un repositorio es de donde proceden los playbooks, el código de Terraform o los scripts de una
plantilla de tarea: un repositorio Git, o un directorio en la máquina que ejecuta las tareas. Antes
de cada tarea, Semaphore clona el repositorio Git, o actualiza la copia que clonó antes, obtiene la
rama y ejecuta la tarea en esa copia.

## Nombre {#repository-name}

El nombre con el que el repositorio aparece en la página Repositorios y en el selector de
repositorio de una plantilla de tarea y de un inventario. El campo es obligatorio.

## URL o ruta {#repository-gitUrl}

Dónde está el código. El campo es obligatorio. Los chips debajo muestran el tipo de dirección que
ha escrito, y un clic en un chip reescribe la dirección a ese tipo.

- **ssh**: `git@github.com:org/repo.git` o `ssh://git@github.com/org/repo.git`. Git inicia sesión
  con el usuario de la dirección y la **Clave SSH** elegida en **Clave de Acceso**.
- **https** y **http**: `https://github.com/org/repo.git`. Un repositorio privado necesita una
  clave **Iniciar sesión con contraseña**; uno público necesita **Ninguno**. Evite `http`: el
  usuario y la contraseña viajan entonces sin cifrar.
- **git**: `git://host/org/repo.git`, el protocolo Git de solo lectura, que no tiene inicio de
  sesión.
- **file**: `file:///srv/git/repo.git`, un repositorio Git en el disco de la máquina que ejecuta la
  tarea, clonado como uno remoto.
- **ruta abs.**: una ruta absoluta como `/srv/playbooks` o `D:\playbooks`. Semaphore no la clona:
  la tarea se ejecuta directamente en ese directorio, que debe existir en la máquina que ejecuta la
  tarea (el servidor de Semaphore o un ejecutor). La rama no se aplica.

Al cambiar la dirección se eliminan las copias que el servidor de Semaphore clonó de la anterior.

## Rama {#repository-gitBranch}

La rama que Semaphore clona y mantiene actualizada, por ejemplo `main`. El campo es obligatorio,
salvo para una ruta absoluta, para la que está bloqueado.

Una plantilla de tarea puede indicar una rama propia, que tiene prioridad sobre esta, y una
plantilla puede permitir que sus tareas elijan una rama al iniciarse. Una etiqueta también
funciona, pero no se puede hacer pull de una etiqueta, así que Semaphore vuelve a clonar el
repositorio antes de cada tarea.

## Single checkout for all templates {#repository-monorepo}

Desactivado por defecto: cada plantilla de tarea recibe su propia copia del repositorio. Marque la
casilla para clonar el repositorio una sola vez y compartir esa copia entre todas las plantillas que
lo usan, para un repositorio demasiado grande para mantener una copia por plantilla.

La copia compartida permanece en la rama de este repositorio: las plantillas que la usan no pueden
tener una rama propia, y sus tareas no pueden elegir otra rama ni otro commit. Bloqueado para una
ruta absoluta.

## Clave de Acceso {#repository-sshKeyId}

La clave del Almacén de Claves con la que Git inicia sesión en el repositorio. El campo es
obligatorio. Ofrece claves de tres tipos; cuál elegir depende de la URL:

- **Clave SSH** para una URL `ssh`. Semaphore carga la clave privada en un agente SSH mientras dura
  la clonación o la actualización. El usuario es el de la URL, como `git` en
  `git@github.com:org/repo.git`.
- **Iniciar sesión con contraseña** para una URL `https`: el usuario y la contraseña, o el token de
  acceso del servidor Git. Con un usuario vacío, la contraseña se envía sola, como nombre de
  usuario, que es lo que algunos servidores esperan de un token. Las mismas credenciales se pasan a
  `ansible-galaxy` para los roles y colecciones del mismo servidor.
- **Ninguno** para un repositorio público, una URL `git` o `file`, o una ruta absoluta.

Una clave que no corresponde a la URL no da a Git ninguna credencial, y la clonación de un
repositorio privado falla entonces.
