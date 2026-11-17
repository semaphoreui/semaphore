---
form: src/components/InventoryForm.vue
---

# Modifica Ansible Inventory

Un inventario indica ad Ansible su quali host eseguire un playbook e contiene le credenziali che
Semaphore usa per connettersi a essi. Un modello di compito fa riferimento a un inventario, e ogni
compito di quel modello viene eseguito su di esso. Il formato di un inventario è descritto nella
[guida agli inventari di Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Nome {#inventoryForm-name}

Il nome con cui l'inventario compare nella pagina Inventario e nel selettore di inventario di un
modello di compito. Scegli un nome che distingua gli host, ad esempio `Production web servers`.

## Tag esecutore {#inventoryForm-runnerTag}

Visibile quando sono disponibili esecutori del progetto.

I compiti che usano questo inventario vengono eseguiti su un esecutore che ha questo tag. Un tag
esecutore impostato sul modello di compito ha la precedenza sul tag dell'inventario. Lascia il campo
vuoto per consentire a qualsiasi esecutore di prendere i compiti.

## Credenziali utente {#inventoryForm-sshKeyId}

La chiave del Negozio chiavi con cui Ansible accede agli host. Il campo è obbligatorio e accetta tre
tipi di chiavi:

- **Chiave SSH**: Ansible si connette con la chiave privata e, se la chiave ne ha uno, con il suo
  nome utente.
- **Accesso con password**: Ansible si connette con il nome utente e la password.
- **Nessuno**: Semaphore non passa alcuna credenziale. Usalo per gli host che non ne hanno bisogno,
  come `localhost`, o quando è l'inventario stesso a definire come connettersi.

## Credenziali Sudo {#inventoryForm-becomeKeyId}

Opzionale. Una chiave di tipo **Accesso con password** dal Negozio chiavi, usata quando un play
eleva i privilegi con `become`: il suo nome utente è l'utente che Ansible diventa, e la sua password
è quella con cui Ansible risponde al prompt.

Lascia il campo vuoto quando i play non elevano i privilegi, o quando l'utente indicato sopra non ha
bisogno di una password per farlo.

## Tipo {#inventoryForm-type}

Da dove proviene l'inventario:

- **Static**: l'inventario viene scritto in questa finestra di dialogo in formato INI.
- **Static YAML**: l'inventario viene scritto in questa finestra di dialogo in formato YAML. Usalo
  anche per la configurazione di un plugin di inventario.
- **File**: l'inventario è un file in un repository Git.
- **Managed**: gli host e le loro variabili sono conservati come record. Apri l'inventario
  dall'elenco per modificarli.

I campi sotto il tipo cambiano in base a esso.

## Percorso del file di inventario {#inventoryForm-inventoryPath}

Visibile quando il tipo è **File**.

Il percorso del file di inventario a partire dalla radice del suo repository, ad esempio
`inventory/production.yml`. Il repository è quello selezionato nel campo Repository oppure, se quel
campo è vuoto, il repository del modello di compito che usa l'inventario.

## Repository {#inventoryForm-repositoryId}

Visibile quando il tipo è **File**. Opzionale.

Il repository che contiene il file di inventario. Semaphore aggiorna la propria copia del repository
prima di ogni compito. Lascia il campo vuoto quando il file si trova nello stesso repository del
playbook: il percorso viene allora cercato nel repository del modello di compito.

## Inventario {#inventoryForm-inventoryCode}

Visibile quando il tipo è **Static** o **Static YAML**.

Il testo dell'inventario. Semaphore lo memorizza e lo scrive in un file temporaneo per ogni compito.
Un inventario statico in formato INI:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Gli stessi host come **Static YAML**:

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
