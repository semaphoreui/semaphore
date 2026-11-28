---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Bearbeiten Secret Storage

Ein Secret-Speicher ist ein externer Secret-Manager, in dem Semaphore Secrets aufbewahrt und aus
dem es sie liest: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager oder Azure
Key Vault. Eine mit einem Speicher erstellte Variablengruppe bewahrt ihre Secrets dort auf, und
ein Speicher mit eingeschalteter Synchronisierung importiert seine Secrets in den
Schlüsselspeicher des Projekts. Die Art des Speichers wird im Menü **New Storage** gewählt und
kann in diesem Dialog nicht geändert werden; die Felder unten hängen von ihr ab.

## Name {#secretStorage-name}

Der Name, unter dem der Speicher auf dem Tab **Storages** des Schlüsselspeichers und in der
Auswahl **Secret storage** einer neuen Variablengruppe aufgeführt wird. Das Feld ist ein
Pflichtfeld.

## Server URL {#secretStorage-vaultURL}

Wird angezeigt, wenn der Speicher ein HashiCorp Vault, ein OpenBao oder ein Devolutions Server
ist.

Die Adresse des Servers mit ihrem Schema, zum Beispiel `https://vault.example.com:8200`. Das Feld
ist ein Pflichtfeld.

## Mount {#secretStorage-vaultMount}

Wird angezeigt, wenn der Speicher ein HashiCorp Vault oder ein OpenBao ist. Optional.

Der Mount-Pfad der Secrets Engine, die die Secrets enthält. Lassen Sie das Feld leer, um
`secret` zu verwenden, die Voreinstellung.

## Namespace {#secretStorage-vaultNamespace}

Wird angezeigt, wenn der Speicher ein HashiCorp Vault oder ein OpenBao ist. Optional.

Der Namespace, in dem die Secrets liegen. Namespaces gibt es in Vault Enterprise und HCP Vault
Dedicated sowie in OpenBao ab Version 2.3. Lassen Sie das Feld leer für einen Server ohne
Namespaces.

## Secret-Quelle {#secretSource-options}

Wird bei jeder Art von Speicher angezeigt, über dem Secret, mit dem Semaphore sich anmeldet: dem
**Token** eines Vault oder eines OpenBao, dem App-Secret eines Devolutions Server, dem Secret Key
von AWS, dem Client Secret von Azure.

Woher Semaphore dieses Secret nimmt:

- **Store in DB**: Sie geben das Secret in diesen Dialog ein; Semaphore verschlüsselt es und
  bewahrt es in seiner Datenbank auf.
- **From ENV**: Sie geben den Namen einer Umgebungsvariablen des Semaphore-Servers ein; das
  Secret wird bei jeder Verwendung des Speichers daraus gelesen.
- **From File**: Sie geben den absoluten Pfad einer Datei auf dem Semaphore-Server ein; das
  Secret wird jedes Mal daraus gelesen, ohne den abschließenden Zeilenumbruch. Die Datei muss im
  Secrets-Verzeichnis des Servers liegen (`dirs.secrets` in der Konfiguration, standardmäßig
  `/tmp/semaphore`), und ein Pfad mit `..` wird abgelehnt.

Der Wechsel der Quelle leert das Feld darunter.

## Token {#secretStorage-vaultToken}

Wird angezeigt, wenn der Speicher ein HashiCorp Vault oder ein OpenBao ist und die Secret-Quelle
**Store in DB** ist.

Das Token, mit dem Semaphore sich am Server authentifiziert. Es muss die Secrets lesen dürfen
und, sofern der Speicher nicht **Read only** ist, sie schreiben und löschen dürfen.

Das Feld ist für einen neuen Speicher ein Pflichtfeld. Wenn Sie einen bestehenden Speicher
öffnen, ist es leer: Lassen Sie es leer, um das gespeicherte Token zu behalten, oder geben Sie
ein neues ein, um es zu ersetzen.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Wird angezeigt, wenn der Speicher ein HashiCorp Vault oder ein OpenBao ist und die Secret-Quelle
**From ENV** oder **From File** ist.

