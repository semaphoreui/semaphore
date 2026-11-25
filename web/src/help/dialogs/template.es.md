---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Editar Plantilla 'Ansible Playbook'

Una plantilla de tarea indica qué ejecuta Semaphore y cómo: el repositorio y el archivo dentro de
él, el inventario y las variables, y qué opciones puede cambiar un usuario cuando se inicia una
tarea. Toda tarea se inicia desde una plantilla. Los campos del diálogo dependen de la aplicación de
la plantilla (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell y otras): un campo
que solo tienen algunas aplicaciones lo indica. Los cambios se aplican a las tareas iniciadas
después de **Guardar**.

## Tipo de plantilla {#templateForm-typeTabs}

Las pestañas de la parte superior definen qué tipo de plantilla es:

- **Tarea**: ejecuta el código. La mayoría de las plantillas son de este tipo.
- **Construir**: crea un artefacto. Cada tarea de construcción recibe una versión, contada a partir
  de **Versión Inicial**.
- **Desplegar**: despliega un artefacto creado por una construcción. Al iniciarla, usted elige la
  tarea de construcción cuya versión despliega.

Una tarea de construcción o de despliegue recibe en su entorno las variables `SEMAPHORE_TASK_TYPE`,
`SEMAPHORE_TASK_TARGET_VERSION` (construcción) y `SEMAPHORE_TASK_INCOMING_VERSION` (despliegue).
Ansible recibe los mismos valores en la variable adicional `semaphore_vars.task_details`.

## Versión Inicial {#templateForm-startVersion}

Se muestra cuando el tipo es **Construir**. Obligatorio.

La versión de la primera tarea de construcción, por ejemplo `1.0.0` o `v0.1`. Cada construcción
siguiente aumenta en uno el último número de la versión anterior y conserva el texto que lo rodea y
sus ceros a la izquierda: `1.0.0`, `1.0.1`, `1.0.2`. Suba aquí el número por encima de la última
construcción para saltar hacia delante, por ejemplo a `2.0.0`; una versión sin ningún número nunca
cambia.

## Plantilla de Construcción {#templateForm-buildTemplateId}

Se muestra cuando el tipo es **Desplegar**. Obligatorio.

La plantilla cuyas tareas crean los artefactos que despliega esta plantilla. La lista contiene las
plantillas de construcción del proyecto y, debajo, las plantillas de despliegue: un despliegue
puede seguir a otro despliegue, y recibe entonces la versión que ese despliegue recibió de su
construcción.

## Autorun {#templateForm-autorun}

Se muestra cuando el tipo es **Desplegar**.

Inicia por sí sola una tarea de esta plantilla cada vez que una tarea de la
**Plantilla de Construcción** tiene éxito, con la versión que creó esa tarea. Una construcción
fallida no inicia nada.

## Nombre {#templateForm-name}

El nombre con el que la plantilla aparece en la página Plantillas de Tareas. El campo es
obligatorio.

## Repositorio {#templateForm-repositoryId}

El repositorio del proyecto que contiene el código que se ejecuta. El campo es obligatorio. Antes de
cada tarea, Semaphore actualiza su copia del repositorio y ejecuta el código desde ella.

La etiqueta a la derecha del campo muestra la rama que obtienen las tareas. Haga clic en ella para
abrir el campo **Rama**.

## Rama {#templateForm-gitBranch}

Se muestra después de hacer clic en la etiqueta de la rama a la derecha de **Repositorio**.

La rama o etiqueta que obtienen las tareas de esta plantilla, en lugar de la rama definida en el
repositorio. Deje el campo vacío para usar la rama del repositorio. Cuando Semaphore puede leer las
ramas del repositorio, el campo las lista.

El campo está deshabilitado cuando el repositorio es un monorepo: todas las plantillas comparten una
única copia de él, en la rama del repositorio.

## Ruta al archivo de playbook {#templateForm-playbook}

