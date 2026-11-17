---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Editar Horario

Un horario inicia una tarea de una plantilla de tarea sin que nadie pulse **Ejecutar**: una y otra
vez según una expresión cron, o una sola vez en un momento determinado. La tarea recibe los
parámetros guardados en este diálogo. Todas las horas están en la zona horaria del servidor de
Semaphore, que se muestra junto a los campos de hora: `UTC` salvo que el administrador establezca
`schedule.timezone` o `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

El nombre con el que el horario aparece en la página Horario. El campo es obligatorio. No cambia lo
que hace el horario.

## Template {#scheduleForm-templateId}

La plantilla de tarea que inicia el horario. El campo es obligatorio. La tarea se ejecuta como si se
hubiera iniciado desde la plantilla, con el inventario, el repositorio, el grupo de variables y el
resto de la configuración propios de la plantilla, y con los parámetros de abajo allí donde la
plantilla permite que una tarea los cambie.

## Parámetros de la tarea {#scheduleForm-taskParams}

Se muestra cuando se ha seleccionado una plantilla.

Los parámetros de cada tarea que inicia el horario. Los campos que aparecen dependen de la
plantilla: sus variables de encuesta y las opciones que permite cambiar a una tarea, como la rama,
el inventario, **Límite**, **Etiquetas**, **Omitir etiquetas** o **CLI args**, y las opciones de
Ansible o de Terraform, OpenTofu y Terragrunt.

El servidor comprueba la plantilla cada vez que se activa el horario: un valor guardado aquí se
ignora en cuanto la plantilla deja de permitir que una tarea lo cambie.

## Mensaje {#taskParams-message}

Opcional. Una nota que se guarda con cada tarea que inicia el horario y se muestra en el historial
de tareas, por ejemplo `Nightly deploy`. No cambia cómo se ejecuta la tarea.

## Variables de Encuesta {#taskParams-survey}

Se muestra cuando la plantilla tiene variables de encuesta.

Los valores de las variables de encuesta de la plantilla para las tareas programadas. Llegan a la
tarea del mismo modo que cuando alguien rellena la encuesta a mano. Una variable obligatoria debe
tener un valor antes de poder guardar el horario.

Los valores predeterminados de la plantilla se copian al abrir el diálogo y se guardan con el
horario. Un valor predeterminado que se cambie más tarde en la plantilla no llega a un horario
guardado antes.

El valor de una variable del tipo secret no se guarda con el horario: la tarea programada se ejecuta
sin él.

## Rama {#taskParams-gitBranch}

Se muestra cuando la plantilla permite que una tarea sobrescriba la rama.

La rama o la etiqueta del repositorio que obtiene la tarea programada. El campo empieza con la rama
de la plantilla. El servidor usa el valor solo mientras la plantilla permita sobrescribirla.

## Inventario {#taskParams-inventoryId}

Se muestra cuando la plantilla permite que una tarea elija el inventario.

El inventario sobre el que se ejecuta la tarea programada, en lugar del inventario de la plantilla.
Para una plantilla de Terraform u OpenTofu el campo se llama **Workspace**, pero el servidor aplica
la elección solo a las plantillas de Ansible y conserva el workspace de la plantilla para las demás.

## Depurar {#taskParams-debug}

Se muestra para una plantilla de Ansible que permite la depuración.

Ejecuta `ansible-playbook` con más salida: `-v` repetido tantas veces como indique
**Nivel de depuración**.

## Nivel de depuración {#taskParams-debugLevel}

Se muestra para una plantilla de Ansible que permite la depuración; activo cuando **Depurar** está
marcado.

Lo detallada que es la salida, de `1` (`-v`) a `6` (`-vvvvvv`). El valor predeterminado es `4`.

## Ejecución en Seco {#taskParams-dryRun}

Se muestra para una plantilla de Ansible salvo que la plantilla lo oculte.

Ejecuta el playbook con `--check`: Ansible informa de lo que cambiaría y no cambia nada.

## Diferencia {#taskParams-diff}

Se muestra para una plantilla de Ansible salvo que la plantilla lo oculte.

Ejecuta el playbook con `--diff`: Ansible muestra los cambios que hace en archivos y plantillas.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Se muestra para una plantilla de Ansible que permite que una tarea lo cambie.

Omite la instalación de los roles y colecciones de los archivos `requirements.yml` antes de que se
ejecute el playbook. Márquelo cuando los requisitos ya estén instalados y la tarea no deba acceder a
Ansible Galaxy.

## Plan {#taskParams-plan}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Ejecuta solo `plan`. La tarea termina después del plan y no aplica nada.

## Destroy {#taskParams-destroy}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Añade `-destroy` al plan y a la aplicación: la tarea destruye los recursos del workspace en lugar de
crearlos o actualizarlos.

## Auto Approve {#taskParams-autoApprove}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Aplica el plan sin esperar una confirmación. Solo funciona cuando la plantilla permite la
aprobación automática; una plantilla configurada para aprobar cada tarea aplica el plan en
cualquier caso.

Sin aprobación automática, una tarea programada se detiene después del plan y espera hasta que
alguien la confirme o la rechace en la página de la tarea.

## Upgrade {#taskParams-upgrade}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Ejecuta `init` con `-upgrade`: los proveedores y módulos se actualizan a las versiones más
recientes que permite la configuración.

## Reconfigure {#taskParams-reconfigure}

Se muestra para una plantilla de Terraform, OpenTofu o Terragrunt.

Ejecuta `init` con `-reconfigure` en lugar de `-migrate-state`: el backend se configura de nuevo y
el estado existente no se migra.

## Run at {#scheduleForm-runAt}

Se muestra para un horario que se ejecuta una sola vez.

La fecha y la hora en que se inicia la tarea, en la zona horaria del servidor que se muestra al
final del campo. Un horario nuevo empieza con la siguiente hora en punto. El momento debe estar en
el futuro: el servidor rechaza una hora que ya ha pasado.

Una vez iniciada la tarea, el horario se desactiva, o se elimina cuando **Delete after run** está
marcado. Cuando Semaphore no está en funcionamiento en ese momento, la tarea no se inicia más tarde:
en el siguiente arranque el horario se desactiva, o se elimina, sin ejecutarse.

## Delete after run {#scheduleForm-deleteAfterRun}

Se muestra para un horario que se ejecuta una sola vez.

Elimina el horario después de que se haya iniciado su tarea. Déjelo sin marcar para conservar el
horario, desactivado, de modo que pueda establecer una nueva hora y volver a activarlo.

## Show cron format {#scheduleForm-rawCron}

Se muestra para un horario cron.

Alterna entre escribir la expresión cron (activado) y construirla a partir de **Timing** y las
casillas de tiempo (desactivado). Ambos editan la misma expresión. El navegador recuerda la
elección.

El interruptor queda bloqueado en activado cuando la expresión no se puede mostrar como casillas:
cuando no es válida o usa una forma como `@daily` o `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Se muestra para un horario cron cuando **Show cron format** está activado.

