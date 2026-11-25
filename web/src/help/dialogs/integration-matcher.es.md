---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Un comparador es una condición sobre la solicitud entrante. Los comparadores deciden qué
integración inicia una solicitud al alias del proyecto, de modo que una URL puede servir a varias
integraciones: una integración se ejecuta solo cuando se cumplen todos sus comparadores. Solo se
usan cuando **Usar alias del proyecto** está activado para la integración; con esta opción
activada, a la integración solo se llega a través del alias del proyecto.

Una integración encontrada por el alias del proyecto sin comparadores nunca se ejecuta: añada al
menos uno.

## Name {#integrationMatcherForm-name}

El nombre con el que el comparador aparece en la página de la integración. No cambia lo que
comprueba el comparador. El campo es obligatorio.

## Match on {#integrationMatcherForm-matchType}

La parte de la solicitud que lee el comparador:

- **Header**: el valor de un encabezado de la solicitud.
- **Body**: el cuerpo de la solicitud.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Se muestra cuando **Match on** es **Body**.

Cómo se lee el cuerpo:

- **JSON**: el cuerpo se analiza como JSON, y **Key** es la ruta a un valor dentro de él.
- **String**: todo el cuerpo se compara como texto, y **Key** no se usa.

## Key {#integrationMatcherForm-key}

Qué leer de la solicitud. El campo es obligatorio.

- Para **Header**: el nombre del encabezado, por ejemplo `X-GitHub-Event`. No distingue entre
  mayúsculas y minúsculas.
- Para **Body** en **JSON**: la ruta al valor, con sus partes unidas por puntos; un elemento de un
  array se escribe como `.[0]`, por ejemplo `repository.full_name` o `commits.[0].id`.
- Para **Body** en **String**: no se usa, pero el formulario necesita un valor; escriba cualquier
  cosa.

## Comparison Method {#integrationMatcherForm-method}

Cómo se compara el valor leído de la solicitud con **Value**:

- **==**: los dos son iguales.
- **!=**: los dos son distintos. Un encabezado o una clave que falta en la solicitud cuenta como
  distinto.
- **Contains**: el valor leído de la solicitud contiene **Value**.

## Value {#integrationMatcherForm-value}

El texto con el que comparar, por ejemplo `push` o `refs/heads/main`. La comparación es exacta y
distingue entre mayúsculas y minúsculas. Un número o un booleano del cuerpo JSON se compara en su
forma de texto: `42`, `true`. El campo es obligatorio.
