---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Integration bearbeiten

Eine Integration startet eine Task-Vorlage oder einen Workflow, wenn ein anderer Dienst eine
Webhook-URL von Semaphore aufruft, etwa bei einem Push-Ereignis von GitHub. Die URL ist ein Alias,
`/api/integrations/` und ein zufälliger Code: Die Aliase des Projekts stehen auf der Seite
Integrationen, die Aliase einer einzelnen Integration auf deren eigener Seite. Diese
Seite enthält auch die Matcher und die Werte, die die Integration aus der Anfrage entnimmt.

## Name {#integrationForm-name}

Der Name, unter dem die Integration auf der Seite Integrationen aufgeführt wird. Tasks, die die
Integration startet, werden ihr im Audit-Log zugeschrieben. Das Feld ist ein Pflichtfeld.

## Runs {#integrationForm-targetToggle}

Wird angezeigt, wenn der Server Workflows unterstützt.

Was die Integration startet: **Template** startet einen Task einer Task-Vorlage, **Workflow**
startet einen Lauf eines Workflows. Der Wechsel zu **Workflow** verwirft die Task-Parameter
darunter.

## Workflow to run {#integrationForm-workflowId}

Wird angezeigt, wenn **Runs** auf **Workflow** steht.

Der Workflow, den eine Anfrage startet. Jeder Wert, den die Integration aus der Anfrage entnimmt,
wird dem Workflow als Antwort auf seine Survey-Variable mit demselben Namen übergeben.
Workflow-Integrationen erfordern die Pro-Edition: Ohne sie ist das Feld deaktiviert, und der
Server lehnt die Änderung ab.

## Task Template to run {#integrationForm-templateId}

Wird angezeigt, wenn **Runs** auf **Template** steht.

Die Task-Vorlage, von der eine Anfrage einen Task startet. Das Feld ist ein Pflichtfeld. Sobald
eine Vorlage ausgewählt ist, erscheinen die Parameter ihrer Tasks darunter.

## Task-Parameter {#integrationForm-taskParams}

Wird angezeigt, wenn eine Task-Vorlage ausgewählt ist.

Die Parameter, die jeder von dieser Integration gestartete Task erhält, als ob Sie ihn selbst aus
der Vorlage gestartet hätten. Hier stehen nur die Optionen, die die Vorlage einen Task
überschreiben lässt. Aus der Anfrage entnommene Werte überschreiben diese: ein Wert vom Typ
**Variables** eine Antwort mit demselben Namen, ein Wert vom Typ **Task Params** einen Parameter
mit demselben Namen.

Wenn die Vorlage CLI-Argumente in einem Task erlaubt, enthält **CLI args** zusätzliche Argumente
für die Befehlszeile des Tasks.

## Nachricht {#taskParams-message}

Optional. Die Nachricht jedes Tasks, den die Integration startet, angezeigt im Verlauf der Tasks.

## Survey-Variablen {#taskParams-survey}

Wird angezeigt, wenn die Vorlage Survey-Variablen hat.

Die Antworten auf den Survey der Vorlage, für jeden Task dieselben. Ein aus der Anfrage
entnommener Wert vom Typ **Variables** ersetzt die Antwort mit demselben Namen.

Antworten auf Variablen vom Typ Secret werden nicht mit der Integration gespeichert: Übergeben Sie
ein Secret stattdessen über die Variablengruppe der Vorlage.

## Branch {#taskParams-gitBranch}

Wird angezeigt, wenn die Vorlage erlaubt, den Branch in einem Task zu überschreiben.

Der Branch oder Tag des Repositorys, den die Tasks auschecken. Lassen Sie ihn unverändert, um den
Branch der Vorlage zu verwenden. Ein aus der Anfrage entnommener Wert namens `git_branch` vom Typ
**Task Params** überschreibt ihn; ein führendes `refs/heads/` wird von diesem Wert entfernt.

## Inventory {#taskParams-inventoryId}

Wird angezeigt, wenn die Vorlage erlaubt, das Inventory in einem Task zu überschreiben.

Das Inventory oder, bei Terraform, der Workspace, gegen das die Tasks statt gegen das der Vorlage
laufen.

## Ansible-Optionen {#taskParams-ansible}

Wird angezeigt, wenn die Vorlage ein Ansible-Playbook ausführt.

Die Optionen von `ansible-playbook` für jeden Task, jede angezeigt, wenn die Vorlage sie erlaubt:
**Limit**, **Tags**, **Tags überspringen**, **Debug**, **Testlauf** (`--check`), **Diff** und
**Skip Galaxy install**.

## Terraform-Optionen {#taskParams-terraform}

Wird angezeigt, wenn die Vorlage Terraform, OpenTofu oder Terragrunt ausführt.

- **Plan**: Der Task führt den Plan aus und hält an, nichts wird angewendet.
- **Destroy**: Der Plan und das Apply zerstören die Ressourcen.
- **Auto Approve**: Der Plan wird ohne Bestätigung angewendet, wenn die Vorlage automatisches
  Akzeptieren erlaubt. Ohne diese Option wartet ein Task, bis jemand ihn in Semaphore bestätigt.
- **Upgrade** und **Reconfigure**: `init` läuft mit `-upgrade` und `-reconfigure`.

## Auth method {#integrationForm-authMethod}

Wie Semaphore prüft, dass eine Anfrage von dem erwarteten Dienst kommt. Eine Anfrage, die die
Prüfung nicht besteht, startet nichts, und der Webhook beantwortet sie genauso wie jede andere
Anfrage.

- **None**: keine Prüfung. Jeder, der die URL kennt, kann den Task starten.
- **GitHub Webhooks**: Der Header `X-Hub-Signature-256` muss die mit dem Secret erstellte
  HMAC-SHA256-Signatur des Bodys enthalten, so wie GitHub sie sendet.
- **Bitbucket Webhooks**: dieselbe Prüfung des Headers `X-Hub-Signature`, so wie Bitbucket ihn
  sendet.
- **Token**: Der in **Auth header** angegebene Header muss das Secret selbst enthalten.
- **HMAC** und **HMAC (SHA-512)**: Der in **Auth header** angegebene Header muss den mit dem
  Secret erstellten HMAC-SHA256 oder HMAC-SHA512 des Bodys enthalten, in hexadezimalen
  Kleinbuchstaben, ohne ein Präfix wie `sha256=`.
- **BasicAuth**: Die Anfrage muss den Login und das Passwort des Schlüssels aus
  **Auth-Secret** in ihrem Header `Authorization` mitführen.

## Auth header {#integrationForm-authHeader}

Wird angezeigt, wenn **Auth method** auf **Token**, **HMAC** oder **HMAC (SHA-512)** steht.

Der Name des Request-Headers, der das Token oder die Signatur enthält, zum Beispiel
`X-Webhook-Token`. Groß- und Kleinschreibung des Namens spielen keine Rolle. Eine Anfrage ohne
diesen Header wird abgelehnt.

## Auth-Secret {#integrationForm-authSecretId}

Wird angezeigt, wenn **Auth method** nicht auf **None** steht.

Der Schlüssel aus dem Schlüsselspeicher, gegen den die Anfrage geprüft wird. Für **BasicAuth**
passt nur ein Schlüssel vom Typ **Login mit Passwort**. Für die anderen Methoden passt auch ein
Schlüssel vom Typ **Zeichenkette**: Die Prüfung verwendet seinen Wert oder das Passwort eines
Schlüssels vom Typ **Login mit Passwort**.

Das Feld wird von jeder Methode außer **None** benötigt: Ohne Schlüssel oder mit einem leeren
Secret wird jede Anfrage abgelehnt.