La expresión cron en cinco campos: minuto, hora, día del mes, mes, día de la semana (`0` es el
domingo). El campo es obligatorio, y el servidor comprueba la expresión mientras escribe. Un horario
nuevo empieza con `* * * * *`, cada minuto.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

El servidor también acepta `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` y
`@every 1h30m`, y un prefijo como `CRON_TZ=Europe/Berlin` para otra zona horaria. La vista previa
de abajo no puede mostrar la próxima ejecución de estas.

## Timing {#scheduleForm-timing}

Se muestra para un horario cron cuando **Show cron format** está desactivado.

Con qué frecuencia se ejecuta la tarea. Las casillas de abajo cambian según esta opción:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** y **Minutes**.
- **Weekly**: **Weekdays**, **Hours** y **Minutes**.
- **Monthly**: **Days**, **Hours** y **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** y **Minutes**.

## Months {#scheduleForm-months}

Se muestra cuando **Timing** es **Yearly**.

Los meses en los que se ejecuta la tarea. Marque varios para que se ejecute en cada uno de ellos.
Sin ninguno marcado, la tarea se ejecuta todos los meses.

## Weekdays {#scheduleForm-weekdays}

Se muestra cuando **Timing** es **Weekly**.

Los días de la semana en los que se ejecuta la tarea. Sin ninguno marcado, la tarea se ejecuta todos
los días.

## Days {#scheduleForm-days}

Se muestra cuando **Timing** es **Monthly** o **Yearly**.

Los días del mes en los que se ejecuta la tarea. Un mes que no tiene ese día se omite: una tarea
configurada para el día 31 no se ejecuta en abril. Sin ninguno marcado, la tarea se ejecuta todos
los días.

## Hours {#scheduleForm-hours}

Se muestra cuando **Timing** es **Daily**, **Weekly**, **Monthly** o **Yearly**.

Las horas a las que se ejecuta la tarea, en la zona horaria que se muestra a la derecha. Sin ninguna
marcada, la tarea se ejecuta cada hora.

## Minutes {#scheduleForm-minutes}

Se muestra para un horario cron cuando **Show cron format** está desactivado.

Los minutos de la hora en los que se ejecuta la tarea, en intervalos de cinco. Sin ninguno marcado,
la tarea se ejecuta cada minuto: marque al menos uno, por ejemplo `:00`, para una tarea que deba
ejecutarse una vez por hora o una vez al día.

## Next run time {#scheduleForm-nextRun}

Cuándo se iniciaría la siguiente tarea, según el horario tal como está en el formulario: en la zona
horaria del servidor y en la hora local de su navegador. Un guion significa que la expresión no es
válida, o es una que la vista previa no puede leer, como `@daily`.

## Habilitado {#scheduleForm-active}

Si el horario inicia tareas. Desmárquelo para conservar un horario sin ejecutarlo. El interruptor de
la lista de horarios hace lo mismo. Un horario nuevo está habilitado.
