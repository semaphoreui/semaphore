---
form: src/components/SurveyVars.vue
---

# Enquête Variabele

Een enquêtevariabele is een vraag die het taakdialoogvenster stelt elke keer dat u het sjabloon
uitvoert. Het antwoord bereikt de taak als een variabele: een extra variabele voor Ansible, een
`-var` voor Terraform en OpenTofu, een argument voor een script, of een omgevingsvariabele. De
variabele wordt met het sjabloon opgeslagen, wanneer u het sjabloondialoogvenster opslaat.
Hetzelfde dialoogvenster bewerkt de enquête van een workflow, waar het type **Secret** niet wordt
aangeboden.

## Naam {#surveyVarDialog-name}

De naam waaronder de taak het antwoord krijgt: de naam van de extra variabele, van de
Terraform-variabele, van het argument `name=value` van een script, of van de omgevingsvariabele.
Spaties aan het begin en eind worden weggelaten. Het veld is verplicht.

Gebruik een naam die geen andere variabele van de enquête heeft, en een die de tool accepteert,
bijvoorbeeld `app_version`. Een antwoord vervangt een variabele met dezelfde naam uit de variabele
groep van het sjabloon.

## Titel {#surveyVarDialog-title}

Het label van het veld in het taakdialoogvenster, bijvoorbeeld `Version to deploy`. Het is ook de
tekst van de chip van de variabele in het sjabloondialoogvenster. Het veld is verplicht.

## Beschrijving {#surveyVarDialog-description}

Optioneel. Een hint die onder het veld van het taakdialoogvenster wordt getoond zolang het de focus
heeft: wat er moet worden ingevuld, of een voorbeeld.

## Type {#surveyVarDialog-type}

Om wat voor antwoord het taakdialoogvenster vraagt:

- **String**: één regel tekst. Dit is de standaard.
- **Integer**: alleen cijfers; het taakdialoogvenster weigert al het andere, ook een minteken. Het
  antwoord bereikt de taak toch als tekst: zet het in het playbook om, bijvoorbeeld met `| int`.
- **Secret**: tekst die het taakdialoogvenster maskeert. Het antwoord wordt niet met de taak
  opgeslagen, dus een herhaalde run vraagt er opnieuw om, en Ansible krijgt het in een met vault
  versleuteld bestand. Een secret heeft geen standaardwaarde.
- **Enum**: één waarde, gekozen uit een lijst die u hieronder definieert.
- **Text**: meerdere regels tekst.
- **Select**: meerdere waarden, gekozen uit een lijst die u hieronder definieert. De taak krijgt
  ze als lijst; in een argument, een `-var` of een omgevingsvariabele wordt de lijst als JSON
  geschreven, bijvoorbeeld `["a","b"]`.

Het wijzigen van het type wist de standaardwaarde, en een type zonder lijst wist de waarden.

## Pass variable as {#surveyVarDialog-target}

Hoe het antwoord de taak bereikt:

- **Extra variable**: de gebruikelijke manier van de tool. Ansible krijgt het als extra variabele,
  Terraform en OpenTofu als `-var name=value`, een Bash-, PowerShell- of Python-script als
  argument `name=value` na het script.
- **Environment variable**: een omgevingsvariabele van het proces, `name=value`, voor elk soort
  sjabloon. De variabele wordt dan niet op de andere manier doorgegeven.

## Sleutel {#surveyVarDialog-valueName}

Wordt getoond wanneer het type **Enum** of **Select** is.

De tekst die het taakdialoogvenster voor deze waarde in de lijst toont. Elke waarde heeft een
sleutel nodig, en de sleutels moeten verschillen: anders weigert het dialoogvenster de variabele
op te slaan.

## Waarde {#surveyVarDialog-valueValue}

Wordt getoond wanneer het type **Enum** of **Select** is.

Wat de taak krijgt wanneer deze waarde wordt gekozen. De waarde mag verschillen van de sleutel: de
sleutel `Production` kan `prod` doorgeven.

## Add Value {#surveyVarDialog-addValue}

Wordt getoond wanneer het type **Enum** of **Select** is.

Voegt een lege rij toe aan de lijst met waarden. Een **Enum** of een **Select** heeft minstens één
waarde nodig. Het prullenbakpictogram van een rij verwijdert de rij.

## Default value {#surveyVarDialog-defaultValue}

Wordt getoond voor elk type behalve **Secret**.

Het antwoord dat het taakdialoogvenster invult wanneer het opent; u kunt het daar wijzigen. Kies
voor een **Enum** een van de waarden, voor een **Select** een willekeurig aantal ervan. Laat het
veld leeg om met een leeg antwoord te beginnen.

De standaardwaarde wordt alleen door het taakdialoogvenster ingevuld. Een taak die op een andere
manier wordt gestart, bijvoorbeeld via de API, krijgt geen standaardwaarde.

## Vereist {#surveyVarDialog-required}

Het taakdialoogvenster start geen taak totdat de variabele een antwoord heeft. Net als de
standaardwaarde wordt de controle alleen door het taakdialoogvenster uitgevoerd, niet door de
server.
