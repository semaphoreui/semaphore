---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Modifica modello 'Ansible Playbook'

Un modello di compito indica cosa esegue Semaphore e come: il repository e il file al suo interno,
l'inventario e le variabili, e quali opzioni un utente può cambiare all'avvio di un compito. Ogni
compito viene avviato da un modello. I campi della finestra di dialogo dipendono dall'app del
modello (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell e altre): un campo che
hanno solo alcune app lo indica. Le modifiche hanno effetto per i compiti avviati dopo **Salva**.

## Template type {#templateForm-typeTabs}

Le schede in alto impostano il tipo di modello:

- **Compito**: esegue il codice. La maggior parte dei modelli è di questo tipo.
- **Costruisci**: crea un artefatto. Ogni compito di build riceve una versione, contata a partire da
  **Versione di partenza**.
- **Distribuisci**: distribuisce un artefatto creato da una build. Quando lo avvii scegli il compito
  di build di cui distribuisce la versione.

Un compito di build o di distribuzione riceve nel suo ambiente le variabili `SEMAPHORE_TASK_TYPE`,
`SEMAPHORE_TASK_TARGET_VERSION` (build) e `SEMAPHORE_TASK_INCOMING_VERSION` (distribuzione). Ansible
riceve gli stessi valori nella variabile extra `semaphore_vars.task_details`.

## Versione di partenza {#templateForm-startVersion}

Visibile quando il tipo è **Costruisci**. Obbligatorio.

La versione del primo compito di build, ad esempio `1.0.0` o `v0.1`. Ogni build successiva aumenta
di uno l'ultimo numero della versione precedente e mantiene il testo intorno e gli zeri iniziali:
`1.0.0`, `1.0.1`, `1.0.2`. Alza qui il numero sopra l'ultima build per saltare avanti, ad esempio a
`2.0.0`; una versione senza numeri non cambia mai.

## Modello di build {#templateForm-buildTemplateId}

Visibile quando il tipo è **Distribuisci**. Obbligatorio.

Il modello i cui compiti creano gli artefatti che questo modello distribuisce. L'elenco contiene i
modelli di build del progetto e, sotto di essi, i modelli di distribuzione: una distribuzione può
seguire un'altra distribuzione, e riceve allora la versione che quella distribuzione ha ricevuto
dalla sua build.

## Esecuzione automatica {#templateForm-autorun}

Visibile quando il tipo è **Distribuisci**.

Avvia da solo un compito di questo modello ogni volta che un compito del **Modello di build**
riesce, con la versione creata da quel compito. Una build fallita non avvia nulla.

## Nome {#templateForm-name}

Il nome con cui il modello compare nella pagina Modelli di compito. Il campo è obbligatorio.

## Repository {#templateForm-repositoryId}

Il repository del progetto che contiene il codice da eseguire. Il campo è obbligatorio. Prima di
ogni compito Semaphore aggiorna la propria copia del repository ed esegue il codice da lì.

L'etichetta a destra del campo mostra il ramo di cui i compiti eseguono il checkout. Fai clic su di
essa per aprire il campo **Ramo**.

## Ramo {#templateForm-gitBranch}

Visibile dopo un clic sull'etichetta del ramo a destra di **Repository**.

Il ramo o il tag di cui i compiti di questo modello eseguono il checkout, al posto del ramo
impostato sul repository. Lascia il campo vuoto per usare il ramo del repository. Quando Semaphore
riesce a leggere i rami del repository, il campo li elenca.

Il campo è disattivato quando il repository è un monorepo: tutti i modelli ne condividono un unico
checkout, sul ramo del repository.

## Percorso del file playbook {#templateForm-playbook}

Cosa eseguire, come percorso a partire dalla radice del repository, ad esempio `deploy/site.yml`. Il
percorso non può essere assoluto e non può portare fuori dal repository. L'elenco offre i file del
repository; puoi anche digitare un percorso che non è nell'elenco. Il pulsante a destra del campo
rilegge l'elenco.

L'etichetta e il significato cambiano con l'app:

- **Ansible**: il playbook. È obbligatorio; l'elenco offre i file `.yml` e `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, la directory con il codice.
  Lascialo vuoto per eseguire il codice nella radice del repository. Digita `/` per elencare il
  contenuto di una directory.
- **Bash**, **Python**, **PowerShell** e altre app: **Script Filename**, lo script da eseguire. È
  obbligatorio. Lo script viene eseguito nella radice del repository.

## Different working directory {#templateForm-workingDirectoryToggle}

Visibile per i modelli Ansible.

Esegue `ansible-playbook` in una directory diversa dalla radice del repository. Selezionalo per
mostrare il campo **Working directory**. Deselezionalo per tornare a eseguire nella radice.