Der Name der Umgebungsvariablen oder der Pfad der Datei, die das Token enthält. Das Feld ist für
einen neuen Speicher ein Pflichtfeld.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Wird angezeigt, wenn der Speicher ein Devolutions Server ist.

Semaphore verbindet sich mit dem Server, ohne sein TLS-Zertifikat zu prüfen. Jeder zwischen
Semaphore und dem Server kann dann die Secrets lesen. Lassen Sie die Option aus und geben Sie dem
Server stattdessen ein Zertifikat, dem Semaphore vertraut.

## Vault ID {#secretStorage-dvlsKey}

Wird angezeigt, wenn der Speicher ein Devolutions Server ist.

Die ID des Vaults von Devolutions Server, der die Secrets enthält. Das Feld ist für einen neuen
Speicher ein Pflichtfeld.

## App Key {#secretStorage-dvlsAppKey}

Wird angezeigt, wenn der Speicher ein Devolutions Server ist.

Der Anwendungsschlüssel der Anwendungsidentität, mit der Semaphore sich anmeldet, derjenigen,
deren Secret in **Secret** eingetragen wird. Das Feld ist für einen neuen Speicher ein
Pflichtfeld.

## Secret {#secretStorage-dvlsSecret}

Wird angezeigt, wenn der Speicher ein Devolutions Server ist und die Secret-Quelle
**Store in DB** ist.

Das App-Secret der Anwendungsidentität aus **App Key**. Das Feld ist für einen neuen Speicher ein
Pflichtfeld. Wenn Sie einen bestehenden Speicher öffnen, ist es leer: Lassen Sie es leer, um das
gespeicherte Secret zu behalten, oder geben Sie ein neues ein, um es zu ersetzen.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Wird angezeigt, wenn der Speicher ein Devolutions Server ist und die Secret-Quelle **From ENV**
oder **From File** ist.

Der Name der Umgebungsvariablen oder der Pfad der Datei, die das App-Secret enthält. Das Feld
ist für einen neuen Speicher ein Pflichtfeld.

## Region {#secretStorage-awsRegion}

Wird angezeigt, wenn der Speicher ein AWS Secrets Manager ist.

Die AWS-Region, in der die Secrets liegen, zum Beispiel `us-east-1`. Das Feld ist ein
Pflichtfeld.

## Endpoint URL {#secretStorage-awsEndpointURL}

Wird angezeigt, wenn der Speicher ein AWS Secrets Manager ist. Optional.

Die Adresse der Secrets-Manager-API, die statt des Standard-Endpoints der Region aufgerufen wird,
zum Beispiel ein VPC-Endpoint. Lassen Sie das Feld leer, um den Standard-Endpoint von AWS zu
verwenden.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Wird angezeigt, wenn der Speicher ein AWS Secrets Manager ist.

Ist die Option eingeschaltet, speichert Semaphore keinen Access Key und nimmt die
AWS-Credentials aus der Umgebung, in der der Server läuft, etwa der IAM-Rolle seiner Instanz. Die
Felder des Access Keys darunter sind dann deaktiviert.

Das Einschalten entfernt beim Speichern den gespeicherten Access Key. Das erneute Ausschalten
erfordert einen neuen **Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Wird angezeigt, wenn der Speicher ein AWS Secrets Manager ist. Deaktiviert, wenn
**Use IAM Role / Instance Profile** eingeschaltet ist.

Die ID des Access Keys, mit dem Semaphore sich anmeldet, zum Beispiel `AKIAIOSFODNN7EXAMPLE`.
Das Feld ist ein Pflichtfeld, wenn die IAM-Rolle ausgeschaltet ist. Die ID ist kein Secret: Sie
wird so gespeichert und angezeigt, wie sie ist.

## Secret Access Key {#secretStorage-awsSecretKey}

Wird angezeigt, wenn der Speicher ein AWS Secrets Manager ist und die Secret-Quelle
**Store in DB** ist. Deaktiviert, wenn **Use IAM Role / Instance Profile** eingeschaltet ist.

