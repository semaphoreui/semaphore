---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nuovo compito

Avvia un compito del modello: un'esecuzione del suo playbook, del suo codice Terraform o del suo
script. Il modello decide quali campi mostra questa finestra di dialogo, e ogni sezione qui sotto
indica quando il suo campo è presente. Il pulsante in basso è **Esegui**, **Costruisci** o
**Distribuisci**, a seconda del tipo del modello.

Quando il modello ha un sondaggio, le sue domande sono campi dopo **Messaggio**. Un campo
contrassegnato da `*` è obbligatorio, un campo Integer accetta solo cifre, e le risposte predefinite
sono già compilate. Semaphore passa le risposte al compito come variabili: variabili extra per
Ansible, `-var` per Terraform, argomenti `name=value` per uno script, o variabili di ambiente per una
domanda impostata così. Una risposta sostituisce una variabile con lo stesso nome nei gruppi di
variabili del modello. Una risposta segreta viene nascosta mentre digiti e non viene conservata con
il compito.

## Commit {#newTask-commit}

Visibile quando esegui di nuovo un compito precedente, e quel compito è stato eseguito su un commit
noto.

Il compito esegue il checkout di questo commit del repository invece del commit più recente del
ramo. Chiudi il riquadro per eseguire il commit più recente.

Il commit viene usato solo quando il modello consente di sovrascrivere il ramo, e mai per un
repository contrassegnato come monorepo. Altrimenti il compito esegue il commit più recente.

## Versione di build {#newTask-buildTaskId}

Visibile quando il tipo del modello è **Distribuisci**.

La build da distribuire: uno degli ultimi 20 compiti riusciti del modello di build, elencati per
versione e messaggio. È selezionato il più recente. Il campo è obbligatorio.

La versione della build arriva al compito: come `semaphore_vars.task_details.incoming_version` in
Ansible, come variabile di ambiente `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` in uno script.
L'elenco è vuoto quando non puoi vedere il modello di build.

## Messaggio {#newTask-message}

Opzionale. Una nota sul perché il compito viene eseguito, ad esempio `Hotfix for the login page`.
Viene mostrata con il compito nella cronologia dei compiti e non cambia nulla nell'esecuzione.

