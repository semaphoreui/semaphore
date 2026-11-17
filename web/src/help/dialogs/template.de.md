---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Vorlage bearbeiten 'Ansible Playbook'

Eine Task-Vorlage legt fest, was Semaphore ausführt und wie: das Repository und die Datei darin,
das Inventory und die Variablen und welche Optionen ein Benutzer beim Start eines Tasks ändern
darf. Jeder Task wird aus einer Vorlage gestartet. Die Felder des Dialogs hängen von der App der
Vorlage ab (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell und andere): Ein
Feld, das nur manche Apps haben, sagt das. Änderungen wirken für Tasks, die nach **Speichern**
gestartet werden.

## Vorlagentyp {#templateForm-typeTabs}

Die Tabs oben legen fest, welche Art von Vorlage es ist:

- **Task**: führt den Code aus. Die meisten Vorlagen sind von dieser Art.
- **Build**: erstellt ein Artefakt. Jeder Build-Task erhält eine Version, gezählt ab
  **Startversion**.
- **Deployment**: liefert ein von einem Build erstelltes Artefakt aus. Beim Start wählen Sie den
  Build-Task, dessen Version ausgeliefert wird.

Ein Build- oder Deployment-Task erhält die Variablen `SEMAPHORE_TASK_TYPE`,
`SEMAPHORE_TASK_TARGET_VERSION` (Build) und `SEMAPHORE_TASK_INCOMING_VERSION` (Deployment) in
seiner Umgebung. Ansible erhält dieselben Werte in der zusätzlichen Variablen
`semaphore_vars.task_details`.

## Startversion {#templateForm-startVersion}

Wird angezeigt, wenn der Typ **Build** ist. Pflichtfeld.

Die Version des ersten Build-Tasks, zum Beispiel `1.0.0` oder `v0.1`. Jeder weitere Build erhöht
die letzte Zahl der vorherigen Version um eins und behält den Text darum herum und ihre führenden
Nullen bei: `1.0.0`, `1.0.1`, `1.0.2`. Setzen Sie die Zahl hier über den letzten Build, um
vorzuspringen, zum Beispiel auf `2.0.0`; eine Version ohne Zahl ändert sich nie.

## Build-Vorlage {#templateForm-buildTemplateId}

Wird angezeigt, wenn der Typ **Deployment** ist. Pflichtfeld.

Die Vorlage, deren Tasks die Artefakte erstellen, die diese Vorlage ausliefert. Die Liste enthält
die Build-Vorlagen des Projekts und darunter die Deployment-Vorlagen: Ein Deployment kann auf ein
anderes Deployment folgen und erhält dann die Version, die jenes Deployment von seinem Build
erhalten hat.

## Auto-Run {#templateForm-autorun}

Wird angezeigt, wenn der Typ **Deployment** ist.

Startet von selbst einen Task dieser Vorlage, jedes Mal wenn ein Task der **Build-Vorlage**
erfolgreich ist, mit der Version, die dieser Task erstellt hat. Ein fehlgeschlagener Build startet
nichts.

## Name {#templateForm-name}

Der Name, unter dem die Vorlage auf der Seite Task-Vorlagen aufgeführt wird. Das Feld ist ein
Pflichtfeld.

## Repository {#templateForm-repositoryId}

Das Repository des Projekts, das den auszuführenden Code enthält. Das Feld ist ein Pflichtfeld.
Vor jedem Task aktualisiert Semaphore seine Kopie des Repositorys und führt den Code von dort aus.

Die Beschriftung rechts neben dem Feld zeigt den Branch, den die Tasks auschecken. Klicken Sie
darauf, um das Feld **Branch** zu öffnen.

## Branch {#templateForm-gitBranch}

Wird nach einem Klick auf die Branch-Beschriftung rechts neben **Repository** angezeigt.

Der Branch oder Tag, den die Tasks dieser Vorlage auschecken, statt des am Repository
festgelegten Branches. Lassen Sie das Feld leer, um den Branch des Repositorys zu verwenden. Wenn
Semaphore die Branches des Repositorys lesen kann, listet das Feld sie auf.

Das Feld ist deaktiviert, wenn das Repository ein Monorepo ist: Alle Vorlagen teilen sich einen
Checkout davon, auf dem Branch des Repositorys.

## Pfad zur Playbook-Datei {#templateForm-playbook}

Was ausgeführt wird, als Pfad ab dem Stammverzeichnis des Repositorys, zum Beispiel
`deploy/site.yml`. Der Pfad darf nicht absolut sein und nicht aus dem Repository herausführen. Die
Liste bietet die Dateien des Repositorys an; Sie können auch einen Pfad eingeben, der nicht in der
Liste steht. Die Schaltfläche rechts neben dem Feld liest die Liste erneut ein.

