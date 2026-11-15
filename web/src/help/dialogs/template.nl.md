---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Sjabloon Bewerken 'Ansible Playbook'

Een taaksjabloon zegt wat Semaphore uitvoert en hoe: de repository en het bestand daarin, de
inventaris en de variabelen, en welke opties een gebruiker mag wijzigen wanneer een taak start.
Elke taak wordt vanuit een sjabloon gestart. De velden van het dialoogvenster hangen af van de app
van het sjabloon (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell en andere):
een veld dat alleen sommige apps hebben, vermeldt dat. Wijzigingen gelden voor taken die na
**Opslaan** worden gestart.

## Sjabloontype {#templateForm-typeTabs}

De tabbladen bovenaan bepalen wat voor sjabloon het is:

- **Taak**: voert de code uit. De meeste sjablonen zijn van dit soort.
- **Bouwen**: maakt een artifact. Elke bouwtaak krijgt een versie, geteld vanaf **Start Versie**.
- **Deployen**: deployt een artifact dat door een build is gemaakt. Wanneer u het start, kiest u
  de bouwtaak waarvan het de versie deployt.

Een bouw- of deploytaak krijgt de variabelen `SEMAPHORE_TASK_TYPE`, `SEMAPHORE_TASK_TARGET_VERSION`
(bouwen) en `SEMAPHORE_TASK_INCOMING_VERSION` (deployen) in zijn omgeving. Ansible krijgt dezelfde
waarden in de extra variabele `semaphore_vars.task_details`.

## Start Versie {#templateForm-startVersion}

Wordt getoond wanneer het type **Bouwen** is. Verplicht.

De versie van de eerste bouwtaak, bijvoorbeeld `1.0.0` of `v0.1`. Elke volgende build verhoogt het
laatste getal van de vorige versie met één en behoudt de tekst eromheen en de voorloopnullen:
`1.0.0`, `1.0.1`, `1.0.2`. Verhoog het getal hier boven de laatste build om vooruit te springen,
bijvoorbeeld naar `2.0.0`; een versie zonder getal erin verandert nooit.

## Bouw Sjabloon {#templateForm-buildTemplateId}

Wordt getoond wanneer het type **Deployen** is. Verplicht.

Het sjabloon waarvan de taken de artifacts maken die dit sjabloon deployt. De lijst bevat de
bouwsjablonen van het project en, daaronder, de deploysjablonen: een deploy kan op een andere
deploy volgen, en krijgt dan de versie die die deploy van zijn build kreeg.

## Autorun {#templateForm-autorun}

Wordt getoond wanneer het type **Deployen** is.

Start vanzelf een taak van dit sjabloon elke keer dat een taak van het **Bouw Sjabloon** slaagt,
met de versie die die taak heeft gemaakt. Een mislukte build start niets.

## Naam {#templateForm-name}

De naam waaronder het sjabloon wordt vermeld op de pagina Taak Sjablonen. Het veld is verplicht.

## Repository {#templateForm-repositoryId}

De repository van het project die de uit te voeren code bevat. Het veld is verplicht. Vóór elke
taak werkt Semaphore zijn kopie van de repository bij en voert de code daarvandaan uit.

Het label rechts van het veld toont de branch die de taken uitchecken. Klik erop om het veld
**Tak** te openen.

## Tak {#templateForm-gitBranch}

Wordt getoond na een klik op het branchlabel rechts van **Repository**.

De branch of tag die de taken van dit sjabloon uitchecken, in plaats van de branch die op de
repository is ingesteld. Laat het veld leeg om de branch van de repository te gebruiken. Wanneer
Semaphore de branches van de repository kan lezen, vermeldt het veld ze.

Het veld is uitgeschakeld wanneer de repository een monorepo is: alle sjablonen delen er één
checkout van, op de branch van de repository.

## Pad naar playbook bestand {#templateForm-playbook}

Wat er moet worden uitgevoerd, als pad vanaf de hoofdmap van de repository, bijvoorbeeld
`deploy/site.yml`. Het pad kan niet absoluut zijn en kan niet buiten de repository leiden. De lijst
biedt de bestanden van de repository aan; u kunt ook een pad typen dat niet in de lijst staat. De
knop rechts van het veld leest de lijst opnieuw in.

