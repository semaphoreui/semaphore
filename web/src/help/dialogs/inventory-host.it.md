---
form: src/components/InventoryHostForm.vue
---

# Modifica host

Un host di un inventario gestito. Un compito che usa l'inventario viene eseguito sugli host che sono
attivati nell'elenco degli host.

## Nome {#inventoryHostForm-name}

Il nome con cui Ansible conosce l'host e l'indirizzo a cui si connette, a meno che le variabili non
impostino `ansible_host`. Il campo è obbligatorio. Un nome è unico nel suo inventario e distingue
maiuscole e minuscole: `web1` e `Web1` sono due host. Non può contenere spazi, virgole, parentesi
quadre o due punti, ed è lungo al massimo 250 caratteri.

## Descrizione {#inventoryHostForm-description}

Facoltativo. Una nota per le persone, mostrata nell'elenco degli host. Ansible non la riceve.

## Variabili {#inventoryHostForm-variables}

Facoltativo. Le variabili di questo host, scritte come mapping YAML o JSON:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Una variabile dell'host ha la precedenza su una variabile dell'inventario con lo stesso nome.
Ancore, alias e chiavi di merge di YAML non sono accettati.
