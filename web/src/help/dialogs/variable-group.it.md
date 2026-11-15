---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Modifica gruppo di variabili

Un gruppo di variabili contiene le variabili extra, le variabili di ambiente e i segreti che
Semaphore passa a ogni compito di un modello di compito che usa il gruppo. Ansible riceve le
variabili extra come `--extra-vars`, Terraform e OpenTofu come `-var`; le variabili di ambiente
vengono impostate per il processo del compito. Un modello può usare più gruppi: quando due di essi
definiscono lo stesso nome, prevale il gruppo che viene dopo nel modello.

## Nome gruppo {#varGroup-name}

Il nome con cui il gruppo compare nella pagina Gruppi di variabili e nel selettore del gruppo di
variabili di un modello di compito. Il campo è obbligatorio.

## Secret storage {#varGroup-secretStorageId}

Visibile quando il gruppo è nuovo e il server supporta gli archivi di segreti. Opzionale.

L'archivio di segreti in cui sono conservati i segreti di questo gruppo, ad esempio un HashiCorp
Vault. Lascia il campo vuoto per conservare i segreti nel database di Semaphore, cifrati.

L'archivio si sceglie una sola volta, alla creazione del gruppo: un gruppo esistente mantiene il suo
archivio, e il campo non viene più mostrato. Un archivio contrassegnato come **Read only** non
accetta nuovi segreti: salvarvi un segreto fallisce.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Visibile quando il gruppo è nuovo e il server supporta gli archivi di segreti. Disponibile una
volta scelto un **Secret storage**.

L'inizio del percorso che ogni segreto di questo gruppo riceve nell'archivio. Semaphore vi aggiunge
10 caratteri casuali: con il prefisso `semaphore/prod/` un segreto viene memorizzato come
`semaphore/prod/` seguito da quei caratteri. Il prefisso viene preso così com'è, quindi terminalo
con `/` quando i segreti devono stare sotto di esso come in una cartella.

Il prefisso si imposta una sola volta, alla creazione del gruppo. La scheda **Secrets** lo mostra
come modello del percorso di origine.

## Table, JSON or YAML {#varGroup-extraVarsMode}

Come modifichi le variabili extra:

- **Table**: una riga per variabile, ognuna con un tipo. Aggiungi una riga con il pulsante **+**
  accanto al selettore.
- **JSON**: tutte le variabili come un unico oggetto JSON.
- **YAML**: tutte le variabili come un'unica mappatura YAML.

La finestra di dialogo si apre in **Table**. Il selettore converte ciò che hai digitato; quando non
può essere letto, la finestra di dialogo resta nella modalità in cui eri e mostra l'errore.
Qualunque sia la modalità, le variabili vengono salvate come un unico oggetto di nomi e valori.

## Extra variables in JSON {#varGroup-extraVarsJson}

Visibile quando le variabili extra sono modificate come **JSON**.

