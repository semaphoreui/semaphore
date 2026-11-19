---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Editar Grupo de Variables

Un grupo de variables contiene las variables adicionales, las variables de entorno y los secretos
que Semaphore pasa a cada tarea de una plantilla de tarea que usa el grupo. Ansible recibe las
variables adicionales como `--extra-vars`, Terraform y OpenTofu como `-var`; las variables de
entorno se establecen para el proceso de la tarea. Una plantilla puede usar varios grupos: cuando
dos de ellos definen el mismo nombre, gana el grupo que aparece después en la plantilla.

## Nombre del Grupo {#varGroup-name}

El nombre con el que el grupo aparece en la página Grupos de Variables y en el selector de grupo de
variables de una plantilla de tarea. El campo es obligatorio.

## Secret storage {#varGroup-secretStorageId}

Se muestra cuando el grupo es nuevo y el servidor admite almacenes de secretos. Opcional.

El almacén de secretos en el que se guardan los secretos de este grupo, por ejemplo un HashiCorp
Vault. Deje el campo vacío para guardar los secretos en la base de datos de Semaphore, cifrados.

El almacén se elige una sola vez, al crear el grupo: un grupo existente conserva su almacén, y el
campo ya no se muestra. Un almacén marcado como **Read only** no acepta secretos nuevos: guardar un
secreto en él falla.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Se muestra cuando el grupo es nuevo y el servidor admite almacenes de secretos. Disponible una vez
elegido un **Secret storage**.

El comienzo de la ruta que recibe cada secreto de este grupo en el almacén. Semaphore le añade 10
caracteres aleatorios: con el prefijo `semaphore/prod/` un secreto se guarda como `semaphore/prod/`
seguido de esos caracteres. El prefijo se toma tal cual, así que termínelo en `/` cuando los
secretos deban estar debajo de él como en una carpeta.

El prefijo se establece una sola vez, al crear el grupo. La pestaña **Secrets** lo muestra como el
patrón de la ruta de origen.

## Tabla, JSON o YAML {#varGroup-extraVarsMode}

Cómo se editan las variables adicionales:

- **Table**: una fila por variable, cada una con un tipo. Añada una fila con el botón **+** junto al
  interruptor.
- **JSON**: todas las variables como un único objeto JSON.
- **YAML**: todas las variables como un único mapeo YAML.

El diálogo se abre en **Table**. Un cambio de modo convierte lo que ha escrito; cuando no se puede
leer, el diálogo permanece en el modo en el que estaba y muestra el error. Sea cual sea el modo,
las variables se guardan como un único objeto de nombres y valores.

## Variables adicionales en JSON {#varGroup-extraVarsJson}

Se muestra cuando las variables adicionales se editan como **JSON**.

