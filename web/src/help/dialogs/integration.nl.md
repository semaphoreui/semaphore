---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Uitvoering bewerken

Een uitvoering is wat een integratie start wanneer een andere dienst een webhook-URL van Semaphore
aanroept, bijvoorbeeld bij een push-gebeurtenis van GitHub: een taak van een taaksjabloon of een
workflow. De URL en de controle van het verzoek horen bij het endpoint van de integratie, niet bij
de uitvoering. De pagina van een uitvoering bevat de matchers en de waarden die ze uit het verzoek
haalt.

## Naam {#integrationForm-name}

De naam waaronder de integratie wordt vermeld op de pagina Integraties. Taken die de integratie
start, worden in het auditlogboek aan haar toegeschreven. Het veld is verplicht.

## Integratie {#integrationForm-groupId}

De integratie waartoe de uitvoering behoort. De uitvoering start bij verzoeken die het endpoint van
die integratie bereiken en door de autorisatie komen. Het veld is verplicht.

## Doel {#integrationForm-targetToggle}

Wordt getoond wanneer de server workflows ondersteunt.

Wat de integratie start: **Template** start een taak van een taaksjabloon, **Workflow** start een
run van een workflow. Overschakelen naar **Workflow** laat de taakparameters hieronder vervallen.

## Workflow to run {#integrationForm-workflowId}

Wordt getoond wanneer **Doel** op **Workflow** staat.

De workflow die een verzoek start. Elke waarde die de integratie uit het verzoek neemt, wordt aan
de workflow doorgegeven als het antwoord op zijn enquêtevariabele met dezelfde naam.
Workflow-integraties vereisen de Pro-editie: zonder die editie is het veld uitgeschakeld en
weigert de server de wijziging.

## Uit te voeren taaksjabloon {#integrationForm-templateId}

Wordt getoond wanneer **Doel** op **Template** staat.

Het taaksjabloon waarvan een verzoek een taak start. Het veld is verplicht. Zodra een sjabloon is
gekozen, verschijnen de parameters van zijn taken eronder.

## Taakparameters {#integrationForm-taskParams}

Wordt getoond wanneer een taaksjabloon is gekozen.

De parameters die elke taak krijgt die deze integratie start, alsof u de taak zelf vanuit het
sjabloon startte. Alleen de opties die het sjabloon een taak laat overschrijven, staan hier.
Waarden die uit het verzoek worden genomen, gaan voor deze: een waarde van het type **Variables**
voor een antwoord met dezelfde naam, een waarde van het type **Task Params** voor een parameter met
dezelfde naam.

Wanneer het sjabloon CLI-argumenten in een taak toestaat, bevat **CLI args** extra argumenten voor
de opdrachtregel van de taak.

## Bericht {#taskParams-message}

Optioneel. Het bericht van elke taak die de integratie start, getoond in de geschiedenis van
taken.

## Enquête Variabelen {#taskParams-survey}

Wordt getoond wanneer het sjabloon enquêtevariabelen heeft.

De antwoorden op de enquête van het sjabloon, voor elke taak hetzelfde. Een waarde van het type
**Variables** die uit het verzoek wordt genomen, vervangt het antwoord met dezelfde naam.

Antwoorden op variabelen van het type secret worden niet met de integratie opgeslagen: geef een
geheim in plaats daarvan door via de variabele groep van het sjabloon.

## Tak {#taskParams-gitBranch}

Wordt getoond wanneer het sjabloon toestaat de branch in een taak te overschrijven.

De branch of tag van de repository die de taken uitchecken. Laat het veld zoals het is om de
branch van het sjabloon te gebruiken. Een waarde met de naam `git_branch` van het type
**Task Params**, uit het verzoek genomen, overschrijft het; een voorafgaande `refs/heads/` wordt
uit die waarde verwijderd.

## Inventaris {#taskParams-inventoryId}

Wordt getoond wanneer het sjabloon toestaat de inventaris in een taak te overschrijven.

De inventaris, of voor Terraform de workspace, waarop de taken worden uitgevoerd in plaats van die
van het sjabloon.

## Ansible-opties {#taskParams-ansible}

Wordt getoond wanneer het sjabloon een Ansible-playbook uitvoert.

De opties van `ansible-playbook` voor elke taak, elk getoond wanneer het sjabloon het toestaat:
**Limiet**, **Tags**, **Sla tags over**, **Debug**, **Droge Run** (`--check`), **Verschil** en
**Skip Galaxy install**.

## Terraform-opties {#taskParams-terraform}

Wordt getoond wanneer het sjabloon Terraform, OpenTofu of Terragrunt uitvoert.

- **Plan**: de taak voert het plan uit en stopt, er wordt niets toegepast.
- **Destroy**: het plan en de apply vernietigen de resources.
- **Auto Approve**: het plan wordt zonder bevestiging toegepast, wanneer het sjabloon automatisch
  goedkeuren toestaat. Zonder deze optie wacht een taak totdat iemand haar in Semaphore bevestigt.
- **Upgrade** en **Reconfigure**: `init` draait met `-upgrade` en `-reconfigure`.