Qué ejecutar, como una ruta desde la raíz del repositorio, por ejemplo `deploy/site.yml`. La ruta no
puede ser absoluta ni salir del repositorio. La lista ofrece los archivos del repositorio; también
puede escribir una ruta que no esté en la lista. El botón a la derecha del campo vuelve a leer la
lista.

La etiqueta y el significado cambian según la aplicación:

- **Ansible**: el playbook. Es obligatorio; la lista ofrece archivos `.yml` y `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, el directorio con el código.
  Déjelo vacío para ejecutar el código en la raíz del repositorio. Escriba `/` para listar lo que
  contiene un directorio.
- **Bash**, **Python**, **PowerShell** y otras aplicaciones: **Script Filename**, el script que se
  ejecuta. Es obligatorio. El script se ejecuta en la raíz del repositorio.

## Different working directory {#templateForm-workingDirectoryToggle}

Se muestra para las plantillas de Ansible.

Ejecuta `ansible-playbook` en un directorio distinto de la raíz del repositorio. Márquelo para
mostrar el campo **Working directory**. Desmárquelo para volver a ejecutar en la raíz.

## Working directory {#templateForm-workingDirectory}

Se muestra cuando **Different working directory** está marcado. Obligatorio.

El directorio en el que se ejecutan `ansible-playbook` y `ansible-galaxy`, como una ruta desde la
raíz del repositorio, por ejemplo `ansible`. Úselo cuando el `ansible.cfg` del proyecto está en un
subdirectorio, ya que Ansible lee el archivo del directorio actual. La ruta del playbook sigue
siendo relativa a la raíz del repositorio. La ruta no puede ser absoluta ni salir del repositorio.

## Inventario {#templateForm-inventoryId}

Se muestra para las plantillas de Ansible, Terraform, OpenTofu y Terragrunt.

- **Ansible**: el inventario sobre el que se ejecuta el playbook, con las credenciales para los
  hosts. El campo es obligatorio.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, el workspace en el que se ejecuta el
  código. Semaphore lo selecciona antes del plan y lo crea cuando no existe. Una plantilla nueva
  guardada sin workspace recibe uno llamado `default`.

La lista contiene los inventarios del proyecto creados para la aplicación de la plantilla.

## Grupos de Variables {#templateForm-environmentIds}

Opcional. Los grupos de variables cuyas variables, variables de entorno y secretos recibe cada tarea
de esta plantilla. Puede elegir varios: se combinan en el orden del campo, y un grupo posterior
gana sobre uno anterior cuando ambos definen el mismo nombre.

Cómo llegan las variables al código depende de la aplicación: variables adicionales para Ansible,
`-var` para Terraform y OpenTofu, argumentos `name=value` para un script. Las variables de entorno
se establecen para todas las aplicaciones.

## Vista {#templateForm-viewId}

Opcional. La vista en la que aparece la plantilla: las vistas son las pestañas de la página
Plantillas de Tareas, que agrupan las plantillas de un proyecto. Deje el campo vacío para no poner
la plantilla en ninguna vista.

## Etiqueta del ejecutor {#templateForm-runnerTag}

Se muestra cuando hay ejecutores del proyecto disponibles.

Las tareas de esta plantilla se ejecutan solo en un ejecutor que lleva esta etiqueta, aunque el
servidor pudiera ejecutarlas por sí mismo. La etiqueta tiene prioridad sobre la etiqueta del
ejecutor del inventario. Deje el campo vacío para usar la etiqueta del inventario, o ninguna.

## Executor image {#templateForm-executorImage}

La imagen de contenedor que el ejecutor inicia para las tareas de esta plantilla, por ejemplo
`semaphoreui/job:latest`. Deje el campo vacío para usar la imagen de la configuración del ejecutor.

Solo los ejecutores Docker y Kubernetes ejecutan las tareas en un contenedor, por lo que el campo
solo está habilitado cuando uno de ellos está disponible. Una tarea que se ejecuta sin contenedor lo
ignora.

## Variables de Encuesta {#surveyVars}

Los valores que se piden a un usuario cuando se inicia una tarea de esta plantilla. Haga clic en
**+** para añadir una variable, en un chip para cambiarla, en la cruz de un chip para quitarla;
arrastre los chips para definir el orden en que se hacen las preguntas.

Las respuestas llegan al código como las variables de un grupo de variables, o como variables de
entorno, según lo que defina cada variable. Ganan sobre un grupo de variables que define el mismo
nombre.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Permite que las tareas de esta plantilla se ejecuten al mismo tiempo. Sin esta opción, una tarea de
la plantilla espera en la cola hasta que termine la que está en ejecución. Los límites de tareas
paralelas del proyecto y del servidor se aplican en cualquier caso.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Se muestra cuando el administrador ha activado la emisión de JWT en el servidor.

Da a cada tarea de esta plantilla un JSON Web Token firmado por Semaphore, en la variable de entorno
`SEMAPHORE_JWT`. El código puede mostrar el token a otro servicio para demostrar qué tarea es: el
token indica la tarea, el proyecto, la plantilla y el usuario que la inició. El servicio comprueba
la firma con las claves públicas en `/.well-known/jwks.json` del servidor de Semaphore.

## JWT audience {#templateForm-jwtAudience}

Se muestra cuando **Issue JWT to task runner** está marcado. Opcional.

Los servicios a los que va destinado el token, escritos en su claim `aud`, por ejemplo
`https://vault.example.com`. Pulse Intro después de cada valor. Hasta 32 valores; deje el campo
vacío para un token sin audiencia.

