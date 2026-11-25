---
form: src/components/TerraformInventoryForm.vue
---

# Nuevo Terraform Workspace

Un workspace es el inventario de una plantilla de Terraform, OpenTofu o Terragrunt: indica el
workspace que una tarea selecciona después de `init`, y la clave con la que `init` descarga los
módulos privados. El workspace que crea aquí se vincula a esta plantilla y aparece encima de su
estado. Las tareas de la plantilla se ejecutan en su workspace predeterminado, el marcado con una
marca de verificación; **Make default** lo cambia. El mismo diálogo edita un workspace cuando lo
abre con el botón del lápiz.

## Workspace name {#tfWorkspace-inventory}

El nombre del workspace de Terraform, por ejemplo `production`. El campo es obligatorio.

Antes de un plan, Semaphore ejecuta `workspace select -or-create=true` con este nombre, de modo que
el workspace se crea cuando todavía no existe. Cuando `workspace list` falla, por ejemplo porque el
backend no admite workspaces, Semaphore omite la selección.

Semaphore guarda el workspace como `TEMPLATE - NAME`, el nombre de la plantilla y este nombre: ese
nombre completo es el que ve en la página Inventario y al vincular el workspace a una plantilla.
Renombrar un workspace no lo renombra en Terraform: la siguiente tarea selecciona, o crea, un
workspace con el nuevo nombre.

## SSH key for private modules {#tfWorkspace-sshKeyId}

La clave que usa `init` para descargar módulos de repositorios Git privados por SSH. El campo es
obligatorio en este diálogo. La lista ofrece las claves del tipo **Clave SSH**, **Iniciar sesión
con contraseña** y **Ninguno** del Almacén de Claves.

Solo una **Clave SSH** tiene efecto: Semaphore la carga en un agente SSH mientras dura `init`, y Git
la usa para orígenes de módulos como `git::ssh://git@example.com/modules/vpc.git`. Una clave del
tipo **Iniciar sesión con contraseña** no pasa ninguna credencial a Git. Elija **Ninguno** cuando
los módulos no necesiten ninguna clave.
