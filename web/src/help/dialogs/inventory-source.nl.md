---
form: src/components/InventorySourceForm.vue
---

# Bron bewerken

Een bron importeert de hosts en groepen van een beheerde inventaris uit een bestand in een Git-repository. Het bestand wordt bij elke druk op Sync gelezen, en de hosts en groepen die het beschrijft blijven gelijk aan het bestand. De server heeft Ansible en toegang tot de repository nodig.

## Naam {#inventorySourceForm-name}

De naam van de bron, getoond op de kaart en naast de hosts die de bron aanlevert. Het veld is verplicht.

## Repository {#inventorySourceForm-repository}

De repository van het project waaruit het bestand wordt gelezen. Het veld is verplicht. De branch ervan wordt gebruikt.

## Pad {#inventorySourceForm-path}

Het pad van een inventarisbestand in de repository, bijvoorbeeld `inventory/prod.yml`, of van een map met inventarisbestanden. Het veld is verplicht.

Bestanden worden gelezen als statische YAML of INI, scripts en configuratiebestanden van plugins worden geweigerd. In een map moet elk bestand dat Ansible niet overslaat een geldige inventaris zijn, anders mislukt de synchronisatie en verandert er niets. Ansible slaat in een map sommige extensies over, bijvoorbeeld `.cfg`, `.md` en `.txt`. Een repository met een link die naar buiten de repository leidt, wordt geweigerd.
