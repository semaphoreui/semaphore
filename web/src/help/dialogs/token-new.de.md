---
form: src/views/Tokens.vue
---

# New Token

Ein API-Token lässt ein Skript oder einen anderen Dienst die Semaphore-API in Ihrem Namen
aufrufen, mit denselben Rechten wie Ihr Konto. Senden Sie es im Header `Authorization` einer
Anfrage: `Bearer` und das Token. Das Token wird einmal angezeigt, direkt nach seiner Erstellung:
Kopieren Sie es dann.

## Token Name {#token-name}

Optional. Ein Name, der das Token in der Liste der Tokens unterscheidbar macht, zum Beispiel der
Name des Skripts, das es verwendet. Er ist kein Teil des Tokens und ändert nicht, was das Token
darf.

## Expires {#token-expires}

Wann das Token aufhört zu funktionieren. Wählen Sie einen ab jetzt gezählten Zeitraum, **Never**
für ein Token, das funktioniert, bis es gelöscht wird, oder **Custom**, um den Zeitpunkt selbst
zu wählen.

Ein Token, das nie abläuft, funktioniert, bis Sie es aus der Liste löschen. Wählen Sie lieber
einen Zeitraum für ein Token, das Sie einem Skript oder einem anderen Dienst geben.

## Expiration date {#token-expiresAt}

Wird angezeigt, wenn **Expires** auf **Custom** steht.

Datum und Uhrzeit, zu denen das Token aufhört zu funktionieren, in der Zeitzone Ihres Browsers.
Der Zeitpunkt muss in der Zukunft liegen: Der Server lehnt ein Token ab, das bereits abgelaufen
wäre.
