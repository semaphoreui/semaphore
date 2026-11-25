---
form: src/components/KeyForm.vue
---

# Bearbeiten Key

Ein Schlüssel des Schlüsselspeichers enthält ein Credential, das das Projekt verwendet, ohne es
anzuzeigen: den SSH-Schlüssel oder das Passwort, mit dem Ansible sich an den Hosts anmeldet, die
Credentials eines Git-Repositorys, das Passwort eines Ansible Vault, das Secret einer
Integration. Inventorys, Repositories, Task-Vorlagen und Integrationen verweisen über seinen Namen
auf einen Schlüssel. Nach dem Speichern wird das Secret nie wieder an den Browser gesendet.

## Schlüsselname {#key-name}

Der Name, unter dem der Schlüssel im Schlüsselspeicher und in den Schlüsselauswahlen von
Inventorys, Repositories, Task-Vorlagen und Integrationen aufgeführt wird. Das Feld ist ein
Pflichtfeld. Wählen Sie einen Namen, der sagt, was der Schlüssel öffnet, zum Beispiel
`Deploy key of the web servers`.

## Quelle {#key-source}

Wo das Secret des Schlüssels aufbewahrt wird. Die Tabs:

- **Local**: Das Secret wird in diesen Dialog eingegeben und in der Datenbank von Semaphore
  gespeichert, verschlüsselt, wenn der Administrator einen Verschlüsselungsschlüssel für
  Zugangsschlüssel festgelegt hat. Das ist die Voreinstellung.
- **Storage**: Das Secret wird in einem externen Secret-Speicher des Projekts aufbewahrt, der
  unten in **Storage** ausgewählt wird. Den Tab gibt es nur in der Pro-Edition. Die Speicher
  selbst werden auf dem Tab **Storages** des Schlüsselspeichers eingerichtet.
- **Env**: Das Secret wird aus einer Umgebungsvariablen des Semaphore-Serverprozesses gelesen.
  Geben Sie den Namen der Variablen in das Feld unter den Tabs ein.
- **File**: Das Secret wird aus einer Datei auf dem Semaphore-Server gelesen. Geben Sie den
  absoluten Pfad der Datei in das Feld unter den Tabs ein. Die Datei muss im Secrets-Verzeichnis
  des Servers liegen (`dirs.secrets` in seiner Konfiguration, standardmäßig `/tmp/semaphore`);
  ein Pfad außerhalb davon oder einer mit `..` darin schlägt fehl, wenn der Schlüssel verwendet
  wird. Ein abschließender Zeilenumbruch der Datei wird entfernt.

Mit **Env** und **File** schreibt Semaphore das Secret nie: Es liest den Wert jedes Mal, wenn ein
Task den Schlüssel verwendet, sodass Sie das Secret dort ändern, wo es liegt. Der Wert muss zum
**Typ** des Schlüssels passen. Ein Schlüssel vom Typ **Zeichenkette** übernimmt den Wert, wie er
ist. Ein Schlüssel vom Typ **SSH-Schlüssel** oder **Login mit Passwort** erwartet JSON mit den
Feldern dieses Typs, zum Beispiel:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Wenn Sie einen Schlüssel bearbeiten, sind die Tabs gesperrt, bis **Überschreiben** angehakt ist.
Ein aus einem Secret-Speicher synchronisierter Schlüssel behält seine Quelle, und sein Name und
sein Typ können nicht geändert werden.

## Storage {#key-sourceStorageId}

Wird angezeigt, wenn der Tab **Storage** gewählt ist.

Der Secret-Speicher des Projekts, der das Secret enthält: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager oder Azure Key Vault, wie auf dem Tab **Storages** des
Schlüsselspeichers eingerichtet.

- Ein Speicher, der nicht schreibgeschützt ist, erhält das Secret, das Sie in diesen Dialog
  eingeben: Semaphore schreibt es beim Speichern dorthin. Das Löschen des Schlüssels löscht das
  Secret auch aus dem Speicher. Ein in einem solchen Speicher aufbewahrter Schlüssel kann später
  nicht in einen anderen Speicher verschoben werden.
