---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

La chiave di abbonamento sblocca le funzionalità di Semaphore PRO ed Enterprise su questo server. La
finestra di dialogo attiva una chiave e mostra il piano di quella attiva: il suo titolare, il suo
stato, quando scade e quanti utenti Pro, esecutori e altri elementi consente e sono in uso. Viene
aperta da un amministratore dal menu utente dell'edizione Pro. La finestra di dialogo si chiama
**Upgrade to Semaphore PRO** finché un abbonamento non è attivo.

## Enter your PRO or EE key {#subscription-key}

La chiave di abbonamento, così come l'hai ricevuta con il tuo abbonamento PRO o Enterprise.
Incollala qui, o caricala da un file con il pulsante di caricamento, poi premi **Activate New Key**.
Il campo è obbligatorio per l'attivazione.

Il campo è di sola lettura quando la chiave è impostata nella configurazione del server
(`subscription.key` o `subscription.key_file`, o le variabili `SEMAPHORE_SUBSCRIPTION_KEY`
e `SEMAPHORE_SUBSCRIPTION_KEY_FILE`): la chiave può allora essere cambiata solo lì.

## Upload key file {#subscription-upload}

Visibile quando non è ancora stato attivato alcun abbonamento.

Legge la chiave da un file sul tuo computer (`.txt`, `.key`, `.pem` o `.lic`) nel campo sopra. La
chiave non viene attivata finché non premi **Activate New Key**.

## Key actions {#subscription-actionsMenu}

Visibile quando un abbonamento è attivo o è scaduto.

Un menu di azioni sulla chiave del server:

- **Reload**: recupera di nuovo lo stato dell'abbonamento, ad esempio dopo che è stato rinnovato.
- **Upload**: legge una nuova chiave da un file nel campo sopra, come fa il pulsante di
  caricamento.
- **Reset**: rimuove subito la chiave dal server, senza conferma.

## Activate New Key {#subscription-activate}

Attiva su questo server la chiave del campo. Quando la chiave è valida, il piano dell'abbonamento
viene mostrato sotto. Quando il tuo account non è ancora un utente Pro e il piano ha un posto libero
per un utente Pro, la finestra di dialogo chiede se rendere il tuo account un utente Pro.

Una chiave già attivata viene segnalata con un avviso. Il pulsante è disattivato quando la chiave è
impostata nella configurazione del server.

## Buy Pro {#subscription-buy}

Apre la pagina del portale Semaphore in cui si acquista un abbonamento PRO, in una nuova scheda del
browser.

## Get 30-day free trial {#subscription-trial}

Visibile finché nessun abbonamento è attivo.

Apre la pagina del portale Semaphore che rilascia una chiave per una prova gratuita di 30 giorni, in
una nuova scheda del browser. Incolla la chiave che ricevi nel campo sopra.
