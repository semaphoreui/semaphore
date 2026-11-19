---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Un valor extraído toma una parte de la solicitud entrante, como la rama de un evento push, y la pasa
a la tarea que inicia la integración. Cada valor de la integración se toma de cada solicitud que
inicia una tarea.

## Name {#integrationValueForm-name}

El nombre con el que el valor aparece en la página de la integración. No cambia lo que se toma ni a
dónde va. El campo es obligatorio.

## Source of the Value {#integrationValueForm-valueSource}

La parte de la solicitud de la que se toma el valor:

- **Header**: el valor de un encabezado de la solicitud.
- **Body**: el cuerpo de la solicitud.

## Data Type of Body {#integrationValueForm-bodyDataType}

Se muestra cuando **Source of the Value** es **Body**.

Cómo se lee el cuerpo:

- **JSON**: el cuerpo se analiza como JSON, y **Key** es la ruta a un valor dentro de él. El valor
  conserva su tipo JSON: un número, una lista o un objeto llegan a la tarea como tales.
- **String**: todo el cuerpo se toma como texto, y **Key** no se usa.

## Key {#integrationValueForm-key}

Qué tomar de la solicitud. El campo es obligatorio.

- Para **Header**: el nombre del encabezado, por ejemplo `X-GitHub-Event`. No distingue entre
  mayúsculas y minúsculas. Un encabezado que falta en la solicitud da un valor vacío.
- Para **Body** en **JSON**: la ruta al valor, con sus partes unidas por puntos; un elemento de un
  array se escribe como `.[0]`, por ejemplo `repository.full_name` o `commits.[0].id`. Una ruta que
  falta en el cuerpo no da ningún valor, y la tarea conserva el definido en la integración.
- Para **Body** en **String**: no se usa, pero el formulario necesita un valor; escriba cualquier
  cosa.

## Variable Usage {#integrationValueForm-variableType}

A dónde va el valor en una tarea de una plantilla:

- **Variables**: una variable adicional de la tarea, del mismo modo en que se pasa la respuesta a
  una variable de encuesta. Sustituye a una respuesta o a una variable del grupo de variables con el
  mismo nombre.
- **Task Params**: un parámetro de la tarea, del mismo modo en que se guardan las opciones del
  diálogo de la tarea. El nombre `git_branch` establece la rama de la tarea cuando la plantilla
  permite sobrescribir la rama; se elimina un `refs/heads/` inicial.

Una integración que ejecuta un workflow ignora el uso: cada valor se convierte en la respuesta a la
variable de encuesta del workflow indicada en **Variable**.

## Variable {#integrationValueForm-variable}

El nombre con el que se pasa el valor, por ejemplo `branch` para una variable adicional que el
playbook lee como `{{ branch }}`. El campo es obligatorio.
