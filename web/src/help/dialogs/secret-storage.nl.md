---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Bewerken Secret Storage

Een geheimenopslag is een externe geheimenbeheerder waarin Semaphore geheimen bewaart en waaruit
het ze leest: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager of Azure Key
Vault. Een variabele groep die met een opslag is aangemaakt, bewaart haar geheimen daar, en een
opslag waarvoor synchronisatie is ingeschakeld, importeert zijn geheimen in de Sleutel Opslag van
het project. Het soort opslag wordt gekozen in het menu **New Storage** en kan in dit
dialoogvenster niet worden gewijzigd; de velden hieronder hangen ervan af.

## Naam {#secretStorage-name}

De naam waaronder de opslag wordt vermeld op het tabblad **Storages** van de Sleutel Opslag en in
de keuze **Secret storage** van een nieuwe variabele groep. Het veld is verplicht.

## Server URL {#secretStorage-vaultURL}

Wordt getoond wanneer de opslag een HashiCorp Vault, een OpenBao of een Devolutions Server is.

Het adres van de server, met het schema, bijvoorbeeld `https://vault.example.com:8200`. Het veld
is verplicht.

## Mount {#secretStorage-vaultMount}

Wordt getoond wanneer de opslag een HashiCorp Vault of een OpenBao is. Optioneel.

Het mountpad van de secrets engine die de geheimen bevat. Laat het veld leeg om `secret` te
gebruiken, de standaard.

## Namespace {#secretStorage-vaultNamespace}

Wordt getoond wanneer de opslag een HashiCorp Vault of een OpenBao is. Optioneel.

De namespace waarin de geheimen staan. Namespaces bestaan in Vault Enterprise en HCP Vault
Dedicated, en in OpenBao vanaf versie 2.3. Laat het veld leeg voor een server zonder namespaces.

## Bron van het geheim {#secretSource-options}

Wordt getoond voor elk soort opslag, boven het geheim waarmee Semaphore zich aanmeldt: het
**Token** van een Vault of een OpenBao, het app-secret van een Devolutions Server, de secret key
van AWS, het client secret van Azure.

Waar Semaphore dat geheim vandaan haalt:

- **Store in DB**: u typt het geheim in dit dialoogvenster; Semaphore versleutelt het en bewaart
  het in zijn database.
- **From ENV**: u typt de naam van een omgevingsvariabele van de Semaphore-server; het geheim wordt
  eruit gelezen elke keer dat de opslag wordt gebruikt.
- **From File**: u typt het absolute pad van een bestand op de Semaphore-server; het geheim wordt
  er elke keer uit gelezen, zonder het laatste regeleinde. Het bestand moet in de geheimenmap van
  de server staan (`dirs.secrets` in de configuratie, standaard `/tmp/semaphore`), en een pad met
  `..` wordt geweigerd.

Overschakelen naar een andere bron wist het veld eronder.

## Token {#secretStorage-vaultToken}

Wordt getoond wanneer de opslag een HashiCorp Vault of een OpenBao is en de bron van het geheim
**Store in DB** is.

Het token waarmee Semaphore zich bij de server authenticeert. Het moet de geheimen mogen lezen en,
tenzij de opslag **Read only** is, ze mogen schrijven en verwijderen.

Het veld is verplicht voor een nieuwe opslag. Wanneer u een bestaande opslag opent, is het leeg:
laat het leeg om het opgeslagen token te behouden, typ een nieuw token om het te vervangen.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Wordt getoond wanneer de opslag een HashiCorp Vault of een OpenBao is en de bron van het geheim
**From ENV** of **From File** is.

De naam van de omgevingsvariabele, of het pad van het bestand, die het token bevat. Het veld is
verplicht voor een nieuwe opslag.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Wordt getoond wanneer de opslag een Devolutions Server is.

Semaphore maakt verbinding met de server zonder het TLS-certificaat ervan te controleren. Iedereen
tussen Semaphore en de server kan dan de geheimen lezen. Laat dit uit, en geef de server in plaats
daarvan een certificaat dat Semaphore vertrouwt.

## Vault ID {#secretStorage-dvlsKey}

Wordt getoond wanneer de opslag een Devolutions Server is.

De ID van de vault van Devolutions Server die de geheimen bevat. Het veld is verplicht voor een
nieuwe opslag.

## App Key {#secretStorage-dvlsAppKey}

Wordt getoond wanneer de opslag een Devolutions Server is.

De applicatiesleutel van de applicatie-identiteit waarmee Semaphore zich aanmeldt, die waarvan het
geheim in **Secret** komt. Het veld is verplicht voor een nieuwe opslag.

## Secret {#secretStorage-dvlsSecret}

Wordt getoond wanneer de opslag een Devolutions Server is en de bron van het geheim
**Store in DB** is.

Het app-secret van de applicatie-identiteit van **App Key**. Het veld is verplicht voor een nieuwe
opslag. Wanneer u een bestaande opslag opent, is het leeg: laat het leeg om het opgeslagen geheim
te behouden, typ een nieuw geheim om het te vervangen.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Wordt getoond wanneer de opslag een Devolutions Server is en de bron van het geheim **From ENV**
of **From File** is.

De naam van de omgevingsvariabele, of het pad van het bestand, die het app-secret bevat. Het veld
is verplicht voor een nieuwe opslag.

## Region {#secretStorage-awsRegion}

Wordt getoond wanneer de opslag een AWS Secrets Manager is.

De AWS-regio waarin de geheimen staan, bijvoorbeeld `us-east-1`. Het veld is verplicht.

## Endpoint URL {#secretStorage-awsEndpointURL}

