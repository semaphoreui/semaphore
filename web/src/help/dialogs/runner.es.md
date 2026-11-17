---
form: src/components/RunnerForm.vue
---

# Editar Ejecutor

Un ejecutor es un proceso `semaphore runner` independiente que ejecuta tareas en su propia máquina
en lugar de en el servidor de Semaphore. Una tarea va a un ejecutor cuando el servidor está
configurado para ejecutar las tareas en ejecutores, o cuando la plantilla de tarea o su inventario
tiene una etiqueta del ejecutor. El servidor elige entonces un ejecutor que esté habilitado,
registrado, en línea y que no haya alcanzado su límite de tareas paralelas: primero los ejecutores
del proyecto, después los ejecutores globales, en orden aleatorio dentro de cada grupo. Cuando todos
los ejecutores que coinciden están desconectados u ocupados, la tarea espera en la cola; cuando no
coincide ningún ejecutor, la tarea falla.

## Nombre {#runner-name}

Obligatorio. El nombre con el que el ejecutor aparece en la página Ejecutores, por ejemplo el host
en el que se ejecuta. No cambia qué tareas recibe el ejecutor.

## Tags {#runner-tags}

Disponible en Semaphore Pro; en las demás ediciones el campo está bloqueado.

Las etiquetas que eligen las tareas de este ejecutor. Una tarea con una etiqueta del ejecutor va
solo a los ejecutores que llevan esa etiqueta; la etiqueta de la plantilla de tarea tiene prioridad
sobre la etiqueta del inventario. Escriba una etiqueta y pulse Intro, o elija una usada por otros
ejecutores. Las etiquetas vacías y repetidas se descartan al guardar.

Un ejecutor del proyecto necesita al menos una etiqueta salvo que **Is default** esté marcado.

## Is default {#runner-isDefault}

Una tarea sin etiqueta del ejecutor va solo a los ejecutores marcados como predeterminados.
Desmarque la casilla para un ejecutor que solo deba recibir tareas con una de sus etiquetas.

La casilla está marcada para un nuevo ejecutor global y desmarcada para un nuevo ejecutor del
proyecto.

## Register {#runner-registered}

Se muestra cuando crea un ejecutor.

- Marcada (la opción predeterminada): el servidor crea el ejecutor con su token. Después de
  **Crear**, el diálogo muestra el token y los comandos que inician el ejecutor con él.
- Desmarcada: el ejecutor se crea sin ningún token. Más tarde, haga clic en **Regenerate
  registration token** (el icono de dos flechas) en la fila del ejecutor para obtener un token de
  registro de un solo uso, válido durante una hora, y registre el ejecutor con
  `semaphore runner register`.

Un ejecutor no registrado no recibe tareas.

## Webhook {#runner-webhook}

Opcional. Una URL a la que llama el servidor cuando entrega una tarea a este ejecutor, por ejemplo
para iniciar el ejecutor bajo demanda. El servidor envía una solicitud `POST` con un cuerpo JSON:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

La URL debe responder con el estado 200 o 204 en 30 segundos; de lo contrario, la tarea falla. Un
ejecutor con un webhook siempre se considera en línea, aunque no haya consultado al servidor
recientemente. Deje el campo vacío para un ejecutor que funciona todo el tiempo.

## Número máximo de tareas paralelas {#runner-maxParallelTasks}

El número de tareas que el ejecutor puede ejecutar a la vez. Cuando se alcanza el límite, la
siguiente tarea va a otro ejecutor que coincida o espera en la cola. Deje el campo vacío o
establezca `0` para no tener límite.

## Habilitado {#runner-active}

Solo un ejecutor habilitado recibe tareas nuevas. Desmarque la casilla para dejar el ejecutor fuera
de servicio sin eliminarlo; el interruptor de la lista de ejecutores hace lo mismo.

La casilla está desmarcada para un ejecutor nuevo: márquela, o habilite el ejecutor más tarde en la
lista.
