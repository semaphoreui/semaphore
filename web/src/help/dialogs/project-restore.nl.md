---
form: src/components/RestoreProjectForm.vue
---

# Project Herstellen...

Maakt een nieuw project aan uit een back-upbestand dat is gemaakt met **Project Back-up** op de
pagina Instellingen van een project. Het nieuwe project krijgt de taaksjablonen, inventarissen,
repositories, variabele groepen, sleutels, schema's, integraties en de andere zaken die in de
back-up zijn ingesteld, en u wordt de eigenaar ervan. Alleen een beheerder kan een project
herstellen, tenzij de beheerder van de server elke gebruiker projecten laat aanmaken.

## Projectnaam {#restoreProject-name}

De naam van het nieuwe project. Het veld is verplicht, en de naam vervangt de naam die in het
back-upbestand staat.

De naam moet verschillen van de namen van alle projecten op de server: de server weigert een
project te herstellen onder een naam die al in gebruik is.

## Backup file {#restoreProject-file}

Het JSON-bestand van de back-up. Kies het bestand voordat u op **Restore** drukt.

Een back-up laat de geheimen van de sleutels en de geheimen op het tabblad **Secrets** van de
variabele groepen weg: na het herstel vertelt Semaphore u hoeveel sleutels leeg zijn en hun geheimen
opnieuw moeten krijgen, en de geheimen van de variabele groepen moeten ook opnieuw worden ingevoerd.
Een wachtwoord dat rechtstreeks in een ander veld is getypt, zoals het adres van een repository,
komt terug zoals het in het bestand staat. De leden van het project en de geschiedenis van zijn
taken maken geen deel uit van de back-up. Een back-up die door een nieuwere versie van Semaphore is
geschreven, wordt geweigerd.
