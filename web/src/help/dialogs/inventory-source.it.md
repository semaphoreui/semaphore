---
form: src/components/InventorySourceForm.vue
---

# Modifica sorgente

Una sorgente importa gli host e i gruppi di un inventario gestito da un file di un repository Git. Il file viene letto a ogni pressione di Sync, e gli host e i gruppi che descrive restano allineati ad esso. Il server richiede Ansible e l'accesso al repository.

## Nome {#inventorySourceForm-name}

Il nome della sorgente, mostrato sulla sua scheda e accanto agli host che porta. Il campo è obbligatorio.

## Repository {#inventorySourceForm-repository}

Il repository del progetto da cui si legge il file. Il campo è obbligatorio. Si usa il suo branch.

## Percorso {#inventorySourceForm-path}

Il percorso di un file di inventario nel repository, ad esempio `inventory/prod.yml`, o di una directory di file di inventario. Il campo è obbligatorio.

I file si leggono come YAML o INI statici, script e file di configurazione dei plugin vengono rifiutati. In una directory ogni file che Ansible non ignora deve essere un inventario valido, altrimenti la sincronizzazione fallisce e non cambia nulla. Ansible ignora alcune estensioni in una directory, ad esempio `.cfg`, `.md` e `.txt`. Un repository con un collegamento che porta fuori da esso viene rifiutato.
