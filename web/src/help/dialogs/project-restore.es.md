---
form: src/components/RestoreProjectForm.vue
---

# Restaurar Proyecto...

Crea un nuevo proyecto a partir de un archivo de respaldo hecho con **Respaldo del Proyecto** en la
página Configuraciones de un proyecto. El nuevo proyecto recibe las plantillas de tareas,
inventarios, repositorios, grupos de variables, claves, horarios, integraciones y el resto de
elementos configurados en el respaldo, y usted se convierte en su propietario. Solo un
administrador puede restaurar un proyecto, salvo que el administrador del servidor permita a todos
los usuarios crear proyectos.

## Nombre del Proyecto {#restoreProject-name}

El nombre del nuevo proyecto. El campo es obligatorio, y sustituye al nombre escrito en el archivo
de respaldo.

El nombre debe ser distinto de los nombres de todos los proyectos del servidor: el servidor se
niega a restaurar un proyecto con un nombre que ya está en uso.

## Archivo de respaldo {#restoreProject-file}

El archivo JSON del respaldo. Elija el archivo antes de pulsar **Restore**.

Un respaldo omite los secretos de las claves y los secretos de la pestaña **Secrets** de los grupos
de variables: después de la restauración, Semaphore le indica cuántas claves están vacías y
necesitan que se vuelvan a introducir sus secretos, y los secretos de los grupos de variables
también deben introducirse de nuevo. Una contraseña escrita directamente en otro campo, como la
dirección de un repositorio, vuelve tal como está en el archivo. Los miembros del proyecto y el
historial de sus tareas no forman parte del respaldo. Un respaldo escrito por una versión más
reciente de Semaphore se rechaza.