Das Secret des Access Keys. Das Feld ist für einen neuen Speicher und nach dem Ausschalten der
IAM-Rolle ein Pflichtfeld. Andernfalls ist es leer, wenn Sie den Speicher öffnen: Lassen Sie es
leer, um den gespeicherten Schlüssel zu behalten, oder geben Sie einen neuen ein, um ihn zu
ersetzen.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Wird angezeigt, wenn der Speicher ein AWS Secrets Manager ist und die Secret-Quelle **From ENV**
oder **From File** ist. Deaktiviert, wenn **Use IAM Role / Instance Profile** eingeschaltet ist.

Der Name der Umgebungsvariablen oder der Pfad der Datei, die den Secret Access Key enthält. Das
Feld ist in denselben Fällen ein Pflichtfeld wie **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Wird angezeigt, wenn der Speicher ein Azure Key Vault ist.

Die Adresse des Key Vaults, zum Beispiel `https://my-vault.vault.azure.net`. Das Feld ist ein
Pflichtfeld.

## Tenant ID {#secretStorage-azureTenantId}

Wird angezeigt, wenn der Speicher ein Azure Key Vault ist.

Die ID des Microsoft-Entra-Mandanten der Anwendung, als die Semaphore sich anmeldet. Das Feld ist
ein Pflichtfeld.

## Client ID {#secretStorage-azureClientId}

Wird angezeigt, wenn der Speicher ein Azure Key Vault ist.

Die Anwendungs-ID (Client-ID) dieser Anwendung. Sie braucht Zugriff auf die Secrets des Key
Vaults. Das Feld ist ein Pflichtfeld.

## Client Secret {#secretStorage-azureClientSecret}

Wird angezeigt, wenn der Speicher ein Azure Key Vault ist und die Secret-Quelle **Store in DB**
ist.

Das Client Secret der Anwendung. Das Feld ist für einen neuen Speicher ein Pflichtfeld. Wenn Sie
einen bestehenden Speicher öffnen, ist es leer: Lassen Sie es leer, um das gespeicherte Secret zu
behalten, oder geben Sie ein neues ein, um es zu ersetzen.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Wird angezeigt, wenn der Speicher ein Azure Key Vault ist und die Secret-Quelle **From ENV** oder
**From File** ist.

Der Name der Umgebungsvariablen oder der Pfad der Datei, die das Client Secret enthält. Das Feld
ist für einen neuen Speicher ein Pflichtfeld.

## Read only {#secretStorage-readonly}

Ist die Option eingeschaltet, liest Semaphore nur aus dem Speicher. Ein Secret einer
Variablengruppe kann darin weder erstellt noch geändert werden: Das Speichern schlägt fehl. Das
Löschen eines Secrets, eines Schlüssels oder einer Variablengruppe in Semaphore lässt das Secret
im Speicher.

Ist sie ausgeschaltet, schreibt Semaphore die Secrets der Variablengruppen, die den Speicher
verwenden, hinein und löscht sie zusammen mit den Gruppen daraus.

## Sync keys enabled {#secretStorage-syncEnabled}

Schalten Sie die Option ein, um Secrets aus dem Speicher in den Schlüsselspeicher des Projekts zu
importieren, aus den unter **Sync paths** festgelegten Pfaden. Der Import läuft, wenn Sie auf dem
Tab **Storages** auf die Synchronisierungsschaltfläche des Speichers klicken, und von selbst in
dem in **Sync paths** festgelegten Intervall.

Ein importierter Schlüssel ist als synchronisiert markiert: Sein Name und sein Typ können nicht
geändert werden, und sein Löschen in Semaphore lässt das Secret im Speicher. Das Löschen des
Speichers bei eingeschalteter Synchronisierung löscht auch die von ihm importierten Schlüssel.

## Sync paths {#secretStorage-syncPaths}

Wird angezeigt, wenn **Sync keys enabled** eingeschaltet ist.

Öffnet den Dialog mit den Pfaden des Speichers, aus denen importiert wird, und dem Intervall des
automatischen Imports. Die Zahl auf der Schaltfläche ist die Anzahl der Pfade. Die
Synchronisierungsschaltfläche des Speichers bleibt deaktiviert, bis er mindestens einen Pfad hat.

Der Dialog hat nur **Schließen**: Was Sie darin ändern, wird mit dem Speicher gespeichert, durch
**Speichern** dieses Dialogs (**Erstellen** bei einem neuen Speicher).
