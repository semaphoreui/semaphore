---
form: src/components/InventoryHostForm.vue
---

# Host bewerken

Een host van een beheerde inventaris. Een taak die de inventaris gebruikt, wordt uitgevoerd op de
hosts die in de lijst met hosts zijn ingeschakeld.

## Naam {#inventoryHostForm-name}

De naam waaronder Ansible de host kent, en het adres waarmee het verbinding maakt, tenzij de
variabelen `ansible_host` instellen. Het veld is verplicht. Een naam is uniek binnen zijn inventaris
en is hoofdlettergevoelig: `web1` en `Web1` zijn twee hosts. De naam mag geen spaties, komma's,
vierkante haken of dubbele punten bevatten en is hooguit 250 tekens lang.

## Beschrijving {#inventoryHostForm-description}

Optioneel. Een notitie voor mensen, die in de lijst met hosts wordt getoond. Ansible krijgt deze
niet.

## Variabelen {#inventoryHostForm-variables}

Optioneel. Variabelen van deze host, geschreven als een YAML- of JSON-mapping:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Een variabele van de host gaat voor op een variabele van de inventaris met dezelfde naam.
YAML-ankers, -aliassen en merge-sleutels worden niet geaccepteerd.
