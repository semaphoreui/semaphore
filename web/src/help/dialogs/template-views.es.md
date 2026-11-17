---
form: src/components/EditViewsForm.vue
---

# Editar Vistas

Las vistas son las pestañas encima de la lista de plantillas de tareas. Una vista muestra las
plantillas que usted pone en ella, o todas las plantillas del proyecto. El diálogo guarda cada
cambio de inmediato: no hay un botón Guardar para todo el diálogo. Arrastre una vista por su
asa para cambiar el orden de las pestañas.

## Título {#viewsForm-title}

El nombre de la pestaña. Escríbalo y confírmelo con la marca de verificación junto al campo; la cruz
restablece el título guardado. Una vista nueva se guarda en cuanto tiene un título y usted lo
confirma o cambia uno de sus ajustes. Cerrar el diálogo no guarda un título que no haya confirmado.

El engranaje abre los ajustes de la vista debajo de su título.

## Eliminar {#viewsForm-viewDelete}

El icono de la papelera elimina la vista de inmediato, sin preguntar. Las plantillas de la vista no
se eliminan: permanecen en el proyecto y ya no están en ninguna vista.

## Type {#viewsForm-type}

Qué plantillas muestra la pestaña:

- **Custom**: las plantillas cuyo campo **Vista**, en el diálogo de la plantilla, indica esta vista.
  Es la opción predeterminada.
- **All**: todas las plantillas del proyecto, sea cual sea su vista.

Un miembro solo ve las plantillas que tiene permiso para ver, en cualquier vista.

## Sort by {#viewsForm-sortColumn}

La columna por la que se ordenan las plantillas de la pestaña; **Name** es la única. Una pestaña sin
ella también se ordena por nombre, así que el campo importa junto con **Reverse**.

## Reverse {#viewsForm-sortReverse}

Ordena la pestaña en orden descendente. Solo funciona cuando **Sort by** está establecido.

## Hidden {#viewsForm-hidden}

Oculta la pestaña a los miembros que no pueden gestionar los recursos del proyecto. Quienes sí
pueden la siguen viendo, tachada. Ocultar una pestaña no oculta sus plantillas: una vista del tipo
**All** las sigue listando.

## Agregar vista {#viewsForm-add}

Añade una vista al final de la lista, con un título vacío para escribirlo.