Le variabili extra come un unico oggetto JSON. Un nome non può essere vuoto; un valore può essere una
stringa, un numero, un elenco o un oggetto:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Extra variables in YAML {#varGroup-extraVarsYaml}

Visibile quando le variabili extra sono modificate come **YAML**.

Le variabili extra come un'unica mappatura YAML. Il testo viene letto con i tipi che JSON conosce:
una data senza virgolette come `2024-01-01` resta una stringa, e `.inf` o `.nan` vengono rifiutati.
Anche un elenco o uno scalare al livello superiore viene rifiutato.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Variable name {#varGroup-varName}

Visibile quando le variabili extra sono modificate come **Table**, in ogni riga.

Il nome della variabile così come la vede il playbook o il codice Terraform, ad esempio
`app_version`.

## Variable type {#varGroup-varType}

Visibile quando le variabili extra sono modificate come **Table**, in ogni riga.

Come viene letto il valore della riga:

- **String**: il testo così com'è.
- **Number**: un numero, ad esempio `42`.
- **List**: un array JSON, ad esempio `["a", "b"]`.
- **Dict**: un oggetto JSON, ad esempio `{"key": "value"}`.

Un valore che non corrisponde al suo tipo blocca il salvataggio indicando il nome della variabile.

## Variable value {#varGroup-varValue}

Visibile quando le variabili extra sono modificate come **Table**, in ogni riga.

Il valore della variabile, nel formato del suo tipo. Per un **List** o un **Dict**, il pulsante alla
fine del campo apre un editor più grande.

Semaphore passa i valori ad Ansible come dati: un'espressione `{{ … }}` all'interno di un valore non
viene valutata, il playbook riceve il testo così com'è.

## Environment variable name {#varGroup-envName}

Nella scheda **Variables**, sotto **Variabili di ambiente**, in ogni riga. Aggiungi una riga con il
pulsante **+** della sezione.

Il nome di una variabile impostata nell'ambiente del processo che esegue il compito, ad esempio
`AWS_REGION`.

## Environment variable value {#varGroup-envValue}

Nella scheda **Variables**, sotto **Variabili di ambiente**, in ogni riga.

Il valore della variabile, come testo. Viene memorizzato come testo semplice e mostrato a chiunque
apra il gruppo: metti invece password e token nella scheda **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Visibile nella scheda **Secrets** di un gruppo che ha un archivio di segreti.

Attivalo per importare segreti dall'archivio in questo gruppo, dai percorsi impostati in **Sync
paths**. L'importazione viene eseguita quando premi il pulsante di sincronizzazione del gruppo nella
pagina Gruppi di variabili, e da sola all'intervallo impostato in **Sync paths**. Se è disattivato,
non viene importato nulla.

## Sync paths {#varGroup-syncPaths}

Visibile quando **Sync keys enabled** è attivo.

Apre la finestra di dialogo con i percorsi dell'archivio da cui importare e l'intervallo
dell'importazione automatica. Il numero sul pulsante è il numero di percorsi. Il pulsante di
sincronizzazione del gruppo nella pagina Gruppi di variabili resta disattivato finché il gruppo non
ha almeno un percorso.

La finestra di dialogo ha solo **Chiudi**: ciò che vi modifichi viene salvato con il gruppo, da
**Salva** di questa finestra di dialogo.

## Secret variable name {#varGroup-secretVarName}

Nella scheda **Secrets**, sotto **Variabili extra**, in ogni riga. Aggiungi una riga con il pulsante
**+** della sezione.

Il nome di una variabile extra il cui valore è un segreto. Il compito la riceve come riceve le
variabili extra della scheda **Variables**; un segreto prevale su una variabile extra con lo stesso
nome. Ansible riceve i segreti in un file cifrato con Ansible Vault, Terraform e OpenTofu come
`-var`. Come avverte la finestra di dialogo, un segreto passato in questo modo può comunque comparire
in chiaro nel log di Ansible.

## Secret variable value {#varGroup-secretVarValue}

Nella scheda **Secrets**, sotto **Variabili extra**, in ogni riga.

Il valore del segreto. Semaphore lo cifra, o lo scrive nell'archivio di segreti del gruppo, e non lo
mostra mai più: quando riapri il gruppo, il campo è vuoto. Lascialo vuoto per mantenere il valore
memorizzato, digitane uno nuovo per sostituirlo. Eliminare la riga elimina il segreto quando salvi il
gruppo.

## Secret environment variable name {#varGroup-secretEnvName}

Nella scheda **Secrets**, sotto **Variabili di ambiente**, in ogni riga. Aggiungi una riga con il
pulsante **+** della sezione.

Il nome di una variabile di ambiente il cui valore è un segreto, ad esempio `DB_PASSWORD`. Viene
impostata nell'ambiente del processo che esegue il compito, come le variabili di ambiente della
scheda **Variables**.

## Secret environment variable value {#varGroup-secretEnvValue}

Nella scheda **Secrets**, sotto **Variabili di ambiente**, in ogni riga.

Il valore del segreto. Come per una variabile extra segreta, non viene mai più mostrato: lascia il
campo vuoto per mantenere il valore memorizzato, digitane uno nuovo per sostituirlo.
