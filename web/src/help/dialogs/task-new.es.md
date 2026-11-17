---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nueva Tarea

Inicia una tarea de la plantilla: una ejecución de su playbook, código de Terraform o script. La
plantilla decide qué campos muestra este diálogo, y cada sección de abajo indica cuándo aparece su
campo. El botón de la parte inferior es **Ejecutar**, **Construir** o **Desplegar**, según el tipo
de la plantilla.

Cuando la plantilla tiene una encuesta, sus preguntas son campos después de **Mensaje**. Un campo
marcado con `*` es obligatorio, un campo Integer solo admite dígitos, y las respuestas
predeterminadas vienen rellenadas. Semaphore pasa las respuestas a la tarea como variables:
variables adicionales para Ansible, `-var` para Terraform, argumentos `name=value` para un script,
o variables de entorno para una pregunta configurada así. Una respuesta sustituye a una variable con
el mismo nombre de los grupos de variables de la plantilla. Una respuesta secreta se oculta mientras
escribe y no se guarda con la tarea.

## Commit {#newTask-commit}

Se muestra cuando vuelve a ejecutar una tarea anterior, y esa tarea se ejecutó sobre un commit
conocido.

La tarea obtiene este commit del repositorio en lugar del commit más reciente de la rama. Cierre la
casilla para ejecutar el commit más reciente.

El commit solo se usa cuando la plantilla permite sobrescribir la rama, y nunca para un repositorio
marcado como monorepo. En los demás casos, la tarea ejecuta el commit más reciente.

## Versión de Construcción {#newTask-buildTaskId}

Se muestra cuando el tipo de la plantilla es **Desplegar**.

La construcción que se va a desplegar: una de las últimas 20 tareas exitosas de la plantilla de
construcción, listadas por versión y mensaje. Se selecciona la más reciente. El campo es
obligatorio.

La versión de la construcción llega a la tarea: como `semaphore_vars.task_details.incoming_version`
en Ansible, como la variable de entorno `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` en un script. La
lista está vacía cuando no tiene permiso para ver la plantilla de construcción.

## Mensaje {#newTask-message}

Opcional. Una nota sobre por qué se ejecuta la tarea, por ejemplo `Hotfix for the login page`. Se
muestra con la tarea en el historial de tareas y no cambia nada en la ejecución.

