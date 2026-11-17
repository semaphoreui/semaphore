---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Projectinstellingen

De instellingen van het project als geheel: de naam, hoeveel van de taken tegelijk mogen draaien
en of het project waarschuwingen verstuurt. De Gevarenzone onder het formulier bevat de acties
die het hele project raken: de back-up, de cache en het verwijderen ervan. Wijzigingen in het
formulier worden van kracht na **Opslaan**.

## Projectnaam {#newProject-name}

De naam waaronder het project wordt weergegeven in het projectmenu bovenaan de zijbalk. Het veld
is verplicht.

## Maximaal aantal parallelle taken {#newProject-maxParallelTasks}

Hoeveel taken van dit project tegelijkertijd mogen draaien. Een taak die boven de limiet wordt
gestart, wacht in de wachtrij totdat een van de lopende taken klaar is.

Laat het veld leeg of stel het in op `0` voor geen eigen limiet. De limiet van de hele server,
ingesteld door de beheerder, geldt in beide gevallen.

## Telegram Chat ID {#newProject-tg}

Optioneel. De Telegram-chat die de waarschuwingen van dit project ontvangt. Wanneer het veld leeg
is, gaan de waarschuwingen naar de chat die voor de hele server is ingesteld.

Het veld heeft alleen effect wanneer de beheerder Telegram-waarschuwingen voor de server heeft
ingeschakeld en waarschuwingen voor dit project zijn toegestaan.

## Sta waarschuwingen voor dit project toe {#newProject-alert}

Schakelt waarschuwingen in voor de taken van dit project. Waar de waarschuwingen naartoe gaan,
stelt de beheerder in voor de hele server: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams,
DingTalk of Gotify.

- Een e-mail wordt verstuurd wanneer een taak mislukt, naar de leden van het project die de taak
  mogen zien en waarschuwingen in hun account hebben ingeschakeld.
- Een bericht naar de messengers wordt verstuurd wanneer een taak slaagt, mislukt of op een
  bevestiging wacht.

Een taaksjabloon kan voorkomen dat zijn geslaagde of zijn mislukte taken waarschuwingen versturen.

## Test Alerts {#settings-testAlerts}

Verstuurt een testwaarschuwing via elk kanaal dat de server heeft ingeschakeld, zodat u kunt
controleren of de waarschuwingen van dit project aankomen.

De knop werkt alleen wanneer **Sta waarschuwingen voor dit project toe** is aangevinkt en
opgeslagen. Vink het aan en klik eerst op **Opslaan**.

## Project Back-up {#settings-exportProject}

Downloadt een JSON-bestand met de configuratie van het project: taaksjablonen, inventarissen,
repositories, variabele groepen, sleutels, schema's, integraties en de andere zaken die erin
zijn ingesteld. Het bestand kan weer in een project worden omgezet met **Project Herstellen** in
het projectmenu bovenaan de zijbalk.

Het bestand laat de geheimen van de sleutels en de geheimen op het tabblad **Secrets** van de
variabele groepen weg: die moeten na een herstel opnieuw worden ingevoerd. Al het andere wordt
gekopieerd zoals het is, ook een wachtwoord dat rechtstreeks in een ander veld is getypt: het adres
van een repository met een wachtwoord erin, de tekst van een statische inventaris, de waarden van
gewone variabelen en omgevingsvariabelen. Controleer het bestand op zulke waarden voordat u het
deelt. De geschiedenis van taken maakt geen deel uit van de back-up.

## Cache wissen {#settings-clearCache}

Verwijdert de bestanden die Semaphore voor dit project op de server bewaart, zoals de kopieën van
de repositories. Ze worden opnieuw gedownload wanneer de volgende taak start.

Gebruik dit wanneer een taak mislukt door een verouderde of beschadigde kopie van een repository.
De actie kan niet ongedaan worden gemaakt en raakt de instellingen van het project niet.

## Project Verwijderen {#settings-deleteProject}

Verwijdert het project met alles wat erin staat: taaksjablonen, de geschiedenis van taken,
inventarissen, repositories, variabele groepen, sleutels en schema's. Semaphore vraagt eerst om
een bevestiging.

Een verwijderd project kan niet worden teruggehaald. Download een back-up voordat u een project
verwijdert dat u misschien nog nodig hebt.
