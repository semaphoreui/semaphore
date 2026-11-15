---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

I nodi di un cluster condividono in Redis lo stato del pool dei compiti: quali compiti attendono,
quali sono in esecuzione e quale nodo possiede ciascuno di essi. Questa finestra di dialogo elimina
da Redis i gruppi di record che selezioni, per recuperare un cluster bloccato, ad esempio con un
compito terminato che risulta ancora in esecuzione. Nulla viene eliminato dal database: i compiti
mantengono il loro stato, il loro log e la loro cronologia.

Qualunque cosa selezioni, ogni nodo scarta anche i compiti che tiene in memoria e li rilegge da
Redis e dal database. All'apertura della finestra di dialogo è selezionato solo **Queue**; seleziona
solo i gruppi bloccati. Il risultato mostra quanti record sono stati rimossi.

## Queue {#clusterClearDialog-queue}

I compiti che attendono che un nodo li avvii. I compiti cancellati non vengono avviati: restano nel
database con lo stato Waiting. Fermali, o esegui di nuovo i loro modelli.

## Running {#clusterClearDialog-running}

I compiti che il cluster considera in esecuzione. La cancellazione non li ferma: un compito continua
a essere eseguito sul suo esecutore. Ma il cluster ne perde traccia: il compito non conta più per i
limiti di compiti paralleli del server e del suo esecutore, quindi possono partire più compiti di
quanti i limiti consentano, e un compito che un esecutore non ha ancora preso non gli viene più
consegnato.

## Active by project {#clusterClearDialog-active}

I compiti non terminati di ogni progetto. Fanno rispettare il limite di compiti paralleli di un
progetto e impediscono che parta un secondo compito di un modello mentre uno è in esecuzione,
quando il modello non consente compiti paralleli. Cancella questo gruppo quando i compiti di un
progetto attendono all'infinito dietro compiti che sono terminati. I compiti in esecuzione in quel
momento non contano più, quindi nuovi compiti del progetto possono partire accanto a essi.

## Aliases {#clusterClearDialog-aliases}

I nomi casuali che i compiti Terraform, OpenTofu e Terragrunt ricevono per la loro durata, associati
al compito. Un compito in esecuzione raggiunge il suo stato Terraform sul server tramite questo
nome: dopo la cancellazione, un compito in esecuzione di queste app non può più essere trovato
tramite esso.

## Claims {#clusterClearDialog-claims}

Quale nodo possiede ciascun compito. Un nodo rinnova le rivendicazioni dei suoi compiti ogni minuto;
una rivendicazione non rinnovata scade dopo cinque minuti. Un compito senza rivendicazione è
considerato orfano: un compito che non ha ancora raggiunto un esecutore viene rimesso in coda per un
altro nodo, e un compito che un esecutore sta già eseguendo viene lasciato terminare lì.

## Runtime fields {#clusterClearDialog-runtimeFields}

I dettagli di ogni compito che il database non contiene: l'utente che lo ha avviato, la versione con
cui è stato avviato, il suo alias, il suo esecutore e il suo stato, e il progetto di ogni compito.
Senza di essi, un nodo che non ha avviato un compito ne mostra meno informazioni, e il pulitore dei
compiti orfani salta un compito in esecuzione, quindi un record bloccato non viene più rimosso da
solo.
