---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Un valore estratto prende una parte della richiesta in arrivo, come il ramo di un evento push, e la
passa al compito che l'integrazione avvia. Ogni valore dell'integrazione viene preso da ogni
richiesta che avvia un compito.

## Name {#integrationValueForm-name}

Il nome con cui il valore compare nella pagina dell'integrazione. Non cambia cosa viene preso né
dove va. Il campo è obbligatorio.

## Source of the Value {#integrationValueForm-valueSource}

La parte della richiesta da cui viene preso il valore:

- **Header**: il valore di un'intestazione della richiesta.
- **Body**: il corpo della richiesta.

## Data Type of Body {#integrationValueForm-bodyDataType}

Visibile quando **Source of the Value** è **Body**.

Come viene letto il corpo:

- **JSON**: il corpo viene analizzato come JSON, e **Key** è il percorso di un valore al suo
  interno. Il valore mantiene il suo tipo JSON: un numero, un elenco o un oggetto arrivano al
  compito come tali.
- **String**: l'intero corpo viene preso come testo, e **Key** non viene usato.

## Key {#integrationValueForm-key}

Cosa prendere dalla richiesta. Il campo è obbligatorio.

- Per **Header**: il nome dell'intestazione, ad esempio `X-GitHub-Event`. Maiuscole e minuscole non
  contano. Un'intestazione che manca nella richiesta dà un valore vuoto.
- Per **Body** in **JSON**: il percorso del valore, con le sue parti unite da punti; un elemento di
  un array si scrive come `.[0]`, ad esempio `repository.full_name` o `commits.[0].id`. Un percorso
  che manca nel corpo non dà alcun valore, e il compito mantiene quello impostato
  nell'integrazione.
- Per **Body** in **String**: non usato, ma il modulo richiede un valore; digita qualsiasi cosa.

## Variable Usage {#integrationValueForm-variableType}

Dove va il valore in un compito di un modello:

- **Variables**: una variabile extra del compito, nel modo in cui viene passata la risposta a una
  variabile del sondaggio. Sostituisce una risposta o una variabile del gruppo di variabili con lo
  stesso nome.
- **Task Params**: un parametro del compito, nel modo in cui vengono memorizzate le opzioni della
  finestra di dialogo del compito. Il nome `git_branch` imposta il ramo del compito quando il
  modello consente di sovrascrivere il ramo; un prefisso `refs/heads/` viene rimosso.

Un'integrazione che esegue un workflow ignora l'uso: ogni valore diventa la risposta alla variabile
del sondaggio del workflow indicata in **Variable**.

## Variable {#integrationValueForm-variable}

Il nome con cui viene passato il valore, ad esempio `branch` per una variabile extra che il playbook
legge come `{{ branch }}`. Il campo è obbligatorio.
