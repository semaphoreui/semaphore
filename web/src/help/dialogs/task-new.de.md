---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Neuer Task

Startet einen Task der Vorlage: einen Lauf ihres Playbooks, ihres Terraform-Codes oder ihres
Skripts. Die Vorlage bestimmt, welche Felder dieser Dialog zeigt, und jeder Abschnitt unten sagt,
wann sein Feld vorhanden ist. Die Schaltfläche unten heißt je nach Typ der Vorlage
**Ausführen**, **Build** oder **Deployment**.

Hat die Vorlage einen Survey, stehen seine Fragen als Felder nach **Nachricht**. Ein mit `*`
markiertes Feld ist ein Pflichtfeld, ein Integer-Feld nimmt nur Ziffern an, und die
Standardantworten sind eingetragen. Semaphore übergibt die Antworten als Variablen an den Task:
als Extra-Vars für Ansible, als `-var` für Terraform, als Argumente `name=value` für ein Skript
oder als Umgebungsvariablen für eine so eingestellte Frage. Eine Antwort ersetzt eine Variable
mit demselben Namen in den Variablengruppen der Vorlage. Eine geheime Antwort wird bei der
Eingabe verborgen und nicht mit dem Task gespeichert.

## Commit {#newTask-commit}

Wird angezeigt, wenn Sie einen früheren Task erneut ausführen und dieser Task auf einem bekannten
Commit lief.

Der Task checkt diesen Commit des Repositorys aus statt des neuesten Commits des Branches.
Schließen Sie das Feld, um den neuesten Commit auszuführen.

Der Commit wird nur verwendet, wenn die Vorlage das Überschreiben des Branches erlaubt, und nie
bei einem Repository, das als Monorepo markiert ist. Andernfalls führt der Task den neuesten
Commit aus.

## Build-Version {#newTask-buildTaskId}

Wird angezeigt, wenn der Typ der Vorlage **Deployment** ist.

Der Build, der ausgeliefert werden soll: einer der letzten 20 erfolgreichen Tasks der
Build-Vorlage, aufgeführt nach Version und Nachricht. Der neueste ist ausgewählt. Das Feld ist
ein Pflichtfeld.

Die Version des Builds erreicht den Task: in Ansible als
`semaphore_vars.task_details.incoming_version`, in einem Skript als Umgebungsvariable
`SEMAPHORE_TASK_DETAILS_INCOMING_VERSION`. Die Liste ist leer, wenn Sie die Build-Vorlage nicht
sehen dürfen.

## Nachricht {#newTask-message}

Optional. Eine Notiz, warum der Task läuft, zum Beispiel `Hotfix for the login page`. Sie wird mit
dem Task im Verlauf der Tasks angezeigt und ändert nichts am Lauf.

