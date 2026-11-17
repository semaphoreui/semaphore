---
form: src/views/Tokens.vue
---

# New Token

Met een API-token kan een script of een andere dienst de Semaphore-API aanroepen als u, met
dezelfde rechten als uw account. Stuur het mee in de header `Authorization` van een verzoek:
`Bearer` en het token. Het token wordt één keer getoond, direct nadat het is aangemaakt: kopieer
het dan.

## Token Name {#token-name}

Optioneel. Een naam die het token onderscheidt in de lijst met tokens, bijvoorbeeld de naam van
het script dat het gebruikt. De naam maakt geen deel uit van het token en verandert niets aan wat
het token mag.

## Expires {#token-expires}

Wanneer het token ophoudt te werken. Kies een periode vanaf nu, **Never** voor een token dat werkt
totdat het wordt verwijderd, of **Custom** om het moment zelf te kiezen.

Een token dat nooit verloopt, werkt totdat u het uit de lijst verwijdert. Kies bij voorkeur een
periode voor een token dat aan een script of aan een andere dienst wordt gegeven.

## Expiration date {#token-expiresAt}

Wordt getoond wanneer **Expires** op **Custom** staat.

De datum en tijd waarop het token ophoudt te werken, in de tijdzone van uw browser. Het moment
moet in de toekomst liggen: de server weigert een token dat al verlopen zou zijn.