## Working directory {#templateForm-workingDirectory}

Visibile quando **Different working directory** è selezionato. Obbligatorio.

La directory in cui vengono eseguiti `ansible-playbook` e `ansible-galaxy`, come percorso a partire
dalla radice del repository, ad esempio `ansible`. Usalo quando l'`ansible.cfg` del progetto si
trova in una sottodirectory, poiché Ansible legge il file dalla directory corrente. Il percorso del
playbook resta relativo alla radice del repository. Il percorso non può essere assoluto e non può
portare fuori dal repository.

## Inventario {#templateForm-inventoryId}

Visibile per i modelli Ansible, Terraform, OpenTofu e Terragrunt.

- **Ansible**: l'inventario su cui viene eseguito il playbook, con le credenziali per gli host. Il
  campo è obbligatorio.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, il workspace in cui viene eseguito il
  codice. Semaphore lo seleziona prima del piano e lo crea quando non esiste. Un nuovo modello
  salvato senza workspace ne riceve uno chiamato `default`.

L'elenco contiene gli inventari del progetto creati per l'app del modello.

## Gruppo di variabili {#templateForm-environmentIds}

Opzionale. I gruppi di variabili le cui variabili, variabili di ambiente e segreti riceve ogni
compito di questo modello. Puoi sceglierne diversi: vengono uniti nell'ordine del campo, e un gruppo
successivo prevale su uno precedente quando entrambi impostano lo stesso nome.

Il modo in cui le variabili arrivano al codice dipende dall'app: variabili extra per Ansible, `-var`
per Terraform e OpenTofu, argomenti `name=value` per uno script. Le variabili di ambiente vengono
impostate per ogni app.

## Vista {#templateForm-viewId}

Opzionale. La vista in cui compare il modello: le viste sono le schede della pagina Modelli di
compito, che raggruppano i modelli di un progetto. Lascia il campo vuoto per non mettere il modello
in alcuna vista.

## Tag esecutore {#templateForm-runnerTag}

Visibile quando sono disponibili esecutori del progetto.

I compiti di questo modello vengono eseguiti solo su un esecutore che ha questo tag, anche quando il
server potrebbe eseguirli da sé. Il tag sovrascrive il tag esecutore dell'inventario. Lascia il campo
vuoto per usare il tag dell'inventario, o nessuno.

## Executor image {#templateForm-executorImage}

L'immagine del container che l'esecutore avvia per i compiti di questo modello, ad esempio
`semaphoreui/job:latest`. Lascia il campo vuoto per usare l'immagine della configurazione
dell'esecutore.

Solo gli executor Docker e Kubernetes eseguono i compiti in un container, quindi il campo è attivo
solo quando uno di essi è disponibile. Un compito eseguito senza container lo ignora.

## Variabili del sondaggio {#surveyVars}

I valori che vengono chiesti a un utente quando parte un compito di questo modello. Fai clic su
**+** per aggiungere una variabile, su un chip per modificarne una, sulla croce di un chip per
rimuoverla; trascina i chip per impostare l'ordine in cui vengono poste le domande.

Le risposte arrivano al codice come le variabili di un gruppo di variabili, o come variabili di
ambiente, a seconda di come è impostata ciascuna variabile. Prevalgono su un gruppo di variabili che
imposta lo stesso nome.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Consente ai compiti di questo modello di essere eseguiti contemporaneamente. Senza di esso un
compito del modello attende in coda finché quello in esecuzione non termina. I limiti di compiti
paralleli del progetto e del server si applicano in ogni caso.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Visibile quando l'amministratore ha attivato l'emissione di JWT per il server.

Dà a ogni compito di questo modello un JSON Web Token firmato da Semaphore, nella variabile di
ambiente `SEMAPHORE_JWT`. Il codice può mostrare il token a un altro servizio per dimostrare quale
compito è: il token indica il compito, il progetto, il modello e l'utente che lo ha avviato. Il
servizio verifica la firma con le chiavi pubbliche in `/.well-known/jwks.json` del server Semaphore.

## JWT audience {#templateForm-jwtAudience}

Visibile quando **Issue JWT to task runner** è selezionato. Opzionale.

I servizi a cui è destinato il token, scritti nel suo claim `aud`, ad esempio
`https://vault.example.com`. Premi Invio dopo ogni valore. Fino a 32 valori; lascia il campo vuoto
per un token senza audience.

## JWT TTL {#templateForm-jwtTtl}

Visibile quando **Issue JWT to task runner** è selezionato. Opzionale.

Per quanto tempo il token resta valido dall'avvio del compito, come durata del tipo `30s`, `15m` o
`1h`. Lascia il campo vuoto per usare il valore predefinito del server, un'ora a meno che
l'amministratore non ne abbia impostato un altro. Il valore non può superare il massimo del server,
24 ore se non impostato diversamente.

