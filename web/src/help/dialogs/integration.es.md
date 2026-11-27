---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Editar Integración

Una integración inicia una plantilla de tarea, o un workflow, cuando otro servicio llama a una URL
de webhook de Semaphore, como un evento push de GitHub. La URL es un alias, `/api/integrations/` y
un código aleatorio: los alias del proyecto están en la página Integraciones, los alias de una
integración en su propia página. Esa página también contiene los comparadores y los valores que la
integración toma de la solicitud.

## Name {#integrationForm-name}

El nombre con el que la integración aparece en la página Integraciones. Las tareas que inicia la
integración se le atribuyen en el registro de auditoría. El campo es obligatorio.

## Runs {#integrationForm-targetToggle}

Se muestra cuando el servidor admite workflows.

Lo que inicia la integración: **Template** inicia una tarea de una plantilla de tarea, **Workflow**
inicia una ejecución de un workflow. Cambiar a **Workflow** descarta los parámetros de la tarea de
abajo.

## Workflow to run {#integrationForm-workflowId}

Se muestra cuando **Runs** es **Workflow**.

El workflow que inicia una solicitud. Cada valor que la integración toma de la solicitud se pasa al
workflow como la respuesta a su variable de encuesta con el mismo nombre. Las integraciones de
workflows requieren la edición Pro: sin ella el campo está deshabilitado y el servidor rechaza el
cambio.

## Task Template to run {#integrationForm-templateId}

Se muestra cuando **Runs** es **Template**.

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

## Auth method {#integrationForm-authMethod}

Cómo comprueba Semaphore que una solicitud procede del servicio que espera. Una solicitud que no
supera la comprobación no inicia nada, y el webhook le responde igual que a cualquier otra
solicitud.

- **None**: sin comprobación. Cualquiera que conozca la URL puede iniciar la tarea.
- **GitHub Webhooks**: el encabezado `X-Hub-Signature-256` debe contener la firma HMAC-SHA256 del
  cuerpo hecha con el secreto, tal como la envía GitHub.
- **Bitbucket Webhooks**: la misma comprobación sobre el encabezado `X-Hub-Signature`, tal como la
  envía Bitbucket.
- **Token**: el encabezado indicado en **Auth header** debe contener el propio secreto.
- **HMAC** y **HMAC (SHA-512)**: el encabezado indicado en **Auth header** debe contener el
  HMAC-SHA256 o el HMAC-SHA512 del cuerpo hecho con el secreto, en hexadecimal en minúsculas, sin un
  prefijo como `sha256=`.
- **BasicAuth**: la solicitud debe llevar el usuario y la contraseña de la clave de
  **Secreto de autenticación** en su encabezado `Authorization`.

## Auth header {#integrationForm-authHeader}

Se muestra cuando **Auth method** es **Token**, **HMAC** o **HMAC (SHA-512)**.

El nombre del encabezado de la solicitud que lleva el token o la firma, por ejemplo
`X-Webhook-Token`. El nombre no distingue entre mayúsculas y minúsculas. Una solicitud sin ese
encabezado se rechaza.

## Secreto de autenticación {#integrationForm-authSecretId}

Se muestra cuando **Auth method** no es **None**.

La clave del Almacén de Claves con la que se comprueba la solicitud. Para **BasicAuth** solo sirve
una clave del tipo **Iniciar sesión con contraseña**. Para los demás métodos también sirve una
clave del tipo **Cadena**: la comprobación usa su valor, o la contraseña de una clave
**Iniciar sesión con contraseña**.

Todos los métodos excepto **None** necesitan el campo: sin una clave, o con un secreto vacío, se
rechaza cada solicitud.