Het label en de betekenis veranderen met de app:

- **Ansible**: het playbook. Het is verplicht; de lijst biedt `.yml`- en `.yaml`-bestanden aan.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, de map met de code. Laat het
  leeg om de code in de hoofdmap van de repository uit te voeren. Typ `/` om te zien wat een map
  bevat.
- **Bash**, **Python**, **PowerShell** en andere apps: **Script Filename**, het script dat moet
  worden uitgevoerd. Het is verplicht. Het script draait in de hoofdmap van de repository.

## Different working directory {#templateForm-workingDirectoryToggle}

Wordt getoond voor Ansible-sjablonen.

Voert `ansible-playbook` uit in een andere map dan de hoofdmap van de repository. Vink het aan om
het veld **Working directory** te tonen. Vink het uit om weer in de hoofdmap te draaien.

## Working directory {#templateForm-workingDirectory}

Wordt getoond wanneer **Different working directory** is aangevinkt. Verplicht.

De map waarin `ansible-playbook` en `ansible-galaxy` draaien, als pad vanaf de hoofdmap van de
repository, bijvoorbeeld `ansible`. Gebruik dit wanneer de `ansible.cfg` van het project in een
submap staat, omdat Ansible het bestand uit de huidige map leest. Het pad van het playbook blijft
relatief ten opzichte van de hoofdmap van de repository. Het pad kan niet absoluut zijn en kan niet
buiten de repository leiden.

## Inventaris {#templateForm-inventoryId}

Wordt getoond voor Ansible-, Terraform-, OpenTofu- en Terragrunt-sjablonen.

- **Ansible**: de inventaris waarop het playbook wordt uitgevoerd, met de referenties voor de
  hosts. Het veld is verplicht.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, de workspace waarin de code draait.
  Semaphore selecteert hem vóór het plan en maakt hem aan wanneer hij niet bestaat. Een nieuw
  sjabloon dat zonder workspace wordt opgeslagen, krijgt er een met de naam `default`.

De lijst bevat de inventarissen van het project die voor de app van het sjabloon zijn gemaakt.

## Variabele Groepen {#templateForm-environmentIds}

Optioneel. De variabele groepen waarvan elke taak van dit sjabloon de variabelen,
omgevingsvariabelen en geheimen krijgt. U kunt er meerdere kiezen: ze worden samengevoegd in de
volgorde van het veld, en een latere groep gaat voor een eerdere wanneer beide dezelfde naam
instellen.

Hoe de variabelen de code bereiken, hangt af van de app: extra variabelen voor Ansible, `-var`
voor Terraform en OpenTofu, argumenten `name=value` voor een script. Omgevingsvariabelen worden
voor elke app ingesteld.

## Weergave {#templateForm-viewId}

Optioneel. De weergave waarin het sjabloon wordt vermeld: weergaven zijn de tabbladen van de pagina
Taak Sjablonen, die de sjablonen van een project groeperen. Laat het veld leeg om het sjabloon in
geen enkele weergave te plaatsen.

## Runner tag {#templateForm-runnerTag}

Wordt getoond wanneer er projectrunners beschikbaar zijn.

Taken van dit sjabloon draaien alleen op een runner met deze tag, ook wanneer de server ze zelf
zou kunnen uitvoeren. De tag gaat voor de runner tag van de inventaris. Laat het veld leeg om de
tag van de inventaris te gebruiken, of geen.

## Executor image {#templateForm-executorImage}

Het container-image dat de runner start voor de taken van dit sjabloon, bijvoorbeeld
`semaphoreui/job:latest`. Laat het veld leeg om het image uit de configuratie van de runner te
gebruiken.

Alleen de executors Docker en Kubernetes voeren taken in een container uit, dus het veld is alleen
ingeschakeld wanneer een van beide beschikbaar is. Een taak die zonder container draait, negeert
het.

## Enquête Variabelen {#surveyVars}