El playbook puede leerlo como `semaphore_vars.task_details.message`, y un script como la variable de
entorno `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Rama {#newTask-gitBranch}

Se muestra cuando la plantilla permite sobrescribir la rama, salvo que el repositorio sea un
monorepo.

La rama o etiqueta del repositorio que obtiene la tarea. El campo empieza con la rama de la
plantilla. Cuando está vacío, la tarea usa la rama de la plantilla, o la rama del repositorio
cuando la plantilla no tiene ninguna. Un nombre que no es un nombre de rama de Git válido se
rechaza.

## Inventario {#newTask-inventoryId}

Se muestra cuando la plantilla permite elegir el inventario en una tarea y usted puede ver el
proyecto. Para una plantilla de Terraform, OpenTofu o Terragrunt el campo se llama **Workspace**,
pero el servidor usa la elección solo para las plantillas de Ansible: una tarea de Terraform,
OpenTofu o Terragrunt se ejecuta en el workspace de la plantilla, elija lo que elija aquí.

El inventario sobre el que se ejecuta la tarea. El campo empieza con el inventario de la plantilla
y lista los inventarios del proyecto creados para la aplicación de la plantilla. Cuando el
inventario elegido ya no existe en el momento en que se inicia la tarea, la tarea se ejecuta sobre
el inventario de la plantilla.

## Límite {#taskParams-limit}

Se muestra para una plantilla de Ansible que permite sobrescribir el límite.

Los hosts o grupos del inventario sobre los que se ejecuta el playbook, un patrón por chip, por
ejemplo `web` o `10.0.0.11`. Semaphore los pasa como `--limit` unidos por comas. La lista empieza
con el límite de la plantilla. Quite todos los chips para ejecutar sobre todos los hosts del
inventario.

## Etiquetas {#taskParams-tags}

Se muestra para una plantilla de Ansible que permite sobrescribir las etiquetas.

Solo se ejecutan las tareas y los roles del playbook con una de estas etiquetas: Semaphore las pasa
como `--tags`, una etiqueta por chip. La lista empieza con las etiquetas de la plantilla. Quite
todos los chips para ejecutar el playbook completo.

## Omitir etiquetas {#taskParams-skipTags}

Se muestra para una plantilla de Ansible que permite sobrescribir las etiquetas omitidas.

Las tareas y los roles del playbook con una de estas etiquetas no se ejecutan: Semaphore las pasa
como `--skip-tags`, una etiqueta por chip. La lista empieza con las etiquetas omitidas de la
plantilla.

## Depurar {#taskParams-debug}

Se muestra para una plantilla de Ansible que permite la depuración.

Ejecuta `ansible-playbook` con una opción de verbosidad, para que el registro de la tarea muestre
más detalle. La etiqueta muestra la opción: `-v` tantas veces como indique **Nivel de depuración**.

## Nivel de depuración {#taskParams-debugLevel}

Se muestra con **Depurar**, y se habilita cuando **Depurar** está marcado.

Lo detallado que es el registro, de 1 (`-v`) a 6 (`-vvvvvv`). El valor predeterminado es 4
(`-vvvv`), que también muestra los detalles de las conexiones a los hosts.

## Ejecución en Seco {#taskParams-dryRun}

Se muestra para una plantilla de Ansible, salvo que la plantilla lo oculte.

Ejecuta el playbook con `--check`: Ansible informa de lo que cambiaría en los hosts y no cambia
nada.

## Diferencia {#taskParams-diff}

Se muestra para una plantilla de Ansible, salvo que la plantilla lo oculte.

Ejecuta el playbook con `--diff`: el registro muestra los cambios realizados en archivos y
plantillas de los hosts. Marque también **Ejecución en Seco** para ver los cambios sin hacerlos.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Se muestra para una plantilla de Ansible que permite sobrescribir la instalación de Galaxy.

Antes de que se ejecute el playbook, Semaphore instala los roles y colecciones indicados en los
archivos `requirements.yml` del repositorio, cuando un archivo ha cambiado desde la última
instalación. Marque la casilla para omitir ese paso en esta tarea.

La casilla empieza sin marcar, sea cual sea la configuración de la plantilla, y una casilla sin
marcar instala los requisitos.

## Plan {#taskParams-plan}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Ejecuta `plan` y termina la tarea ahí: no se aplica nada. Úselo para revisar primero los cambios.
Sin esta opción, la tarea continúa y aplica el plan cuando tiene cambios.

## Destroy {#taskParams-destroy}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Añade `-destroy` a `plan` y `apply`: la tarea planifica, y después aplica, la eliminación de todos
los recursos del estado del workspace. Marque también **Plan** para ver lo que se destruiría sin
destruirlo.

## Auto Approve {#taskParams-autoApprove}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Aplica el plan en cuanto se crea. Sin esta opción, la tarea se detiene después del plan y espera a
que alguien lo confirme o lo rechace en la página de la tarea.

La casilla solo tiene efecto cuando la plantilla permite la aprobación automática. Una plantilla
configurada para aprobar siempre aplica sin preguntar, diga lo que diga la casilla.

## Upgrade {#taskParams-upgrade}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Ejecuta `init` con `-upgrade`: los proveedores y módulos se actualizan a las versiones más
recientes que permiten sus restricciones de versión.

## Reconfigure {#taskParams-reconfigure}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Ejecuta `init` con `-reconfigure`: el backend se configura de nuevo, y se ignora la configuración
guardada del backend. Sin esta opción, Semaphore ejecuta `init -migrate-state`, que traslada el
estado existente a un backend que ha cambiado.

## CLI args {#newTask-args}

Se muestra cuando la plantilla permite sobrescribir los argumentos CLI.

Argumentos adicionales para la línea de comandos de la tarea, uno por chip: `--forks=10` es un
chip, `--forks` y `10` son dos. Haga clic en un chip para editarlo, arrástrelo para cambiar el
orden, y use el último chip para añadir uno.

La lista empieza con los argumentos de la plantilla. Semaphore siempre pasa los argumentos de la
plantilla y añade esta lista después de ellos, así que, en cuanto cambia la lista, los argumentos de
la plantilla que queden en ella se pasan dos veces. Quítelos y conserve solo los argumentos que
añade.
