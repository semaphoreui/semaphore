---
form: src/components/RunnerForm.vue
---

# Runner Bewerken

Een runner is een afzonderlijk `semaphore runner`-proces dat taken op zijn eigen machine uitvoert
in plaats van op de Semaphore-server. Een taak gaat naar een runner wanneer de server is ingesteld
om taken op runners uit te voeren, of wanneer het taaksjabloon of de inventaris ervan een runner
tag heeft. De server kiest dan een runner die is ingeschakeld, geregistreerd, online en niet aan
zijn limiet van parallelle taken zit: eerst runners van het project, dan globale runners, binnen
elke groep in willekeurige volgorde. Wanneer elke passende runner offline of bezet is, wacht de
taak in de wachtrij; wanneer er helemaal geen runner past, mislukt de taak.

## Naam {#runner-name}

Verplicht. De naam waaronder de runner wordt vermeld op de pagina Runners, bijvoorbeeld de host
waarop hij draait. De naam verandert niets aan welke taken de runner krijgt.

## Tags {#runner-tags}

Beschikbaar in Semaphore Pro; in andere edities is het veld vergrendeld.

De tags die de taken van deze runner kiezen. Een taak met een runner tag gaat alleen naar de
runners die die tag hebben; de tag van het taaksjabloon gaat voor de tag van de inventaris. Typ
een tag en druk op Enter, of kies er een die door andere runners wordt gebruikt. Lege en dubbele
tags worden bij het opslaan weggelaten.

Een projectrunner heeft minstens één tag nodig, tenzij **Is default** is aangevinkt.

## Is default {#runner-isDefault}

Een taak zonder runner tag gaat alleen naar de runners die als standaard zijn gemarkeerd. Vink het
vakje uit voor een runner die alleen taken met een van zijn tags mag aannemen.

Het vakje is aangevinkt voor een nieuwe globale runner en uitgevinkt voor een nieuwe projectrunner.

## Register {#runner-registered}

Wordt getoond wanneer u een runner aanmaakt.

- Aangevinkt (de standaard): de server maakt de runner aan met zijn token. Na **Aanmaken** toont
  het dialoogvenster het token en de opdrachten die de runner ermee starten.
- Uitgevinkt: de runner wordt zonder token aangemaakt. Klik later op **Regenerate registration
  token** (het pictogram met twee pijlen) in de rij van de runner om een eenmalig
  registratietoken te krijgen, één uur geldig, en registreer de runner met
  `semaphore runner register`.

Een niet-geregistreerde runner krijgt geen taken.

## Webhook {#runner-webhook}

Optioneel. Een URL die de server aanroept wanneer hij een taak aan deze runner geeft, bijvoorbeeld
om de runner op aanvraag te starten. De server verstuurt een `POST`-verzoek met een JSON-body:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

De URL moet binnen 30 seconden antwoorden met de status 200 of 204; anders mislukt de taak. Een
runner met een webhook wordt altijd als online beschouwd, ook wanneer hij de server de laatste tijd
niet heeft gepolld. Laat het veld leeg voor een runner die altijd draait.

## Maximaal aantal parallelle taken {#runner-maxParallelTasks}

Het aantal taken dat de runner tegelijk mag uitvoeren. Wanneer de limiet is bereikt, gaat de
volgende taak naar een andere passende runner of wacht ze in de wachtrij. Laat het veld leeg of
stel `0` in voor geen limiet.

## Ingeschakeld {#runner-active}

Alleen een ingeschakelde runner krijgt nieuwe taken. Vink het vakje uit om de runner buiten dienst
te stellen zonder hem te verwijderen; de schakelaar in de lijst met runners doet hetzelfde.

Voor een nieuwe runner is het vakje uitgevinkt: vink het aan, of schakel de runner later in de
lijst in.
