---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Variabele Groep Bewerken

Een variabele groep bevat de extra variabelen, de omgevingsvariabelen en de geheimen die Semaphore
doorgeeft aan elke taak van een taaksjabloon dat de groep gebruikt. Ansible krijgt de extra
variabelen als `--extra-vars`, Terraform en OpenTofu als `-var`; de omgevingsvariabelen worden
ingesteld voor het proces van de taak. Een sjabloon kan meerdere groepen gebruiken: wanneer twee
ervan dezelfde naam definiëren, wint de groep die later in het sjabloon komt.

## Groepsnaam {#varGroup-name}

De naam waaronder de groep wordt vermeld op de pagina Variabele Groepen en in de keuze van de
variabele groep van een taaksjabloon. Het veld is verplicht.

## Secret storage {#varGroup-secretStorageId}

Wordt getoond wanneer de groep nieuw is en de server geheimenopslagen ondersteunt. Optioneel.

De geheimenopslag waarin de geheimen van deze groep worden bewaard, bijvoorbeeld een HashiCorp
Vault. Laat het veld leeg om de geheimen versleuteld in de database van Semaphore te bewaren.

De opslag wordt één keer gekozen, wanneer de groep wordt aangemaakt: een bestaande groep behoudt
haar opslag, en het veld wordt niet meer getoond. Een opslag die als **Read only** is gemarkeerd,
neemt geen nieuwe geheimen aan: het opslaan van een geheim erin mislukt.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Wordt getoond wanneer de groep nieuw is en de server geheimenopslagen ondersteunt. Beschikbaar
zodra een **Secret storage** is gekozen.

Het begin van het pad dat elk geheim van deze groep in de opslag krijgt. Semaphore voegt er 10
willekeurige tekens aan toe: met het voorvoegsel `semaphore/prod/` wordt een geheim opgeslagen als
`semaphore/prod/` gevolgd door die tekens. Het voorvoegsel wordt genomen zoals het is, dus laat
het eindigen op `/` wanneer de geheimen eronder als in een map moeten staan.

Het voorvoegsel wordt één keer ingesteld, wanneer de groep wordt aangemaakt. Het tabblad
**Secrets** toont het als het patroon van het bronpad.

## Tabel, JSON of YAML {#varGroup-extraVarsMode}

Hoe u de extra variabelen bewerkt:

- **Table**: één rij per variabele, elk met een type. Voeg een rij toe met de knop **+** naast de
  schakelaar.
- **JSON**: alle variabelen als één JSON-object.
- **YAML**: alle variabelen als één YAML-mapping.

Het dialoogvenster opent in **Table**. Een omschakeling zet om wat u hebt getypt; wanneer dat niet
kan worden gelezen, blijft het dialoogvenster in de modus waarin u was en toont het de fout. In
welke modus ook, de variabelen worden opgeslagen als één object van namen en waarden.

## Extra variabelen in JSON {#varGroup-extraVarsJson}

Wordt getoond wanneer de extra variabelen als **JSON** worden bewerkt.

