---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Modifica programma

Un programma avvia un compito di un modello di compito senza che nessuno prema **Esegui**: più e più
volte in base a un'espressione cron, oppure una sola volta in un momento stabilito. Il compito
riceve i parametri salvati in questa finestra di dialogo. Tutti gli orari sono nel fuso orario del
server Semaphore, mostrato accanto ai campi dell'ora: `UTC` a meno che l'amministratore non imposti
`schedule.timezone` o `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

Il nome con cui il programma compare nella pagina Programma. Il campo è obbligatorio. Non cambia
ciò che fa il programma.

## Template {#scheduleForm-templateId}

Il modello di compito che il programma avvia. Il campo è obbligatorio. Il compito viene eseguito
come se fosse stato avviato dal modello, con l'inventario, il repository, il gruppo di variabili e
le altre impostazioni del modello stesso, e con i parametri qui sotto dove il modello consente a un
compito di cambiarli.

## Task parameters {#scheduleForm-taskParams}

Visibile quando è selezionato un modello.

I parametri di ogni compito che il programma avvia. Quali campi compaiono dipende dal modello: le
sue variabili del sondaggio, e le opzioni che consente a un compito di cambiare, come il ramo,
l'inventario, **Limite**, **Tag**, **Salta tag** o **Argomenti CLI**, e le opzioni di Ansible o di
Terraform, OpenTofu e Terragrunt.

Il server controlla il modello ogni volta che il programma scatta: un valore salvato qui viene
ignorato quando il modello non consente più a un compito di cambiarlo.

## Messaggio {#taskParams-message}

Opzionale. Una nota memorizzata con ogni compito che il programma avvia e mostrata nella cronologia
dei compiti, ad esempio `Nightly deploy`. Non cambia il modo in cui viene eseguito il compito.

## Variabili del sondaggio {#taskParams-survey}

Visibile quando il modello ha variabili del sondaggio.

I valori delle variabili del sondaggio del modello per i compiti programmati. Arrivano al compito
nello stesso modo di quando qualcuno compila il sondaggio a mano. Una variabile obbligatoria deve
avere un valore prima che il programma possa essere salvato.

I valori predefiniti del modello vengono copiati all'apertura della finestra di dialogo e salvati
con il programma. Un valore predefinito cambiato in seguito nel modello non arriva a un programma
salvato in precedenza.

Il valore di una variabile di tipo secret non viene salvato con il programma: il compito programmato
viene eseguito senza di esso.

## Ramo {#taskParams-gitBranch}

Visibile quando il modello consente a un compito di sovrascrivere il ramo.

Il ramo o il tag del repository di cui il compito programmato esegue il checkout. Il campo parte con
il ramo del modello. Il server usa il valore solo finché il modello consente la sovrascrittura.

## Inventario {#taskParams-inventoryId}

Visibile quando il modello consente a un compito di scegliere l'inventario.

L'inventario su cui viene eseguito il compito programmato, al posto dell'inventario del modello. Per
un modello Terraform o OpenTofu il campo si chiama **Workspace**, ma il server applica la scelta
solo ai modelli Ansible e per gli altri mantiene il workspace del modello.

## Debug {#taskParams-debug}

Visibile per un modello Ansible che consente il debug.

Esegue `ansible-playbook` con più output: `-v` ripetuto tante volte quante imposta **Debug level**.

## Debug level {#taskParams-debugLevel}

Visibile per un modello Ansible che consente il debug; attivo quando **Debug** è selezionato.

Quanto è dettagliato l'output, da `1` (`-v`) a `6` (`-vvvvvv`). Il valore predefinito è `4`.

## Esecuzione simulata {#taskParams-dryRun}

Visibile per un modello Ansible a meno che il modello non la nasconda.

Esegue il playbook con `--check`: Ansible riporta cosa cambierebbe e non cambia nulla.

## Differenza {#taskParams-diff}

Visibile per un modello Ansible a meno che il modello non la nasconda.

Esegue il playbook con `--diff`: Ansible mostra le modifiche che apporta a file e template.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Visibile per un modello Ansible che consente a un compito di cambiarla.

Salta l'installazione dei ruoli e delle collezioni dei file `requirements.yml` prima dell'esecuzione
del playbook. Selezionala quando i requisiti sono già installati e il compito non deve raggiungere
Ansible Galaxy.

## Plan {#taskParams-plan}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Esegue solo `plan`. Il compito termina dopo il piano e non applica nulla.

## Destroy {#taskParams-destroy}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Aggiunge `-destroy` al piano e all'applicazione: il compito distrugge le risorse del workspace invece
di crearle o aggiornarle.

## Auto Approve {#taskParams-autoApprove}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Applica il piano senza attendere una conferma. Funziona solo quando il modello consente
l'approvazione automatica; un modello impostato per approvare ogni compito applica il piano in ogni
caso.

Senza approvazione automatica un compito programmato si ferma dopo il piano e attende finché qualcuno
non lo conferma o lo rifiuta nella pagina del compito.

## Upgrade {#taskParams-upgrade}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Esegue `init` con `-upgrade`: i provider e i moduli vengono aggiornati alle versioni più recenti
consentite dalla configurazione.

## Reconfigure {#taskParams-reconfigure}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Esegue `init` con `-reconfigure` invece di `-migrate-state`: il backend viene configurato di nuovo e
lo stato esistente non viene migrato.

## Run at {#scheduleForm-runAt}

Visibile per un programma che viene eseguito una sola volta.

La data e l'ora in cui parte il compito, nel fuso orario del server mostrato alla fine del campo. Un
nuovo programma parte con l'ora intera successiva. Il momento deve essere nel futuro: il server
rifiuta un orario già passato.

Dopo l'avvio del compito, il programma viene disattivato, o eliminato quando **Delete after run** è
selezionato. Quando Semaphore non è in esecuzione in quel momento, il compito non viene avviato più
tardi: al successivo avvio il programma viene disattivato, o eliminato, senza esecuzione.

## Delete after run {#scheduleForm-deleteAfterRun}

Visibile per un programma che viene eseguito una sola volta.

Elimina il programma dopo l'avvio del suo compito. Lascialo deselezionato per mantenere il programma,
disattivato, così da poter impostare un nuovo orario e riattivarlo.

## Show cron format {#scheduleForm-rawCron}

Visibile per un programma cron.

Passa dalla digitazione dell'espressione cron (attivo) alla sua costruzione da **Timing** e dalle
caselle dell'orario (disattivo). Entrambi modificano la stessa espressione. Il browser ricorda la
scelta.

L'interruttore è bloccato su attivo quando l'espressione non può essere mostrata come caselle:
quando non è valida, o usa una forma come `@daily` o `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Visibile per un programma cron quando **Show cron format** è attivo.

