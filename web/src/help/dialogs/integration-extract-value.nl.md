---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Een geëxtraheerde waarde neemt een stuk van het binnenkomende verzoek, zoals de branch van een
push-event, en geeft het door aan de taak die de integratie start. Elke waarde van de integratie
wordt genomen uit elk verzoek dat een taak start.

## Name {#integrationValueForm-name}

De naam waaronder de waarde wordt vermeld op de pagina van de integratie. De naam verandert niets
aan wat er wordt genomen of waar het naartoe gaat. Het veld is verplicht.

## Source of the Value {#integrationValueForm-valueSource}

Het deel van het verzoek waaruit de waarde wordt genomen:

- **Header**: de waarde van een header van het verzoek.
- **Body**: de body van het verzoek.

## Data Type of Body {#integrationValueForm-bodyDataType}

Wordt getoond wanneer **Source of the Value** op **Body** staat.

Hoe de body wordt gelezen:

- **JSON**: de body wordt als JSON geparsed, en **Key** is het pad naar één waarde daarin. De
  waarde behoudt haar JSON-type: een getal, een lijst of een object komen als zodanig bij de taak
  aan.
- **String**: de hele body wordt als tekst genomen, en **Key** wordt niet gebruikt.

## Key {#integrationValueForm-key}

Wat er uit het verzoek moet worden genomen. Het veld is verplicht.

- Voor **Header**: de naam van de header, bijvoorbeeld `X-GitHub-Event`. Hoofdletters en kleine
  letters maken niet uit. Een header die in het verzoek ontbreekt, geeft een lege waarde.
- Voor **Body** in **JSON**: het pad naar de waarde, met de delen gescheiden door punten; een
  element van een array wordt geschreven als `.[0]`, bijvoorbeeld `repository.full_name` of
  `commits.[0].id`. Een pad dat in de body ontbreekt, geeft geen waarde, en de taak behoudt de
  waarde die in de integratie is ingesteld.
- Voor **Body** in **String**: niet gebruikt, maar het formulier heeft een waarde nodig; typ iets
  willekeurigs.

## Variable Usage {#integrationValueForm-variableType}

Waar de waarde terechtkomt in een taak van een sjabloon:

- **Variables**: een extra variabele van de taak, zoals het antwoord op een enquêtevariabele
  wordt doorgegeven. De waarde vervangt een antwoord of een variabele van de variabele groep met
  dezelfde naam.
- **Task Params**: een parameter van de taak, zoals de opties van het taakdialoogvenster worden
  opgeslagen. De naam `git_branch` stelt de branch van de taak in wanneer het sjabloon toestaat de
  branch te overschrijven; een voorafgaande `refs/heads/` wordt verwijderd.

Een integratie die een workflow uitvoert, negeert het gebruik: elke waarde wordt het antwoord op
de enquêtevariabele van de workflow die in **Variable** wordt genoemd.

## Variable {#integrationValueForm-variable}

De naam waaronder de waarde wordt doorgegeven, bijvoorbeeld `branch` voor een extra variabele die
het playbook leest als `{{ branch }}`. Het veld is verplicht.