## JWT TTL {#templateForm-jwtTtl}

Se muestra cuando **Issue JWT to task runner** está marcado. Opcional.

Cuánto tiempo sigue siendo válido el token desde el inicio de la tarea, como una duración del tipo
`30s`, `15m` o `1h`. Deje el campo vacío para usar el valor predeterminado del servidor, una hora
salvo que el administrador haya establecido otro. El valor no puede superar el máximo del servidor,
24 horas salvo que se establezca otro.

## Quiero ejecutar una tarea por cron solo para nuevos commits de algún repositorio {#templateForm-cronVisible}

Inicia una tarea de esta plantilla cuando aparece un commit nuevo en un repositorio. Semaphore
pregunta al repositorio por su último commit con el intervalo elegido e inicia una tarea solo
cuando el commit ha cambiado desde la última comprobación.

Marcar la casilla y guardar crea el horario de la comprobación; desmarcarla y guardar elimina ese
horario. Los demás horarios de la plantilla se conservan.

## Repositorio {#templateForm-cronRepositoryId}

Se muestra cuando
**Quiero ejecutar una tarea por cron solo para nuevos commits de algún repositorio** está marcado.
Obligatorio.

El repositorio que se vigila en busca de commits nuevos. Semaphore vigila la rama definida en ese
repositorio, no la **Rama** de la plantilla.

## intervaloDeVerificación {#templateForm-cronFormat}

Se muestra cuando
**Quiero ejecutar una tarea por cron solo para nuevos commits de algún repositorio** está marcado.

Con qué frecuencia comprueba Semaphore si hay un commit nuevo en el repositorio: desde cada minuto
hasta una vez al día. El valor predeterminado es cada minuto.

## Suprimir alertas de éxito {#templateForm-suppressSuccessAlerts}

Impide que las tareas de esta plantilla que tienen éxito envíen alertas. Las alertas de las tareas
fallidas se siguen enviando. Solo tiene efecto cuando las alertas están activadas para el proyecto.

## Suprimir alertas de error {#templateForm-suppressErrorAlerts}

Impide que las tareas de esta plantilla que fallan envíen alertas. Las alertas de las tareas
exitosas se siguen enviando. Solo tiene efecto cuando las alertas están activadas para el proyecto.

