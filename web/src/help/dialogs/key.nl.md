---
form: src/components/KeyForm.vue
---

# Bewerken Key

Een sleutel in de Sleutel Opslag bevat een referentie die het project gebruikt zonder haar te
tonen: de SSH-sleutel of het wachtwoord waarmee Ansible op de hosts inlogt, de referenties van een
Git-repository, het wachtwoord van een Ansible-vault, het geheim van een integratie.
Inventarissen, repositories, taaksjablonen en integraties verwijzen met de naam naar een sleutel.
Na het opslaan wordt het geheim nooit meer naar de browser teruggestuurd.

## Sleutelnaam {#key-name}

De naam waaronder de sleutel wordt vermeld in de Sleutel Opslag en in de sleutelkeuzes van
inventarissen, repositories, taaksjablonen en integraties. Het veld is verplicht. Kies een naam die
zegt wat de sleutel opent, bijvoorbeeld `Deploy key of the web servers`.

## Bron {#key-source}

Waar het geheim van de sleutel wordt bewaard. De tabbladen:

- **Local**: het geheim wordt in dit dialoogvenster getypt en in de database van Semaphore
  opgeslagen, versleuteld wanneer de beheerder een versleutelingssleutel voor toegangssleutels
  heeft ingesteld. Dit is de standaard.
- **Storage**: het geheim wordt bewaard in een externe geheimenopslag van het project, gekozen in
  **Storage** hieronder. Het tabblad is er alleen in de Pro-editie. De opslagen zelf worden
  ingesteld op het tabblad **Storages** van de Sleutel Opslag.
- **Env**: het geheim wordt gelezen uit een omgevingsvariabele van het serverproces van Semaphore.
  Typ de naam van de variabele in het veld onder de tabbladen.
- **File**: het geheim wordt gelezen uit een bestand op de Semaphore-server. Typ het absolute pad
  van het bestand in het veld onder de tabbladen. Het bestand moet in de geheimenmap van de server
  staan (`dirs.secrets` in de configuratie, standaard `/tmp/semaphore`); een pad daarbuiten, of
  een pad met `..` erin, mislukt wanneer de sleutel wordt gebruikt. Een regeleinde aan het eind
  van het bestand wordt weggelaten.

Met **Env** en **File** schrijft Semaphore het geheim nooit: het leest de waarde elke keer dat een
taak de sleutel gebruikt, dus u wijzigt het geheim op de plek waar het staat. De waarde moet passen
bij het **Type** van de sleutel. Een sleutel van het type **Tekenreeks** neemt de waarde zoals ze
is. Een sleutel van het type **SSH Sleutel** of **Inloggen met wachtwoord** verwacht JSON met de
velden van dat type, bijvoorbeeld:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Wanneer u een sleutel bewerkt, zijn de tabbladen vergrendeld totdat **Overschrijven** is
aangevinkt. Een sleutel die uit een geheimenopslag is gesynchroniseerd, behoudt zijn bron, en zijn
naam en type kunnen niet worden gewijzigd.

## Storage {#key-sourceStorageId}

Wordt getoond wanneer het tabblad **Storage** is gekozen.

De geheimenopslag van het project die het geheim bevat: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager of Azure Key Vault, zoals ingesteld op het tabblad **Storages** van de
Sleutel Opslag.

- Een opslag die niet alleen-lezen is, ontvangt het geheim dat u in dit dialoogvenster typt:
  Semaphore schrijft het daar weg wanneer u opslaat. Het verwijderen van de sleutel verwijdert ook
  het geheim uit de opslag. Een sleutel die in zo'n opslag wordt bewaard, kan later niet naar een
  andere opslag worden verplaatst.
- Een alleen-lezen opslag wordt alleen gelezen. Het dialoogvenster toont dan geen velden voor het
  geheim, en Semaphore leest het uit de opslag elke keer dat een taak de sleutel gebruikt.

## Source Key {#key-sourceStorageKey}

Wordt getoond wanneer een **Storage** is geselecteerd.

Waar het geheim in die opslag staat, bijvoorbeeld het pad naar het geheim in HashiCorp Vault.

## Type {#key-type}

Wat voor referentie de sleutel bevat. Het veld is verplicht, en de velden eronder veranderen mee
met het type.

- **SSH Sleutel**: een privésleutel, met een optionele login en passphrase. Ansible maakt er
  verbinding mee met de hosts, en Git kloont er een repository met een `ssh`-URL mee.
- **Inloggen met wachtwoord**: een login en een wachtwoord. Ansible maakt er verbinding mee met de
  hosts of wordt er een andere gebruiker mee (**Sudo Referenties** van een inventaris); Git kloont
  er een `https`-repository mee; het wachtwoord opent een Ansible-vault van een taaksjabloon.
- **Tekenreeks**: één geheime waarde zonder login: een token, een API-sleutel, een wachtwoord.
  Gebruik dit voor het geheim van een integratie. Het kan geen SSH-sleutel, sudo-referenties, het
  wachtwoord van een vault of de toegangssleutel van een repository zijn.
- **Geen**: geen referentie. Gebruik dit voor een repository die geen login nodig heeft (een
  openbare `https`-repository, een `file`-URL of een lokaal pad) en voor een inventaris waarvan de
  hosts geen referenties nodig hebben of de verbinding zelf definiëren.

## Gebruikersnaam {#key-loginPasswordLogin}

