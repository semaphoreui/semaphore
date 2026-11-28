---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Endpoint bewerken

Een endpoint is een ingang voor webhooks: de URL's die andere diensten aanroepen en de manier waarop
hun verzoeken worden gecontroleerd. Integraties hangen aan een endpoint en starten hun uitvoeringen
bij de verzoeken die het accepteert. De URL's van een endpoint, de aliassen, voegt u toe op het
tabblad 'Endpoints'.

## Naam {#endpointForm-name}

De naam waaronder het endpoint staat op het tabblad 'Endpoints' en in het formulier van een
integratie. Het veld is verplicht.

## Beschrijving {#endpointForm-description}

Optioneel. Een notitie over waar het endpoint voor dient; ze staat naast de naam.

## Auth-methode {#integrationAuthForm-authMethod}

Hoe Semaphore controleert dat een verzoek afkomstig is van de dienst die u verwacht. Een verzoek
dat niet door de controle komt, start niets, en de webhook beantwoordt het net als elk ander
verzoek.

- **Geen**: geen controle. Iedereen die de URL kent, kan de taak starten.
- **GitHub Webhooks**: de header `X-Hub-Signature-256` moet de HMAC-SHA256-handtekening van de
  body bevatten, gemaakt met het geheim, zoals GitHub die verstuurt.
- **Bitbucket Webhooks**: dezelfde controle op de header `X-Hub-Signature`, zoals Bitbucket die
  verstuurt.
- **Token**: de header in **Auth-header** moet het geheim zelf bevatten.
- **HMAC** en **HMAC (SHA-512)**: de header in **Auth-header** moet de HMAC-SHA256 of HMAC-SHA512
  van de body bevatten, gemaakt met het geheim, in hexadecimale kleine letters, zonder voorvoegsel
  zoals `sha256=`.
- **BasicAuth**: het verzoek moet de login en het wachtwoord van de sleutel in **Auth-secret** in
  zijn header `Authorization` meesturen.

## Auth-header {#integrationAuthForm-authHeader}

Wordt getoond wanneer **Auth-methode** op **Token**, **HMAC** of **HMAC (SHA-512)** staat.

De naam van de header van het verzoek die het token of de handtekening bevat, bijvoorbeeld
`X-Webhook-Token`. Hoofdletters en kleine letters in de naam maken niet uit. Een verzoek zonder die
header wordt geweigerd.

## Auth-secret {#integrationAuthForm-authSecretId}

Wordt getoond wanneer **Auth-methode** niet op **Geen** staat.

De sleutel uit de Sleutel Opslag waartegen het verzoek wordt gecontroleerd. Voor **BasicAuth**
past alleen een sleutel van het type **Inloggen met wachtwoord**. Voor de andere methoden past ook
een sleutel van het type **Tekenreeks**: de controle gebruikt de waarde ervan, of het wachtwoord
van een sleutel van het type **Inloggen met wachtwoord**.

Het veld is voor elke methode behalve **Geen** nodig: zonder sleutel, of met een leeg geheim,
wordt elk verzoek geweigerd.
