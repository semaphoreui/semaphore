---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Integration bearbeiten

Eine Integration ist eine Gruppe von Läufen, die von einem Endpunkt aus starten. Wenn eine Anfrage
an eine URL dieses Endpunkts die Autorisierung besteht, startet jeder Lauf der Integration, dessen
Matcher die Anfrage annehmen.

## Name {#integrationGroupForm-name}

Der Name, unter dem die Integration auf der Seite „Integrationen“ aufgeführt wird. Das Feld ist
erforderlich.

## Endpunkt {#integrationGroupForm-endpointId}

Der Endpunkt, dessen URLs die Läufe dieser Integration starten. Mehrere Integrationen können einen
Endpunkt verwenden. Das Feld ist erforderlich.

## Autorisierung überschreiben {#integrationGroupForm-authOverride}

Lassen Sie die Option aus, damit Anfragen so geprüft werden wie beim Endpunkt. Schalten Sie sie ein,
um die Anfragen dieser Integration mit einer eigenen Methode und einem eigenen Secret zu prüfen, die
in den darunter erscheinenden Feldern festgelegt werden; die Einstellungen des Endpunkts gelten dann
nicht für diese Integration.

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
