---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Integratie Bewerken

Een integratie start een taaksjabloon, of een workflow, wanneer een andere dienst een webhook-URL
van Semaphore aanroept, zoals een push-event van GitHub. De URL is een alias: `/api/integrations/`
en een willekeurige code. De aliassen van het project staan op de pagina Integraties, de aliassen
van één integratie op haar eigen pagina. Die pagina bevat ook de matchers en de waarden die de
integratie uit het verzoek neemt.

## Name {#integrationForm-name}

De naam waaronder de integratie wordt vermeld op de pagina Integraties. Taken die de integratie
start, worden in het auditlogboek aan haar toegeschreven. Het veld is verplicht.

## Runs {#integrationForm-targetToggle}

Wordt getoond wanneer de server workflows ondersteunt.

Wat de integratie start: **Template** start een taak van een taaksjabloon, **Workflow** start een
run van een workflow. Overschakelen naar **Workflow** laat de taakparameters hieronder vervallen.

## Workflow to run {#integrationForm-workflowId}

Wordt getoond wanneer **Runs** op **Workflow** staat.

De workflow die een verzoek start. Elke waarde die de integratie uit het verzoek neemt, wordt aan
de workflow doorgegeven als het antwoord op zijn enquêtevariabele met dezelfde naam.
Workflow-integraties vereisen de Pro-editie: zonder die editie is het veld uitgeschakeld en
weigert de server de wijziging.

## Task Template to run {#integrationForm-templateId}

Wordt getoond wanneer **Runs** op **Template** staat.

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

## Auth method {#integrationForm-authMethod}

Hoe Semaphore controleert dat een verzoek afkomstig is van de dienst die u verwacht. Een verzoek
dat niet door de controle komt, start niets, en de webhook beantwoordt het net als elk ander
verzoek.

- **None**: geen controle. Iedereen die de URL kent, kan de taak starten.
- **GitHub Webhooks**: de header `X-Hub-Signature-256` moet de HMAC-SHA256-handtekening van de
  body bevatten, gemaakt met het geheim, zoals GitHub die verstuurt.
- **Bitbucket Webhooks**: dezelfde controle op de header `X-Hub-Signature`, zoals Bitbucket die
  verstuurt.
- **Token**: de header in **Auth header** moet het geheim zelf bevatten.
- **HMAC** en **HMAC (SHA-512)**: de header in **Auth header** moet de HMAC-SHA256 of HMAC-SHA512
  van de body bevatten, gemaakt met het geheim, in hexadecimale kleine letters, zonder voorvoegsel
  zoals `sha256=`.
- **BasicAuth**: het verzoek moet de login en het wachtwoord van de sleutel in **Auth-secret** in
  zijn header `Authorization` meesturen.

## Auth header {#integrationForm-authHeader}

Wordt getoond wanneer **Auth method** op **Token**, **HMAC** of **HMAC (SHA-512)** staat.

De naam van de header van het verzoek die het token of de handtekening bevat, bijvoorbeeld
`X-Webhook-Token`. Hoofdletters en kleine letters in de naam maken niet uit. Een verzoek zonder die
header wordt geweigerd.

## Auth-secret {#integrationForm-authSecretId}

Wordt getoond wanneer **Auth method** niet op **None** staat.

De sleutel uit de Sleutel Opslag waartegen het verzoek wordt gecontroleerd. Voor **BasicAuth**
past alleen een sleutel van het type **Inloggen met wachtwoord**. Voor de andere methoden past ook
een sleutel van het type **Tekenreeks**: de controle gebruikt de waarde ervan, of het wachtwoord
van een sleutel van het type **Inloggen met wachtwoord**.

Het veld is voor elke methode behalve **None** nodig: zonder sleutel, of met een leeg geheim,
wordt elk verzoek geweigerd.
