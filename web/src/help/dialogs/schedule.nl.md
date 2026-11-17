---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Schema Bewerken

Een schema start een taak van een taaksjabloon zonder dat iemand op **Uitvoeren** drukt: steeds
opnieuw volgens een cron-expressie, of één keer op een vastgesteld moment. De taak krijgt de
parameters die in dit dialoogvenster zijn opgeslagen. Alle tijden zijn in de tijdzone van de
Semaphore-server, getoond naast de tijdvelden: `UTC`, tenzij de beheerder `schedule.timezone` of
`SEMAPHORE_SCHEDULE_TIMEZONE` instelt.

## Name {#scheduleForm-name}

De naam waaronder het schema wordt vermeld op de pagina Schema. Het veld is verplicht. De naam
verandert niets aan wat het schema doet.

## Template {#scheduleForm-templateId}

Het taaksjabloon dat het schema start. Het veld is verplicht. De taak draait alsof ze vanuit het
sjabloon is gestart, met de eigen inventaris, repository, variabele groep en andere instellingen
van het sjabloon, en met de parameters hieronder waar het sjabloon een taak die laat wijzigen.

## Taakparameters {#scheduleForm-taskParams}

Wordt getoond wanneer een sjabloon is geselecteerd.

De parameters van elke taak die het schema start. Welke velden verschijnen, hangt af van het
sjabloon: zijn enquêtevariabelen, en de opties die het een taak laat wijzigen, zoals de branch, de
inventaris, **Limiet**, **Tags**, **Sla tags over** of **CLI args**, en de opties van Ansible of van
Terraform, OpenTofu en Terragrunt.

De server controleert het sjabloon elke keer dat het schema afgaat: een waarde die hier is
opgeslagen, wordt genegeerd zodra het sjabloon een taak niet meer toestaat die te wijzigen.

## Bericht {#taskParams-message}

Optioneel. Een notitie die bij elke taak die het schema start wordt opgeslagen en in de
geschiedenis van taken wordt getoond, bijvoorbeeld `Nightly deploy`. De notitie verandert niets aan
hoe de taak draait.

## Enquête Variabelen {#taskParams-survey}

Wordt getoond wanneer het sjabloon enquêtevariabelen heeft.

De waarden van de enquêtevariabelen van het sjabloon voor de geplande taken. Ze bereiken de taak
op dezelfde manier als wanneer iemand de enquête met de hand invult. Een verplichte variabele moet
een waarde hebben voordat het schema kan worden opgeslagen.

De standaardwaarden van het sjabloon worden overgenomen wanneer het dialoogvenster opent en worden
met het schema opgeslagen. Een standaardwaarde die later in het sjabloon wordt gewijzigd, bereikt
een eerder opgeslagen schema niet.

De waarde van een variabele van het type secret wordt niet met het schema opgeslagen: de geplande
taak draait zonder die waarde.

## Tak {#taskParams-gitBranch}

Wordt getoond wanneer het sjabloon een taak toestaat de branch te overschrijven.

De branch of de tag van de repository die de geplande taak uitcheckt. Het veld begint met de
branch van het sjabloon. De server gebruikt de waarde alleen zolang het sjabloon het
overschrijven toestaat.

## Inventaris {#taskParams-inventoryId}

Wordt getoond wanneer het sjabloon een taak toestaat de inventaris te kiezen.

De inventaris waarop de geplande taak wordt uitgevoerd, in plaats van de inventaris van het
sjabloon. Voor een Terraform- of OpenTofu-sjabloon heet het veld **Workspace**, maar de server past
de keuze alleen toe op Ansible-sjablonen en behoudt voor de andere de workspace van het sjabloon.

## Debug {#taskParams-debug}

Wordt getoond voor een Ansible-sjabloon dat debuggen toestaat.

Voert `ansible-playbook` uit met meer uitvoer: `-v` zo vaak herhaald als **Debug level** aangeeft.

## Debug level {#taskParams-debugLevel}

Wordt getoond voor een Ansible-sjabloon dat debuggen toestaat; actief wanneer **Debug** is
aangevinkt.

Hoe uitgebreid de uitvoer is, van `1` (`-v`) tot `6` (`-vvvvvv`). De standaard is `4`.

## Droge Run {#taskParams-dryRun}

Wordt getoond voor een Ansible-sjabloon, tenzij het sjabloon het verbergt.

Voert het playbook uit met `--check`: Ansible meldt wat het zou wijzigen en wijzigt niets.

## Verschil {#taskParams-diff}

Wordt getoond voor een Ansible-sjabloon, tenzij het sjabloon het verbergt.

Voert het playbook uit met `--diff`: Ansible toont de wijzigingen die het aanbrengt in bestanden en
templates.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Wordt getoond voor een Ansible-sjabloon dat een taak toestaat dit te wijzigen.

Slaat de installatie van de rollen en collecties uit de `requirements.yml`-bestanden over voordat
het playbook draait. Vink het aan wanneer de vereisten al zijn geïnstalleerd en de taak Ansible
Galaxy niet mag benaderen.

## Plan {#taskParams-plan}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voert alleen `plan` uit. De taak eindigt na het plan en past niets toe.

## Destroy {#taskParams-destroy}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voegt `-destroy` toe aan het plan en de apply: de taak vernietigt de resources van de workspace in
plaats van ze aan te maken of bij te werken.

