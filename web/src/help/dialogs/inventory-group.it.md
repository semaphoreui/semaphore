---
form: src/components/InventoryGroupForm.vue
---

# Modifica gruppo

Un gruppo di host di un inventario gestito. Un playbook viene eseguito su una parte degli host
tramite il nome del gruppo, e una variabile del gruppo si imposta una volta per tutti i suoi host.
Gli host si mettono in un gruppo nella scheda Hosts, nella finestra dell'host o con le azioni sulla
selezione.

## Nome {#inventoryGroupForm-name}

Il nome con cui Ansible conosce il gruppo, per esempio in `hosts: web` di un playbook. Il campo è
obbligatorio. Un nome è unico nel suo inventario e distingue maiuscole e minuscole: `web` e `Web`
sono due gruppi. Non può contenere spazi, virgole, parentesi quadre o due punti, è lungo al massimo
250 caratteri e non può essere `all` o `ungrouped`. Ansible avvisa per i nomi con caratteri diversi
da lettere, cifre e trattini bassi, come `web-eu`, ma il nome funziona comunque.

## Descrizione {#inventoryGroupForm-description}

Facoltativo. Una nota per le persone, mostrata nell'elenco dei gruppi. Ansible non la riceve.

## Child groups {#inventoryGroupForm-children}

Facoltativo. I gruppi dentro questo gruppo. Gli host di un gruppo figlio appartengono anche a questo
gruppo. Un gruppo non può essere il proprio antenato.

## Variabili {#inventoryGroupForm-variables}

Facoltativo. Le variabili di questo gruppo, scritte come mapping YAML o JSON:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Ogni host del gruppo le riceve. Una variabile di un host ha la precedenza su una variabile del
gruppo con lo stesso nome, e una variabile del gruppo ha la precedenza su una variabile
dell'inventario. Ancore, alias e chiavi di merge di YAML non sono accettati.