Das Playbook kann sie als `semaphore_vars.task_details.message` lesen, ein Skript als
Umgebungsvariable `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Branch {#newTask-gitBranch}

Wird angezeigt, wenn die Vorlage das Überschreiben des Branches erlaubt, sofern das Repository
kein Monorepo ist.

Der Branch oder Tag des Repositorys, den der Task auscheckt. Das Feld beginnt mit dem Branch der
Vorlage. Ist es leer, verwendet der Task den Branch der Vorlage oder, wenn die Vorlage keinen hat,
den Branch des Repositorys. Ein Name, der kein gültiger Git-Branchname ist, wird abgelehnt.

## Inventory {#newTask-inventoryId}

Wird angezeigt, wenn die Vorlage erlaubt, das Inventory in einem Task zu wählen, und Sie das Projekt
sehen dürfen. Bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage heißt das Feld **Workspace**,
aber der Server verwendet die Auswahl nur für Ansible-Vorlagen: Ein Terraform-, OpenTofu- oder
Terragrunt-Task läuft im Workspace der Vorlage, ganz gleich, was Sie hier wählen.

Das Inventory, gegen das der Task läuft. Das Feld beginnt mit dem Inventory der Vorlage und
listet die Inventorys des Projekts auf, die für die App der Vorlage gemacht sind. Ist das gewählte
Inventory beim Start des Tasks nicht mehr vorhanden, läuft der Task gegen das Inventory der
Vorlage.

## Limit {#taskParams-limit}

Wird bei einer Ansible-Vorlage angezeigt, die das Überschreiben des Limits erlaubt.

Die Hosts oder Gruppen des Inventorys, auf denen das Playbook läuft, ein Muster pro Chip, zum
Beispiel `web` oder `10.0.0.11`. Semaphore übergibt sie als `--limit`, durch Kommas verbunden. Die
Liste beginnt mit dem Limit der Vorlage. Entfernen Sie alle Chips, um auf allen Hosts des
Inventorys zu laufen.

## Tags {#taskParams-tags}

Wird bei einer Ansible-Vorlage angezeigt, die das Überschreiben der Tags erlaubt.

Nur die Tasks und Rollen des Playbooks mit einem dieser Tags laufen: Semaphore übergibt sie als
`--tags`, ein Tag pro Chip. Die Liste beginnt mit den Tags der Vorlage. Entfernen Sie alle Chips,
um das gesamte Playbook auszuführen.

## Tags überspringen {#taskParams-skipTags}

Wird bei einer Ansible-Vorlage angezeigt, die das Überschreiben der übersprungenen Tags erlaubt.

Die Tasks und Rollen des Playbooks mit einem dieser Tags laufen nicht: Semaphore übergibt sie als
`--skip-tags`, ein Tag pro Chip. Die Liste beginnt mit den übersprungenen Tags der Vorlage.

## Debug {#taskParams-debug}

Wird bei einer Ansible-Vorlage angezeigt, die Debugging erlaubt.

Führt `ansible-playbook` mit einem Ausführlichkeits-Flag aus, sodass das Log des Tasks mehr
Details zeigt. Die Beschriftung zeigt das Flag: `-v` so oft, wie **Debug-Level** angibt.

## Debug-Level {#taskParams-debugLevel}

Wird mit **Debug** angezeigt und ist aktiviert, wenn **Debug** angehakt ist.

Wie ausführlich das Log ist, von 1 (`-v`) bis 6 (`-vvvvvv`). Der Standardwert ist 4 (`-vvvv`), der
auch die Details der Verbindungen zu den Hosts zeigt.

## Testlauf {#taskParams-dryRun}

Wird bei einer Ansible-Vorlage angezeigt, sofern die Vorlage es nicht ausblendet.

Führt das Playbook mit `--check` aus: Ansible meldet, was es auf den Hosts ändern würde, und
ändert nichts.

## Diff {#taskParams-diff}

Wird bei einer Ansible-Vorlage angezeigt, sofern die Vorlage es nicht ausblendet.

Führt das Playbook mit `--diff` aus: Das Log zeigt die Änderungen an Dateien und Templates auf den
Hosts. Haken Sie zusätzlich **Testlauf** an, um die Änderungen zu sehen, ohne sie vorzunehmen.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Wird bei einer Ansible-Vorlage angezeigt, die das Überschreiben der Galaxy-Installation erlaubt.

Bevor das Playbook läuft, installiert Semaphore die Rollen und Collections, die in den Dateien
`requirements.yml` des Repositorys aufgeführt sind, wenn sich eine Datei seit der letzten
Installation geändert hat. Haken Sie das Kästchen an, um diesen Schritt für diesen Task zu
überspringen.

Das Kästchen beginnt nicht angehakt, unabhängig von der Einstellung der Vorlage, und ein nicht
angehaktes Kästchen installiert die Anforderungen.

## Plan {#taskParams-plan}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Führt `plan` aus und beendet den Task dort: Nichts wird angewendet. Verwenden Sie diese Option,
um die Änderungen zuerst zu prüfen. Ohne sie wendet der Task anschließend den Plan an, wenn er
Änderungen enthält.

## Destroy {#taskParams-destroy}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Fügt `plan` und `apply` die Option `-destroy` hinzu: Der Task plant das Entfernen jeder Ressource
im State des Workspace und wendet es dann an. Haken Sie zusätzlich **Plan** an, um zu sehen, was
zerstört würde, ohne es zu zerstören.

## Auto Approve {#taskParams-autoApprove}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Wendet den Plan an, sobald er erstellt ist. Ohne diese Option hält der Task nach dem Plan an und
wartet, bis jemand ihn auf der Seite des Tasks bestätigt oder ablehnt.

Das Kästchen wirkt nur, wenn die Vorlage automatisches Akzeptieren erlaubt. Eine Vorlage, die
immer akzeptiert, wendet ohne Nachfrage an, unabhängig vom Kästchen.

## Upgrade {#taskParams-upgrade}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Führt `init` mit `-upgrade` aus: Die Provider und Module werden auf die neuesten Versionen
aktualisiert, die ihre Versionsbeschränkungen erlauben.

## Reconfigure {#taskParams-reconfigure}

Wird bei einer Terraform-, OpenTofu- oder Terragrunt-Vorlage angezeigt.

Führt `init` mit `-reconfigure` aus: Das Backend wird neu eingerichtet, und die gespeicherte
Konfiguration des Backends wird ignoriert. Ohne diese Option führt Semaphore
`init -migrate-state` aus, was den bestehenden State in ein geändertes Backend verschiebt.

## CLI args {#newTask-args}

Wird angezeigt, wenn die Vorlage das Überschreiben der CLI-Argumente erlaubt.

Zusätzliche Argumente für die Befehlszeile des Tasks, eines pro Chip: `--forks=10` ist ein Chip,
`--forks` und `10` sind zwei. Klicken Sie auf einen Chip, um ihn zu bearbeiten, ziehen Sie ihn, um
die Reihenfolge zu ändern, und verwenden Sie den letzten Chip, um einen hinzuzufügen.

Die Liste beginnt mit den Argumenten der Vorlage. Semaphore übergibt immer die Argumente der
Vorlage und fügt diese Liste danach an, sobald Sie die Liste also ändern, werden die darin
verbliebenen Argumente der Vorlage zweimal übergeben. Entfernen Sie sie und behalten Sie nur die
Argumente, die Sie hinzufügen.
