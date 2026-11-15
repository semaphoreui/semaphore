---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Modifica Secret Storage

Un archivio di segreti è un gestore di segreti esterno in cui Semaphore conserva i segreti e da cui
li legge: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager o Azure Key Vault. Un
gruppo di variabili creato con un archivio vi conserva i suoi segreti, e un archivio con la
sincronizzazione attiva importa i suoi segreti nel Negozio chiavi del progetto. Il tipo di archivio
si sceglie nel menu **New Storage** e non può essere cambiato in questa finestra di dialogo; i campi
qui sotto dipendono da esso.

## Nome {#secretStorage-name}

Il nome con cui l'archivio compare nella scheda **Storages** del Negozio chiavi e nel selettore
**Secret storage** di un nuovo gruppo di variabili. Il campo è obbligatorio.

## Server URL {#secretStorage-vaultURL}

Visibile quando l'archivio è un HashiCorp Vault, un OpenBao o un Devolutions Server.

L'indirizzo del server, con il suo schema, ad esempio `https://vault.example.com:8200`. Il campo è
obbligatorio.

## Mount {#secretStorage-vaultMount}

Visibile quando l'archivio è un HashiCorp Vault o un OpenBao. Opzionale.

Il percorso di mount del motore dei segreti che contiene i segreti. Lascia il campo vuoto per usare
`secret`, il valore predefinito.

## Namespace {#secretStorage-vaultNamespace}

Visibile quando l'archivio è un HashiCorp Vault o un OpenBao. Opzionale.

Il namespace in cui si trovano i segreti. I namespace esistono in Vault Enterprise e HCP Vault
Dedicated, e in OpenBao dalla versione 2.3. Lascia il campo vuoto per un server senza namespace.

## Secret source {#secretSource-options}

Visibile per ogni tipo di archivio, sopra il segreto con cui Semaphore si autentica: il **Token** di
un Vault o di un OpenBao, l'app secret di un Devolutions Server, la secret key di AWS, il client
secret di Azure.

Da dove Semaphore prende quel segreto:

- **Store in DB**: digiti il segreto in questa finestra di dialogo; Semaphore lo cifra e lo conserva
  nel suo database.
- **From ENV**: digiti il nome di una variabile di ambiente del server Semaphore; il segreto viene
  letto da essa ogni volta che l'archivio viene usato.
- **From File**: digiti il percorso assoluto di un file sul server Semaphore; il segreto viene letto
  da esso ogni volta, senza l'interruzione di riga finale. Il file deve trovarsi all'interno della
  directory dei segreti del server (`dirs.secrets` nella configurazione, `/tmp/semaphore` per
  impostazione predefinita), e un percorso con `..` viene rifiutato.

Cambiare l'origine svuota il campo sottostante.

## Token {#secretStorage-vaultToken}

Visibile quando l'archivio è un HashiCorp Vault o un OpenBao e l'origine del segreto è
**Store in DB**.

Il token con cui Semaphore si autentica sul server. Deve poter leggere i segreti e, a meno che
l'archivio non sia **Read only**, scriverli ed eliminarli.

Il campo è obbligatorio per un nuovo archivio. Quando apri un archivio esistente, è vuoto: lascialo
vuoto per mantenere il token memorizzato, digitane uno nuovo per sostituirlo.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Visibile quando l'archivio è un HashiCorp Vault o un OpenBao e l'origine del segreto è
**From ENV** o **From File**.

Il nome della variabile di ambiente, o il percorso del file, che contiene il token. Il campo è
obbligatorio per un nuovo archivio.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Visibile quando l'archivio è un Devolutions Server.

Semaphore si connette al server senza verificarne il certificato TLS. Chiunque si trovi tra
Semaphore e il server può allora leggere i segreti. Lascialo disattivato, e dai invece al server un
certificato di cui Semaphore si fida.

## Vault ID {#secretStorage-dvlsKey}

Visibile quando l'archivio è un Devolutions Server.

L'ID del vault di Devolutions Server che contiene i segreti. Il campo è obbligatorio per un nuovo
archivio.

## App Key {#secretStorage-dvlsAppKey}

Visibile quando l'archivio è un Devolutions Server.

La chiave dell'applicazione dell'identità applicativa con cui Semaphore si autentica, quella il cui
segreto va in **Secret**. Il campo è obbligatorio per un nuovo archivio.

## Secret {#secretStorage-dvlsSecret}

Visibile quando l'archivio è un Devolutions Server e l'origine del segreto è **Store in DB**.

L'app secret dell'identità applicativa di **App Key**. Il campo è obbligatorio per un nuovo
archivio. Quando apri un archivio esistente, è vuoto: lascialo vuoto per mantenere il segreto
memorizzato, digitane uno nuovo per sostituirlo.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Visibile quando l'archivio è un Devolutions Server e l'origine del segreto è **From ENV** o
**From File**.

Il nome della variabile di ambiente, o il percorso del file, che contiene l'app secret. Il campo è
obbligatorio per un nuovo archivio.

## Region {#secretStorage-awsRegion}

Visibile quando l'archivio è un AWS Secrets Manager.