L'espressione cron in cinque campi: minuto, ora, giorno del mese, mese, giorno della settimana (`0`
è la domenica). Il campo è obbligatorio, e il server controlla l'espressione mentre digiti. Un nuovo
programma parte con `* * * * *`, ogni minuto.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

Il server accetta anche `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` e `@every 1h30m`, e un
prefisso come `CRON_TZ=Europe/Berlin` per un altro fuso orario. L'anteprima qui sotto non può
mostrare la prossima esecuzione di queste forme.

## Timing {#scheduleForm-timing}

Visibile per un programma cron quando **Show cron format** è disattivo.

Ogni quanto viene eseguito il compito. Le caselle qui sotto cambiano in base a esso:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** e **Minutes**.
- **Weekly**: **Weekdays**, **Hours** e **Minutes**.
- **Monthly**: **Days**, **Hours** e **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** e **Minutes**.

## Months {#scheduleForm-months}

Visibile quando **Timing** è **Yearly**.

I mesi in cui viene eseguito il compito. Selezionane diversi per eseguirlo in ciascuno di essi. Se
non ne è selezionato nessuno, il compito viene eseguito in ogni mese.

## Weekdays {#scheduleForm-weekdays}

Visibile quando **Timing** è **Weekly**.

I giorni della settimana in cui viene eseguito il compito. Se non ne è selezionato nessuno, il
compito viene eseguito ogni giorno.

## Days {#scheduleForm-days}

Visibile quando **Timing** è **Monthly** o **Yearly**.

I giorni del mese in cui viene eseguito il compito. Un mese senza quel giorno viene saltato: un
compito impostato sul 31 non viene eseguito ad aprile. Se non ne è selezionato nessuno, il compito
viene eseguito ogni giorno.

## Hours {#scheduleForm-hours}

Visibile quando **Timing** è **Daily**, **Weekly**, **Monthly** o **Yearly**.

Le ore in cui viene eseguito il compito, nel fuso orario mostrato a destra. Se non ne è selezionata
nessuna, il compito viene eseguito ogni ora.

## Minutes {#scheduleForm-minutes}

Visibile per un programma cron quando **Show cron format** è disattivo.

I minuti dell'ora in cui viene eseguito il compito, a passi di cinque. Se non ne è selezionato
nessuno, il compito viene eseguito ogni minuto: selezionane almeno uno, ad esempio `:00`, per un
compito che deve essere eseguito una volta all'ora o una volta al giorno.

## Next run time {#scheduleForm-nextRun}

Quando partirebbe il compito successivo, secondo il programma così come è nel modulo: nel fuso
orario del server e nell'ora locale del tuo browser. Un trattino significa che l'espressione non è
valida, o è una che l'anteprima non sa leggere, come `@daily`.

## Abilitato {#scheduleForm-active}

Se il programma avvia compiti. Deselezionalo per mantenere un programma senza eseguirlo.
L'interruttore nell'elenco dei programmi fa lo stesso. Un nuovo programma è abilitato.
