---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Modifica integrazione

Un'integrazione avvia un modello di compito, o un workflow, quando un altro servizio chiama un URL
webhook di Semaphore, ad esempio con un evento push di GitHub. L'URL è un alias, `/api/integrations/`
e un codice casuale: gli alias del progetto sono nella pagina Integrazioni, gli alias di una singola
integrazione nella sua pagina. Quella pagina contiene anche i corrispondenti e i valori che
l'integrazione prende dalla richiesta.

## Name {#integrationForm-name}

Il nome con cui l'integrazione compare nella pagina Integrazioni. I compiti che l'integrazione avvia
le vengono attribuiti nel registro di audit. Il campo è obbligatorio.

## Runs {#integrationForm-targetToggle}

Visibile quando il server supporta i workflow.

Cosa avvia l'integrazione: **Template** avvia un compito di un modello di compito, **Workflow**
avvia un'esecuzione di un workflow. Passare a **Workflow** scarta i parametri del compito qui
sotto.

## Workflow to run {#integrationForm-workflowId}

Visibile quando **Runs** è **Workflow**.

Il workflow che una richiesta avvia. Ogni valore che l'integrazione prende dalla richiesta viene
passato al workflow come risposta alla sua variabile del sondaggio con lo stesso nome. Le
integrazioni di workflow richiedono l'edizione Pro: senza di essa il campo è disattivato e il server
rifiuta la modifica.

## Task Template to run {#integrationForm-templateId}

Visibile quando **Runs** è **Template**.

Il modello di compito di cui una richiesta avvia un compito. Il campo è obbligatorio. Una volta
scelto un modello, i parametri dei suoi compiti compaiono sotto di esso.

## Task parameters {#integrationForm-taskParams}

Visibile quando è scelto un modello di compito.

I parametri che riceve ogni compito avviato da questa integrazione, come se lo avviassi tu stesso dal
modello. Qui ci sono solo le opzioni che il modello consente a un compito di sovrascrivere. I valori
presi dalla richiesta sovrascrivono questi: un valore di tipo **Variables** una risposta con lo
stesso nome, un valore di tipo **Task Params** un parametro con lo stesso nome.

Quando il modello consente argomenti CLI in un compito, **Argomenti CLI** contiene argomenti
aggiuntivi per la riga di comando del compito.

## Messaggio {#taskParams-message}

Opzionale. Il messaggio di ogni compito avviato dall'integrazione, mostrato nella cronologia dei
compiti.

## Variabili del sondaggio {#taskParams-survey}

Visibile quando il modello ha variabili del sondaggio.

Le risposte al sondaggio del modello, le stesse per ogni compito. Un valore di tipo **Variables**
preso dalla richiesta sostituisce la risposta con lo stesso nome.

Le risposte alle variabili di tipo secret non vengono salvate con l'integrazione: passa invece un
segreto tramite il gruppo di variabili del modello.

## Ramo {#taskParams-gitBranch}

Visibile quando il modello consente di sovrascrivere il ramo in un compito.

Il ramo o il tag del repository di cui i compiti eseguono il checkout. Lascialo così com'è per usare
il ramo del modello. Un valore chiamato `git_branch` di tipo **Task Params**, preso dalla richiesta,
lo sovrascrive; un prefisso `refs/heads/` viene rimosso da quel valore.

## Inventario {#taskParams-inventoryId}

Visibile quando il modello consente di sovrascrivere l'inventario in un compito.

L'inventario, o per Terraform il workspace, su cui vengono eseguiti i compiti al posto di quello del
modello.

## Opzioni Ansible {#taskParams-ansible}

Visibile quando il modello esegue un playbook Ansible.

Le opzioni di `ansible-playbook` per ogni compito, ognuna mostrata quando il modello la consente:
**Limite**, **Tag**, **Salta tag**, **Debug**, **Esecuzione simulata** (`--check`), **Differenza** e
**Skip Galaxy install**.

## Opzioni Terraform {#taskParams-terraform}

Visibile quando il modello esegue Terraform, OpenTofu o Terragrunt.

- **Plan**: il compito esegue il piano e si ferma, non viene applicato nulla.
- **Destroy**: il piano e l'applicazione distruggono le risorse.
- **Auto Approve**: il piano viene applicato senza conferma, quando il modello consente
  l'approvazione automatica. Senza di essa un compito attende finché qualcuno non lo conferma in
  Semaphore.
- **Upgrade** e **Reconfigure**: `init` viene eseguito con `-upgrade` e `-reconfigure`.

## Auth method {#integrationForm-authMethod}

Come Semaphore verifica che una richiesta provenga dal servizio che ti aspetti. Una richiesta che
non supera il controllo non avvia nulla, e il webhook le risponde come a qualsiasi altra richiesta.

- **None**: nessun controllo. Chiunque conosca l'URL può avviare il compito.
- **GitHub Webhooks**: l'intestazione `X-Hub-Signature-256` deve contenere la firma HMAC-SHA256 del
  corpo creata con il segreto, così come la invia GitHub.
- **Bitbucket Webhooks**: lo stesso controllo sull'intestazione `X-Hub-Signature`, così come la
  invia Bitbucket.
- **Token**: l'intestazione indicata in **Auth header** deve contenere il segreto stesso.
- **HMAC** e **HMAC (SHA-512)**: l'intestazione indicata in **Auth header** deve contenere l'HMAC-SHA256
  o l'HMAC-SHA512 del corpo creato con il segreto, in esadecimale minuscolo, senza un prefisso
  come `sha256=`.
- **BasicAuth**: la richiesta deve portare il nome utente e la password della chiave indicata in
  **Segreto di autenticazione** nella sua intestazione `Authorization`.

## Auth header {#integrationForm-authHeader}

Visibile quando **Auth method** è **Token**, **HMAC** o **HMAC (SHA-512)**.

Il nome dell'intestazione della richiesta che porta il token o la firma, ad esempio
`X-Webhook-Token`. Maiuscole e minuscole nel nome non contano. Una richiesta senza quell'intestazione
viene rifiutata.

## Segreto di autenticazione {#integrationForm-authSecretId}

Visibile quando **Auth method** non è **None**.

La chiave del Negozio chiavi con cui viene verificata la richiesta. Per **BasicAuth** va bene solo
una chiave di tipo **Accesso con password**. Per gli altri metodi va bene anche una chiave di tipo
**Stringa**: il controllo usa il suo valore, o la password di una chiave **Accesso con password**.

Il campo è necessario per ogni metodo tranne **None**: senza una chiave, o con un segreto vuoto, ogni
richiesta viene rifiutata.