- Ein schreibgeschützter Speicher wird nur gelesen. Der Dialog zeigt dann keine Felder für das
  Secret, und Semaphore liest es jedes Mal aus dem Speicher, wenn ein Task den Schlüssel
  verwendet.

## Source Key {#key-sourceStorageKey}

Wird angezeigt, wenn ein **Storage** ausgewählt ist.

Wo das Secret in diesem Speicher liegt, zum Beispiel der Pfad zum Secret in HashiCorp Vault.

## Typ {#key-type}

Welche Art von Credential der Schlüssel enthält. Das Feld ist ein Pflichtfeld, und die Felder
darunter ändern sich mit dem Typ.

- **SSH-Schlüssel**: ein privater Schlüssel mit optionalem Login und optionaler Passphrase.
  Ansible verbindet sich damit mit den Hosts, und Git klont damit ein Repository mit einer
  `ssh`-URL.
- **Login mit Passwort**: ein Login und ein Passwort. Ansible verbindet sich damit mit den Hosts
  oder wird damit zu einem anderen Benutzer (**Sudo-Credentials** eines Inventorys); Git klont
  damit ein `https`-Repository; das Passwort öffnet einen Ansible Vault einer Task-Vorlage.
- **Zeichenkette**: ein einzelner geheimer Wert ohne Login: ein Token, ein API-Schlüssel, ein
  Passwort. Verwenden Sie ihn für das Secret einer Integration. Er kann kein SSH-Schlüssel, keine
  Sudo-Credentials, kein Vault-Passwort und kein Zugangsschlüssel eines Repositorys sein.
- **Keine**: kein Credential. Verwenden Sie diesen Typ für ein Repository, das keinen Login
  braucht (ein öffentliches `https`-Repository, eine `file`-URL oder ein lokaler Pfad), und für
  ein Inventory, dessen Hosts keine Credentials brauchen oder die Verbindung selbst festlegen.

## Benutzername {#key-loginPasswordLogin}

Wird angezeigt, wenn der Typ **Login mit Passwort** ist. Optional.

Der Benutzername des Credentials.

- Ansible verbindet sich als dieser Benutzer mit den Hosts (`--user`); als **Sudo-Credentials**
  eines Inventorys ist es der Benutzer, zu dem Ansible wird (`--become-user`). Ist das Feld leer,
  verwendet Ansible den Benutzer, den das Inventory oder seine eigene Konfiguration nennt.
- Für ein `https`-Repository meldet sich Git mit diesem Benutzer und dem Passwort an. Lassen Sie
  das Feld leer für ein Token, das der Git-Server allein akzeptiert: Das Passwort wird dann als
  Benutzername gesendet.

## Passwort {#key-loginPasswordPassword}

Wird angezeigt, wenn der Typ **Login mit Passwort** ist. Pflichtfeld.

Das Passwort des Credentials. Semaphore beantwortet damit die Passwortabfragen von Ansible: das
SSH-Passwort (`--ask-pass`) für Benutzer-Credentials, das `become`-Passwort
(`--ask-become-pass`) für Sudo-Credentials und das Vault-Passwort für einen Ansible Vault. Für ein
`https`-Repository ist es das Passwort oder das Zugangstoken des Git-Servers.

## Wert {#key-stringValue}

Wird angezeigt, wenn der Typ **Zeichenkette** ist. Pflichtfeld.

Der geheime Wert selbst, zum Beispiel das Token, gegen das eine Integration die empfangenen
Anfragen prüft. Er wird so gespeichert, wie er eingegeben wird.

## Benutzername {#key-sshLogin}

Wird angezeigt, wenn der Typ **SSH-Schlüssel** ist. Optional.

Der Benutzer, als der Ansible sich mit den Hosts verbindet (`--user`). Ist das Feld leer,
verwendet Ansible den Benutzer, den das Inventory oder seine eigene Konfiguration nennt.

