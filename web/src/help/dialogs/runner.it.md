---
form: src/components/RunnerForm.vue
---

# Modifica esecutore

Un esecutore è un processo `semaphore runner` separato che esegue i compiti sulla propria macchina
invece che sul server Semaphore. Un compito va a un esecutore quando il server è configurato per
eseguire i compiti sugli esecutori, o quando il modello di compito o il suo inventario ha un tag
esecutore. Il server sceglie allora un esecutore abilitato, registrato, online e non al suo limite
di compiti paralleli: prima gli esecutori del progetto, poi quelli globali, in ordine casuale
all'interno di ogni gruppo. Quando ogni esecutore corrispondente è offline o occupato, il compito
attende in coda; quando nessun esecutore corrisponde, il compito fallisce.

## Nome {#runner-name}

Obbligatorio. Il nome con cui l'esecutore compare nella pagina Esecutori, ad esempio l'host su cui
gira. Non cambia quali compiti riceve l'esecutore.

## Tags {#runner-tags}

Disponibile in Semaphore Pro; nelle altre edizioni il campo è bloccato.

I tag che scelgono i compiti di questo esecutore. Un compito con un tag esecutore va solo agli
esecutori che hanno quel tag; il tag del modello di compito prevale sul tag dell'inventario. Digita
un tag e premi Invio, o scegline uno usato da altri esecutori. I tag vuoti e ripetuti vengono
scartati al salvataggio.

Un esecutore del progetto ha bisogno di almeno un tag, a meno che **Is default** non sia
selezionato.

## Is default {#runner-isDefault}

Un compito senza tag esecutore va solo agli esecutori contrassegnati come predefiniti. Deseleziona
la casella per un esecutore che deve prendere solo compiti con uno dei suoi tag.

La casella è selezionata per un nuovo esecutore globale e deselezionata per un nuovo esecutore del
progetto.

## Register {#runner-registered}

Visibile quando crei un esecutore.

- Selezionata (il valore predefinito): il server crea l'esecutore con il suo token. Dopo **Crea** la
  finestra di dialogo mostra il token e i comandi che avviano l'esecutore con esso.
- Deselezionata: l'esecutore viene creato senza alcun token. In seguito, fai clic su **Regenerate
  registration token** (l'icona con due frecce) nella riga dell'esecutore per ottenere un token di
  registrazione monouso, valido per un'ora, e registra l'esecutore con `semaphore runner register`.

Un esecutore non registrato non riceve compiti.

## Webhook {#runner-webhook}

Opzionale. Un URL che il server chiama quando assegna un compito a questo esecutore, ad esempio per
avviare l'esecutore su richiesta. Il server invia una richiesta `POST` con un corpo JSON:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

L'URL deve rispondere con lo stato 200 o 204 entro 30 secondi; altrimenti il compito fallisce. Un
esecutore con un webhook è sempre considerato online, anche quando non ha interrogato il server di
recente. Lascia il campo vuoto per un esecutore sempre in funzione.

## Numero massimo di compiti paralleli {#runner-maxParallelTasks}

Il numero di compiti che l'esecutore può eseguire contemporaneamente. Quando il limite è raggiunto,
il compito successivo va a un altro esecutore corrispondente o attende in coda. Lascia il campo
vuoto o imposta `0` per non avere limiti.

## Abilitato {#runner-active}

Solo un esecutore abilitato riceve nuovi compiti. Deseleziona la casella per mettere fuori servizio
l'esecutore senza eliminarlo; l'interruttore nell'elenco degli esecutori fa lo stesso.

La casella è deselezionata per un nuovo esecutore: selezionala, o abilita l'esecutore in seguito
nell'elenco.
