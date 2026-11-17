---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Impostazioni del progetto

Le impostazioni del progetto nel suo insieme: il suo nome, quanti dei suoi compiti possono essere
eseguiti contemporaneamente e se invia avvisi. La Zona pericolosa sotto il modulo contiene le azioni
che riguardano l'intero progetto: il suo backup, la sua cache e la sua eliminazione. Le modifiche
del modulo hanno effetto dopo **Salva**.

## Nome progetto {#newProject-name}

Il nome con cui il progetto compare nel menu dei progetti in cima alla barra laterale. Il campo è
obbligatorio.

## Numero massimo di compiti paralleli {#newProject-maxParallelTasks}

Quanti compiti di questo progetto possono essere eseguiti allo stesso tempo. Un compito avviato
oltre il limite attende in coda finché uno dei compiti in esecuzione non termina.

Lascia il campo vuoto o impostalo a `0` per non avere un limite proprio. Il limite dell'intero
server, impostato dal suo amministratore, si applica in ogni caso.

## ID chat Telegram {#newProject-tg}

Opzionale. La chat Telegram che riceve gli avvisi di questo progetto. Quando il campo è vuoto, gli
avvisi vanno alla chat impostata per l'intero server.

Il campo ha effetto solo quando l'amministratore ha attivato gli avvisi Telegram sul server e gli
avvisi sono consentiti per questo progetto.

## Consenti avvisi per questo progetto {#newProject-alert}

Attiva gli avvisi per i compiti di questo progetto. La destinazione degli avvisi è impostata per
l'intero server dal suo amministratore: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams,
DingTalk o Gotify.

- Un'e-mail viene inviata quando un compito fallisce, ai membri del progetto che possono vedere il
  compito e hanno attivato gli avvisi nel proprio account.
- Un messaggio ai servizi di messaggistica viene inviato quando un compito riesce, fallisce o attende
  una conferma.

Un modello di compito può impedire ai propri compiti riusciti o falliti di inviare avvisi.

## Test Alerts {#settings-testAlerts}

Invia un avviso di prova attraverso ogni canale attivato sul server, così puoi verificare che gli
avvisi di questo progetto arrivino.

Il pulsante funziona solo quando **Consenti avvisi per questo progetto** è selezionato e salvato.
Selezionalo e premi prima **Salva**.

## Backup progetto {#settings-exportProject}

Scarica un file JSON con la configurazione del progetto: modelli di compito, inventari, repository,
gruppi di variabili, chiavi, programmazioni, integrazioni e le altre cose configurate al suo
interno. Il file può essere riconvertito in un progetto con **Ripristina progetto** nel menu dei
progetti in cima alla barra laterale.

Il file esclude i segreti delle chiavi e i segreti nella scheda **Secrets** dei gruppi di variabili:
devono essere inseriti di nuovo dopo un ripristino. Tutto il resto viene copiato così com'è, anche
una password digitata direttamente in un altro campo: l'indirizzo di un repository che contiene una
password, il testo di un inventario statico, i valori delle variabili ordinarie e delle variabili di
ambiente. Controlla che il file non contenga valori del genere prima di condividerlo. La cronologia
dei compiti non fa parte del backup.

## Pulisci cache {#settings-clearCache}

Elimina i file che Semaphore conserva sul server per questo progetto, come le sue copie dei
repository. Vengono scaricati di nuovo all'avvio del compito successivo.

Usalo quando un compito fallisce a causa di una copia di un repository obsoleta o danneggiata.
L'azione non può essere annullata e non modifica le impostazioni del progetto.

## Elimina progetto {#settings-deleteProject}

Elimina il progetto con tutto il suo contenuto: modelli di compito, cronologia dei compiti,
inventari, repository, gruppi di variabili, chiavi e programmazioni. Semaphore chiede prima una
conferma.

Un progetto eliminato non può essere recuperato. Scarica un backup prima di eliminare un progetto di
cui potresti avere ancora bisogno.
