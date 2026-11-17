---
form: src/components/AppForm.vue
---

# Edit App

Una aplicación es una herramienta con la que Semaphore ejecuta tareas: Ansible, Terraform,
OpenTofu, Terragrunt, Bash, Python, PowerShell o un comando propio. Cada plantilla de tarea se crea
para una aplicación, y este diálogo define cómo la inicia el servidor y cómo se muestra la
aplicación en la interfaz.

La configuración se guarda en la base de datos y se aplica de inmediato en el servidor que la
guarda; los demás nodos de un clúster la leen cuando se reinician. Un campo que vacíe conserva su
valor anterior hasta que el servidor se reinicie.

## ID {#appForm-id}

Obligatorio. La clave con la que se guarda la aplicación; una plantilla de tarea hace referencia a
su aplicación mediante esta clave. Las aplicaciones integradas son `ansible`, `terraform`, `tofu`,
`terragrunt`, `pulumi`, `bash`, `python` y `powershell`. Cualquier otro ID añade una aplicación
propia: Semaphore la ejecuta como un script, con el comando indicado en **Path** o, cuando **Path**
está vacío, con el comando que se llama como el ID.

Use un ID sin puntos. Cambiar el ID de una aplicación existente no la renombra: la configuración se
guarda como una aplicación nueva con el nuevo ID, y la aplicación anterior se mantiene.

## Icon {#appForm-icon}

Opcional. El nombre de un icono del conjunto Material Design Icons sin el prefijo `mdi-`, por
ejemplo `language-python`. Déjelo vacío para conservar el icono integrado de la aplicación;
una aplicación propia muestra entonces un signo de interrogación.

## Nombre {#appForm-title}

Opcional. El nombre de la aplicación en la lista de aplicaciones y en el menú **Nueva plantilla**.
Déjelo vacío para conservar el nombre integrado, como `Ansible Playbook`; una aplicación propia no
tiene entonces nombre, así que asígnele uno.

## Path {#appForm-path}

Opcional. El comando que Semaphore ejecuta en lugar del predeterminado, ya sea un nombre que se
busca en el `PATH` del servidor o una ruta completa, por ejemplo `/usr/local/bin/terraform-1.9`.
Los predeterminados son `terraform`, `tofu`, `terragrunt`, `bash`, `python3` y `powershell`; una
aplicación propia ejecuta el comando que se llama como su ID.

Ansible ignora este campo: Semaphore siempre ejecuta `ansible-playbook` y `ansible-galaxy`.

## Priority {#appForm-priority}

La posición de la aplicación en la lista de aplicaciones y en el menú **Nueva plantilla**: una
aplicación con un número mayor aparece primero. Las aplicaciones integradas van desde 1000 para
Ansible hasta 500 para Python. Una aplicación integrada con una prioridad de `0` o menor recupera
su prioridad integrada cuando se inicia el servidor.

## Argumentos {#appForm-args}

Opcional. Argumentos que se colocan en la línea de comandos justo después del comando, uno por
chip. Haga clic en **+** para añadir uno, haga clic en un chip para cambiarlo y arrastre los chips
para cambiar su orden.

- Terraform, OpenTofu y Terragrunt: los argumentos van antes de los argumentos que Semaphore añade
  a cada comando, como `init` y `apply`.
- Aplicaciones de script: los argumentos van antes del script de la plantilla y sustituyen a los
  predeterminados: PowerShell se ejecuta con `-File` por defecto, así que ponga `-File` al final de
  su lista.
- Ansible ignora este campo.

## Active {#appForm-active}

Una aplicación activa se ofrece en el menú **Nueva plantilla**. Desmarque la casilla para ocultar
la aplicación de ese menú: las plantillas ya creadas con ella conservan su aplicación y se siguen
ejecutando.

La casilla está desmarcada para una aplicación nueva: márquela para que la aplicación esté
disponible.