De extra variabelen als één JSON-object. Een naam mag niet leeg zijn; een waarde mag een string,
een getal, een lijst of een object zijn:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Extra variabelen in YAML {#varGroup-extraVarsYaml}

Wordt getoond wanneer de extra variabelen als **YAML** worden bewerkt.

De extra variabelen als één YAML-mapping. De tekst wordt gelezen met de typen die JSON kent: een
datum zonder aanhalingstekens zoals `2024-01-01` blijft een string, en `.inf` of `.nan` worden
geweigerd. Een lijst of een scalaire waarde op het hoogste niveau wordt ook geweigerd.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Naam van de variabele {#varGroup-varName}

Wordt getoond wanneer de extra variabelen als **Table** worden bewerkt, in elke rij.

De naam van de variabele zoals het playbook of de Terraform-code die ziet, bijvoorbeeld
`app_version`.

## Type van de variabele {#varGroup-varType}

Wordt getoond wanneer de extra variabelen als **Table** worden bewerkt, in elke rij.

Hoe de waarde van de rij wordt gelezen:

- **String**: de tekst zoals hij is.
- **Number**: een getal, bijvoorbeeld `42`.
- **List**: een JSON-array, bijvoorbeeld `["a", "b"]`.
- **Dict**: een JSON-object, bijvoorbeeld `{"key": "value"}`.

Een waarde die niet bij haar type past, stopt het opslaan met de naam van de variabele.

## Waarde van de variabele {#varGroup-varValue}

Wordt getoond wanneer de extra variabelen als **Table** worden bewerkt, in elke rij.

De waarde van de variabele, in het formaat van haar type. Voor een **List** of een **Dict** opent
de knop aan het eind van het veld een grotere editor.

Semaphore geeft de waarden als data aan Ansible door: een expressie `{{ … }}` in een waarde wordt
niet geëvalueerd, het playbook krijgt de tekst zoals hij is.

## Naam van de omgevingsvariabele {#varGroup-envName}

Op het tabblad **Variables**, onder **Omgevingsvariabelen**, in elke rij. Voeg een rij toe met de
knop **+** van de sectie.

De naam van een variabele die wordt ingesteld in de omgeving van het proces dat de taak uitvoert,
bijvoorbeeld `AWS_REGION`.

## Waarde van de omgevingsvariabele {#varGroup-envValue}

Op het tabblad **Variables**, onder **Omgevingsvariabelen**, in elke rij.

De waarde van de variabele, als tekst. Ze wordt als platte tekst opgeslagen en getoond aan iedereen
die de groep opent: zet wachtwoorden en tokens in plaats daarvan op het tabblad **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Wordt getoond op het tabblad **Secrets** van een groep die een geheimenopslag heeft.

Schakel dit in om geheimen uit de opslag in deze groep te importeren, vanaf de paden die onder
**Sync paths** zijn ingesteld. De import draait wanneer u op de synchronisatieknop van de groep
drukt op de pagina Variabele Groepen, en vanzelf met het interval dat in **Sync paths** is
ingesteld. Wanneer dit uit staat, wordt er niets geïmporteerd.

## Sync paths {#varGroup-syncPaths}

Wordt getoond wanneer **Sync keys enabled** aan staat.

Opent het dialoogvenster met de paden van de opslag waaruit wordt geïmporteerd en het interval van
de automatische import. Het getal op de knop is het aantal paden. De synchronisatieknop van de
groep op de pagina Variabele Groepen blijft uitgeschakeld totdat de groep minstens één pad heeft.

Het dialoogvenster heeft alleen **Sluiten**: wat u erin wijzigt, wordt met de groep opgeslagen,
via **Opslaan** van dit dialoogvenster.

## Naam van de geheime variabele {#varGroup-secretVarName}

Op het tabblad **Secrets**, onder **Extra variabelen**, in elke rij. Voeg een rij toe met de knop
**+** van de sectie.

De naam van een extra variabele waarvan de waarde een geheim is. De taak krijgt hem zoals ze de
extra variabelen van het tabblad **Variables** krijgt; een geheim gaat voor een extra variabele
met dezelfde naam. Ansible krijgt de geheimen in een bestand dat met Ansible Vault is versleuteld,
Terraform en OpenTofu als `-var`. Zoals het dialoogvenster waarschuwt, kan een geheim dat op deze
manier wordt doorgegeven toch in platte tekst in het log van Ansible verschijnen.

## Waarde van de geheime variabele {#varGroup-secretVarValue}

Op het tabblad **Secrets**, onder **Extra variabelen**, in elke rij.

De waarde van het geheim. Semaphore versleutelt hem, of schrijft hem in de geheimenopslag van de
groep, en toont hem nooit meer: wanneer u de groep later opent, is het veld leeg. Laat het leeg om
de opgeslagen waarde te behouden, typ een nieuwe om haar te vervangen. Het verwijderen van de rij
verwijdert het geheim wanneer u de groep opslaat.

## Naam van de geheime omgevingsvariabele {#varGroup-secretEnvName}

Op het tabblad **Secrets**, onder **Omgevingsvariabelen**, in elke rij. Voeg een rij toe met de
knop **+** van de sectie.

De naam van een omgevingsvariabele waarvan de waarde een geheim is, bijvoorbeeld `DB_PASSWORD`. Hij
wordt ingesteld in de omgeving van het proces dat de taak uitvoert, zoals de omgevingsvariabelen
van het tabblad **Variables**.

## Waarde van de geheime omgevingsvariabele {#varGroup-secretEnvValue}

Op het tabblad **Secrets**, onder **Omgevingsvariabelen**, in elke rij.

De waarde van het geheim. Net als bij een geheime extra variabele wordt ze nooit meer getoond:
laat het veld leeg om de opgeslagen waarde te behouden, typ een nieuwe om haar te vervangen.
