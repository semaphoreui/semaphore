---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Editar ejecución

Una ejecución es lo que una integración inicia cuando otro servicio llama a una URL de webhook de
Semaphore, por ejemplo en un evento push de GitHub: una tarea de una plantilla de tarea o un
workflow. La URL y la comprobación de la solicitud pertenecen al endpoint de la integración, no a la
ejecución. La página de una ejecución contiene sus matchers y los valores que toma de la solicitud.

## Nombre {#integrationForm-name}

El nombre con el que la integración aparece en la página Integraciones. Las tareas que inicia la
integración se le atribuyen en el registro de auditoría. El campo es obligatorio.

## Integración {#integrationForm-groupId}

La integración a la que pertenece la ejecución. La ejecución se inicia con las solicitudes que
llegan al endpoint de esa integración y superan su autorización. El campo es obligatorio.

## Destino {#integrationForm-targetToggle}

Se muestra cuando el servidor admite workflows.

Lo que inicia la integración: **Template** inicia una tarea de una plantilla de tarea, **Workflow**
inicia una ejecución de un workflow. Cambiar a **Workflow** descarta los parámetros de la tarea de
abajo.

## Workflow to run {#integrationForm-workflowId}

Se muestra cuando **Destino** es **Workflow**.

El workflow que inicia una solicitud. Cada valor que la integración toma de la solicitud se pasa al
workflow como la respuesta a su variable de encuesta con el mismo nombre. Las integraciones de
workflows requieren la edición Pro: sin ella el campo está deshabilitado y el servidor rechaza el
cambio.

## Plantilla de tarea a ejecutar {#integrationForm-templateId}

Se muestra cuando **Destino** es **Template**.

La plantilla de tarea de la que una solicitud inicia una tarea. El campo es obligatorio. Una vez
elegida una plantilla, los parámetros de sus tareas aparecen debajo.

## Parámetros de la tarea {#integrationForm-taskParams}

Se muestra cuando se ha elegido una plantilla de tarea.

Los parámetros que recibe cada tarea que inicia esta integración, como si usted la iniciara desde
la plantilla. Aquí solo están las opciones que la plantilla permite sobrescribir en una tarea. Los
valores tomados de la solicitud tienen prioridad sobre estos: un valor del tipo **Variables** sobre
una respuesta con el mismo nombre, un valor del tipo **Task Params** sobre un parámetro con el mismo
nombre.

Cuando la plantilla permite argumentos CLI en una tarea, **CLI args** contiene argumentos
adicionales para la línea de comandos de la tarea.

## Mensaje {#taskParams-message}

Opcional. El mensaje de cada tarea que inicia la integración, que se muestra en el historial de
tareas.

## Variables de Encuesta {#taskParams-survey}

Se muestra cuando la plantilla tiene variables de encuesta.

Las respuestas a la encuesta de la plantilla, las mismas para cada tarea. Un valor del tipo
**Variables** tomado de la solicitud sustituye a la respuesta con el mismo nombre.

Las respuestas a variables del tipo secret no se guardan con la integración: pase un secreto a
través del grupo de variables de la plantilla.

## Rama {#taskParams-gitBranch}

Se muestra cuando la plantilla permite sobrescribir la rama en una tarea.

La rama o etiqueta del repositorio que obtienen las tareas. Déjela como está para usar la rama de la
plantilla. Un valor llamado `git_branch` del tipo **Task Params**, tomado de la solicitud, tiene
prioridad sobre ella; se elimina un `refs/heads/` inicial de ese valor.

## Inventario {#taskParams-inventoryId}

Se muestra cuando la plantilla permite sobrescribir el inventario en una tarea.

El inventario, o en Terraform el workspace, sobre el que se ejecutan las tareas en lugar del de la
plantilla.

## Opciones de Ansible {#taskParams-ansible}

Se muestra cuando la plantilla ejecuta un playbook de Ansible.

Las opciones de `ansible-playbook` para cada tarea, cada una se muestra cuando la plantilla la
permite: **Límite**, **Etiquetas**, **Omitir etiquetas**, **Depurar**, **Ejecución en Seco**
(`--check`), **Diferencia** y **Skip Galaxy install**.

## Opciones de Terraform {#taskParams-terraform}

Se muestra cuando la plantilla ejecuta Terraform, OpenTofu o Terragrunt.

- **Plan**: la tarea ejecuta el plan y se detiene, no se aplica nada.
- **Destroy**: el plan y la aplicación destruyen los recursos.
- **Auto Approve**: el plan se aplica sin confirmación, cuando la plantilla permite la aprobación
  automática. Sin esta opción, una tarea espera hasta que alguien la confirme en Semaphore.
- **Upgrade** y **Reconfigure**: `init` se ejecuta con `-upgrade` y `-reconfigure`.