Il playbook può leggerla come `semaphore_vars.task_details.message`, e uno script come variabile di
ambiente `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Ramo {#newTask-gitBranch}

Visibile quando il modello consente di sovrascrivere il ramo, a meno che il repository non sia un
monorepo.

Il ramo o il tag del repository di cui il compito esegue il checkout. Il campo parte con il ramo del
modello. Quando è vuoto, il compito usa il ramo del modello, o il ramo del repository quando il
modello non ne ha uno. Un nome che non è un nome di ramo Git valido viene rifiutato.

## Inventario {#newTask-inventoryId}

Visibile quando il modello consente di scegliere l'inventario in un compito e puoi vedere il
progetto. Per un modello Terraform, OpenTofu o Terragrunt il campo si chiama **Workspace**, ma il
server usa la scelta solo per i modelli Ansible: un compito Terraform, OpenTofu o Terragrunt viene
eseguito nel workspace del modello, qualunque cosa tu scelga qui.

L'inventario su cui viene eseguito il compito. Il campo parte con l'inventario del modello ed elenca
gli inventari del progetto creati per l'app del modello. Quando l'inventario scelto non esiste più
al momento dell'avvio del compito, il compito viene eseguito sull'inventario del modello.

## Limite {#taskParams-limit}

Visibile per un modello Ansible che consente di sovrascrivere il limite.

Gli host o i gruppi dell'inventario su cui viene eseguito il playbook, un pattern per chip, ad
esempio `web` o `10.0.0.11`. Semaphore li passa come `--limit` uniti da virgole. L'elenco parte con
il limite del modello. Rimuovi tutti i chip per eseguire su tutti gli host dell'inventario.

## Tag {#taskParams-tags}

Visibile per un modello Ansible che consente di sovrascrivere i tag.

Vengono eseguiti solo i task e i ruoli del playbook con uno di questi tag: Semaphore li passa come
`--tags`, un tag per chip. L'elenco parte con i tag del modello. Rimuovi tutti i chip per eseguire
l'intero playbook.

## Salta tag {#taskParams-skipTags}

Visibile per un modello Ansible che consente di sovrascrivere i tag saltati.

I task e i ruoli del playbook con uno di questi tag non vengono eseguiti: Semaphore li passa come
`--skip-tags`, un tag per chip. L'elenco parte con i tag saltati del modello.

## Debug {#taskParams-debug}

Visibile per un modello Ansible che consente il debug.

Esegue `ansible-playbook` con un flag di verbosità, così il log del compito mostra più dettagli.
L'etichetta mostra il flag: `-v` tante volte quante indica **Debug level**.

## Debug level {#taskParams-debugLevel}

Visibile con **Debug**, e attivo quando **Debug** è selezionato.

Quanto è dettagliato il log, da 1 (`-v`) a 6 (`-vvvvvv`). Il valore predefinito è 4 (`-vvvv`), che
mostra anche i dettagli delle connessioni agli host.

## Esecuzione simulata {#taskParams-dryRun}

Visibile per un modello Ansible, a meno che il modello non la nasconda.

Esegue il playbook con `--check`: Ansible riporta cosa cambierebbe sugli host e non cambia nulla.

## Differenza {#taskParams-diff}

Visibile per un modello Ansible, a meno che il modello non la nasconda.

Esegue il playbook con `--diff`: il log mostra le modifiche apportate a file e template sugli host.
Seleziona anche **Esecuzione simulata** per vedere le modifiche senza apportarle.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Visibile per un modello Ansible che consente di sovrascrivere l'installazione Galaxy.

Prima dell'esecuzione del playbook, Semaphore installa i ruoli e le collezioni elencati nei file
`requirements.yml` del repository, quando un file è cambiato dall'ultima installazione. Seleziona la
casella per saltare quel passaggio per questo compito.

La casella parte deselezionata qualunque cosa imposti il modello, e una casella deselezionata
installa i requisiti.

## Plan {#taskParams-plan}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Esegue `plan` e termina lì il compito: non viene applicato nulla. Usalo per rivedere prima le
modifiche. Senza di esso, il compito prosegue applicando il piano quando contiene modifiche.

## Destroy {#taskParams-destroy}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Aggiunge `-destroy` a `plan` e `apply`: il compito pianifica, e poi applica, la rimozione di ogni
risorsa nello stato del workspace. Seleziona anche **Plan** per vedere cosa verrebbe distrutto senza
distruggerlo.

## Auto Approve {#taskParams-autoApprove}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Applica il piano appena viene creato. Senza di esso, il compito si ferma dopo il piano e attende che
qualcuno lo confermi o lo rifiuti nella pagina del compito.

La casella ha effetto solo quando il modello consente l'approvazione automatica. Un modello
impostato per approvare applica sempre senza chiedere, qualunque cosa indichi la casella.

## Upgrade {#taskParams-upgrade}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Esegue `init` con `-upgrade`: i provider e i moduli vengono aggiornati alle versioni più recenti
consentite dai loro vincoli di versione.

## Reconfigure {#taskParams-reconfigure}

Visibile per un modello Terraform, OpenTofu o Terragrunt.

Esegue `init` con `-reconfigure`: il backend viene configurato di nuovo, e la configurazione salvata
del backend viene ignorata. Senza di esso, Semaphore esegue `init -migrate-state`, che sposta lo
stato esistente in un backend modificato.

## Argomenti CLI {#newTask-args}

Visibile quando il modello consente di sovrascrivere gli argomenti CLI.

Argomenti aggiuntivi per la riga di comando del compito, uno per chip: `--forks=10` è un chip,
`--forks` e `10` sono due. Fai clic su un chip per modificarlo, trascinalo per cambiare l'ordine, e
usa l'ultimo chip per aggiungerne uno.

L'elenco parte con gli argomenti del modello. Semaphore passa sempre gli argomenti del modello e
aggiunge questo elenco dopo di essi, quindi una volta che modifichi l'elenco, gli argomenti del
modello rimasti al suo interno vengono passati due volte. Rimuovili e mantieni solo gli argomenti
che aggiungi.
