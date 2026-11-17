---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Modifica utente

Un utente è un account che accede a Semaphore. La scheda **Settings** contiene il profilo
dell'account e i suoi flag; **Salva** li scrive. La scheda **Security** contiene la password,
l'autenticazione a due fattori e gli account di accesso collegati, e la scheda **Roles** i ruoli
globali dell'utente: entrambe agiscono subito, senza **Salva**. Ciò che un utente può fare
all'interno di un progetto si imposta nella pagina **Team** di quel progetto, non qui.

## Nome {#user-name}

Il nome completo dell'utente, mostrato nella barra laterale e negli elenchi degli utenti e dei
membri del progetto. Il campo è obbligatorio.

Per un utente **Esterno**, Semaphore prende il nome dal provider di identità all'accesso, quando il
provider ne invia uno, quindi una modifica fatta qui potrebbe non durare oltre l'accesso successivo.

## Nome utente {#user-username}

Il login dell'utente. Il campo è obbligatorio. L'utente accede con il nome utente o con l'email, e
con la password.

Il nome utente di un utente **Esterno** non può essere cambiato: il server rifiuta di salvarlo.

## Email {#user-email}

L'indirizzo email dell'utente. Il campo è obbligatorio. Semaphore vi invia gli avvisi dei compiti
dell'utente, e un'app di autenticazione lo mostra come nome dell'account. È privato: lo vedono solo
l'utente e chi può gestire gli utenti.

Per un utente **Esterno**, l'email proviene dal provider di identità: il campo è di sola lettura, e
Semaphore aggiorna l'indirizzo a ogni accesso.

## Password {#user-password}

Visibile quando viene creato un nuovo utente.

La password con cui l'utente accede. Il campo è obbligatorio, tranne che per un utente **Esterno**,
che non ha una password in Semaphore. Per cambiare la password di un utente esistente, usa **Change
password** nella scheda **Security**.

## Invia avvisi {#user-alert}

Invia all'utente un'email quando un compito fallisce in un progetto in cui l'utente può vedere il
compito. Il progetto deve consentire gli avvisi, e l'amministratore del server deve aver attivato gli
avvisi via email. Un modello di compito può impedire ai suoi compiti falliti di inviare avvisi.

Il flag non influisce sugli avvisi di un progetto inviati a Telegram, Slack e agli altri servizi di
messaggistica.

## Utente admin {#user-admin}

Visibile quando sei un amministratore.

Un amministratore può fare tutto sul server: gestire gli utenti, i ruoli globali e le impostazioni
del server, e aprire ogni progetto con pieni diritti.

Non puoi togliere il flag al tuo account, e il server mantiene almeno un amministratore: rifiuta di
togliere il flag all'ultimo.

## Pro user {#user-pro}

Visibile nell'edizione Pro. Solo un amministratore può cambiare il flag.

Assegna all'utente un posto dell'abbonamento Pro. L'abbonamento ha un numero fisso di utenti Pro:
quando è raggiunto, il server rifiuta di salvare il flag. Senza un abbonamento attivo il flag non
ha effetto.

## Esterno {#user-external}

Visibile quando sei un amministratore. Può essere impostato solo alla creazione dell'utente.

Un utente esterno accede tramite LDAP o OpenID Connect, non con una password di Semaphore: il campo
della password è disattivato, e un accesso con password viene rifiutato. Il nome e l'email vengono
presi dal provider di identità a ogni accesso. Lascia la casella vuota per un utente che accede con
una password.

## Change password {#user-changePassword}

Visibile nella scheda **Security** di un utente esistente che non è **Esterno**, quando il server
consente l'accesso con password.

Apre una finestra di dialogo che imposta una nuova password. Per cambiare la tua password, devi
indicare quella attuale; per reimpostare la password di un altro utente, non serve.

## Time-based one-time password {#user-totp}

Visibile nella scheda **Security** quando l'amministratore ha attivato l'autenticazione a due
fattori per il server.

Attiva l'autenticazione a due fattori per questo utente. Quando attivi l'interruttore, Semaphore
crea subito un segreto e lo mostra come codice QR: scansionalo con un'app di autenticazione. Da quel
momento, un accesso chiede il codice dell'app dopo la password.

Quando il server consente il recupero, sotto il codice QR viene mostrato un codice di recupero.
Viene mostrato una sola volta: copialo e conservalo al sicuro. Disattivare l'interruttore rimuove
subito il segreto.

## LDAP provider {#user-ldapProvider}

Visibile nella scheda **Security** del tuo account quando il server ha un provider LDAP a cui il tuo
account non è ancora collegato.

Il server LDAP a cui collegare il tuo account, quando ce ne sono diversi.

## LDAP username {#user-ldapUsername}

Visibile con **LDAP provider**.

Il tuo nome utente sul server LDAP. Il collegamento ti consente di accedere a questo account anche
tramite LDAP.

## LDAP password {#user-ldapPassword}

Visibile con **LDAP provider**.

La tua password sul server LDAP. Semaphore la usa una volta, per verificare che l'account LDAP sia
tuo, e non la memorizza.

## Link {#user-ldapLink}

Visibile con **LDAP provider**.

Collega il tuo account all'account LDAP indicato sopra. Il collegamento viene rifiutato quando le
credenziali sono errate, quando l'account LDAP è collegato a un altro utente, o quando il tuo
account è già collegato a quel provider. Un account collegato compare sotto **Linked accounts**;
chiudi il suo chip per scollegarlo.

## Global role {#global-role-select}

Visibile nella scheda **Roles** di un utente esistente, nell'edizione Pro, quando puoi gestire i
ruoli globali.

Un ruolo globale da assegnare all'utente, tra i ruoli che l'utente non ha ancora. Un ruolo globale
concede permessi sull'intero server: gestire gli utenti, i ruoli globali o le impostazioni del
server, o leggere il registro di audit. I ruoli globali stessi si creano nella scheda **Ruoli
globali** della pagina Utenti.

## Assign {#globalRoleAssign-assign}

Visibile con **Global role**.

Assegna subito all'utente il ruolo selezionato. I ruoli che l'utente ha sono elencati sotto,
ognuno con un pulsante per toglierlo. **Effective global permissions** mostra cosa può fare
l'utente alla fine, e da dove proviene ogni permesso.