La regione AWS in cui si trovano i segreti, ad esempio `us-east-1`. Il campo è obbligatorio.

## Endpoint URL {#secretStorage-awsEndpointURL}

Visibile quando l'archivio è un AWS Secrets Manager. Opzionale.

L'indirizzo dell'API di Secrets Manager da chiamare al posto dell'endpoint predefinito della
regione, ad esempio un endpoint VPC. Lascia il campo vuoto per usare l'endpoint AWS predefinito.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Visibile quando l'archivio è un AWS Secrets Manager.

Quando è attivo, Semaphore non memorizza alcuna chiave di accesso e prende le credenziali AWS
dall'ambiente in cui gira il server, come il ruolo IAM della sua istanza. I campi della chiave di
accesso qui sotto sono allora disattivati.

Attivarlo rimuove la chiave di accesso memorizzata quando salvi. Disattivarlo di nuovo richiede una
nuova **Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Visibile quando l'archivio è un AWS Secrets Manager. Disattivato quando
**Use IAM Role / Instance Profile** è attivo.

L'ID della chiave di accesso con cui Semaphore si autentica, ad esempio `AKIAIOSFODNN7EXAMPLE`. Il
campo è obbligatorio quando il ruolo IAM è disattivato. L'ID non è un segreto: viene memorizzato e
mostrato così com'è.

## Secret Access Key {#secretStorage-awsSecretKey}

Visibile quando l'archivio è un AWS Secrets Manager e l'origine del segreto è **Store in DB**.
Disattivato quando **Use IAM Role / Instance Profile** è attivo.

Il segreto della chiave di accesso. Il campo è obbligatorio per un nuovo archivio e dopo che il
ruolo IAM è stato disattivato. Altrimenti è vuoto quando apri l'archivio: lascialo vuoto per
mantenere la chiave memorizzata, digitane una nuova per sostituirla.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Visibile quando l'archivio è un AWS Secrets Manager e l'origine del segreto è **From ENV** o
**From File**. Disattivato quando **Use IAM Role / Instance Profile** è attivo.

Il nome della variabile di ambiente, o il percorso del file, che contiene la secret access key. Il
campo è obbligatorio negli stessi casi di **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Visibile quando l'archivio è un Azure Key Vault.

L'indirizzo del key vault, ad esempio `https://my-vault.vault.azure.net`. Il campo è obbligatorio.

## Tenant ID {#secretStorage-azureTenantId}

Visibile quando l'archivio è un Azure Key Vault.

L'ID del tenant Microsoft Entra dell'applicazione con cui Semaphore si autentica. Il campo è
obbligatorio.

## Client ID {#secretStorage-azureClientId}

Visibile quando l'archivio è un Azure Key Vault.

L'ID applicazione (client) di quell'applicazione. Deve avere accesso ai segreti del key vault. Il
campo è obbligatorio.

## Client Secret {#secretStorage-azureClientSecret}

Visibile quando l'archivio è un Azure Key Vault e l'origine del segreto è **Store in DB**.

Il client secret dell'applicazione. Il campo è obbligatorio per un nuovo archivio. Quando apri un
archivio esistente, è vuoto: lascialo vuoto per mantenere il segreto memorizzato, digitane uno nuovo
per sostituirlo.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Visibile quando l'archivio è un Azure Key Vault e l'origine del segreto è **From ENV** o
**From File**.

Il nome della variabile di ambiente, o il percorso del file, che contiene il client secret. Il campo
è obbligatorio per un nuovo archivio.

## Read only {#secretStorage-readonly}

Quando è attivo, Semaphore legge soltanto dall'archivio. Un segreto di un gruppo di variabili non
può esservi creato o modificato: salvarlo fallisce. Eliminare un segreto, una chiave o un gruppo di
variabili in Semaphore lascia il suo segreto nell'archivio.

Quando è disattivato, Semaphore vi scrive i segreti dei gruppi di variabili che usano l'archivio e
li elimina da esso insieme ai gruppi.

## Sync keys enabled {#secretStorage-syncEnabled}

Attivalo per importare segreti dall'archivio nel Negozio chiavi del progetto, dai percorsi impostati
in **Sync paths**. L'importazione viene eseguita quando premi il pulsante di sincronizzazione
dell'archivio nella scheda **Storages**, e da sola all'intervallo impostato in **Sync paths**.

Una chiave importata è contrassegnata come sincronizzata: il suo nome e il suo tipo non possono
essere cambiati, ed eliminarla in Semaphore lascia il segreto nell'archivio. Eliminare l'archivio
mentre la sincronizzazione è attiva elimina anche le chiavi che ha importato.

## Sync paths {#secretStorage-syncPaths}

Visibile quando **Sync keys enabled** è attivo.

Apre la finestra di dialogo con i percorsi dell'archivio da cui importare e l'intervallo
dell'importazione automatica. Il numero sul pulsante è il numero di percorsi. Il pulsante di
sincronizzazione dell'archivio resta disattivato finché non ha almeno un percorso.

La finestra di dialogo ha solo **Chiudi**: ciò che vi modifichi viene salvato con l'archivio, da
**Salva** di questa finestra di dialogo (**Crea** per un nuovo archivio).
