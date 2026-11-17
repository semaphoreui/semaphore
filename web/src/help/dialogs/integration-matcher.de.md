---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Ein Matcher ist eine Bedingung für die eingehende Anfrage. Matcher entscheiden, welche
Integration eine Anfrage an den Alias des Projekts startet, sodass eine URL mehrere Integrationen
bedienen kann: Eine Integration läuft nur, wenn jeder ihrer Matcher zutrifft. Sie werden nur
verwendet, wenn **Projektalias verwenden** für die Integration eingeschaltet ist; ist es
eingeschaltet, wird die Integration nur über den Projektalias erreicht.

Eine über den Projektalias gefundene Integration ohne Matcher läuft nie: Fügen Sie mindestens
einen hinzu.

## Name {#integrationMatcherForm-name}

Der Name, unter dem der Matcher auf der Seite der Integration aufgeführt wird. Er ändert nicht,
was der Matcher prüft. Das Feld ist ein Pflichtfeld.

## Match on {#integrationMatcherForm-matchType}

Der Teil der Anfrage, den der Matcher liest:

- **Header**: der Wert eines Request-Headers.
- **Body**: der Body der Anfrage.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Wird angezeigt, wenn **Match on** auf **Body** steht.

Wie der Body gelesen wird:

- **JSON**: Der Body wird als JSON geparst, und **Key** ist der Pfad zu einem Wert darin.
- **String**: Der gesamte Body wird als Text verglichen, und **Key** wird nicht verwendet.

## Key {#integrationMatcherForm-key}

Was aus der Anfrage gelesen wird. Das Feld ist ein Pflichtfeld.

- Für **Header**: der Name des Headers, zum Beispiel `X-GitHub-Event`. Groß- und
  Kleinschreibung spielen keine Rolle.
- Für **Body** mit **JSON**: der Pfad zum Wert, dessen Teile durch Punkte verbunden sind; ein
  Element eines Arrays wird als `.[0]` geschrieben, zum Beispiel `repository.full_name` oder
  `commits.[0].id`.
- Für **Body** mit **String**: nicht verwendet, aber das Formular braucht einen Wert; geben Sie
  irgendetwas ein.

## Comparison Method {#integrationMatcherForm-method}

Wie der aus der Anfrage gelesene Wert mit **Value** verglichen wird:

- **==**: Die beiden sind gleich.
- **!=**: Die beiden unterscheiden sich. Ein Header oder ein Schlüssel, der in der Anfrage
  fehlt, gilt als verschieden.
- **Contains**: Der aus der Anfrage gelesene Wert enthält **Value**.

## Value {#integrationMatcherForm-value}

Der Text, mit dem verglichen wird, zum Beispiel `push` oder `refs/heads/main`. Der Vergleich ist
exakt und unterscheidet Groß- und Kleinschreibung. Eine Zahl oder ein boolescher Wert des
JSON-Bodys wird in seiner Textform verglichen: `42`, `true`. Das Feld ist ein Pflichtfeld.
