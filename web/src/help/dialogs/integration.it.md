---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Modifica esecuzione

Un’esecuzione è ciò che un’integrazione avvia quando un altro servizio chiama un URL webhook di
Semaphore, ad esempio per un evento push di GitHub: un task di un template di task oppure un
workflow. L’URL e il controllo della richiesta appartengono all’endpoint dell’integrazione, non
all’esecuzione. La pagina di un’esecuzione contiene i suoi matcher e i valori che ricava dalla
richiesta.

## Nome {#integrationForm-name}

Il nome con cui l'integrazione compare nella pagina Integrazioni. I compiti che l'integrazione avvia
le vengono attribuiti nel registro di audit. Il campo è obbligatorio.

## Integrazione {#integrationForm-groupId}

L’integrazione a cui appartiene l’esecuzione. L’esecuzione parte con le richieste che raggiungono
l’endpoint di quell’integrazione e ne superano l’autorizzazione. Il campo è obbligatorio.

## Destinazione {#integrationForm-targetToggle}

Visibile quando il server supporta i workflow.

Cosa avvia l'integrazione: **Template** avvia un compito di un modello di compito, **Workflow**
avvia un'esecuzione di un workflow. Passare a **Workflow** scarta i parametri del compito qui
sotto.

## Workflow to run {#integrationForm-workflowId}

Visibile quando **Destinazione** è **Workflow**.

Il workflow che una richiesta avvia. Ogni valore che l'integrazione prende dalla richiesta viene
passato al workflow come risposta alla sua variabile del sondaggio con lo stesso nome. Le
integrazioni di workflow richiedono l'edizione Pro: senza di essa il campo è disattivato e il server
rifiuta la modifica.

## Template di task da eseguire {#integrationForm-templateId}

Visibile quando **Destinazione** è **Template**.

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
