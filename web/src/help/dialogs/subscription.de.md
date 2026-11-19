---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

Der Subscription-Schlüssel schaltet die Funktionen von Semaphore PRO und Enterprise auf diesem
Server frei. Der Dialog aktiviert einen Schlüssel und zeigt den Plan des aktiven Schlüssels:
seinen Inhaber, seinen Status, wann er abläuft und wie viele Pro-Benutzer, Runner und andere
Elemente er erlaubt und wie viele davon verwendet werden. Ein Administrator öffnet ihn über das
Benutzermenü der Pro-Edition. Der Dialog heißt **Upgrade to Semaphore PRO**, bis eine
Subscription aktiv ist.

## Enter your PRO or EE key {#subscription-key}

Der Subscription-Schlüssel, so wie Sie ihn mit Ihrer PRO- oder Enterprise-Subscription erhalten
haben. Fügen Sie ihn hier ein oder laden Sie ihn mit der Upload-Schaltfläche aus einer Datei, und
klicken Sie dann auf **Activate New Key**. Das Feld ist für die Aktivierung ein Pflichtfeld.

Das Feld ist schreibgeschützt, wenn der Schlüssel in der Konfiguration des Servers festgelegt ist
(`subscription.key` oder `subscription.key_file` oder die Variablen
`SEMAPHORE_SUBSCRIPTION_KEY` und `SEMAPHORE_SUBSCRIPTION_KEY_FILE`): Der Schlüssel kann dann nur
dort geändert werden.

## Schlüsseldatei hochladen {#subscription-upload}

Wird angezeigt, solange noch keine Subscription aktiviert wurde.

Liest den Schlüssel aus einer Datei auf Ihrem Computer (`.txt`, `.key`, `.pem` oder `.lic`) in
das Feld darüber ein. Der Schlüssel wird erst aktiviert, wenn Sie auf **Activate New Key**
klicken.

## Schlüsselaktionen {#subscription-actionsMenu}

Wird angezeigt, wenn eine Subscription aktiv oder abgelaufen ist.

Ein Menü mit Aktionen für den Schlüssel des Servers:

- **Reload**: ruft den Zustand der Subscription erneut ab, zum Beispiel nach einer Verlängerung.
- **Upload**: liest einen neuen Schlüssel aus einer Datei in das Feld darüber ein, wie die
  Upload-Schaltfläche.
- **Reset**: entfernt den Schlüssel sofort vom Server, ohne Bestätigung.

## Activate New Key {#subscription-activate}

Aktiviert den Schlüssel aus dem Feld auf diesem Server. Ist der Schlüssel gültig, wird der Plan
der Subscription darunter angezeigt. Ist Ihr Konto noch kein Pro-Benutzer und hat der Plan einen
freien Platz für einen Pro-Benutzer, fragt der Dialog, ob Ihr Konto zu einem Pro-Benutzer gemacht
werden soll.

Ein bereits aktivierter Schlüssel wird als Warnung gemeldet. Die Schaltfläche ist deaktiviert,
wenn der Schlüssel in der Konfiguration des Servers festgelegt ist.

## Buy Pro {#subscription-buy}

Öffnet in einem neuen Tab des Browsers die Seite des Semaphore-Portals, auf der eine
PRO-Subscription gekauft wird.

## Get 30-day free trial {#subscription-trial}

Wird angezeigt, solange keine Subscription aktiv ist.

Öffnet in einem neuen Tab des Browsers die Seite des Semaphore-Portals, die einen Schlüssel für
eine kostenlose Testversion von 30 Tagen ausstellt. Fügen Sie den erhaltenen Schlüssel in das
Feld darüber ein.