Wordt getoond wanneer het type **Inloggen met wachtwoord** is. Optioneel.

De gebruikersnaam van de referentie.

- Ansible maakt als deze gebruiker verbinding met de hosts (`--user`); als **Sudo Referenties**
  van een inventaris is het de gebruiker die Ansible wordt (`--become-user`). Wanneer het veld
  leeg is, gebruikt Ansible de gebruiker die de inventaris of de eigen configuratie van Ansible
  noemt.
- Voor een `https`-repository logt Git in met deze gebruiker en het wachtwoord. Laat het veld leeg
  voor een token dat de Git-server zonder gebruiker accepteert: het wachtwoord wordt dan als
  gebruikersnaam verstuurd.

## Wachtwoord {#key-loginPasswordPassword}

Wordt getoond wanneer het type **Inloggen met wachtwoord** is. Verplicht.

Het wachtwoord van de referentie. Semaphore beantwoordt er de wachtwoordprompts van Ansible mee:
het SSH-wachtwoord (`--ask-pass`) voor gebruikersreferenties, het `become`-wachtwoord
(`--ask-become-pass`) voor sudo-referenties, en het vault-wachtwoord voor een Ansible-vault. Voor
een `https`-repository is het het wachtwoord of het toegangstoken van de Git-server.

## Waarde {#key-stringValue}

Wordt getoond wanneer het type **Tekenreeks** is. Verplicht.

De geheime waarde zelf, bijvoorbeeld het token waartegen een integratie de verzoeken controleert
die ze ontvangt. De waarde wordt opgeslagen zoals ze is getypt.

## Gebruikersnaam {#key-sshLogin}

Wordt getoond wanneer het type **SSH Sleutel** is. Optioneel.

De gebruiker als wie Ansible verbinding maakt met de hosts (`--user`). Wanneer het veld leeg is,
gebruikt Ansible de gebruiker die de inventaris of de eigen configuratie van Ansible noemt.

Zet voor een Git-repository de gebruiker in plaats daarvan in de URL van de repository, zoals in
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Wordt getoond wanneer het type **SSH Sleutel** is. Optioneel.

De passphrase waarmee de privésleutel is versleuteld. Laat het veld leeg voor een sleutel zonder
passphrase. Een verkeerde passphrase laat elke taak die de sleutel gebruikt mislukken wanneer de
sleutel wordt geladen. Een sleutel die door Semaphore is gegenereerd, heeft geen passphrase: wat
dit veld dan bevat, wordt weggegooid.

## Generate SSH Key {#key-generateSshKey}

Wordt getoond wanneer het type **SSH Sleutel** is.

Semaphore maakt bij het opslaan een nieuw sleutelpaar aan: een 2048-bits RSA-privésleutel, die in
de sleutel blijft, en de bijbehorende publieke sleutel, die na het opslaan één keer in een
dialoogvenster wordt getoond en daarna onder **Public key** telkens wanneer u de sleutel bewerkt.
Voeg de publieke sleutel toe aan `~/.ssh/authorized_keys` op de hosts, of als deploy key van de
Git-repository.

Het veld van de privésleutel is vergrendeld zolang het vakje is aangevinkt. Vink bij een bestaande
sleutel eerst **Overschrijven** aan: het oude sleutelpaar wordt vervangen. Een sleutel waarvan het
geheim uit **Env**, **File** of een alleen-lezen opslag wordt gelezen, kan niet worden gegenereerd,
omdat Semaphore nergens heeft om hem weg te schrijven.

## Privésleutel {#key-sshPrivateKey}

Wordt getoond wanneer het type **SSH Sleutel** is. Verplicht, tenzij **Generate SSH Key** is
aangevinkt.

De privésleutel, in het OpenSSH- of PEM-formaat, inclusief de regels `BEGIN` en `END`. Voor een
taak laadt Semaphore hem in een eigen SSH-agent, die Ansible en Git gebruiken zolang de taak duurt.

## Public key {#key-publicKey}

Wordt getoond wanneer u een SSH-sleutel bewerkt die Semaphore heeft gegenereerd, zolang
**Overschrijven** niet is aangevinkt.

De publieke helft van het gegenereerde sleutelpaar. **Show** klapt hem uit, de kopieerknop zet hem
op het klembord. Voeg hem toe aan `~/.ssh/authorized_keys` op de hosts, of als deploy key van de
Git-repository.

Het is de publieke sleutel van het paar dat Semaphore het laatst heeft gegenereerd. Nadat u de
privésleutel met de hand hebt vervangen met **Overschrijven**, past de sleutel die hier wordt
getoond er niet meer bij.

## Overschrijven {#key-overrideSecret}

Wordt getoond wanneer u een bestaande sleutel bewerkt.

Semaphore toont een opgeslagen geheim nooit meer, dus de velden van het geheim zijn leeg en
vergrendeld. Vink het vakje aan om het geheim te vervangen: de bron, het type en de velden van het
geheim gaan open, en alles daarin wordt opnieuw opgeslagen, dus voer het hele geheim opnieuw in,
niet alleen het deel dat verandert. Zonder dit vakje wijzigt opslaan alleen de naam van de
sleutel.

Het opslaan van een sleutel verwijdert ook de kopieën die de Semaphore-server bewaart van de
repositories die de sleutel gebruiken: de volgende taak kloont ze opnieuw.
