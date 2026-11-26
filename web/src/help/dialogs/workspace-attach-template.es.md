---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Vincula este workspace a una plantilla, de modo que aparezca entre los workspaces de la plantilla
con su estado y sus alias. Un workspace pertenece a una sola plantilla a la vez: una vez vinculado,
sale de la página Inventario y se gestiona en la plantilla. Vincularlo no lo convierte en el
workspace predeterminado de la plantilla: para ello, abra la plantilla y pulse **Make default**.

## Template {#inventoryAttach-templateId}

La plantilla a la que vincular el workspace. El campo es obligatorio. La lista muestra las
plantillas del proyecto creadas para la misma aplicación que el workspace, Terraform u OpenTofu.

El workspace conserva su nombre, su clave y su estado. Para trasladarlo más tarde a otra plantilla,
pulse **Detach** en la plantilla a la que está vinculado: vuelve a la página Inventario. El
workspace predeterminado de una plantilla no se puede desvincular.
