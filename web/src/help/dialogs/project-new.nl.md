---
form: src/components/ProjectForm.vue
---

# Nieuw Project

Een project bevat de taaksjablonen, inventarissen, repositories, variabele groepen en sleutels van
één stuk werk, en de mensen die eraan werken. U wordt de eigenaar van het project dat u aanmaakt.
Alleen een beheerder kan een project aanmaken, tenzij de beheerder van de server dat aan elke
gebruiker toestaat. Elke instelling hieronder kan later worden gewijzigd op de pagina Instellingen
van het project.

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

## Demo {#newProject-demo}

Vult het nieuwe project met voorbeelden om Semaphore mee uit te proberen: de openbare repository
`https://github.com/semaphoreui/semaphore-demo.git`, inventarissen, een sleutel met een
vault-wachtwoord en taaksjablonen voor Ansible, OpenTofu, Terragrunt, Bash en PowerShell,
ingedeeld in de weergaven **Build**, **Deploy** en **Tools**.

Laat de schakelaar uit voor een project zonder voorbeelden: het krijgt alleen de sleutel **Geen**
en de weergave **Alle**. De voorbeelden zijn gewone onderdelen van het project: bewerk of
verwijder ze zoals u wilt.