Die Beschriftung und die Bedeutung ändern sich mit der App:

- **Ansible**: das Playbook. Es ist ein Pflichtfeld; die Liste bietet `.yml`- und `.yaml`-Dateien
  an.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, das Verzeichnis mit dem
  Code. Lassen Sie das Feld leer, um den Code im Stammverzeichnis des Repositorys auszuführen.
  Geben Sie `/` ein, um aufzulisten, was ein Verzeichnis enthält.
- **Bash**, **Python**, **PowerShell** und andere Apps: **Script Filename**, das auszuführende
  Skript. Es ist ein Pflichtfeld. Das Skript läuft im Stammverzeichnis des Repositorys.

## Different working directory {#templateForm-workingDirectoryToggle}

Wird bei Ansible-Vorlagen angezeigt.

Führt `ansible-playbook` in einem anderen Verzeichnis als dem Stammverzeichnis des Repositorys
aus. Haken Sie das Kästchen an, um das Feld **Working directory** anzuzeigen. Entfernen Sie den
Haken, um wieder im Stammverzeichnis auszuführen.

## Working directory {#templateForm-workingDirectory}

Wird angezeigt, wenn **Different working directory** angehakt ist. Pflichtfeld.

Das Verzeichnis, in dem `ansible-playbook` und `ansible-galaxy` laufen, als Pfad ab dem
Stammverzeichnis des Repositorys, zum Beispiel `ansible`. Verwenden Sie es, wenn die
`ansible.cfg` des Projekts in einem Unterverzeichnis liegt, da Ansible die Datei aus dem
aktuellen Verzeichnis liest. Der Playbook-Pfad bleibt relativ zum Stammverzeichnis des
Repositorys. Der Pfad darf nicht absolut sein und nicht aus dem Repository herausführen.

## Inventory {#templateForm-inventoryId}

Wird bei Ansible-, Terraform-, OpenTofu- und Terragrunt-Vorlagen angezeigt.

- **Ansible**: das Inventory, gegen das das Playbook läuft, mit den Credentials für die Hosts. Das
  Feld ist ein Pflichtfeld.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, der Workspace, in dem der Code
  läuft. Semaphore wählt ihn vor dem Plan aus und erstellt ihn, wenn er nicht existiert. Eine neue
  Vorlage, die ohne Workspace gespeichert wird, erhält einen namens `default`.

Die Liste enthält die Inventorys des Projekts, die für die App der Vorlage gemacht sind.

## Variablengruppen {#templateForm-environmentIds}

Optional. Die Variablengruppen, deren Variablen, Umgebungsvariablen und Secrets jeder Task dieser
Vorlage erhält. Sie können mehrere wählen: Sie werden in der Reihenfolge des Feldes
zusammengeführt, und eine spätere Gruppe hat Vorrang vor einer früheren, wenn beide denselben
Namen setzen.

Wie die Variablen den Code erreichen, hängt von der App ab: als zusätzliche Variablen für
Ansible, als `-var` für Terraform und OpenTofu, als Argumente `name=value` für ein Skript.
Umgebungsvariablen werden für jede App gesetzt.

## Ansicht {#templateForm-viewId}

Optional. Die Ansicht, in der die Vorlage aufgeführt wird: Ansichten sind die Tabs der Seite
Task-Vorlagen, die die Vorlagen eines Projekts gruppieren. Lassen Sie das Feld leer, um die
Vorlage keiner Ansicht zuzuordnen.

## Runner-Tag {#templateForm-runnerTag}

Wird angezeigt, wenn Projekt-Runner verfügbar sind.

Tasks dieser Vorlage laufen nur auf einem Runner, der diesen Tag trägt, auch wenn der Server sie
selbst ausführen könnte. Der Tag überschreibt den Runner-Tag des Inventorys. Lassen Sie das Feld
leer, um den Tag des Inventorys oder keinen zu verwenden.

## Executor image {#templateForm-executorImage}

Das Container-Image, das der Runner für die Tasks dieser Vorlage startet, zum Beispiel
`semaphoreui/job:latest`. Lassen Sie das Feld leer, um das Image aus der Runner-Konfiguration zu
verwenden.

Nur die Executors Docker und Kubernetes führen Tasks in einem Container aus, daher ist das Feld
nur aktiviert, wenn einer von ihnen verfügbar ist. Ein Task, der ohne Container läuft, ignoriert
es.

## Survey-Variablen {#surveyVars}

