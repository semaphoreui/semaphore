---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nieuwe Taak

Start een taak van het sjabloon: één run van zijn playbook, Terraform-code of script. Het sjabloon
bepaalt welke velden dit dialoogvenster toont, en elke sectie hieronder zegt wanneer het veld er
is. De knop onderaan is **Uitvoeren**, **Bouwen** of **Deployen**, naar het type van het sjabloon.

Wanneer het sjabloon een enquête heeft, zijn de vragen ervan velden na **Bericht**. Een veld dat
met `*` is gemarkeerd, is verplicht, een veld van het type Integer accepteert alleen cijfers, en
de standaardantwoorden zijn ingevuld. Semaphore geeft de antwoorden als variabelen aan de taak
door: extra vars voor Ansible, `-var` voor Terraform, argumenten `name=value` voor een script, of
omgevingsvariabelen voor een vraag die zo is ingesteld. Een antwoord vervangt een variabele met
dezelfde naam in de variabele groepen van het sjabloon. Een geheim antwoord wordt tijdens het
typen verborgen en wordt niet bij de taak bewaard.

## Commit {#newTask-commit}

Wordt getoond wanneer u een eerdere taak opnieuw uitvoert, en die taak op een bekende commit
draaide.

De taak checkt deze commit van de repository uit in plaats van de nieuwste commit van de branch.
Sluit het vak om de nieuwste commit uit te voeren.

De commit wordt alleen gebruikt wanneer het sjabloon het overschrijven van de branch toestaat, en
nooit voor een repository die als monorepo is gemarkeerd. Anders voert de taak de nieuwste commit
uit.

## Bouw Versie {#newTask-buildTaskId}

Wordt getoond wanneer het type van het sjabloon **Deployen** is.

De build die moet worden gedeployed: een van de laatste 20 geslaagde taken van het bouwsjabloon,
vermeld met versie en bericht. De nieuwste is geselecteerd. Het veld is verplicht.

De versie van de build bereikt de taak: als `semaphore_vars.task_details.incoming_version` in
Ansible, als de omgevingsvariabele `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` in een script. De
lijst is leeg wanneer u het bouwsjabloon niet mag zien.

## Bericht {#newTask-message}

Optioneel. Een notitie over waarom de taak draait, bijvoorbeeld `Hotfix for the login page`. De
notitie wordt bij de taak in de geschiedenis van taken getoond en verandert niets aan de run.

