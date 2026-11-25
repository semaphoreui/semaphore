---
form: src/components/SurveyVars.vue
---

# Variabile del sondaggio

Una variabile del sondaggio è una domanda che la finestra di dialogo del compito pone ogni volta che
esegui il modello. La risposta arriva al compito come variabile: una variabile extra per Ansible, un
`-var` per Terraform e OpenTofu, un argomento per uno script, o una variabile di ambiente. La
variabile viene salvata con il modello, quando salvi la finestra di dialogo del modello. La stessa
finestra di dialogo modifica il sondaggio di un workflow, dove il tipo **Secret** non è offerto.

## Nome {#surveyVarDialog-name}

Il nome con cui il compito riceve la risposta: il nome della variabile extra, della variabile
Terraform, dell'argomento `name=value` di uno script o della variabile di ambiente. Gli spazi alle
estremità vengono rimossi. Il campo è obbligatorio.

Usa un nome che nessun'altra variabile del sondaggio ha, e che lo strumento accetta, ad esempio
`app_version`. Una risposta sostituisce una variabile con lo stesso nome del gruppo di variabili del
modello.

## Titolo {#surveyVarDialog-title}

L'etichetta del campo nella finestra di dialogo del compito, ad esempio `Version to deploy`. È anche
il testo del chip della variabile nella finestra di dialogo del modello. Il campo è obbligatorio.

## Descrizione {#surveyVarDialog-description}

Opzionale. Un suggerimento mostrato sotto il campo della finestra di dialogo del compito mentre ha
il focus: cosa inserire, o un esempio.

## Tipo {#surveyVarDialog-type}

Che tipo di risposta chiede la finestra di dialogo del compito:

- **String**: una riga di testo. È il valore predefinito.
- **Integer**: solo cifre; la finestra di dialogo del compito rifiuta qualsiasi altra cosa, anche un
  segno meno. La risposta arriva comunque al compito come testo: convertila nel playbook, ad esempio
  con `| int`.
- **Secret**: testo che la finestra di dialogo del compito maschera. La risposta non viene salvata
  con il compito, quindi una riesecuzione la chiede di nuovo, e Ansible la riceve in un file cifrato
  con vault. Un segreto non ha un valore predefinito.
- **Enum**: un valore scelto da un elenco che definisci sotto.
- **Text**: più righe di testo.
- **Select**: più valori scelti da un elenco che definisci sotto. Il compito li riceve come elenco;
  in un argomento, in un `-var` o in una variabile di ambiente l'elenco viene scritto come JSON, ad
  esempio `["a","b"]`.

Cambiare il tipo cancella il valore predefinito, e un tipo senza elenco cancella i valori.

## Pass variable as {#surveyVarDialog-target}

Come la risposta arriva al compito:

- **Extra variable**: il modo abituale dello strumento. Ansible la riceve come variabile extra,
  Terraform e OpenTofu come `-var name=value`, uno script Bash, PowerShell o Python come argomento
  `name=value` dopo lo script.
- **Environment variable**: una variabile di ambiente del processo, `name=value`, per ogni tipo di
  modello. La variabile allora non viene passata nell'altro modo.

## Chiave {#surveyVarDialog-valueName}

Visibile quando il tipo è **Enum** o **Select**.

Il testo che la finestra di dialogo del compito mostra per questo valore nell'elenco. Ogni valore ha
bisogno di una chiave, e le chiavi devono essere diverse: altrimenti la finestra di dialogo rifiuta
di salvare la variabile.

## Valore {#surveyVarDialog-valueValue}

Visibile quando il tipo è **Enum** o **Select**.

Cosa riceve il compito quando viene scelto questo valore. Può essere diverso dalla sua chiave: la
chiave `Production` può passare `prod`.

## Add Value {#surveyVarDialog-addValue}

Visibile quando il tipo è **Enum** o **Select**.

Aggiunge una riga vuota all'elenco dei valori. Un **Enum** o un **Select** ha bisogno di almeno un
valore. L'icona del cestino di una riga la rimuove.

## Default value {#surveyVarDialog-defaultValue}

Visibile per ogni tipo tranne **Secret**.

La risposta che la finestra di dialogo del compito compila all'apertura; puoi cambiarla lì. Per un
**Enum** scegli uno dei valori, per un **Select** un numero qualsiasi di essi. Lascia il campo vuoto
per iniziare con una risposta vuota.

Il valore predefinito viene compilato solo dalla finestra di dialogo del compito. Un compito avviato
in un altro modo, ad esempio tramite l'API, non riceve alcun valore predefinito.

## Obbligatorio {#surveyVarDialog-required}

La finestra di dialogo del compito non avvia un compito finché la variabile non ha una risposta.
Come il valore predefinito, il controllo viene fatto solo dalla finestra di dialogo del compito, non
dal server.
