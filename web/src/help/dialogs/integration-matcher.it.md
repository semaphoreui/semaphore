---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Un corrispondente è una condizione sulla richiesta in arrivo. I corrispondenti decidono quale
integrazione avvia una richiesta all'alias del progetto, così un solo URL può servire più
integrazioni: un'integrazione viene eseguita solo quando ognuno dei suoi corrispondenti è
soddisfatto. Vengono usati solo quando **Usa alias progetto** è attivo per l'integrazione; con
questa opzione attiva, l'integrazione è raggiungibile solo tramite l'alias del progetto.

Un'integrazione trovata tramite l'alias del progetto senza corrispondenti non viene mai eseguita:
aggiungine almeno uno.

## Name {#integrationMatcherForm-name}

Il nome con cui il corrispondente compare nella pagina dell'integrazione. Non cambia ciò che il
corrispondente verifica. Il campo è obbligatorio.

## Match on {#integrationMatcherForm-matchType}

La parte della richiesta che il corrispondente legge:

- **Header**: il valore di un'intestazione della richiesta.
- **Body**: il corpo della richiesta.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Visibile quando **Match on** è **Body**.

Come viene letto il corpo:

- **JSON**: il corpo viene analizzato come JSON, e **Key** è il percorso di un valore al suo
  interno.
- **String**: l'intero corpo viene confrontato come testo, e **Key** non viene usato.

## Key {#integrationMatcherForm-key}

Cosa leggere dalla richiesta. Il campo è obbligatorio.

- Per **Header**: il nome dell'intestazione, ad esempio `X-GitHub-Event`. Maiuscole e minuscole non
  contano.
- Per **Body** in **JSON**: il percorso del valore, con le sue parti unite da punti; un elemento di
  un array si scrive come `.[0]`, ad esempio `repository.full_name` o `commits.[0].id`.
- Per **Body** in **String**: non usato, ma il modulo richiede un valore; digita qualsiasi cosa.

## Comparison Method {#integrationMatcherForm-method}

Come il valore letto dalla richiesta viene confrontato con **Value**:

- **==**: i due sono uguali.
- **!=**: i due sono diversi. Un'intestazione o una chiave che manca nella richiesta conta come
  diversa.
- **Contains**: il valore letto dalla richiesta contiene **Value**.

## Value {#integrationMatcherForm-value}

Il testo con cui confrontare, ad esempio `push` o `refs/heads/main`. Il confronto è esatto e
distingue tra maiuscole e minuscole. Un numero o un booleano del corpo JSON viene confrontato nella
sua forma testuale: `42`, `true`. Il campo è obbligatorio.