De waarden waarnaar een gebruiker wordt gevraagd wanneer een taak van dit sjabloon start. Klik op
**+** om een variabele toe te voegen, op een chip om er een te wijzigen, op het kruisje van een
chip om hem te verwijderen; sleep de chips om de volgorde te bepalen waarin de vragen worden
gesteld.

De antwoorden bereiken de code zoals de variabelen van een variabele groep, of als
omgevingsvariabelen, zoals elke variabele instelt. Ze gaan voor een variabele groep die dezelfde
naam instelt.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Laat taken van dit sjabloon tegelijkertijd draaien. Zonder deze optie wacht een taak van het
sjabloon in de wachtrij totdat de lopende taak klaar is. De limieten van parallelle taken van het
project en van de server gelden in beide gevallen.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Wordt getoond wanneer de beheerder het uitgeven van JWT's voor de server heeft ingeschakeld.

Geeft elke taak van dit sjabloon een door Semaphore ondertekend JSON Web Token, in de
omgevingsvariabele `SEMAPHORE_JWT`. De code kan het token aan een andere dienst tonen om te
bewijzen welke taak het is: het token noemt de taak, het project, het sjabloon en de gebruiker die
de taak startte. De dienst controleert de handtekening met de publieke sleutels op
`/.well-known/jwks.json` van de Semaphore-server.

## JWT audience {#templateForm-jwtAudience}

Wordt getoond wanneer **Issue JWT to task runner** is aangevinkt. Optioneel.

De diensten waarvoor het token is bedoeld, geschreven in de claim `aud`, bijvoorbeeld
`https://vault.example.com`. Druk na elke waarde op Enter. Maximaal 32 waarden; laat het veld leeg
voor een token zonder audience.

## JWT TTL {#templateForm-jwtTtl}

Wordt getoond wanneer **Issue JWT to task runner** is aangevinkt. Optioneel.

Hoe lang het token geldig blijft vanaf de start van de taak, als duur zoals `30s`, `15m` of `1h`.
Laat het veld leeg om de standaard van de server te gebruiken, één uur tenzij de beheerder een
andere heeft ingesteld. De waarde kan het maximum van de server niet overschrijden, 24 uur tenzij
anders ingesteld.

## Ik wil een taak via cron alleen uitvoeren voor nieuwe commits van een bepaalde repository {#templateForm-cronVisible}

Start een taak van dit sjabloon wanneer er een nieuwe commit in een repository verschijnt.
Semaphore vraagt de repository met het gekozen interval om zijn laatste commit en start alleen een
taak wanneer de commit sinds de vorige controle is veranderd.

Het vakje aanvinken en opslaan maakt het schema van de controle aan; het uitvinken en opslaan
verwijdert dat schema. De andere schema's van het sjabloon blijven behouden.

## Repository {#templateForm-cronRepositoryId}

Wordt getoond wanneer **Ik wil een taak via cron alleen uitvoeren voor nieuwe commits van een
bepaalde repository** is aangevinkt. Verplicht.

De repository die op nieuwe commits wordt gecontroleerd. Semaphore volgt de branch die op die
repository is ingesteld, niet de **Tak** van het sjabloon.

## controleerInterval {#templateForm-cronFormat}

Wordt getoond wanneer **Ik wil een taak via cron alleen uitvoeren voor nieuwe commits van een
bepaalde repository** is aangevinkt.

Hoe vaak Semaphore de repository op een nieuwe commit controleert: van elke minuut tot één keer per
dag. De standaard is elke minuut.

## Suppress success alerts {#templateForm-suppressSuccessAlerts}

Voorkomt dat de taken van dit sjabloon die slagen waarschuwingen versturen. De waarschuwingen van
mislukte taken worden nog steeds verstuurd. Dit heeft alleen effect wanneer waarschuwingen voor
het project zijn ingeschakeld.

## Suppress error alerts {#templateForm-suppressErrorAlerts}

Voorkomt dat de taken van dit sjabloon die mislukken waarschuwingen versturen. De waarschuwingen
van geslaagde taken worden nog steeds verstuurd. Dit heeft alleen effect wanneer waarschuwingen
voor het project zijn ingeschakeld.

## CLI args {#templateForm-args}

