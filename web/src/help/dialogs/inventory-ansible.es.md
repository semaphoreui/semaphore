---
form: src/components/InventoryForm.vue
---

# Editar Ansible Inventory

Un inventario indica a Ansible en qué hosts ejecutar un playbook y contiene las credenciales que
Semaphore usa para conectarse a ellos. Una plantilla de tarea hace referencia a un inventario, y
cada tarea de esa plantilla se ejecuta sobre él. El formato de un inventario se describe en la
[guía de inventarios de Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Nombre {#inventoryForm-name}

El nombre con el que el inventario aparece en la página Inventario y en el selector de inventario de
una plantilla de tarea. Elija un nombre que distinga los hosts, por ejemplo `Production web servers`.

## Etiqueta del ejecutor {#inventoryForm-runnerTag}

Se muestra cuando hay ejecutores del proyecto disponibles.

Las tareas que usan este inventario se ejecutan en un ejecutor que tiene esta etiqueta. Una etiqueta
del ejecutor definida en la plantilla de tarea tiene prioridad sobre la etiqueta del inventario.
Deje el campo vacío para que cualquier ejecutor pueda tomar las tareas.

## Credenciales de Usuario {#inventoryForm-sshKeyId}

La clave del Almacén de Claves con la que Ansible inicia sesión en los hosts. El campo es
obligatorio y acepta tres tipos de claves:

- **Clave SSH**: Ansible se conecta con la clave privada y, si la clave tiene uno, con su usuario.
- **Iniciar sesión con contraseña**: Ansible se conecta con el usuario y la contraseña.
- **Ninguno**: Semaphore no pasa ninguna credencial. Úselo para hosts que no las necesitan, como
  `localhost`, o cuando el propio inventario define cómo conectarse.

## Credenciales de Sudo {#inventoryForm-becomeKeyId}

Opcional. Una clave del tipo **Iniciar sesión con contraseña** del Almacén de Claves, que se usa
cuando un play eleva privilegios con `become`: su usuario es el usuario en el que se convierte
Ansible, y su contraseña es la que Ansible introduce cuando se le solicita.

Deje el campo vacío cuando los plays no eleven privilegios, o cuando el usuario anterior no necesite
contraseña para hacerlo.

## Tipo {#inventoryForm-type}

De dónde procede el inventario:

- **Static**: el inventario se escribe en este diálogo en formato INI.
- **Static YAML**: el inventario se escribe en este diálogo en formato YAML. Úselo también para la
  configuración de un plugin de inventario.
- **File**: el inventario es un archivo en un repositorio Git.
- **Managed**: los hosts y sus variables se guardan como registros. Abra el inventario desde la
  lista para editarlos.

Los campos debajo del tipo cambian según este.

## Ruta al archivo de Inventario {#inventoryForm-inventoryPath}

Se muestra cuando el tipo es **File**.

La ruta al archivo de inventario desde la raíz de su repositorio, por ejemplo
`inventory/production.yml`. El repositorio es el seleccionado en el campo Repositorio o, cuando ese
campo está vacío, el repositorio de la plantilla de tarea que usa el inventario.

## Repositorio {#inventoryForm-repositoryId}

Se muestra cuando el tipo es **File**. Opcional.

El repositorio que contiene el archivo de inventario. Semaphore actualiza su copia del repositorio
antes de cada tarea. Deje el campo vacío cuando el archivo esté en el mismo repositorio que el
playbook: en ese caso la ruta se busca en el repositorio de la plantilla de tarea.

## Inventario {#inventoryForm-inventoryCode}

Se muestra cuando el tipo es **Static** o **Static YAML**.

El texto del inventario. Semaphore lo almacena y lo escribe en un archivo temporal para cada tarea.
Un inventario estático en formato INI:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Los mismos hosts como **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
