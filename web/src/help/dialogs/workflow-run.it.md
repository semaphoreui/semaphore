---
form: src/components/WorkflowRunDialog.vue
---

# Esegui workflow

Avvia un'esecuzione del workflow dopo che hai risposto al suo sondaggio. La finestra di dialogo si
apre solo per un workflow con un sondaggio: un workflow senza sondaggio parte subito. Dopo **Esegui**
si apre la pagina della nuova esecuzione; quando il server rifiuta i valori, la finestra di dialogo
resta aperta con le tue risposte.

## Variables {#workflowRunDialog-survey}

Un campo per ogni variabile del sondaggio del workflow, nell'ordine del sondaggio, con il suo valore
predefinito già compilato. Un campo contrassegnato da `*` è obbligatorio, un campo Integer accetta
solo cifre e un campo elenco accetta un valore oppure, quando la variabile ne consente diversi, un
numero qualsiasi di valori. **Esegui** non avvia nulla finché ogni campo non è valido.

I valori vengono memorizzati con l'esecuzione e mostrati nella sua pagina sotto **Variables**. Un
compito del workflow non li riceve da solo: una variabile del suo modello riceve un valore solo
quando il suo nodo la collega a una variabile del workflow. Il solo nome non trasmette nulla.

Un sondaggio di workflow non contiene segreti, perché l'esecuzione conserva i suoi valori in chiaro.
Metti invece password e token in una chiave o in un gruppo di variabili.