Het playbook kan haar lezen als `semaphore_vars.task_details.message`, en een script als de
omgevingsvariabele `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Tak {#newTask-gitBranch}

Wordt getoond wanneer het sjabloon het overschrijven van de branch toestaat, tenzij de repository
een monorepo is.

De branch of tag van de repository die de taak uitcheckt. Het veld begint met de branch van het
sjabloon. Wanneer het leeg is, gebruikt de taak de branch van het sjabloon, of de branch van de
repository wanneer het sjabloon er geen heeft. Een naam die geen geldige Git-branchnaam is, wordt
geweigerd.

## Inventaris {#newTask-inventoryId}

Wordt getoond wanneer het sjabloon toestaat de inventaris in een taak te kiezen en u het project mag
zien. Voor een Terraform-, OpenTofu- of Terragrunt-sjabloon heet het veld **Workspace**, maar de
server gebruikt de keuze alleen voor Ansible-sjablonen: een Terraform-, OpenTofu- of Terragrunt-taak
draait in de workspace van het sjabloon, wat u hier ook kiest.

De inventaris waarop de taak wordt uitgevoerd. Het veld begint met de inventaris van het sjabloon
en vermeldt de inventarissen van het project die voor de app van het sjabloon zijn gemaakt.
Wanneer de gekozen inventaris niet meer bestaat op het moment dat de taak start, draait de taak op
de inventaris van het sjabloon.

## Limiet {#taskParams-limit}

Wordt getoond voor een Ansible-sjabloon dat het overschrijven van de limiet toestaat.

De hosts of groepen van de inventaris waarop het playbook draait, één patroon per chip,
bijvoorbeeld `web` of `10.0.0.11`. Semaphore geeft ze door als `--limit`, gescheiden door komma's.
De lijst begint met de limiet van het sjabloon. Verwijder alle chips om op alle hosts van de
inventaris te draaien.

## Tags {#taskParams-tags}

Wordt getoond voor een Ansible-sjabloon dat het overschrijven van de tags toestaat.

Alleen de taken en rollen van het playbook met een van deze tags draaien: Semaphore geeft ze door
als `--tags`, één tag per chip. De lijst begint met de tags van het sjabloon. Verwijder alle chips
om het hele playbook uit te voeren.

## Sla tags over {#taskParams-skipTags}

Wordt getoond voor een Ansible-sjabloon dat het overschrijven van de overgeslagen tags toestaat.

De taken en rollen van het playbook met een van deze tags draaien niet: Semaphore geeft ze door als
`--skip-tags`, één tag per chip. De lijst begint met de overgeslagen tags van het sjabloon.

## Debug {#taskParams-debug}

Wordt getoond voor een Ansible-sjabloon dat debuggen toestaat.

Voert `ansible-playbook` uit met een verbosity-vlag, zodat het log van de taak meer details toont.
Het label toont de vlag: `-v` zo vaak als **Debug level** aangeeft.

## Debug level {#taskParams-debugLevel}

Wordt getoond met **Debug**, en is actief wanneer **Debug** is aangevinkt.

Hoe uitgebreid het log is, van 1 (`-v`) tot 6 (`-vvvvvv`). De standaard is 4 (`-vvvv`), wat ook de
details van de verbindingen met de hosts toont.

## Droge Run {#taskParams-dryRun}

Wordt getoond voor een Ansible-sjabloon, tenzij het sjabloon het verbergt.

Voert het playbook uit met `--check`: Ansible meldt wat het op de hosts zou wijzigen en wijzigt
niets.

## Verschil {#taskParams-diff}

Wordt getoond voor een Ansible-sjabloon, tenzij het sjabloon het verbergt.

Voert het playbook uit met `--diff`: het log toont de wijzigingen die in bestanden en templates op
de hosts worden aangebracht. Vink ook **Droge Run** aan om de wijzigingen te zien zonder ze aan te
brengen.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Wordt getoond voor een Ansible-sjabloon dat het overschrijven van de Galaxy-installatie toestaat.

Voordat het playbook draait, installeert Semaphore de rollen en collecties die in de
`requirements.yml`-bestanden van de repository staan, wanneer een bestand sinds de laatste
installatie is gewijzigd. Vink het vakje aan om die stap voor deze taak over te slaan.

Het vakje begint uitgevinkt, wat het sjabloon ook instelt, en een uitgevinkt vakje installeert de
vereisten.

## Plan {#taskParams-plan}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voert `plan` uit en beëindigt de taak daar: er wordt niets toegepast. Gebruik dit om de
wijzigingen eerst te bekijken. Zonder deze optie gaat de taak door met het toepassen van het plan
wanneer het wijzigingen bevat.

## Destroy {#taskParams-destroy}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voegt `-destroy` toe aan `plan` en `apply`: de taak plant, en past daarna toe, het verwijderen van
elke resource in de state van de workspace. Vink ook **Plan** aan om te zien wat er zou worden
vernietigd zonder het te vernietigen.

## Auto Approve {#taskParams-autoApprove}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Past het plan toe zodra het is gemaakt. Zonder deze optie stopt de taak na het plan en wacht ze
totdat iemand haar op de pagina van de taak bevestigt of afwijst.

Het vakje heeft alleen effect wanneer het sjabloon automatisch goedkeuren toestaat. Een sjabloon
dat is ingesteld om altijd goed te keuren, past toe zonder te vragen, wat het vakje ook zegt.

## Upgrade {#taskParams-upgrade}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voert `init` uit met `-upgrade`: de providers en modules worden bijgewerkt naar de nieuwste versies
die hun versiebeperkingen toestaan.

## Reconfigure {#taskParams-reconfigure}

Wordt getoond voor een Terraform-, OpenTofu- of Terragrunt-sjabloon.

Voert `init` uit met `-reconfigure`: de backend wordt opnieuw ingesteld, en de opgeslagen
configuratie van de backend wordt genegeerd. Zonder deze optie voert Semaphore
`init -migrate-state` uit, wat de bestaande state naar een gewijzigde backend verplaatst.

## CLI args {#newTask-args}

Wordt getoond wanneer het sjabloon het overschrijven van de CLI-argumenten toestaat.

Extra argumenten voor de opdrachtregel van de taak, één per chip: `--forks=10` is één chip,
`--forks` en `10` zijn er twee. Klik op een chip om hem te bewerken, sleep hem om de volgorde te
wijzigen, en gebruik de laatste chip om er een toe te voegen.

De lijst begint met de argumenten van het sjabloon. Semaphore geeft altijd de argumenten van het
sjabloon door en voegt deze lijst erachter toe, dus zodra u de lijst wijzigt, worden de argumenten
van het sjabloon die erin zijn blijven staan twee keer doorgegeven. Verwijder ze en houd alleen de
argumenten die u toevoegt.