## CLI args {#templateForm-args}

Argumentos que se añaden al comando de cada tarea. Cada chip es un argumento, que se pasa tal cual,
sin shell: ponga `--forks=10` en un chip, o `--forks` y `10` en dos. Arrastre los chips para cambiar
su orden. El botón a la derecha de la lista abre los argumentos como un array JSON:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: se añaden a `ansible-playbook`, antes del playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: se añaden a `plan` y `apply`, no a `init`.
- Un script: se añaden después del nombre del script.

## Argumentos CLI {#templateForm-allowOverrideArgsInTask}

En **Solicitudes**. Permite al usuario añadir argumentos propios cuando se inicia una tarea. Se
añaden después de los argumentos de la plantilla.

## Rama {#templateForm-allowOverrideBranchInTask}

En **Solicitudes**. Permite al usuario elegir otra rama, etiqueta o commit cuando se inicia una
tarea, en lugar de la rama de la plantilla. Sin esta opción, la tarea siempre obtiene la rama de la
plantilla.

La casilla está deshabilitada cuando el repositorio es un monorepo.

## Inventario {#templateForm-allowOverrideInventory}

En **Solicitudes**. Se muestra para las plantillas de Ansible, Terraform, OpenTofu y Terragrunt.

Permite al usuario elegir otro inventario cuando se inicia una tarea. Solo las plantillas de Ansible
usan el inventario elegido allí: una tarea de Terraform, OpenTofu o Terragrunt se ejecuta en el
workspace de la plantilla aunque la casilla esté marcada.

## Límite {#templateForm-limit}

Se muestra para las plantillas de Ansible.

Los hosts o grupos del inventario sobre los que se ejecuta el playbook, que se pasan como
`--limit`. Cada chip es un patrón, por ejemplo `web` o `db01`; deje la lista vacía para ejecutar
sobre todos los hosts.

## Etiquetas {#templateForm-tags}

Se muestra para las plantillas de Ansible.

Ejecuta solo las tareas y los roles con estas etiquetas, que se pasan como `--tags`. Deje la lista
vacía para ejecutarlo todo.

## Omitir etiquetas {#templateForm-skipTags}

Se muestra para las plantillas de Ansible.

Omite las tareas y los roles con estas etiquetas, que se pasan como `--skip-tags`.

## Bóvedas {#templateVaults}

Se muestra para las plantillas de Ansible.

Las contraseñas de Ansible Vault que el playbook necesita para descifrar sus archivos. Haga clic en
**+** para añadir una, en un chip para cambiarla, en la cruz de un chip para quitarla. Cada bóveda
tiene un nombre, su vault ID, y da la contraseña desde una clave del Almacén de Claves o desde un
script cliente del repositorio.

## Aprobación automática {#templateForm-autoApprove}

Se muestra para las plantillas de Terraform, OpenTofu y Terragrunt.

Aplica el plan sin preguntar. Sin esta opción, una tarea se detiene después del plan y espera hasta
que un usuario la confirme o la rechace; un plan sin cambios termina la tarea sin nada que aplicar.

## Override backend settings {#templateForm-overrideBackend}

Se muestra para las plantillas de Terraform, OpenTofu y Terragrunt cuando Semaphore puede guardar
el estado de Terraform.

Guarda el estado del código en Semaphore. Antes de `init`, Semaphore escribe en el directorio del
código un archivo que define el backend `http`, y lo elimina después de la tarea:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Se muestra con **Override backend settings**; habilitado cuando esa casilla está marcada. Opcional.

El nombre del archivo en el que Semaphore escribe el backend, por ejemplo `semaphore_backend.tf`. El
valor predeterminado es `backend.tf`. El nombre debe terminar en `.tf` y contener solo letras,
dígitos, `_`, `-` y `.`. Elija un nombre que no tenga ningún archivo de su código: Semaphore
sobrescribe ese archivo y lo elimina después de la tarea.

