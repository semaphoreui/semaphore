---
form: src/components/WorkflowRunDialog.vue
---

# Run Workflow

Inicia una ejecución del workflow una vez que ha respondido a su encuesta. El diálogo solo se abre
para un workflow con encuesta: un workflow sin ella se inicia de inmediato. Después de **Ejecutar**
se abre la página de la nueva ejecución; cuando el servidor rechaza los valores, el diálogo
permanece abierto con sus respuestas.

## Variables {#workflowRunDialog-survey}

Un campo por cada variable de la encuesta del workflow, en el orden de la encuesta, con su valor
predeterminado rellenado. Un campo marcado con `*` es obligatorio, un campo Integer solo admite
dígitos, y un campo de lista admite un valor o, cuando la variable permite varios, cualquier número
de ellos. **Ejecutar** no inicia nada hasta que todos los campos son válidos.

Los valores se guardan con la ejecución y se muestran en su página en **Variables**. Una tarea del
workflow no los recibe por sí sola: una variable de su plantilla recibe un valor solo cuando su nodo
la vincula a una variable del workflow. El nombre por sí solo no entrega nada.

Una encuesta de workflow no contiene secretos, porque la ejecución guarda sus valores en claro.
Ponga las contraseñas y los tokens en una clave o en un grupo de variables.
