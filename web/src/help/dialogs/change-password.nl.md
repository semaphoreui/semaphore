---
form: src/components/ChangePasswordForm.vue
---

# Wachtwoord Wijzigen

Stelt een nieuw wachtwoord in om met een gebruikersnaam en een wachtwoord bij Semaphore in te
loggen. Het nieuwe wachtwoord wordt van kracht wanneer u op **Opslaan** drukt. Een externe
gebruiker, die inlogt via LDAP of OpenID Connect, heeft hier geen wachtwoord om te wijzigen.

## Current password {#changePassword-currentPassword}

Wordt getoond wanneer u uw eigen wachtwoord wijzigt.

Uw huidige wachtwoord. Het veld is verplicht: Semaphore controleert het voordat het nieuwe wordt
ingesteld, zodat een gestolen sessie of API-token niet in uw wachtwoord kan worden omgezet. De
server weigert de wijziging wanneer het wachtwoord niet overeenkomt.

Een beheerder of een gebruikersbeheerder die het wachtwoord van een andere gebruiker opnieuw
instelt, ziet dit veld niet.

## Wachtwoord {#changePassword-password}

Het nieuwe wachtwoord. Het veld is verplicht. Klik op het oogpictogram om te tonen wat u hebt
getypt.

Semaphore slaat alleen een hash van het wachtwoord op.