Die Werte, nach denen ein Benutzer gefragt wird, wenn ein Task dieser Vorlage startet. Klicken Sie
auf **+**, um eine Variable hinzuzufügen, auf einen Chip, um eine zu ändern, auf das Kreuz eines
Chips, um sie zu entfernen; ziehen Sie die Chips, um die Reihenfolge der Fragen festzulegen.

Die Antworten erreichen den Code wie die Variablen einer Variablengruppe oder als
Umgebungsvariablen, je nach Einstellung der Variablen. Sie haben Vorrang vor einer
Variablengruppe, die denselben Namen setzt.

## Parallele Tasks erlauben {#templateForm-allowParallelTasks}

Lässt Tasks dieser Vorlage gleichzeitig laufen. Ohne diese Option wartet ein Task der Vorlage in
der Warteschlange, bis der laufende endet. Die Limits paralleler Tasks des Projekts und des
Servers gelten in beiden Fällen.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Wird angezeigt, wenn der Administrator die Ausgabe von JWTs für den Server eingeschaltet hat.

Gibt jedem Task dieser Vorlage ein von Semaphore signiertes JSON Web Token in der
Umgebungsvariablen `SEMAPHORE_JWT`. Der Code kann das Token einem anderen Dienst vorlegen, um zu
beweisen, welcher Task er ist: Das Token nennt den Task, das Projekt, die Vorlage und den
Benutzer, der ihn gestartet hat. Der Dienst prüft die Signatur mit den öffentlichen Schlüsseln
unter `/.well-known/jwks.json` des Semaphore-Servers.

## JWT audience {#templateForm-jwtAudience}

Wird angezeigt, wenn **Issue JWT to task runner** angehakt ist. Optional.

Die Dienste, für die das Token bestimmt ist, eingetragen in seinen Claim `aud`, zum Beispiel
`https://vault.example.com`. Drücken Sie nach jedem Wert Enter. Bis zu 32 Werte; lassen Sie das
Feld leer für ein Token ohne Audience.

## JWT TTL {#templateForm-jwtTtl}

Wird angezeigt, wenn **Issue JWT to task runner** angehakt ist. Optional.

Wie lange das Token ab dem Start des Tasks gültig bleibt, als Dauer wie `30s`, `15m` oder `1h`.
Lassen Sie das Feld leer, um den Standardwert des Servers zu verwenden, eine Stunde, sofern der
Administrator keinen anderen festgelegt hat. Der Wert darf das Maximum des Servers nicht
überschreiten, 24 Stunden, sofern nicht anders festgelegt.

## Ich möchte einen Task per Cron nur für neue Commits eines Repositories ausführen {#templateForm-cronVisible}

Startet einen Task dieser Vorlage, wenn in einem Repository ein neuer Commit erscheint. Semaphore
fragt das Repository im gewählten Intervall nach seinem letzten Commit und startet nur dann einen
Task, wenn sich der Commit seit der letzten Prüfung geändert hat.

Das Anhaken des Kästchens und Speichern erstellt den Zeitplan der Prüfung; das Entfernen des
Hakens und Speichern löscht diesen Zeitplan. Die anderen Zeitpläne der Vorlage bleiben erhalten.

## Repository {#templateForm-cronRepositoryId}

Wird angezeigt, wenn **Ich möchte einen Task per Cron nur für neue Commits eines Repositories
ausführen** angehakt ist. Pflichtfeld.

Das Repository, das auf neue Commits überwacht wird. Semaphore überwacht den an diesem Repository
festgelegten Branch, nicht den **Branch** der Vorlage.

## Prüfintervall {#templateForm-cronFormat}

Wird angezeigt, wenn **Ich möchte einen Task per Cron nur für neue Commits eines Repositories
ausführen** angehakt ist.

Wie oft Semaphore das Repository auf einen neuen Commit prüft: von jeder Minute bis einmal am
Tag. Der Standardwert ist jede Minute.

## Erfolgsbenachrichtigungen unterdrücken {#templateForm-suppressSuccessAlerts}

Verhindert, dass die erfolgreichen Tasks dieser Vorlage Benachrichtigungen versenden. Die
Benachrichtigungen fehlgeschlagener Tasks werden weiterhin versendet. Die Option wirkt nur, wenn
Benachrichtigungen für das Projekt eingeschaltet sind.

## Fehlerbenachrichtigungen unterdrücken {#templateForm-suppressErrorAlerts}

Verhindert, dass die fehlgeschlagenen Tasks dieser Vorlage Benachrichtigungen versenden. Die
Benachrichtigungen erfolgreicher Tasks werden weiterhin versendet. Die Option wirkt nur, wenn
Benachrichtigungen für das Projekt eingeschaltet sind.

