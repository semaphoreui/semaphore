---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

Los nodos de un clúster comparten en Redis el estado del grupo de tareas: qué tareas esperan, cuáles
se ejecutan y qué nodo es propietario de cada una. Este diálogo elimina de Redis los grupos de
registros que marque, para recuperar un clúster que se ha atascado, por ejemplo con una tarea
terminada que sigue apareciendo como en ejecución. No se elimina nada de la base de datos: las
tareas conservan su estado, su registro y su historial.

Marque lo que marque, cada nodo descarta además las tareas que guarda en memoria y las vuelve a leer
de Redis y de la base de datos. Al abrirse el diálogo solo está marcado **Queue**; marque solo los
grupos que estén atascados. El resultado muestra cuántos registros se han eliminado.

## Queue {#clusterClearDialog-queue}

Las tareas que esperan a que un nodo las inicie. Las tareas eliminadas no se inician: permanecen en
la base de datos con el estado Waiting. Deténgalas o vuelva a ejecutar sus plantillas.

## Running {#clusterClearDialog-running}

Las tareas que el clúster cuenta como en ejecución. Eliminarlas no las detiene: una tarea sigue
ejecutándose en su ejecutor. Pero el clúster pierde su seguimiento: la tarea ya no cuenta para los
límites de tareas paralelas del servidor y de su ejecutor, por lo que pueden iniciarse más tareas
de las que permiten los límites, y una tarea que un ejecutor aún no ha recogido ya no se le
entrega.

## Active by project {#clusterClearDialog-active}

Las tareas sin terminar de cada proyecto. Hacen cumplir el límite de tareas paralelas de un
proyecto e impiden que se inicie una segunda tarea de una plantilla mientras una se ejecuta, cuando
la plantilla no permite tareas paralelas. Elimine este grupo cuando las tareas de un proyecto
esperan indefinidamente detrás de tareas que ya han terminado. Las tareas que se ejecutan en ese
momento dejan de contar, por lo que pueden iniciarse nuevas tareas del proyecto junto a ellas.

## Aliases {#clusterClearDialog-aliases}

Los nombres aleatorios que reciben las tareas de Terraform, OpenTofu y Terragrunt durante su vida,
asociados a la tarea. Una tarea en ejecución accede a su estado de Terraform en el servidor mediante
este nombre: después de eliminarlos, una tarea en ejecución de estas aplicaciones ya no se puede
encontrar por él.

## Claims {#clusterClearDialog-claims}

Qué nodo es propietario de cada tarea. Un nodo renueva las reclamaciones de sus tareas cada minuto;
una reclamación que no se renueva caduca a los cinco minutos. Una tarea sin reclamación cuenta como
huérfana: una tarea que aún no ha llegado a un ejecutor vuelve a la cola para otro nodo, y una tarea
que un ejecutor ya ejecuta se deja terminar allí.

## Runtime fields {#clusterClearDialog-runtimeFields}

Los datos de cada tarea que la base de datos no guarda: el usuario que la inició, la versión con la
que se inició, su alias, su ejecutor y su estado, y el proyecto de cada tarea. Sin ellos, un nodo
que no inició una tarea muestra menos información sobre ella, y el limpiador de tareas huérfanas
omite una tarea en ejecución, por lo que un registro atascado ya no se elimina por sí solo.