Argumenten die aan de opdracht van elke taak worden toegevoegd. Elke chip is één argument, dat
wordt doorgegeven zoals het is, zonder shell: zet `--forks=10` in één chip, of `--forks` en `10`
in twee. Sleep de chips om hun volgorde te wijzigen. De knop rechts van de lijst opent de
argumenten als JSON-array:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: toegevoegd aan `ansible-playbook`, vóór het playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: toegevoegd aan `plan` en `apply`, niet aan `init`.
- Een script: toegevoegd na de naam van het script.

## CLI-argumenten {#templateForm-allowOverrideArgsInTask}

Onder **Vragen**. Laat de gebruiker eigen argumenten toevoegen wanneer een taak start. Ze worden na
de argumenten van het sjabloon toegevoegd.

## Tak {#templateForm-allowOverrideBranchInTask}

Onder **Vragen**. Laat de gebruiker een andere branch, tag of commit kiezen wanneer een taak start,
in plaats van de branch van het sjabloon. Zonder deze optie checkt de taak altijd de branch van het
sjabloon uit.

Het vakje is uitgeschakeld wanneer de repository een monorepo is.

## Inventaris {#templateForm-allowOverrideInventory}

Onder **Vragen**. Wordt getoond voor Ansible-, Terraform-, OpenTofu- en Terragrunt-sjablonen.

Laat de gebruiker een andere inventaris kiezen wanneer een taak start. Alleen Ansible-sjablonen
gebruiken de inventaris die daar is gekozen: een Terraform-, OpenTofu- of Terragrunt-taak draait in
de workspace van het sjabloon, ook wanneer het vakje is aangevinkt.

## Limiet {#templateForm-limit}

Wordt getoond voor Ansible-sjablonen.

De hosts of groepen uit de inventaris waarop het playbook draait, doorgegeven als `--limit`. Elke
chip is één patroon, bijvoorbeeld `web` of `db01`; laat de lijst leeg om op elke host te draaien.

## Tags {#templateForm-tags}

Wordt getoond voor Ansible-sjablonen.

Voert alleen de taken en rollen met deze tags uit, doorgegeven als `--tags`. Laat de lijst leeg om
alles uit te voeren.

## Sla tags over {#templateForm-skipTags}

Wordt getoond voor Ansible-sjablonen.

Slaat de taken en rollen met deze tags over, doorgegeven als `--skip-tags`.

## Kluis {#templateVaults}

Wordt getoond voor Ansible-sjablonen.

De Ansible Vault-wachtwoorden die het playbook nodig heeft om zijn bestanden te ontsleutelen. Klik
op **+** om er een toe te voegen, op een chip om hem te wijzigen, op het kruisje van een chip om
hem te verwijderen. Elke kluis heeft een naam, zijn vault ID, en geeft het wachtwoord uit een
sleutel van de Sleutel Opslag of uit een client script van de repository.

## Automatisch goedkeuren {#templateForm-autoApprove}

Wordt getoond voor Terraform-, OpenTofu- en Terragrunt-sjablonen.

Past het plan toe zonder te vragen. Zonder deze optie stopt een taak na het plan en wacht ze totdat
een gebruiker haar bevestigt of afwijst; een plan zonder wijzigingen beëindigt de taak zonder iets
toe te passen.

## Override backend settings {#templateForm-overrideBackend}

Wordt getoond voor Terraform-, OpenTofu- en Terragrunt-sjablonen wanneer Semaphore de
Terraform-state kan bewaren.

Bewaart de state van de code in Semaphore. Vóór `init` schrijft Semaphore een bestand dat de
`http`-backend instelt in de map van de code, en verwijdert het na de taak:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Wordt getoond met **Override backend settings**; ingeschakeld wanneer dat vakje is aangevinkt.
Optioneel.

De naam van het bestand waarin Semaphore de backend schrijft, bijvoorbeeld `semaphore_backend.tf`.
De standaard is `backend.tf`. De naam moet eindigen op `.tf` en mag alleen letters, cijfers, `_`,
`-` en `.` bevatten. Kies een naam die geen bestand van uw code heeft: Semaphore overschrijft dat
bestand en verwijdert het na de taak.

