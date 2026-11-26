---
form: src/components/SurveyVars.vue
---

# Variable de Encuesta

Una variable de encuesta es una pregunta que el diálogo de la tarea hace cada vez que ejecuta la
plantilla. La respuesta llega a la tarea como una variable: una variable adicional para Ansible, un
`-var` para Terraform y OpenTofu, un argumento para un script o una variable de entorno. La
variable se guarda con la plantilla, cuando guarda el diálogo de la plantilla. El mismo diálogo
edita la encuesta de un workflow, donde no se ofrece el tipo **Secret**.

## Nombre {#surveyVarDialog-name}

El nombre con el que la tarea recibe la respuesta: el nombre de la variable adicional, de la
variable de Terraform, del argumento `name=value` de un script o de la variable de entorno. Los
espacios de los extremos se eliminan. El campo es obligatorio.

Use un nombre que no tenga ninguna otra variable de la encuesta y que la herramienta acepte, por
ejemplo `app_version`. Una respuesta sustituye a una variable con el mismo nombre del grupo de
variables de la plantilla.

## Título {#surveyVarDialog-title}

La etiqueta del campo en el diálogo de la tarea, por ejemplo `Version to deploy`. También es el
texto del chip de la variable en el diálogo de la plantilla. El campo es obligatorio.

## Descripción {#surveyVarDialog-description}

Opcional. Una indicación que se muestra debajo del campo del diálogo de la tarea mientras tiene el
foco: qué introducir, o un ejemplo.

## Tipo {#surveyVarDialog-type}

Qué tipo de respuesta pide el diálogo de la tarea:

- **String**: una línea de texto. Es la opción predeterminada.
- **Integer**: solo dígitos; el diálogo de la tarea rechaza cualquier otra cosa, también un signo
  menos. La respuesta llega igualmente a la tarea como texto: conviértala en el playbook, por
  ejemplo con `| int`.
- **Secret**: texto que el diálogo de la tarea oculta. La respuesta no se guarda con la tarea, así
  que una nueva ejecución la vuelve a pedir, y Ansible la recibe en un archivo cifrado con vault. Un
  secreto no tiene valor predeterminado.
- **Enum**: un valor elegido de una lista que define más abajo.
- **Text**: varias líneas de texto.
- **Select**: varios valores elegidos de una lista que define más abajo. La tarea los recibe como
  una lista; en un argumento, un `-var` o una variable de entorno la lista se escribe como JSON, por
  ejemplo `["a","b"]`.

Cambiar el tipo vacía el valor predeterminado, y un tipo sin lista vacía los valores.

## Pass variable as {#surveyVarDialog-target}

Cómo llega la respuesta a la tarea:

- **Extra variable**: la forma habitual de la herramienta. Ansible la recibe como una variable
  adicional, Terraform y OpenTofu como `-var name=value`, un script de Bash, PowerShell o Python
  como un argumento `name=value` después del script.
- **Environment variable**: una variable de entorno del proceso, `name=value`, para cualquier tipo
  de plantilla. La variable no se pasa entonces de la otra forma.

## Clave {#surveyVarDialog-valueName}

Se muestra cuando el tipo es **Enum** o **Select**.

El texto que el diálogo de la tarea muestra para este valor en la lista. Cada valor necesita una
clave, y las claves deben ser distintas: de lo contrario, el diálogo se niega a guardar la
variable.

## Valor {#surveyVarDialog-valueValue}

Se muestra cuando el tipo es **Enum** o **Select**.

Lo que recibe la tarea cuando se elige este valor. Puede ser distinto de su clave: la clave
`Production` puede pasar `prod`.

## Add Value {#surveyVarDialog-addValue}

Se muestra cuando el tipo es **Enum** o **Select**.

Añade una fila vacía a la lista de valores. Un **Enum** o un **Select** necesita al menos un valor.
El icono de la papelera de una fila la elimina.

## Default value {#surveyVarDialog-defaultValue}

Se muestra para todos los tipos excepto **Secret**.

La respuesta que el diálogo de la tarea rellena al abrirse; puede cambiarla allí. Para un **Enum**
elija uno de los valores, para un **Select** cualquier número de ellos. Deje el campo vacío para
empezar con una respuesta vacía.

Solo el diálogo de la tarea rellena el valor predeterminado. Una tarea iniciada de otra forma, a
través de la API por ejemplo, no recibe ningún valor predeterminado.

## Requerido {#surveyVarDialog-required}

El diálogo de la tarea no inicia una tarea hasta que la variable tiene una respuesta. Al igual que
el valor predeterminado, la comprobación la hace solo el diálogo de la tarea, no el servidor.