## Auto Approve {#taskParams-autoApprove}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Past het plan toe zonder op een bevestiging te wachten. Dit werkt alleen wanneer het sjabloon
automatisch goedkeuren toestaat; een sjabloon dat is ingesteld om elke taak goed te keuren, past
het plan in beide gevallen toe.

Zonder automatisch goedkeuren stopt een geplande taak na het plan en wacht ze totdat iemand haar op
de taakpagina bevestigt of afwijst.

## Upgrade {#taskParams-upgrade}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voert `init` uit met `-upgrade`: de providers en modules worden bijgewerkt naar de nieuwste versies
die de configuratie toestaat.

## Reconfigure {#taskParams-reconfigure}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voert `init` uit met `-reconfigure` in plaats van `-migrate-state`: de backend wordt opnieuw
ingesteld en de bestaande state wordt niet gemigreerd.

## Run at {#scheduleForm-runAt}

Wordt getoond voor een schema dat één keer draait.

De datum en tijd waarop de taak start, in de tijdzone van de server die aan het eind van het veld
wordt getoond. Een nieuw schema begint met het volgende hele uur. Het moment moet in de toekomst
liggen: de server weigert een tijdstip dat al voorbij is.

Nadat de taak is gestart, wordt het schema uitgeschakeld, of verwijderd wanneer **Delete after
run** is aangevinkt. Wanneer Semaphore op dat moment niet draait, wordt de taak niet later
gestart: bij de volgende start wordt het schema zonder run uitgeschakeld, of verwijderd.

## Delete after run {#scheduleForm-deleteAfterRun}

Wordt getoond voor een schema dat één keer draait.

Verwijdert het schema nadat zijn taak is gestart. Laat het uitgevinkt om het schema te bewaren,
uitgeschakeld, zodat u een nieuwe tijd kunt instellen en het weer kunt inschakelen.

## Show cron format {#scheduleForm-rawCron}

Wordt getoond voor een cron-schema.

Schakelt tussen het typen van de cron-expressie (aan) en het opbouwen ervan met **Timing** en de
tijdvakjes (uit). Beide bewerken dezelfde expressie. De browser onthoudt de keuze.

De schakelaar staat vast op aan wanneer de expressie niet als vakjes kan worden getoond: wanneer
ze ongeldig is, of een vorm gebruikt zoals `@daily` of `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Wordt getoond voor een cron-schema wanneer **Show cron format** aan staat.

De cron-expressie in vijf velden: minuut, uur, dag van de maand, maand, dag van de week (`0` is
zondag). Het veld is verplicht, en de server controleert de expressie terwijl u typt. Een nieuw
schema begint met `* * * * *`, elke minuut.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

De server accepteert ook `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` en
`@every 1h30m`, en een voorvoegsel zoals `CRON_TZ=Europe/Berlin` voor een andere tijdzone. De
voorvertoning eronder kan de volgende run daarvan niet tonen.

## Timing {#scheduleForm-timing}

Wordt getoond voor een cron-schema wanneer **Show cron format** uit staat.

Hoe vaak de taak draait. De vakjes eronder veranderen mee:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** en **Minutes**.
- **Weekly**: **Weekdays**, **Hours** en **Minutes**.
- **Monthly**: **Days**, **Hours** en **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** en **Minutes**.

## Months {#scheduleForm-months}

Wordt getoond wanneer **Timing** op **Yearly** staat.

De maanden waarin de taak draait. Vink er meerdere aan om in elk ervan te draaien. Wanneer niets is
aangevinkt, draait de taak in elke maand.

## Weekdays {#scheduleForm-weekdays}

Wordt getoond wanneer **Timing** op **Weekly** staat.

De dagen van de week waarop de taak draait. Wanneer niets is aangevinkt, draait de taak op elke
dag.

## Days {#scheduleForm-days}

Wordt getoond wanneer **Timing** op **Monthly** of **Yearly** staat.

De dagen van de maand waarop de taak draait. Een maand zonder die dag wordt overgeslagen: een taak
die op de 31e is ingesteld, draait niet in april. Wanneer niets is aangevinkt, draait de taak op
elke dag.

## Hours {#scheduleForm-hours}

Wordt getoond wanneer **Timing** op **Daily**, **Weekly**, **Monthly** of **Yearly** staat.

De uren waarop de taak draait, in de tijdzone die rechts wordt getoond. Wanneer niets is
aangevinkt, draait de taak in elk uur.

## Minutes {#scheduleForm-minutes}

Wordt getoond voor een cron-schema wanneer **Show cron format** uit staat.

De minuten van het uur waarop de taak draait, in stappen van vijf. Wanneer niets is aangevinkt,
draait de taak elke minuut: vink er minstens één aan, bijvoorbeeld `:00`, voor een taak die één
keer per uur of één keer per dag moet draaien.

## Next run time {#scheduleForm-nextRun}

Wanneer de volgende taak zou starten, volgens het schema zoals het in het formulier staat: in de
tijdzone van de server en in de lokale tijd van uw browser. Een streepje betekent dat de expressie
ongeldig is, of een expressie is die de voorvertoning niet kan lezen, zoals `@daily`.

## Ingeschakeld {#scheduleForm-active}

Of het schema taken start. Vink het uit om een schema te bewaren zonder het uit te voeren. De
schakelaar in de lijst met schema's doet hetzelfde. Een nieuw schema is ingeschakeld.