Für ein Git-Repository tragen Sie den Benutzer stattdessen in die Repository-URL ein, wie in
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Wird angezeigt, wenn der Typ **SSH-Schlüssel** ist. Optional.

Die Passphrase, mit der der private Schlüssel verschlüsselt ist. Lassen Sie das Feld leer für
einen Schlüssel ohne Passphrase. Eine falsche Passphrase lässt jeden Task, der den Schlüssel
verwendet, beim Laden des Schlüssels fehlschlagen. Ein von Semaphore erzeugter Schlüssel hat keine
Passphrase: Was dieses Feld enthält, wird dann verworfen.

## Generate SSH Key {#key-generateSshKey}

Wird angezeigt, wenn der Typ **SSH-Schlüssel** ist.

Semaphore erzeugt beim Speichern ein neues Schlüsselpaar: einen privaten RSA-Schlüssel mit
2048 Bit, der im Schlüssel bleibt, und seinen öffentlichen Schlüssel, der nach dem Speichern
einmal in einem Dialog und danach unter **Public key** angezeigt wird, wann immer Sie den
Schlüssel bearbeiten. Fügen Sie den öffentlichen Schlüssel auf den Hosts in
`~/.ssh/authorized_keys` hinzu oder als Deploy-Key des Git-Repositorys.

Das Feld für den privaten Schlüssel ist gesperrt, solange das Kästchen angehakt ist. Bei einem
bestehenden Schlüssel haken Sie zuerst **Überschreiben** an: Das alte Schlüsselpaar wird ersetzt.
Ein Schlüssel, dessen Secret aus **Env**, **File** oder einem schreibgeschützten Speicher gelesen
wird, kann nicht erzeugt werden, da Semaphore ihn nirgendwohin schreiben kann.

## Privater Schlüssel {#key-sshPrivateKey}

Wird angezeigt, wenn der Typ **SSH-Schlüssel** ist. Pflichtfeld, außer wenn
**Generate SSH Key** angehakt ist.

Der private Schlüssel im Format OpenSSH oder PEM, einschließlich seiner Zeilen `BEGIN` und `END`.
Für einen Task lädt Semaphore ihn in einen eigenen SSH-Agenten, den Ansible und Git für die Dauer
des Tasks verwenden.

## Public key {#key-publicKey}

Wird angezeigt, wenn Sie einen von Semaphore erzeugten SSH-Schlüssel bearbeiten, solange
**Überschreiben** nicht angehakt ist.

Die öffentliche Hälfte des erzeugten Schlüsselpaars. **Show** klappt sie auf, die
Kopierschaltfläche legt sie in die Zwischenablage. Fügen Sie sie auf den Hosts in
`~/.ssh/authorized_keys` hinzu oder als Deploy-Key des Git-Repositorys.

Es ist der öffentliche Schlüssel des Paars, das Semaphore zuletzt erzeugt hat. Nachdem Sie den
privaten Schlüssel mit **Überschreiben** von Hand ersetzt haben, passt der hier angezeigte
Schlüssel nicht mehr zu ihm.

## Überschreiben {#key-overrideSecret}

Wird angezeigt, wenn Sie einen bestehenden Schlüssel bearbeiten.

Semaphore zeigt ein gespeichertes Secret nie wieder an, daher sind die Felder des Secrets leer und
gesperrt. Haken Sie das Kästchen an, um das Secret zu ersetzen: Die Quelle, der Typ und die Felder
des Secrets werden freigegeben, und alles darin wird neu gespeichert, geben Sie also das gesamte
Secret erneut ein, nicht nur den geänderten Teil. Ohne diese Option ändert das Speichern nur den
Namen des Schlüssels.

Das Speichern eines Schlüssels entfernt außerdem die Kopien, die der Semaphore-Server von den
Repositories aufbewahrt, die ihn verwenden: Der nächste Task klont sie erneut.