## Voglio eseguire un compito tramite cron solo per nuovi commit di un repository {#templateForm-cronVisible}

Avvia un compito di questo modello quando compare un nuovo commit in un repository. Semaphore chiede
al repository il suo ultimo commit all'intervallo scelto e avvia un compito solo quando il commit è
cambiato dall'ultimo controllo.

Selezionare la casella e salvare crea il programma del controllo; deselezionarla e salvare elimina
quel programma. Gli altri programmi del modello vengono mantenuti.

## Repository {#templateForm-cronRepositoryId}

Visibile quando **Voglio eseguire un compito tramite cron solo per nuovi commit di un repository** è
selezionato. Obbligatorio.

Il repository da osservare per i nuovi commit. Semaphore osserva il ramo impostato su quel
repository, non il **Ramo** del modello.

## intervalloDiControllo {#templateForm-cronFormat}

Visibile quando **Voglio eseguire un compito tramite cron solo per nuovi commit di un repository** è
selezionato.

Ogni quanto Semaphore controlla il repository alla ricerca di un nuovo commit: da ogni minuto a una
volta al giorno. Il valore predefinito è ogni minuto.

## Sopprimi avvisi di successo {#templateForm-suppressSuccessAlerts}

Impedisce ai compiti riusciti di questo modello di inviare avvisi. Gli avvisi dei compiti falliti
vengono comunque inviati. Ha effetto solo quando gli avvisi sono attivi per il progetto.

## Sopprimi avvisi di errore {#templateForm-suppressErrorAlerts}

Impedisce ai compiti falliti di questo modello di inviare avvisi. Gli avvisi dei compiti riusciti
vengono comunque inviati. Ha effetto solo quando gli avvisi sono attivi per il progetto.

## Argomenti CLI {#templateForm-args}

Argomenti aggiunti al comando di ogni compito. Ogni chip è un argomento, passato così com'è, senza
una shell: metti `--forks=10` in un chip, oppure `--forks` e `10` in due. Trascina i chip per
cambiarne l'ordine. Il pulsante a destra dell'elenco apre gli argomenti come array JSON:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: aggiunti a `ansible-playbook`, prima del playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: aggiunti a `plan` e `apply`, non a `init`.
- Uno script: aggiunti dopo il nome dello script.

## Argomenti CLI {#templateForm-allowOverrideArgsInTask}

Sotto **Richieste**. Consente all'utente di aggiungere propri argomenti all'avvio di un compito.
Vengono aggiunti dopo gli argomenti del modello.

## Ramo {#templateForm-allowOverrideBranchInTask}

Sotto **Richieste**. Consente all'utente di scegliere un altro ramo, tag o commit all'avvio di un
compito, al posto del ramo del modello. Senza di esso il compito esegue sempre il checkout del ramo
del modello.

La casella è disattivata quando il repository è un monorepo.

## Inventario {#templateForm-allowOverrideInventory}

Sotto **Richieste**. Visibile per i modelli Ansible, Terraform, OpenTofu e Terragrunt.

Consente all'utente di scegliere un altro inventario all'avvio di un compito. Solo i modelli Ansible
usano l'inventario scelto lì: un compito Terraform, OpenTofu o Terragrunt viene eseguito nel
workspace del modello anche quando la casella è selezionata.

## Limite {#templateForm-limit}

Visibile per i modelli Ansible.

Gli host o i gruppi dell'inventario su cui eseguire il playbook, passati come `--limit`. Ogni chip è
un pattern, ad esempio `web` o `db01`; lascia l'elenco vuoto per eseguire su ogni host.

## Tag {#templateForm-tags}

Visibile per i modelli Ansible.

Esegue solo i task e i ruoli con questi tag, passati come `--tags`. Lascia l'elenco vuoto per
eseguire tutto.

## Salta tag {#templateForm-skipTags}

Visibile per i modelli Ansible.

Salta i task e i ruoli con questi tag, passati come `--skip-tags`.

## Vaults {#templateVaults}

Visibile per i modelli Ansible.

Le password di Ansible Vault di cui il playbook ha bisogno per decifrare i suoi file. Fai clic su
**+** per aggiungerne una, su un chip per modificarla, sulla croce di un chip per rimuoverla. Ogni
vault ha un nome, il suo ID vault, e fornisce la password da una chiave del Negozio chiavi o da uno
script client del repository.

## Approvazione automatica {#templateForm-autoApprove}

Visibile per i modelli Terraform, OpenTofu e Terragrunt.

Applica il piano senza chiedere. Senza di essa un compito si ferma dopo il piano e attende finché un
utente non lo conferma o lo rifiuta; un piano senza modifiche termina il compito senza nulla da
applicare.

