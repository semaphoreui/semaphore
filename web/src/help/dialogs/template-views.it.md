---
form: src/components/EditViewsForm.vue
---

# Modifica viste

Le viste sono le schede sopra l'elenco dei modelli di compito. Una vista mostra i modelli che vi
inserisci, oppure tutti i modelli del progetto. La finestra di dialogo salva subito ogni modifica:
non c'è un pulsante Salva per l'intera finestra. Trascina una vista per la sua maniglia per cambiare
l'ordine delle schede.

## Titolo {#viewsForm-title}

Il nome della scheda. Digitalo e confermalo con il segno di spunta accanto al campo; la croce
ripristina il titolo salvato. Una nuova vista viene salvata quando ha un titolo e lo confermi o
cambi una delle sue impostazioni. Chiudere la finestra di dialogo non salva un titolo che non hai
confermato.

L'ingranaggio apre le impostazioni della vista sotto il suo titolo.

## Delete {#viewsForm-viewDelete}

L'icona del cestino elimina subito la vista, senza chiedere conferma. I modelli della vista non
vengono eliminati: restano nel progetto e non sono più in alcuna vista.

## Type {#viewsForm-type}

Quali modelli mostra la scheda:

- **Custom**: i modelli il cui campo **Vista**, nella finestra di dialogo del modello, indica questa
  vista. È il valore predefinito.
- **All**: tutti i modelli del progetto, qualunque sia la loro vista.

Un membro vede solo i modelli che può vedere, in qualsiasi vista.

## Sort by {#viewsForm-sortColumn}

La colonna in base alla quale sono ordinati i modelli della scheda; **Name** è l'unica. Anche una
scheda senza di essa è ordinata per nome, quindi il campo conta insieme a **Reverse**.

## Reverse {#viewsForm-sortReverse}

Ordina la scheda in ordine decrescente. Funziona solo quando **Sort by** è impostato.

## Hidden {#viewsForm-hidden}

Nasconde la scheda ai membri che non possono gestire le risorse del progetto. Chi può gestirle la
vede ancora, barrata. Nascondere una scheda non nasconde i suoi modelli: una vista di tipo **All**
continua a elencarli.

## Show task counts {#viewsForm-showTemplateStats}

Mostra sulla scheda quanti dei suoi modelli sono in ciascuno stato, in un chip di segmenti colorati:
verde per un ultimo task riuscito, rosso per uno fallito, grigio per uno arrestato o per un modello
mai avviato, arancione per un task che attende una conferma, blu per uno in coda o in esecuzione. È
disattivato per impostazione predefinita, e i conteggi vengono calcolati solo per le viste in cui è
attivo.

## Aggiungi vista {#viewsForm-add}

Aggiunge una vista alla fine dell'elenco, con un titolo vuoto da digitare.
