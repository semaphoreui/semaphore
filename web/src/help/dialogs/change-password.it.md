---
form: src/components/ChangePasswordForm.vue
---

# Cambia password

Imposta una nuova password per accedere a Semaphore con un nome utente e una password. La nuova
password ha effetto quando premi **Salva**. Un utente esterno, che accede tramite LDAP o OpenID
Connect, non ha qui alcuna password da cambiare.

## Current password {#changePassword-currentPassword}

Visibile quando cambi la tua password.

La tua password attuale. Il campo è obbligatorio: Semaphore la verifica prima di impostare quella
nuova, così che una sessione o un token API rubati non possano essere trasformati nella tua
password. Il server rifiuta la modifica quando la password non corrisponde.

Un amministratore o un gestore degli utenti che reimposta la password di un altro utente non vede
questo campo.

## Password {#changePassword-password}

La nuova password. Il campo è obbligatorio. Fai clic sull'icona dell'occhio per mostrare ciò che hai
digitato.

Semaphore memorizza solo un hash della password.