## Override backend settings {#templateForm-overrideBackend}

Visibile per i modelli Terraform, OpenTofu e Terragrunt quando Semaphore può conservare lo stato di
Terraform.

Conserva lo stato del codice in Semaphore. Prima di `init`, Semaphore scrive nella directory del
codice un file che imposta il backend `http`, e lo elimina dopo il compito:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Visibile con **Override backend settings**; attivo quando quella casella è selezionata. Opzionale.

Il nome del file in cui Semaphore scrive il backend, ad esempio `semaphore_backend.tf`. Il valore
predefinito è `backend.tf`. Il nome deve terminare con `.tf` e contenere solo lettere, cifre, `_`,
`-` e `.`. Scegli un nome che nessun file del tuo codice ha: Semaphore sovrascrive quel file e lo
elimina dopo il compito.

## Limite {#templateForm-allowOverrideLimit}

Sotto **Richieste Ansible**. Consente all'utente di impostare gli host su cui eseguire all'avvio di
un compito, al posto del **Limite** del modello.

## Tag {#templateForm-allowOverrideTags}

Sotto **Richieste Ansible**. Consente all'utente di impostare i tag all'avvio di un compito, al posto
dei **Tag** del modello.

## Salta tag {#templateForm-allowOverrideSkipTags}

Sotto **Richieste Ansible**. Consente all'utente di impostare i tag da saltare all'avvio di un
compito, al posto dei **Salta tag** del modello.

## Debug {#templateForm-allowDebug}

Sotto **Richieste Ansible**. Consente all'utente di attivare un output dettagliato all'avvio di un
compito, da `-v` a `-vvvvvv`.

## Esecuzione simulata {#templateForm-showDryRun}

Sotto **Richieste Ansible**. Selezionato per impostazione predefinita. Offre l'opzione **Esecuzione
simulata** all'avvio di un compito, che esegue il playbook con `--check`: Ansible riporta cosa
cambierebbe e non cambia nulla. Deseleziona la casella per togliere l'opzione.

## Differenza {#templateForm-showDiff}

Sotto **Richieste Ansible**. Selezionato per impostazione predefinita. Offre l'opzione **Differenza**
all'avvio di un compito, che esegue il playbook con `--diff` per mostrare le modifiche apportate ai
file. Deseleziona la casella per togliere l'opzione.

## Approvazione automatica {#templateForm-allowAutoApprove}

Sotto le richieste di Terraform, OpenTofu e Terragrunt. Consente all'utente di selezionare **Auto
Approve** all'avvio di un compito, per applicare il piano senza conferma. La finestra di dialogo del
compito offre quell'opzione in ogni caso, ma senza questa casella Semaphore la ignora: il compito
attende una conferma dopo il piano, a meno che **Approvazione automatica** del modello non sia
selezionata.

## Galaxy install options {#templateForm-galaxy}

Visibile per i modelli Ansible. Fai clic per mostrare le opzioni dell'installazione Galaxy;
l'etichetta conta le opzioni che differiscono dai valori predefiniti.

Prima dell'esecuzione del playbook, Semaphore installa con `ansible-galaxy` i ruoli e le collezioni
richiesti dal repository. Cerca `requirements.yml`, `roles/requirements.yml` e
`collections/requirements.yml` accanto al playbook e nella radice del repository, ed esegue
l'installazione solo quando un file è cambiato dall'ultimo compito.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Sotto **Galaxy install options**. Salta l'installazione Galaxy: il playbook viene eseguito con i
ruoli e le collezioni installati in precedenza. Usalo quando i requisiti sono installati in un altro
modo o l'installazione è lenta e non necessaria.

## Role install args {#templateForm-galaxyRoleArgs}

Sotto **Galaxy install options**. Argomenti aggiunti a `ansible-galaxy role install`, uno per chip,
ad esempio `--ignore-certs` o `--server=https://galaxy.example.com`. Semaphore accetta solo i flag
dell'installazione che è sicuro passare: un flag sconosciuto, `--token`, `--api-key`, `-p` o
`--roles-path` impedisce il salvataggio del modello.

## Collection install args {#templateForm-galaxyCollectionArgs}

Sotto **Galaxy install options**. Argomenti aggiunti a `ansible-galaxy collection install`, uno per
chip, ad esempio `--pre` o `--upgrade`. Vale la stessa regola di **Role install args**: `--token`,
`--api-key`, `-p` e `--collections-path` non sono accettati.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Sotto **Galaxy install options**, in **Prompts**. Consente all'utente di scegliere all'avvio di un
compito se saltare l'installazione Galaxy. La scelta del compito prevale su **Skip Galaxy install**
del modello.
