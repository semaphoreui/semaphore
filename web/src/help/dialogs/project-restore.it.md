---
form: src/components/RestoreProjectForm.vue
---

# Ripristina progetto...

Crea un nuovo progetto da un file di backup creato con **Backup progetto** nella pagina Impostazioni
di un progetto. Il nuovo progetto riceve i modelli di compito, gli inventari, i repository, i gruppi
di variabili, le chiavi, i programmi, le integrazioni e le altre cose configurate nel backup, e tu
ne diventi il proprietario. Solo un amministratore può ripristinare un progetto, a meno che
l'amministratore del server non consenta a ogni utente di creare progetti.

## Nome progetto {#restoreProject-name}

Il nome del nuovo progetto. Il campo è obbligatorio e sostituisce il nome scritto nel file di
backup.

Il nome deve essere diverso dai nomi di tutti i progetti sul server: il server rifiuta di
ripristinare un progetto con un nome già in uso.

## Backup file {#restoreProject-file}

Il file JSON del backup. Scegli il file prima di premere **Restore**.

Un backup esclude i segreti delle chiavi e i segreti nella scheda **Secrets** dei gruppi di
variabili: dopo il ripristino, Semaphore ti dice quante chiavi sono vuote e richiedono di inserire
di nuovo i loro segreti, e anche i segreti dei gruppi di variabili devono essere inseriti di nuovo.
Una password digitata direttamente in un altro campo, come l'indirizzo di un repository, torna così
com'è nel file. I membri del progetto e la cronologia dei suoi compiti non fanno parte del backup.
Un backup scritto da una versione più recente di Semaphore viene rifiutato.
