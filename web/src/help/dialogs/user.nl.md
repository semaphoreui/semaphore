---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Gebruiker Bewerken

Een gebruiker is een account dat bij Semaphore inlogt. Het tabblad **Settings** bevat het profiel
van het account en zijn vlaggen; **Opslaan** schrijft ze weg. Het tabblad **Security** bevat het
wachtwoord, de tweestapsverificatie en de gekoppelde aanmeldaccounts, en het tabblad **Roles** de
globale rollen van de gebruiker: beide werken direct, zonder **Opslaan**. Wat een gebruiker binnen
een project mag doen, wordt ingesteld op de pagina **Team** van dat project, niet hier.

## Naam {#user-name}

De volledige naam van de gebruiker, getoond in de zijbalk en in de lijsten met gebruikers en met
projectleden. Het veld is verplicht.

Voor een gebruiker die **Extern** is, neemt Semaphore de naam bij het inloggen over van de
identiteitsprovider, wanneer de provider er een verstuurt, dus een wijziging die hier wordt
gemaakt, blijft misschien niet behouden na de volgende aanmelding.

## Gebruikersnaam {#user-username}

De login van de gebruiker. Het veld is verplicht. De gebruiker logt in met de gebruikersnaam of met
het e-mailadres, en het wachtwoord.

De gebruikersnaam van een gebruiker die **Extern** is, kan niet worden gewijzigd: de server
weigert hem op te slaan.

## E-mail {#user-email}

Het e-mailadres van de gebruiker. Het veld is verplicht. Semaphore stuurt de taakwaarschuwingen van
de gebruiker naar dit adres, en een authenticator-app toont het als de naam van het account. Het
adres is privé: alleen de gebruiker en wie gebruikers mag beheren, zien het.

Voor een gebruiker die **Extern** is, komt het e-mailadres van de identiteitsprovider: het veld is
alleen-lezen, en Semaphore werkt het adres bij elke aanmelding bij.

## Wachtwoord {#user-password}

Wordt getoond wanneer een nieuwe gebruiker wordt aangemaakt.

Het wachtwoord waarmee de gebruiker inlogt. Het veld is verplicht, behalve voor een gebruiker die
**Extern** is, die geen wachtwoord in Semaphore heeft. Gebruik **Change password** op het tabblad
**Security** om het wachtwoord van een bestaande gebruiker te wijzigen.

## Stuur waarschuwingen {#user-alert}

Stuurt de gebruiker een e-mail wanneer een taak mislukt in een project waarin de gebruiker de taak
mag zien. Het project moet waarschuwingen toestaan, en de beheerder van de server moet
e-mailwaarschuwingen hebben ingeschakeld. Een taaksjabloon kan voorkomen dat zijn mislukte taken
waarschuwingen versturen.

De vlag heeft geen invloed op de waarschuwingen van een project die naar Telegram, Slack en de
andere messengers worden verstuurd.

## Admin gebruiker {#user-admin}

Wordt getoond wanneer u beheerder bent.

Een beheerder mag alles op de server: de gebruikers, de globale rollen en de instellingen van de
server beheren, en elk project met volledige rechten openen.

U kunt de vlag niet van uw eigen account afnemen, en de server houdt minstens één beheerder: hij
weigert de vlag van de laatste af te nemen.

## Pro user {#user-pro}

Wordt getoond in de Pro-editie. Alleen een beheerder kan de vlag wijzigen.

Geeft de gebruiker een plaats in het Pro-abonnement. Het abonnement heeft een vast aantal
Pro-gebruikers: wanneer dat is bereikt, weigert de server de vlag op te slaan. Zonder actief
abonnement heeft de vlag geen effect.

## Extern {#user-external}

Wordt getoond wanneer u beheerder bent. De vlag kan alleen worden ingesteld wanneer de gebruiker
wordt aangemaakt.

Een externe gebruiker logt in via LDAP of OpenID Connect, niet met een wachtwoord van Semaphore:
het wachtwoordveld is uitgeschakeld, en een aanmelding met een wachtwoord wordt geweigerd. De naam
en het e-mailadres worden bij elke aanmelding van de identiteitsprovider overgenomen. Laat het
vakje leeg voor een gebruiker die met een wachtwoord inlogt.

## Change password {#user-changePassword}

Wordt getoond op het tabblad **Security** van een bestaande gebruiker die niet **Extern** is,
wanneer de server inloggen met een wachtwoord toestaat.

Opent een dialoogvenster dat een nieuw wachtwoord instelt. Om uw eigen wachtwoord te wijzigen,
moet u het huidige opgeven; om het wachtwoord van een andere gebruiker opnieuw in te stellen, niet.

## Time-based one-time password {#user-totp}

Wordt getoond op het tabblad **Security** wanneer de beheerder tweestapsverificatie voor de server
heeft ingeschakeld.

Schakelt tweestapsverificatie in voor deze gebruiker. Wanneer u de schakelaar aanzet, maakt
Semaphore direct een geheim aan en toont het als QR-code: scan die met een authenticator-app. Vanaf
dan vraagt een aanmelding na het wachtwoord om de code van de app.

Wanneer de server herstel toestaat, wordt onder de QR-code een herstelcode getoond. Die wordt maar
één keer getoond: kopieer hem en bewaar hem veilig. Het uitzetten van de schakelaar verwijdert het
geheim direct.

## LDAP-provider {#user-ldapProvider}

Wordt getoond op het tabblad **Security** van uw eigen account wanneer de server een LDAP-provider
heeft waaraan uw account nog niet is gekoppeld.

De LDAP-server waaraan u uw account koppelt, wanneer er meerdere zijn.

## LDAP username {#user-ldapUsername}

Wordt getoond met **LDAP-provider**.

Uw gebruikersnaam op de LDAP-server. Door te koppelen kunt u ook via LDAP bij dit account
inloggen.

## LDAP password {#user-ldapPassword}

Wordt getoond met **LDAP-provider**.

Uw wachtwoord op de LDAP-server. Semaphore gebruikt het één keer, om te controleren dat het
LDAP-account van u is, en slaat het niet op.

## Link {#user-ldapLink}

Wordt getoond met **LDAP-provider**.

Koppelt uw account aan het LDAP-account hierboven. De koppeling wordt geweigerd wanneer de
referenties onjuist zijn, wanneer het LDAP-account aan een andere gebruiker is gekoppeld, of
wanneer uw account al aan die provider is gekoppeld. Een gekoppeld account verschijnt onder
**Linked accounts**; sluit de chip ervan om het te ontkoppelen.

## Global role {#global-role-select}

Wordt getoond op het tabblad **Roles** van een bestaande gebruiker, in de Pro-editie, wanneer u
globale rollen mag beheren.

Een globale rol om aan de gebruiker te geven, uit de rollen die de gebruiker nog niet heeft. Een
globale rol verleent rechten over de hele server: gebruikers, globale rollen of de instellingen van
de server beheren, of het auditlogboek lezen. De globale rollen zelf worden aangemaakt op het
tabblad **Globale rollen** van de pagina Gebruikers.

## Assign {#globalRoleAssign-assign}

Wordt getoond met **Global role**.

Geeft de gebruiker direct de geselecteerde rol. De rollen die de gebruiker heeft, staan eronder,
elk met een knop om de rol af te nemen. **Effective global permissions** toont wat de gebruiker
uiteindelijk mag doen, en waar elk recht vandaan komt.