## CLI args {#templateForm-args}

Argumente, die dem Befehl jedes Tasks hinzugefügt werden. Jeder Chip ist ein Argument, das
unverändert und ohne Shell übergeben wird: Setzen Sie `--forks=10` in einen Chip oder `--forks`
und `10` in zwei. Ziehen Sie die Chips, um ihre Reihenfolge zu ändern. Die Schaltfläche rechts
neben der Liste öffnet die Argumente als JSON-Array:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: an `ansible-playbook` angefügt, vor dem Playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: an `plan` und `apply` angefügt, nicht an `init`.
- Ein Skript: nach dem Skriptnamen angefügt.

## CLI-Argumente {#templateForm-allowOverrideArgsInTask}

Unter **Prompts**. Lässt den Benutzer beim Start eines Tasks eigene Argumente hinzufügen. Sie
werden nach den Argumenten der Vorlage angefügt.

## Branch überschreiben {#templateForm-allowOverrideBranchInTask}

Unter **Prompts**. Lässt den Benutzer beim Start eines Tasks einen anderen Branch, Tag oder Commit
wählen statt des Branches der Vorlage. Ohne diese Option checkt der Task immer den Branch der
Vorlage aus.

Das Kästchen ist deaktiviert, wenn das Repository ein Monorepo ist.

## Inventory {#templateForm-allowOverrideInventory}

Unter **Prompts**. Wird bei Ansible-, Terraform-, OpenTofu- und Terragrunt-Vorlagen angezeigt.

Lässt den Benutzer beim Start eines Tasks ein anderes Inventory wählen. Nur Ansible-Vorlagen
verwenden das dort gewählte Inventory: Ein Terraform-, OpenTofu- oder Terragrunt-Task läuft im
Workspace der Vorlage, auch wenn das Kästchen angehakt ist.

## Limit {#templateForm-limit}

Wird bei Ansible-Vorlagen angezeigt.

Die Hosts oder Gruppen aus dem Inventory, auf denen das Playbook läuft, übergeben als `--limit`.
Jeder Chip ist ein Muster, zum Beispiel `web` oder `db01`; lassen Sie die Liste leer, um auf jedem
Host zu laufen.

## Tags {#templateForm-tags}

Wird bei Ansible-Vorlagen angezeigt.

Führt nur die Tasks und Rollen mit diesen Tags aus, übergeben als `--tags`. Lassen Sie die Liste
leer, um alles auszuführen.

## Tags überspringen {#templateForm-skipTags}

Wird bei Ansible-Vorlagen angezeigt.

Überspringt die Tasks und Rollen mit diesen Tags, übergeben als `--skip-tags`.

## Vaults {#templateVaults}

Wird bei Ansible-Vorlagen angezeigt.

Die Ansible-Vault-Passwörter, die das Playbook zum Entschlüsseln seiner Dateien braucht. Klicken
Sie auf **+**, um eines hinzuzufügen, auf einen Chip, um es zu ändern, auf das Kreuz eines Chips,
um es zu entfernen. Jeder Vault hat einen Namen, seine Vault-ID, und liefert das Passwort aus
einem Schlüssel des Schlüsselspeichers oder aus einem Client-Skript des Repositorys.

## Auto-Akzeptieren {#templateForm-autoApprove}

Wird bei Terraform-, OpenTofu- und Terragrunt-Vorlagen angezeigt.

Wendet den Plan ohne Nachfrage an. Ohne diese Option hält ein Task nach dem Plan an und wartet,
bis ein Benutzer ihn bestätigt oder ablehnt; ein Plan ohne Änderungen beendet den Task, ohne dass
etwas anzuwenden ist.

## Backend-Einstellungen überschreiben {#templateForm-overrideBackend}

Wird bei Terraform-, OpenTofu- und Terragrunt-Vorlagen angezeigt, wenn Semaphore den
Terraform-State aufbewahren kann.

