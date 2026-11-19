---
form: src/components/EnvironmentForm.vue, src/components/SecretStorageForm.vue, src/components/SecretStorageSyncOptionsForm.vue
skip: varGroup-name, varGroup-secretStorageId, varGroup-secretStorageKeyPrefix, varGroup-extraVarsJson, varGroup-extraVarsYaml, varGroup-varName, varGroup-varType, varGroup-varValue, varGroup-envName, varGroup-envValue, varGroup-syncEnabled, varGroup-secretVarName, varGroup-secretVarValue, varGroup-secretEnvName, varGroup-secretEnvValue, secretStorage-name, secretStorage-vaultURL, secretStorage-vaultMount, secretStorage-vaultNamespace, secretStorage-vaultToken, secretStorage-vaultTokenSource, secretStorage-dvlsInsecureTls, secretStorage-dvlsKey, secretStorage-dvlsAppKey, secretStorage-dvlsSecret, secretStorage-dvlsEnv, secretStorage-awsRegion, secretStorage-awsEndpointURL, secretStorage-awsUseIamRole, secretStorage-awsAccessKeyId, secretStorage-awsSecretKey, secretStorage-awsSecretKeySource, secretStorage-azureVaultURL, secretStorage-azureTenantId, secretStorage-azureClientId, secretStorage-azureClientSecret, secretStorage-azureClientSecretSource, secretStorage-readonly, secretStorage-syncEnabled
---

# Sync paths

De paden van een geheimenopslag waaruit Semaphore geheimen importeert, en hoe vaak het dat uit
zichzelf doet. Geopend vanuit een geheimenopslag vult de import de Sleutel Opslag van het project;
geopend vanuit een variabele groep vult ze die groep. Wat u hier wijzigt, blijft behouden wanneer
u op **Sluiten** drukt en wordt met de opslag of de groep opgeslagen, via de knop **Opslaan**
ervan.

## Auto-sync interval {#varGroupSync-syncInterval}

Wordt getoond wanneer het dialoogvenster vanuit een variabele groep wordt geopend.

Hoe vaak de groep uit zichzelf wordt geïmporteerd, in minuten. `0`, de standaard, schakelt de
automatische import uit: de groep wordt dan alleen geïmporteerd wanneer u op de
synchronisatieknop ervan drukt op de pagina Variabele Groepen.

De server controleert één keer per minuut en importeert wanneer het interval sinds de laatste
poging is verstreken, geslaagd of niet; de eerste controle nadat u een interval hebt ingesteld,
importeert direct. Een mislukte import wordt na het volgende interval opnieuw geprobeerd.

## Auto-sync interval {#secretStorageSync-syncInterval}

Wordt getoond wanneer het dialoogvenster vanuit een geheimenopslag wordt geopend.

Hoe vaak de opslag uit zichzelf wordt geïmporteerd, in minuten. `0`, de standaard, schakelt de
automatische import uit: de opslag wordt dan alleen geïmporteerd wanneer u op de
synchronisatieknop ervan drukt op het tabblad **Storages** van de Sleutel Opslag.

De server controleert één keer per minuut en importeert wanneer het interval sinds de laatste
poging is verstreken, geslaagd of niet; de eerste controle nadat u een interval hebt ingesteld,
importeert direct. Een mislukte import wordt na het volgende interval opnieuw geprobeerd.

## Pad {#secretStorageSync-path}

Wordt getoond in elke rij van de lijst. Voeg een rij toe met **Add path**.

Een pad in de opslag waaruit de geheimen worden geïmporteerd, geschreven zoals de opslag het
noemt. Elke rij is één pad; de import leest ze allemaal.

## separator {#secretStorageSync-separator}

Wordt getoond in elke rij van de lijst. Optioneel.

Een scheidingsteken dat de import van dit pad gebruikt. Laat het veld leeg, tenzij de namen in uw
opslag er een nodig hebben.

## prefix {#secretStorageSync-prefix}

Wordt getoond in elke rij van de lijst. Optioneel.

Een voorvoegsel dat de import van dit pad gebruikt. Laat het veld leeg wanneer de geïmporteerde
geheimen er geen nodig hebben.

## Pad verwijderen {#secretStorageSync-removePath}

Wordt getoond in elke rij van de lijst.

Verwijdert het pad uit de lijst. De wijziging wordt van kracht wanneer de opslag of de groep wordt
opgeslagen.

## Add path {#secretStorageSync-addPath}

Voegt een lege rij toe aan de lijst. Zolang er geen enkel pad is, meldt het dialoogvenster dat er
geen synchronisatiepaden zijn gedefinieerd, en blijft de synchronisatieknop van de opslag of de
groep uitgeschakeld.