Wordt getoond wanneer de opslag een AWS Secrets Manager is. Optioneel.

Het adres van de Secrets Manager-API dat in plaats van het standaard-endpoint van de regio wordt
aangeroepen, bijvoorbeeld een VPC-endpoint. Laat het veld leeg om het standaard-endpoint van AWS
te gebruiken.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Wordt getoond wanneer de opslag een AWS Secrets Manager is.

Wanneer dit aan staat, bewaart Semaphore geen toegangssleutel en haalt het de AWS-referenties uit
de omgeving waarin de server draait, zoals de IAM-rol van zijn instance. De velden van de
toegangssleutel hieronder zijn dan uitgeschakeld.

Het inschakelen verwijdert bij het opslaan de opgeslagen toegangssleutel. Om het weer uit te
schakelen is een nieuwe **Secret Access Key** nodig.

## Access Key ID {#secretStorage-awsAccessKeyId}

Wordt getoond wanneer de opslag een AWS Secrets Manager is. Uitgeschakeld wanneer
**Use IAM Role / Instance Profile** aan staat.

De ID van de toegangssleutel waarmee Semaphore zich aanmeldt, bijvoorbeeld
`AKIAIOSFODNN7EXAMPLE`. Het veld is verplicht wanneer de IAM-rol uit staat. De ID is geen geheim:
hij wordt opgeslagen en getoond zoals hij is.

## Secret Access Key {#secretStorage-awsSecretKey}

Wordt getoond wanneer de opslag een AWS Secrets Manager is en de bron van het geheim
**Store in DB** is. Uitgeschakeld wanneer **Use IAM Role / Instance Profile** aan staat.

Het geheim van de toegangssleutel. Het veld is verplicht voor een nieuwe opslag en nadat de
IAM-rol is uitgeschakeld. Anders is het leeg wanneer u de opslag opent: laat het leeg om de
opgeslagen sleutel te behouden, typ een nieuwe om hem te vervangen.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Wordt getoond wanneer de opslag een AWS Secrets Manager is en de bron van het geheim **From ENV**
of **From File** is. Uitgeschakeld wanneer **Use IAM Role / Instance Profile** aan staat.

De naam van de omgevingsvariabele, of het pad van het bestand, die de secret access key bevat. Het
veld is verplicht in dezelfde gevallen als **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Wordt getoond wanneer de opslag een Azure Key Vault is.

Het adres van de key vault, bijvoorbeeld `https://my-vault.vault.azure.net`. Het veld is
verplicht.

## Tenant ID {#secretStorage-azureTenantId}

Wordt getoond wanneer de opslag een Azure Key Vault is.

De ID van de Microsoft Entra-tenant van de applicatie als welke Semaphore zich aanmeldt. Het veld
is verplicht.

## Client ID {#secretStorage-azureClientId}

Wordt getoond wanneer de opslag een Azure Key Vault is.

De applicatie-ID (client-ID) van die applicatie. Ze heeft toegang nodig tot de geheimen van de key
vault. Het veld is verplicht.

## Client Secret {#secretStorage-azureClientSecret}

Wordt getoond wanneer de opslag een Azure Key Vault is en de bron van het geheim **Store in DB**
is.

Het client secret van de applicatie. Het veld is verplicht voor een nieuwe opslag. Wanneer u een
bestaande opslag opent, is het leeg: laat het leeg om het opgeslagen geheim te behouden, typ een
nieuw geheim om het te vervangen.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Wordt getoond wanneer de opslag een Azure Key Vault is en de bron van het geheim **From ENV** of
**From File** is.

De naam van de omgevingsvariabele, of het pad van het bestand, die het client secret bevat. Het
veld is verplicht voor een nieuwe opslag.

## Read only {#secretStorage-readonly}

Wanneer dit aan staat, leest Semaphore alleen uit de opslag. Een geheim van een variabele groep kan
er niet in worden aangemaakt of gewijzigd: het opslaan ervan mislukt. Het verwijderen van een
geheim, een sleutel of een variabele groep in Semaphore laat het geheim ervan in de opslag staan.

Wanneer dit uit staat, schrijft Semaphore de geheimen van de variabele groepen die de opslag
gebruiken erin weg en verwijdert ze eruit samen met de groepen.

## Sync keys enabled {#secretStorage-syncEnabled}

Schakel dit in om geheimen uit de opslag te importeren in de Sleutel Opslag van het project, vanaf
de paden die onder **Sync paths** zijn ingesteld. De import draait wanneer u op de
synchronisatieknop van de opslag op het tabblad **Storages** drukt, en vanzelf met het interval
dat in **Sync paths** is ingesteld.

Een geïmporteerde sleutel wordt als gesynchroniseerd gemarkeerd: zijn naam en type kunnen niet
worden gewijzigd, en het verwijderen ervan in Semaphore laat het geheim in de opslag staan. Het
verwijderen van de opslag terwijl synchronisatie aan staat, verwijdert ook de sleutels die ze
heeft geïmporteerd.

## Sync paths {#secretStorage-syncPaths}

Wordt getoond wanneer **Sync keys enabled** aan staat.

Opent het dialoogvenster met de paden van de opslag waaruit wordt geïmporteerd en het interval van
de automatische import. Het getal op de knop is het aantal paden. De synchronisatieknop van de
opslag blijft uitgeschakeld totdat er minstens één pad is.

Het dialoogvenster heeft alleen **Sluiten**: wat u erin wijzigt, wordt met de opslag opgeslagen,
via **Opslaan** van dit dialoogvenster (**Aanmaken** voor een nieuwe opslag).
