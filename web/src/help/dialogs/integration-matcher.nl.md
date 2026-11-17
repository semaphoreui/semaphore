---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Een matcher is een voorwaarde voor het binnenkomende verzoek. Matchers bepalen welke integratie
een verzoek aan de alias van het project start, zodat één URL meerdere integraties kan bedienen:
een integratie wordt alleen uitgevoerd wanneer al haar matchers kloppen. Ze worden alleen gebruikt
wanneer **Gebruik projectalias** voor de integratie is ingeschakeld; wanneer dat zo is, is de
integratie alleen via de projectalias bereikbaar.

Een integratie die via de projectalias wordt gevonden en geen matchers heeft, wordt nooit
uitgevoerd: voeg er minstens één toe.

## Name {#integrationMatcherForm-name}

De naam waaronder de matcher wordt vermeld op de pagina van de integratie. De naam verandert niets
aan wat de matcher controleert. Het veld is verplicht.

## Match on {#integrationMatcherForm-matchType}

Het deel van het verzoek dat de matcher leest:

- **Header**: de waarde van een header van het verzoek.
- **Body**: de body van het verzoek.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Wordt getoond wanneer **Match on** op **Body** staat.

Hoe de body wordt gelezen:

- **JSON**: de body wordt als JSON geparsed, en **Key** is het pad naar één waarde daarin.
- **String**: de hele body wordt als tekst vergeleken, en **Key** wordt niet gebruikt.

## Key {#integrationMatcherForm-key}

Wat er uit het verzoek moet worden gelezen. Het veld is verplicht.

- Voor **Header**: de naam van de header, bijvoorbeeld `X-GitHub-Event`. Hoofdletters en kleine
  letters maken niet uit.
- Voor **Body** in **JSON**: het pad naar de waarde, met de delen gescheiden door punten; een
  element van een array wordt geschreven als `.[0]`, bijvoorbeeld `repository.full_name` of
  `commits.[0].id`.
- Voor **Body** in **String**: niet gebruikt, maar het formulier heeft een waarde nodig; typ iets
  willekeurigs.

## Comparison Method {#integrationMatcherForm-method}

Hoe de waarde die uit het verzoek is gelezen, wordt vergeleken met **Value**:

- **==**: de twee zijn gelijk.
- **!=**: de twee verschillen. Een header of een sleutel die in het verzoek ontbreekt, geldt als
  verschillend.
- **Contains**: de waarde die uit het verzoek is gelezen, bevat **Value**.

## Value {#integrationMatcherForm-value}

De tekst waarmee wordt vergeleken, bijvoorbeeld `push` of `refs/heads/main`. De vergelijking is
exact en hoofdlettergevoelig. Een getal of een boolean uit de JSON-body wordt in zijn tekstvorm
vergeleken: `42`, `true`. Het veld is verplicht.
