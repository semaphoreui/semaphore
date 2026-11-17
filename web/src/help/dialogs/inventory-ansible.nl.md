---
form: src/components/InventoryForm.vue
---

# Bewerken Ansible Inventory

Een inventaris vertelt Ansible op welke hosts een playbook moet worden uitgevoerd, en bevat de
referenties waarmee Semaphore verbinding met die hosts maakt. Een taaksjabloon verwijst naar een
inventaris, en elke taak van dat sjabloon wordt op die inventaris uitgevoerd. Het formaat van een
inventaris staat beschreven in de
[Ansible-handleiding over inventarissen](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Naam {#inventoryForm-name}

De naam waaronder de inventaris wordt weergegeven op de pagina Inventaris en in de
inventariskeuze van een taaksjabloon. Kies een naam die de hosts van elkaar onderscheidt,
bijvoorbeeld `Production web servers`.

## Runner tag {#inventoryForm-runnerTag}

Wordt getoond wanneer er projectrunners beschikbaar zijn.

Taken die deze inventaris gebruiken, worden uitgevoerd op een runner met deze tag. Een runner tag
die op het taaksjabloon is ingesteld, gaat voor de tag van de inventaris. Laat het veld leeg om
elke runner de taken te laten uitvoeren.

## Gebruikersreferenties {#inventoryForm-sshKeyId}

De sleutel uit de Sleutel Opslag waarmee Ansible zich op de hosts aanmeldt. Het veld is verplicht
en accepteert drie soorten sleutels:

- **SSH Sleutel**: Ansible maakt verbinding met de privésleutel en, als de sleutel er een heeft,
  met de bijbehorende login.
- **Inloggen met wachtwoord**: Ansible maakt verbinding met de login en het wachtwoord.
- **Geen**: Semaphore geeft geen referenties door. Gebruik dit voor hosts die er geen nodig
  hebben, zoals `localhost`, of wanneer de inventaris zelf bepaalt hoe verbinding wordt gemaakt.

## Sudo Referenties {#inventoryForm-becomeKeyId}

Optioneel. Een sleutel van het type **Inloggen met wachtwoord** uit de Sleutel Opslag, die wordt
gebruikt wanneer een play met `become` zijn rechten verhoogt: de login ervan is de gebruiker die
Ansible wordt, en het wachtwoord ervan is het wachtwoord waarmee Ansible de prompt beantwoordt.

Laat het veld leeg wanneer de plays geen rechten verhogen, of wanneer de gebruiker hierboven
daarvoor geen wachtwoord nodig heeft.

## Type {#inventoryForm-type}

Waar de inventaris vandaan komt:

- **Static**: de inventaris wordt in dit dialoogvenster ingevoerd in het INI-formaat.
- **Static YAML**: de inventaris wordt in dit dialoogvenster ingevoerd in het YAML-formaat.
  Gebruik dit ook voor de configuratie van een inventarisplugin.
- **File**: de inventaris is een bestand in een Git-repository.
- **Managed**: de hosts en hun variabelen worden als records bewaard. Open de inventaris vanuit de
  lijst om ze te bewerken.

De velden onder het type veranderen mee met het type.

## Pad naar Inventarisbestand {#inventoryForm-inventoryPath}

Wordt getoond wanneer het type **File** is.

Het pad naar het inventarisbestand vanaf de hoofdmap van de repository, bijvoorbeeld
`inventory/production.yml`. De repository is de repository die in het veld Repository is
geselecteerd of, wanneer dat veld leeg is, de repository van het taaksjabloon dat de inventaris
gebruikt.

## Repository {#inventoryForm-repositoryId}

Wordt getoond wanneer het type **File** is. Optioneel.

De repository die het inventarisbestand bevat. Semaphore werkt zijn kopie van de repository vóór
elke taak bij. Laat het veld leeg wanneer het bestand in dezelfde repository staat als het
playbook: het pad wordt dan opgezocht in de repository van het taaksjabloon.

## Inventaris {#inventoryForm-inventoryCode}

Wordt getoond wanneer het type **Static** of **Static YAML** is.

De tekst van de inventaris. Semaphore slaat deze op en schrijft hem voor elke taak naar een
tijdelijk bestand. Een statische inventaris in het INI-formaat:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Dezelfde hosts als **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
