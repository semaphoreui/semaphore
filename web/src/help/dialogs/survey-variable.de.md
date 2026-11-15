---
form: src/components/SurveyVars.vue
---

# Survey-Variable

Eine Survey-Variable ist eine Frage, die der Task-Dialog bei jedem Ausführen der Vorlage stellt.
Die Antwort erreicht den Task als Variable: als zusätzliche Variable für Ansible, als `-var` für
Terraform und OpenTofu, als Argument für ein Skript oder als Umgebungsvariable. Die Variable wird
mit der Vorlage gespeichert, wenn Sie den Vorlagendialog speichern. Derselbe Dialog bearbeitet
den Survey eines Workflows, bei dem der Typ **Secret** nicht angeboten wird.

## Name {#surveyVarDialog-name}

Der Name, unter dem der Task die Antwort erhält: der Name der zusätzlichen Variablen, der
Terraform-Variablen, des Arguments `name=value` eines Skripts oder der Umgebungsvariablen.
Leerzeichen an den Enden werden entfernt. Das Feld ist ein Pflichtfeld.

Verwenden Sie einen Namen, den keine andere Variable des Surveys hat und den das Werkzeug
akzeptiert, zum Beispiel `app_version`. Eine Antwort ersetzt eine Variable mit demselben Namen
aus der Variablengruppe der Vorlage.

## Titel {#surveyVarDialog-title}

Die Beschriftung des Feldes im Task-Dialog, zum Beispiel `Version to deploy`. Sie ist auch der
Text des Chips der Variablen im Vorlagendialog. Das Feld ist ein Pflichtfeld.

## Beschreibung {#surveyVarDialog-description}

Optional. Ein Hinweis, der unter dem Feld des Task-Dialogs angezeigt wird, solange es den Fokus
hat: was einzugeben ist, oder ein Beispiel.

## Typ {#surveyVarDialog-type}

Welche Art von Antwort der Task-Dialog verlangt:

- **String**: eine Zeile Text. Das ist die Voreinstellung.
- **Integer**: nur Ziffern; der Task-Dialog lehnt alles andere ab, auch ein Minuszeichen. Die
  Antwort erreicht den Task trotzdem als Text: Wandeln Sie sie im Playbook um, zum Beispiel mit
  `| int`.
- **Secret**: Text, den der Task-Dialog maskiert. Die Antwort wird nicht mit dem Task
  gespeichert, sodass eine erneute Ausführung sie wieder abfragt, und Ansible erhält sie in einer
  mit Vault verschlüsselten Datei. Ein Secret hat keinen Standardwert.
- **Enum**: ein Wert, ausgewählt aus einer Liste, die Sie unten festlegen.
- **Text**: mehrere Zeilen Text.
- **Select**: mehrere Werte, ausgewählt aus einer Liste, die Sie unten festlegen. Der Task erhält
  sie als Liste; in einem Argument, einem `-var` oder einer Umgebungsvariablen wird die Liste als
  JSON geschrieben, zum Beispiel `["a","b"]`.

Das Ändern des Typs leert den Standardwert, und ein Typ ohne Liste leert die Werte.

## Pass variable as {#surveyVarDialog-target}

Wie die Antwort den Task erreicht:

- **Extra variable**: der übliche Weg des Werkzeugs. Ansible erhält sie als zusätzliche Variable,
  Terraform und OpenTofu als `-var name=value`, ein Bash-, PowerShell- oder Python-Skript als
  Argument `name=value` nach dem Skript.
- **Environment variable**: eine Umgebungsvariable des Prozesses, `name=value`, für jede Art von
  Vorlage. Die Variable wird dann nicht auf dem anderen Weg übergeben.

## Schlüssel {#surveyVarDialog-valueName}

Wird angezeigt, wenn der Typ **Enum** oder **Select** ist.

Der Text, den der Task-Dialog für diesen Wert in der Liste anzeigt. Jeder Wert braucht einen
Schlüssel, und die Schlüssel müssen sich unterscheiden: Andernfalls lehnt der Dialog das
Speichern der Variablen ab.

## Wert {#surveyVarDialog-valueValue}

Wird angezeigt, wenn der Typ **Enum** oder **Select** ist.

Was der Task erhält, wenn dieser Wert ausgewählt wird. Er kann sich von seinem Schlüssel
unterscheiden: Der Schlüssel `Production` kann `prod` übergeben.

## Add Value {#surveyVarDialog-addValue}

Wird angezeigt, wenn der Typ **Enum** oder **Select** ist.

Fügt der Liste der Werte eine leere Zeile hinzu. Ein **Enum** oder ein **Select** braucht
mindestens einen Wert. Das Papierkorbsymbol einer Zeile entfernt sie.

## Standardwert {#surveyVarDialog-defaultValue}

Wird bei jedem Typ außer **Secret** angezeigt.

Die Antwort, die der Task-Dialog beim Öffnen einträgt; Sie können sie dort ändern. Wählen Sie bei
einem **Enum** einen der Werte, bei einem **Select** beliebig viele davon. Lassen Sie das Feld
leer, um mit einer leeren Antwort zu beginnen.

Der Standardwert wird nur vom Task-Dialog eingetragen. Ein auf anderem Weg gestarteter Task, zum
Beispiel über die API, erhält keinen Standardwert.

## Erforderlich {#surveyVarDialog-required}

Der Task-Dialog startet keinen Task, bis die Variable eine Antwort hat. Wie der Standardwert wird
diese Prüfung nur vom Task-Dialog vorgenommen, nicht vom Server.
