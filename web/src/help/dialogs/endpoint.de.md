---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Endpunkt bearbeiten

Ein Endpunkt ist ein Einstiegspunkt für Webhooks: die URLs, die andere Dienste aufrufen, und die
Art, wie ihre Anfragen geprüft werden. Integrationen hängen an einem Endpunkt und starten ihre Läufe
bei den Anfragen, die er annimmt. Die URLs eines Endpunkts, seine Aliase, werden auf dem Tab
„Endpunkte“ hinzugefügt.

## Name {#endpointForm-name}

Der Name, unter dem der Endpunkt auf dem Tab „Endpunkte“ und im Formular einer Integration
aufgeführt wird. Das Feld ist erforderlich.

## Beschreibung {#endpointForm-description}

Optional. Eine Notiz dazu, wofür der Endpunkt gedacht ist; sie wird neben seinem Namen angezeigt.

## Auth-Methode {#integrationAuthForm-authMethod}

Wie Semaphore prüft, dass eine Anfrage von dem erwarteten Dienst kommt. Eine Anfrage, die die
Prüfung nicht besteht, startet nichts, und der Webhook beantwortet sie genauso wie jede andere
Anfrage.

- **Keine**: keine Prüfung. Jeder, der die URL kennt, kann den Task starten.
- **GitHub Webhooks**: Der Header `X-Hub-Signature-256` muss die mit dem Secret erstellte
  HMAC-SHA256-Signatur des Bodys enthalten, so wie GitHub sie sendet.
- **Bitbucket Webhooks**: dieselbe Prüfung des Headers `X-Hub-Signature`, so wie Bitbucket ihn
  sendet.
- **Token**: Der in **Auth-Header** angegebene Header muss das Secret selbst enthalten.
- **HMAC** und **HMAC (SHA-512)**: Der in **Auth-Header** angegebene Header muss den mit dem
  Secret erstellten HMAC-SHA256 oder HMAC-SHA512 des Bodys enthalten, in hexadezimalen
  Kleinbuchstaben, ohne ein Präfix wie `sha256=`.
- **BasicAuth**: Die Anfrage muss den Login und das Passwort des Schlüssels aus
  **Auth-Secret** in ihrem Header `Authorization` mitführen.

## Auth-Header {#integrationAuthForm-authHeader}

Wird angezeigt, wenn **Auth-Methode** auf **Token**, **HMAC** oder **HMAC (SHA-512)** steht.

Der Name des Request-Headers, der das Token oder die Signatur enthält, zum Beispiel
`X-Webhook-Token`. Groß- und Kleinschreibung des Namens spielen keine Rolle. Eine Anfrage ohne
diesen Header wird abgelehnt.

## Auth-Secret {#integrationAuthForm-authSecretId}

Wird angezeigt, wenn **Auth-Methode** nicht auf **Keine** steht.

Der Schlüssel aus dem Schlüsselspeicher, gegen den die Anfrage geprüft wird. Für **BasicAuth**
passt nur ein Schlüssel vom Typ **Login mit Passwort**. Für die anderen Methoden passt auch ein
Schlüssel vom Typ **Zeichenkette**: Die Prüfung verwendet seinen Wert oder das Passwort eines
Schlüssels vom Typ **Login mit Passwort**.

Das Feld wird von jeder Methode außer **Keine** benötigt: Ohne Schlüssel oder mit einem leeren
Secret wird jede Anfrage abgelehnt.
