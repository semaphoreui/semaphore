---
form: src/components/AppForm.vue
---

# Edit App

Een app is een tool waarmee Semaphore taken uitvoert: Ansible, Terraform, OpenTofu, Terragrunt,
Bash, Python, PowerShell of een eigen opdracht. Elk taaksjabloon is gemaakt voor één app, en dit
dialoogvenster bepaalt hoe de server de app start en hoe de app in de interface wordt getoond.

De instellingen worden in de database opgeslagen en worden direct van kracht op de server die ze
opslaat; de andere nodes van een cluster lezen ze bij hun herstart. Een veld dat u leegmaakt,
behoudt zijn oude waarde totdat de server opnieuw start.

## ID {#appForm-id}

Verplicht. De sleutel waaronder de app wordt opgeslagen; een taaksjabloon verwijst met deze
sleutel naar zijn app. De ingebouwde apps zijn `ansible`, `terraform`, `tofu`, `terragrunt`,
`pulumi`, `bash`, `python` en `powershell`. Elke andere ID voegt een eigen app toe: Semaphore voert
die uit als een script, met de opdracht die in **Path** staat of, wanneer **Path** leeg is, met de
opdracht die net zo heet als de ID.

Gebruik een ID zonder punten. Het wijzigen van de ID van een bestaande app hernoemt die niet: de
instellingen worden opgeslagen als een nieuwe app onder de nieuwe ID, en de oude app blijft
bestaan.

## Icon {#appForm-icon}

Optioneel. De naam van een pictogram uit de set Material Design Icons zonder het voorvoegsel
`mdi-`, bijvoorbeeld `language-python`. Laat het veld leeg om het ingebouwde pictogram van de app
te behouden; een eigen app toont dan een vraagteken.

## Naam {#appForm-title}

Optioneel. De naam van de app in de lijst met applicaties en in het menu **Nieuw sjabloon**. Laat
het veld leeg om de ingebouwde naam te behouden, zoals `Ansible Playbook`; een eigen app heeft dan
geen naam, dus geef hem er een.

## Path {#appForm-path}

Optioneel. De opdracht die Semaphore uitvoert in plaats van de standaardopdracht: een naam die in
het `PATH` van de server wordt opgezocht of een volledig pad, bijvoorbeeld
`/usr/local/bin/terraform-1.9`. De standaardopdrachten zijn `terraform`, `tofu`, `terragrunt`,
`bash`, `python3` en `powershell`; een eigen app voert de opdracht uit die net zo heet als zijn ID.

Ansible negeert dit veld: Semaphore voert altijd `ansible-playbook` en `ansible-galaxy` uit.

## Priority {#appForm-priority}

De plaats van de app in de lijst met applicaties en in het menu **Nieuw sjabloon**: een app met
een hoger getal komt eerst. De ingebouwde apps beginnen bij 1000 voor Ansible en lopen af tot 500
voor Python. Een ingebouwde app met een prioriteit van `0` of minder krijgt bij het starten van de
server zijn ingebouwde prioriteit terug.

## Argumenten {#appForm-args}

Optioneel. Argumenten die direct na de opdracht op de opdrachtregel worden gezet, één per chip.
Klik op **+** om er een toe te voegen, klik op een chip om hem te wijzigen, sleep de chips om hun
volgorde te wijzigen.

- Terraform, OpenTofu en Terragrunt: de argumenten komen vóór de argumenten die Semaphore aan elke
  opdracht toevoegt, zoals `init` en `apply`.
- Script-apps: de argumenten komen vóór het script van het sjabloon en vervangen de
  standaardargumenten: PowerShell draait standaard met `-File`, dus zet `-File` als laatste in uw
  lijst.
- Ansible negeert dit veld.

## Active {#appForm-active}

Een actieve app wordt aangeboden in het menu **Nieuw sjabloon**. Vink het vakje uit om de app in
dat menu te verbergen: sjablonen die er al mee zijn gemaakt, behouden hun app en worden nog steeds
uitgevoerd.

Voor een nieuwe app is het vakje uitgevinkt: vink het aan om de app beschikbaar te maken.
