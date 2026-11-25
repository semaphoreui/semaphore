---
form: src/components/ProjectForm.vue
---

# Nuevo Proyecto

Un proyecto reúne las plantillas de tareas, inventarios, repositorios, grupos de variables y claves
de un trabajo, y a las personas que trabajan en él. Usted se convierte en el propietario del
proyecto que crea. Solo un administrador puede crear un proyecto, salvo que el administrador del
servidor lo permita a todos los usuarios. Cada opción de abajo se puede cambiar más tarde en la
página Configuraciones del proyecto.

## Nombre del Proyecto {#newProject-name}

El nombre con el que el proyecto aparece en el menú de proyectos en la parte superior de la barra
lateral. El campo es obligatorio.

## Número máximo de tareas paralelas {#newProject-maxParallelTasks}

Cuántas tareas de este proyecto pueden ejecutarse al mismo tiempo. Una tarea iniciada por encima del
límite espera en la cola hasta que termine una de las tareas en ejecución.

Deje el campo vacío o establézcalo en `0` para no tener un límite propio. El límite de todo el
servidor, establecido por su administrador, se aplica en cualquier caso.

## ID de Chat de Telegram {#newProject-tg}

Opcional. El chat de Telegram que recibe las alertas de este proyecto. Cuando el campo está vacío,
las alertas van al chat configurado para todo el servidor.

El campo solo tiene efecto cuando el administrador ha activado las alertas de Telegram en el
servidor y las alertas están permitidas para este proyecto.

## Permitir alertas para este proyecto {#newProject-alert}

Activa las alertas para las tareas de este proyecto. El destino de las alertas lo configura para
todo el servidor su administrador: correo electrónico, Telegram, Slack, Rocket.Chat, Microsoft
Teams, DingTalk o Gotify.

- Se envía un correo electrónico cuando una tarea falla, a los miembros del proyecto que pueden ver
  la tarea y tienen las alertas activadas en su cuenta.
- Se envía un mensaje a los servicios de mensajería cuando una tarea tiene éxito, falla o espera una
  confirmación.

## Demo {#newProject-demo}

Llena el nuevo proyecto con ejemplos para probar Semaphore: el repositorio público
`https://github.com/semaphoreui/semaphore-demo.git`, inventarios, una clave con una contraseña de
vault y plantillas de tareas para Ansible, OpenTofu, Terragrunt, Bash y PowerShell, organizadas en
las vistas **Build**, **Deploy** y **Tools**.

Deje el interruptor desactivado para un proyecto sin ejemplos: solo recibe la clave **None** y la
vista **Todo**. Los ejemplos son elementos normales del proyecto: edítelos
o elimínelos como quiera.