Las variables adicionales como un único objeto JSON. Un nombre no puede estar vacío; un valor puede
ser una cadena, un número, una lista o un objeto:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Variables adicionales en YAML {#varGroup-extraVarsYaml}

Se muestra cuando las variables adicionales se editan como **YAML**.

Las variables adicionales como un único mapeo YAML. El texto se lee con los tipos que conoce JSON:
una fecha sin comillas como `2024-01-01` sigue siendo una cadena, y `.inf` o `.nan` se rechazan.
Una lista o un escalar en el nivel superior también se rechazan.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Nombre de la variable {#varGroup-varName}

Se muestra cuando las variables adicionales se editan como **Table**, en cada fila.

El nombre de la variable tal como lo ve el playbook o el código de Terraform, por ejemplo
`app_version`.

## Tipo de variable {#varGroup-varType}

Se muestra cuando las variables adicionales se editan como **Table**, en cada fila.

Cómo se lee el valor de la fila:

- **String**: el texto tal cual.
- **Number**: un número, por ejemplo `42`.
- **List**: un array JSON, por ejemplo `["a", "b"]`.
- **Dict**: un objeto JSON, por ejemplo `{"key": "value"}`.

Un valor que no corresponde a su tipo detiene el guardado con el nombre de la variable.

## Valor de la variable {#varGroup-varValue}

Se muestra cuando las variables adicionales se editan como **Table**, en cada fila.

El valor de la variable, en el formato de su tipo. Para una **List** o un **Dict**, el botón al
final del campo abre un editor más grande.

Semaphore pasa los valores a Ansible como datos: una expresión `{{ … }}` dentro de un valor no se
evalúa, el playbook recibe el texto tal cual.

## Nombre de la variable de entorno {#varGroup-envName}

En la pestaña **Variables**, en **Variables de entorno**, en cada fila. Añada una fila con el botón
**+** de la sección.

El nombre de una variable establecida en el entorno del proceso que ejecuta la tarea, por ejemplo
`AWS_REGION`.

## Valor de la variable de entorno {#varGroup-envValue}

En la pestaña **Variables**, en **Variables de entorno**, en cada fila.

El valor de la variable, como texto. Se guarda como texto sin cifrar y se muestra a todos los que
abren el grupo: ponga las contraseñas y los tokens en la pestaña **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Se muestra en la pestaña **Secrets** de un grupo que tiene un almacén de secretos.

Actívelo para importar secretos del almacén a este grupo, desde las rutas definidas en
**Sync paths**. La importación se ejecuta cuando pulsa el botón de sincronización del grupo en la
página Grupos de Variables, y por sí sola con el intervalo definido en **Sync paths**. Desactivado,
no se importa nada.

## Sync paths {#varGroup-syncPaths}

Se muestra cuando **Sync keys enabled** está activado.

Abre el diálogo con las rutas del almacén desde las que importar y el intervalo de la importación
automática. El número del botón es el número de rutas. El botón de sincronización del grupo en la
página Grupos de Variables permanece deshabilitado hasta que el grupo tenga al menos una ruta.

El diálogo solo tiene **Cerrar**: lo que cambie en él se guarda con el grupo, mediante **Guardar**
de este diálogo.

## Nombre de la variable secreta {#varGroup-secretVarName}

En la pestaña **Secrets**, en **Variables adicionales**, en cada fila. Añada una fila con el botón
**+** de la sección.

El nombre de una variable adicional cuyo valor es un secreto. La tarea la recibe del mismo modo que
las variables adicionales de la pestaña **Variables**; un secreto gana sobre una variable adicional
con el mismo nombre. Ansible recibe los secretos en un archivo cifrado con Ansible Vault, Terraform
y OpenTofu como `-var`. Como advierte el diálogo, un secreto pasado de esta forma puede aparecer
igualmente en texto sin cifrar en el registro de Ansible.

## Valor de la variable secreta {#varGroup-secretVarValue}

En la pestaña **Secrets**, en **Variables adicionales**, en cada fila.

El valor del secreto. Semaphore lo cifra, o lo escribe en el almacén de secretos del grupo, y nunca
lo vuelve a mostrar: cuando abre el grupo más tarde, el campo está vacío. Déjelo vacío para
conservar el valor guardado, escriba uno nuevo para sustituirlo. Al eliminar la fila se elimina el
secreto cuando guarda el grupo.

## Nombre de la variable de entorno secreta {#varGroup-secretEnvName}

En la pestaña **Secrets**, en **Variables de entorno**, en cada fila. Añada una fila con el botón
**+** de la sección.

El nombre de una variable de entorno cuyo valor es un secreto, por ejemplo `DB_PASSWORD`. Se
establece en el entorno del proceso que ejecuta la tarea, igual que las variables de entorno de la
pestaña **Variables**.

## Valor de la variable de entorno secreta {#varGroup-secretEnvValue}

En la pestaña **Secrets**, en **Variables de entorno**, en cada fila.

El valor del secreto. Como para una variable adicional secreta, nunca se vuelve a mostrar: deje el
campo vacío para conservar el valor guardado, escriba uno nuevo para sustituirlo.