## Límite {#templateForm-allowOverrideLimit}

En **Solicitudes de Ansible**. Permite al usuario definir los hosts sobre los que ejecutar cuando se
inicia una tarea, en lugar del **Límite** de la plantilla.

## Etiquetas {#templateForm-allowOverrideTags}

En **Solicitudes de Ansible**. Permite al usuario definir las etiquetas cuando se inicia una tarea,
en lugar de las **Etiquetas** de la plantilla.

## Omitir etiquetas {#templateForm-allowOverrideSkipTags}

En **Solicitudes de Ansible**. Permite al usuario definir las etiquetas que se omiten cuando se
inicia una tarea, en lugar de **Omitir etiquetas** de la plantilla.

## Depurar {#templateForm-allowDebug}

En **Solicitudes de Ansible**. Permite al usuario activar la salida detallada cuando se inicia una
tarea, de `-v` a `-vvvvvv`.

## Ejecución en Seco {#templateForm-showDryRun}

En **Solicitudes de Ansible**. Marcado por defecto. Ofrece la opción **Ejecución en Seco** cuando se
inicia una tarea, que ejecuta el playbook con `--check`: Ansible informa de lo que cambiaría y no
cambia nada. Desmarque la casilla para retirar la opción.

## Diferencia {#templateForm-showDiff}

En **Solicitudes de Ansible**. Marcado por defecto. Ofrece la opción **Diferencia** cuando se inicia
una tarea, que ejecuta el playbook con `--diff` para mostrar los cambios realizados en los archivos.
Desmarque la casilla para retirar la opción.

## Aprobación automática {#templateForm-allowAutoApprove}

En las solicitudes de Terraform, OpenTofu y Terragrunt. Permite al usuario marcar **Auto Approve**
cuando se inicia una tarea, para aplicar el plan sin confirmación. El diálogo de la tarea ofrece
esa opción en cualquier caso, pero sin esta casilla Semaphore la ignora: la tarea espera una
confirmación después del plan, salvo que **Aprobación automática** de la plantilla esté marcado.

## Galaxy install options {#templateForm-galaxy}

Se muestra para las plantillas de Ansible. Haga clic para mostrar las opciones de la instalación de
Galaxy; la etiqueta cuenta las opciones que difieren de los valores predeterminados.

Antes de que se ejecute el playbook, Semaphore instala con `ansible-galaxy` los roles y colecciones
que pide el repositorio. Busca `requirements.yml`, `roles/requirements.yml` y
`collections/requirements.yml` junto al playbook y en la raíz del repositorio, y ejecuta la
instalación solo cuando un archivo ha cambiado desde la última tarea.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

En **Galaxy install options**. Omite la instalación de Galaxy: el playbook se ejecuta con los roles
y colecciones instalados antes. Úselo cuando los requisitos se instalan de otra forma o la
instalación es lenta e innecesaria.

## Role install args {#templateForm-galaxyRoleArgs}

En **Galaxy install options**. Argumentos que se añaden a `ansible-galaxy role install`, uno por
chip, por ejemplo `--ignore-certs` o `--server=https://galaxy.example.com`. Semaphore solo acepta las
opciones de la instalación que se pueden pasar de forma segura: una opción desconocida, `--token`,
`--api-key`, `-p` o `--roles-path` hace que la plantilla no se pueda guardar.

## Collection install args {#templateForm-galaxyCollectionArgs}

En **Galaxy install options**. Argumentos que se añaden a `ansible-galaxy collection install`, uno
por chip, por ejemplo `--pre` o `--upgrade`. Se aplica la misma regla que para
**Role install args**: `--token`, `--api-key`, `-p` y `--collections-path` no se aceptan.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

En **Galaxy install options**, dentro de **Solicitudes**. Permite al usuario elegir, cuando se inicia
una tarea, si se omite la instalación de Galaxy. La elección de la tarea gana sobre
**Skip Galaxy install** de la plantilla.
