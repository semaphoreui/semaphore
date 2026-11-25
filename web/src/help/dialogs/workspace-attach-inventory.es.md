---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Vincula a esta plantilla un workspace que ya existe en el proyecto, de modo que aparezca entre los
workspaces de la plantilla con su estado y sus alias. Un workspace pertenece a una sola plantilla a
la vez. Vincularlo no lo convierte en el workspace predeterminado: las tareas se siguen ejecutando
en el workspace predeterminado hasta que pulse **Make default**.

## Workspace {#workspaceSelect-inventoryId}

El workspace que se va a vincular. El campo es obligatorio. La lista muestra los workspaces del
proyecto creados para la aplicación de esta plantilla (Terraform, OpenTofu o Terragrunt) que no
están vinculados a ninguna plantilla; un workspace desvinculado de su plantilla con **Detach**
vuelve a aparecer aquí.

El workspace conserva su nombre, su clave y su estado. El servidor rechaza un workspace que otra
plantilla haya tomado mientras tanto.