Bewahrt den State des Codes in Semaphore auf. Vor `init` schreibt Semaphore eine Datei, die das
Backend `http` festlegt, in das Verzeichnis des Codes und löscht sie nach dem Task:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend-Dateiname {#templateForm-backendFilename}

Wird mit **Backend-Einstellungen überschreiben** angezeigt; aktiviert, wenn dieses Kästchen
angehakt ist. Optional.

Der Name der Datei, in die Semaphore das Backend schreibt, zum Beispiel `semaphore_backend.tf`.
Der Standardwert ist `backend.tf`. Der Name muss auf `.tf` enden und darf nur Buchstaben, Ziffern,
`_`, `-` und `.` enthalten. Wählen Sie einen Namen, den keine Datei Ihres Codes hat: Semaphore
überschreibt diese Datei und löscht sie nach dem Task.

## Limit {#templateForm-allowOverrideLimit}

Unter **Ansible-Prompts**. Lässt den Benutzer beim Start eines Tasks die Hosts festlegen, auf
denen ausgeführt wird, statt des **Limit** der Vorlage.

## Tags {#templateForm-allowOverrideTags}

Unter **Ansible-Prompts**. Lässt den Benutzer beim Start eines Tasks die Tags festlegen, statt der
**Tags** der Vorlage.

## Tags überspringen {#templateForm-allowOverrideSkipTags}

Unter **Ansible-Prompts**. Lässt den Benutzer beim Start eines Tasks die zu überspringenden Tags
festlegen, statt der Einstellung **Tags überspringen** der Vorlage.

## Debug erlauben {#templateForm-allowDebug}

Unter **Ansible-Prompts**. Lässt den Benutzer beim Start eines Tasks eine ausführliche Ausgabe
einschalten, von `-v` bis `-vvvvvv`.

## Testlauf {#templateForm-showDryRun}

Unter **Ansible-Prompts**. Standardmäßig angehakt. Bietet beim Start eines Tasks die Option
**Testlauf** an, die das Playbook mit `--check` ausführt: Ansible meldet, was es ändern würde, und
ändert nichts. Entfernen Sie den Haken, um die Option zu entfernen.

## Diff {#templateForm-showDiff}

Unter **Ansible-Prompts**. Standardmäßig angehakt. Bietet beim Start eines Tasks die Option
**Diff** an, die das Playbook mit `--diff` ausführt, um die Änderungen an Dateien zu zeigen.
Entfernen Sie den Haken, um die Option zu entfernen.

## Auto-Akzeptieren {#templateForm-allowAutoApprove}

Unter den Prompts von Terraform, OpenTofu und Terragrunt. Lässt den Benutzer beim Start eines
Tasks **Auto Approve** anhaken, um den Plan ohne Bestätigung anzuwenden. Der Task-Dialog bietet
diese Option in jedem Fall an, aber ohne dieses Kästchen ignoriert Semaphore sie: Der Task wartet
nach dem Plan auf eine Bestätigung, es sei denn, **Auto-Akzeptieren** der Vorlage ist angehakt.

## Galaxy install options {#templateForm-galaxy}

Wird bei Ansible-Vorlagen angezeigt. Klicken Sie darauf, um die Optionen der Galaxy-Installation
anzuzeigen; die Beschriftung zählt die Optionen, die von den Voreinstellungen abweichen.

Bevor das Playbook läuft, installiert Semaphore mit `ansible-galaxy` die Rollen und Collections,
die das Repository verlangt. Es sucht `requirements.yml`, `roles/requirements.yml` und
`collections/requirements.yml` neben dem Playbook und im Stammverzeichnis des Repositorys und
führt die Installation nur aus, wenn sich eine Datei seit dem letzten Task geändert hat.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Unter **Galaxy install options**. Überspringt die Galaxy-Installation: Das Playbook läuft mit den
zuvor installierten Rollen und Collections. Verwenden Sie die Option, wenn die Anforderungen auf
anderem Weg installiert werden oder die Installation langsam und nicht nötig ist.

## Role install args {#templateForm-galaxyRoleArgs}

Unter **Galaxy install options**. Argumente, die an `ansible-galaxy role install` angefügt
werden, eines pro Chip, zum Beispiel `--ignore-certs` oder `--server=https://galaxy.example.com`.
Semaphore akzeptiert nur die Flags der Installation, die gefahrlos übergeben werden können: Ein
unbekanntes Flag, `--token`, `--api-key`, `-p` oder `--roles-path` verhindert das Speichern der
Vorlage.

## Collection install args {#templateForm-galaxyCollectionArgs}

Unter **Galaxy install options**. Argumente, die an `ansible-galaxy collection install` angefügt
werden, eines pro Chip, zum Beispiel `--pre` oder `--upgrade`. Es gilt dieselbe Regel wie für
**Role install args**: `--token`, `--api-key`, `-p` und `--collections-path` werden nicht
akzeptiert.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Unter **Galaxy install options**, in **Prompts**. Lässt den Benutzer beim Start eines Tasks
wählen, ob die Galaxy-Installation übersprungen wird. Die Wahl des Tasks hat Vorrang vor
**Skip Galaxy install** der Vorlage.
