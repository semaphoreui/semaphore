---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Configuración del proyecto

La configuración del proyecto en su conjunto: su nombre, cuántas de sus tareas pueden ejecutarse a
la vez y si envía alertas. La Zona de Peligro debajo del formulario reúne las acciones que afectan a
todo el proyecto: su respaldo, su caché y su eliminación. Los cambios del formulario se aplican
después de **Guardar**.

## Nombre del Proyecto {#newProject-name}

El nombre con el que el proyecto aparece en el menú de proyectos en la parte superior de la barra
lateral. El campo es obligatorio.

## Número máximo de tareas paralelas {#newProject-maxParallelTasks}

Cuántas tareas de este proyecto pueden ejecutarse al mismo tiempo. Una tarea iniciada por encima del
límite espera en la cola hasta que termine una de las tareas en ejecución.

Deje el campo vacío o establézcalo en `0` para no tener un límite propio. El límite de todo el
servidor, establecido por su administrador, se aplica en cualquier caso.

## ID de Chat de Telegram {#newProject-tg}

Opcional. El chat de Telegram que recibe las alertas de este proyecto. Cuando el campo está vacío,
las alertas van al chat configurado para todo el servidor.

El campo solo tiene efecto cuando el administrador ha activado las alertas de Telegram en el
servidor y las alertas están permitidas para este proyecto.

## Permitir alertas para este proyecto {#newProject-alert}

Activa las alertas para las tareas de este proyecto. El destino de las alertas lo configura para
todo el servidor su administrador: correo electrónico, Telegram, Slack, Rocket.Chat, Microsoft
Teams, DingTalk o Gotify.

- Se envía un correo electrónico cuando una tarea falla, a los miembros del proyecto que pueden ver
  la tarea y tienen las alertas activadas en su cuenta.
- Se envía un mensaje a los servicios de mensajería cuando una tarea tiene éxito, falla o espera una
  confirmación.

Una plantilla de tarea puede impedir que sus tareas exitosas o fallidas envíen alertas.

## Test Alerts {#settings-testAlerts}

Envía una alerta de prueba por cada canal que el servidor tiene activado, para que pueda comprobar
que las alertas de este proyecto llegan.

El botón solo funciona cuando **Permitir alertas para este proyecto** está marcado y guardado.
Márquelo y pulse primero **Guardar**.

## Respaldo del Proyecto {#settings-exportProject}

Descarga un archivo JSON con la configuración del proyecto: plantillas de tareas, inventarios,
repositorios, grupos de variables, claves, programaciones, integraciones y el resto de elementos
configurados en él. El archivo puede volver a convertirse en un proyecto con **Restaurar Proyecto**
en el menú de proyectos en la parte superior de la barra lateral.

El archivo omite los secretos de las claves y los secretos de la pestaña **Secrets** de los grupos
de variables: deben introducirse de nuevo después de una restauración. Todo lo demás se copia tal
cual, también una contraseña escrita directamente en otro campo: la dirección de un repositorio que
contiene una contraseña, el texto de un inventario estático, los valores de las variables normales y
de las variables de entorno. Revise si el archivo contiene tales valores antes de compartirlo. El
historial de tareas no forma parte del respaldo.

## Limpiar caché {#settings-clearCache}

Elimina los archivos que Semaphore guarda en el servidor para este proyecto, como sus copias de los
repositorios. Se descargan de nuevo cuando se inicia la siguiente tarea.

Úselo cuando una tarea falle por una copia de un repositorio desactualizada o dañada. La acción no
se puede deshacer y no afecta a la configuración del proyecto.

## Eliminar Proyecto {#settings-deleteProject}

Elimina el proyecto con todo su contenido: plantillas de tareas, el historial de tareas,
inventarios, repositorios, grupos de variables, claves y programaciones. Semaphore pide primero una
confirmación.

Un proyecto eliminado no se puede recuperar. Descargue un respaldo antes de eliminar un proyecto que
pueda necesitar de nuevo.
