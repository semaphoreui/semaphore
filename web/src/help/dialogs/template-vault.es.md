---
form: src/components/TemplateVaults.vue
---

# Bóveda

Una bóveda da a Ansible la contraseña de los archivos y variables cifrados con Ansible Vault en el
repositorio del playbook. Cada bóveda de la plantilla se convierte en una opción `--vault-id` de
`ansible-playbook`. La bóveda se guarda con la plantilla, cuando guarda el diálogo de la plantilla.

## ID de Bóveda {#templateVaultDialog-name}

Opcional. El vault ID al que corresponde la contraseña, la etiqueta con la que se cifró el
contenido (`ansible-vault encrypt --vault-id prod@prompt`). Déjelo vacío para el vault ID
`default`.

El ID se guarda en minúsculas. Cada bóveda de la plantilla necesita su propio ID, y solo una puede
estar vacía.

## Tipo {#templateVaultDialog-type}

De dónde procede la contraseña:

- **Contraseña**: una clave del Almacén de Claves. Semaphore escribe su contraseña cuando Ansible la
  pide, así que nunca se escribe en un archivo.
- **Script del Cliente**: un script cliente de contraseña de vault que imprime la contraseña;
  Ansible lo ejecuta.

## Contraseña de la Bóveda {#templateVaultDialog-vaultKeyId}

Se muestra cuando el tipo es **Contraseña**.

La clave que contiene la contraseña del vault. Solo se listan las claves del tipo **Iniciar sesión
con contraseña**; Semaphore usa su contraseña e ignora el usuario. El campo es obligatorio.

## Ruta del Script {#templateVaultDialog-script}

Se muestra cuando el tipo es **Script del Cliente**.

La ruta al script, por ejemplo `scripts/vault-pass-client.py`. Una ruta relativa se toma desde el
directorio en el que se ejecuta Ansible: la raíz del repositorio, o el directorio de trabajo de la
plantilla cuando tiene uno. El campo es obligatorio.

El nombre del archivo debe terminar en `-client`, con o sin extensión: solo así Ansible ejecuta el
archivo como script cliente y le pasa el vault ID con `--vault-id`. El script se ejecuta en la
máquina que ejecuta la tarea, como el usuario con el que se ejecuta Semaphore.
