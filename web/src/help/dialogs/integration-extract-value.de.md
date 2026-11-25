---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Ein extrahierter Wert nimmt einen Teil der eingehenden Anfrage, etwa den Branch eines
Push-Ereignisses, und übergibt ihn an den Task, den die Integration startet. Jeder Wert der
Integration wird aus jeder Anfrage entnommen, die einen Task startet.

## Name {#integrationValueForm-name}

Der Name, unter dem der Wert auf der Seite der Integration aufgeführt wird. Er ändert nicht, was
entnommen wird oder wohin es geht. Das Feld ist ein Pflichtfeld.

## Source of the Value {#integrationValueForm-valueSource}

Der Teil der Anfrage, aus dem der Wert entnommen wird:

- **Header**: der Wert eines Request-Headers.
- **Body**: der Body der Anfrage.

## Data Type of Body {#integrationValueForm-bodyDataType}

Wird angezeigt, wenn **Source of the Value** auf **Body** steht.

Wie der Body gelesen wird:

- **JSON**: Der Body wird als JSON geparst, und **Key** ist der Pfad zu einem Wert darin. Der
  Wert behält seinen JSON-Typ: Eine Zahl, eine Liste oder ein Objekt erreichen den Task als
  solche.
- **String**: Der gesamte Body wird als Text übernommen, und **Key** wird nicht verwendet.

## Key {#integrationValueForm-key}

Was aus der Anfrage entnommen wird. Das Feld ist ein Pflichtfeld.

- Für **Header**: der Name des Headers, zum Beispiel `X-GitHub-Event`. Groß- und
  Kleinschreibung spielen keine Rolle. Ein Header, der in der Anfrage fehlt, ergibt einen leeren
  Wert.
- Für **Body** mit **JSON**: der Pfad zum Wert, dessen Teile durch Punkte verbunden sind; ein
  Element eines Arrays wird als `.[0]` geschrieben, zum Beispiel `repository.full_name` oder
  `commits.[0].id`. Ein Pfad, der im Body fehlt, ergibt keinen Wert, und der Task behält den in
  der Integration festgelegten.
- Für **Body** mit **String**: nicht verwendet, aber das Formular braucht einen Wert; geben Sie
  irgendetwas ein.

## Variable Usage {#integrationValueForm-variableType}

Wohin der Wert in einem Task einer Vorlage geht:

- **Variables**: eine zusätzliche Variable des Tasks, so wie die Antwort auf eine
  Survey-Variable übergeben wird. Sie ersetzt eine Antwort oder eine Variable der Variablengruppe
  mit demselben Namen.
- **Task Params**: ein Parameter des Tasks, so wie die Optionen des Task-Dialogs gespeichert
  werden. Der Name `git_branch` setzt den Branch des Tasks, wenn die Vorlage erlaubt, den Branch
  zu überschreiben; ein führendes `refs/heads/` wird entfernt.

Eine Integration, die einen Workflow ausführt, ignoriert die Verwendung: Jeder Wert wird zur
Antwort auf die Survey-Variable des Workflows, die in **Variable** genannt ist.

## Variable {#integrationValueForm-variable}

Der Name, unter dem der Wert übergeben wird, zum Beispiel `branch` für eine zusätzliche Variable,
die das Playbook als `{{ branch }}` liest. Das Feld ist ein Pflichtfeld.
