---
form: src/components/InventorySourceForm.vue
---

# Editar fuente

Una fuente importa los hosts y los grupos de un inventario gestionado desde un archivo de un repositorio Git. El archivo se lee cada vez que se pulsa Sync, y los hosts y grupos que describe se mantienen al día con él. El servidor necesita Ansible y acceso al repositorio.

## Nombre {#inventorySourceForm-name}

El nombre de la fuente, se muestra en su tarjeta y junto a los hosts que aporta. El campo es obligatorio.

## Repositorio {#inventorySourceForm-repository}

El repositorio del proyecto del que se lee el archivo. El campo es obligatorio. Se usa su rama.

## Ruta {#inventorySourceForm-path}

La ruta de un archivo de inventario en el repositorio, por ejemplo `inventory/prod.yml`, o de un directorio de archivos de inventario. El campo es obligatorio.

Los archivos se leen como YAML o INI estáticos, se rechazan los scripts y los archivos de configuración de plugins. En un directorio, cada archivo que Ansible no omite debe ser un inventario válido; si no, la sincronización falla y nada cambia. Ansible omite algunas extensiones en un directorio, por ejemplo `.cfg`, `.md` y `.txt`. Se rechaza un repositorio con un enlace que lleva fuera de él.
