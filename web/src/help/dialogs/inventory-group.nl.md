---
form: src/components/InventoryGroupForm.vue
---

# Groep bewerken

Een groep hosts van een beheerde inventaris. Een playbook wordt via de naam van de groep op een deel
van de hosts uitgevoerd, en een variabele van de groep wordt één keer ingesteld voor al haar hosts.
Hosts komen in een groep op het tabblad Hosts, in het hostvenster of met de acties voor de selectie.

## Naam {#inventoryGroupForm-name}

De naam waaronder Ansible de groep kent, bijvoorbeeld in `hosts: web` van een playbook. Het veld is
verplicht. Een naam is uniek binnen zijn inventaris en is hoofdlettergevoelig: `web` en `Web` zijn
twee groepen. De naam mag geen spaties, komma's, vierkante haken of dubbele punten bevatten, is
hooguit 250 tekens lang en mag niet `all` of `ungrouped` zijn. Ansible waarschuwt voor namen met
andere tekens dan letters, cijfers en underscores, zoals `web-eu`, maar de naam werkt toch. De
tekens `*`, `?`, `!`, `&`, `~` en `@` zijn ook niet toegestaan, omdat Ansible ze als een patroon
leest.

## Beschrijving {#inventoryGroupForm-description}

Optioneel. Een notitie voor mensen, die in de lijst met groepen wordt getoond. Ansible krijgt deze
niet.

## Child groups {#inventoryGroupForm-children}

Optioneel. Groepen binnen deze groep. De hosts van een onderliggende groep horen ook bij deze groep.
Een groep kan niet haar eigen voorouder zijn.

## Variabelen {#inventoryGroupForm-variables}

Optioneel. Variabelen van deze groep, geschreven als een YAML- of JSON-mapping:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Elke host van de groep krijgt ze. Een variabele van een host gaat voor op een variabele van de groep
met dezelfde naam, en een variabele van de groep gaat voor op een variabele van de inventaris.
YAML-ankers, -aliassen en merge-sleutels worden niet geaccepteerd.