## Limiet {#templateForm-allowOverrideLimit}

Onder **Ansible vragen**. Laat de gebruiker de hosts instellen waarop wordt gedraaid wanneer een
taak start, in plaats van de **Limiet** van het sjabloon.

## Tags {#templateForm-allowOverrideTags}

Onder **Ansible vragen**. Laat de gebruiker de tags instellen wanneer een taak start, in plaats van
de **Tags** van het sjabloon.

## Sla tags over {#templateForm-allowOverrideSkipTags}

Onder **Ansible vragen**. Laat de gebruiker de over te slaan tags instellen wanneer een taak start,
in plaats van **Sla tags over** van het sjabloon.

## Debug {#templateForm-allowDebug}

Onder **Ansible vragen**. Laat de gebruiker uitgebreide uitvoer inschakelen wanneer een taak
start, van `-v` tot `-vvvvvv`.

## Droge Run {#templateForm-showDryRun}

Onder **Ansible vragen**. Standaard aangevinkt. Biedt de optie **Droge Run** aan wanneer een taak
start, die het playbook uitvoert met `--check`: Ansible meldt wat het zou wijzigen en wijzigt
niets. Vink het vakje uit om de optie weg te halen.

## Verschil {#templateForm-showDiff}

Onder **Ansible vragen**. Standaard aangevinkt. Biedt de optie **Verschil** aan wanneer een taak
start, die het playbook uitvoert met `--diff` om de wijzigingen in bestanden te tonen. Vink het
vakje uit om de optie weg te halen.

## Automatisch goedkeuren {#templateForm-allowAutoApprove}

Onder de vragen van Terraform, OpenTofu en Terragrunt. Laat de gebruiker **Auto Approve**
aanvinken wanneer een taak start, om het plan zonder bevestiging toe te passen. Het
taakdialoogvenster biedt die optie in elk geval aan, maar zonder dit vakje negeert Semaphore haar:
de taak wacht na het plan op een bevestiging, tenzij **Automatisch goedkeuren** van het sjabloon is
aangevinkt.

## Galaxy install options {#templateForm-galaxy}

Wordt getoond voor Ansible-sjablonen. Klik erop om de opties van de Galaxy-installatie te tonen;
het label telt de opties die van de standaard afwijken.

Voordat het playbook draait, installeert Semaphore met `ansible-galaxy` de rollen en collecties
waar de repository om vraagt. Het zoekt naar `requirements.yml`, `roles/requirements.yml` en
`collections/requirements.yml` naast het playbook en in de hoofdmap van de repository, en voert de
installatie alleen uit wanneer een bestand sinds de vorige taak is gewijzigd.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Onder **Galaxy install options**. Slaat de Galaxy-installatie over: het playbook draait met de
rollen en collecties die eerder zijn geïnstalleerd. Gebruik dit wanneer de vereisten op een andere
manier worden geïnstalleerd of de installatie traag is en niet nodig.

## Role install args {#templateForm-galaxyRoleArgs}

Onder **Galaxy install options**. Argumenten die aan `ansible-galaxy role install` worden
toegevoegd, één per chip, bijvoorbeeld `--ignore-certs` of `--server=https://galaxy.example.com`.
Semaphore accepteert alleen de vlaggen van de installatie die veilig kunnen worden doorgegeven:
met een onbekende vlag, `--token`, `--api-key`, `-p` of `--roles-path` kan het sjabloon niet worden
opgeslagen.

## Collection install args {#templateForm-galaxyCollectionArgs}

Onder **Galaxy install options**. Argumenten die aan `ansible-galaxy collection install` worden
toegevoegd, één per chip, bijvoorbeeld `--pre` of `--upgrade`. Dezelfde regel geldt als voor
**Role install args**: `--token`, `--api-key`, `-p` en `--collections-path` worden niet
geaccepteerd.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Onder **Galaxy install options**, in **Prompts**. Laat de gebruiker bij het starten van een taak
kiezen of de Galaxy-installatie wordt overgeslagen. De keuze van de taak gaat voor **Skip Galaxy
install** van het sjabloon.
