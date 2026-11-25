---
form: src/components/RunnerForm.vue
---

# Runner bearbeiten

Ein Runner ist ein eigener Prozess `semaphore runner`, der Tasks auf seinem eigenen Rechner statt
auf dem Semaphore-Server ausführt. Ein Task geht an einen Runner, wenn der Server so eingestellt
ist, dass er Tasks auf Runnern ausführt, oder wenn die Task-Vorlage oder ihr Inventory einen
Runner-Tag hat. Der Server wählt dann einen Runner, der aktiviert, registriert und online ist und
sein Limit paralleler Tasks nicht erreicht hat: zuerst Runner des Projekts, dann globale Runner,
innerhalb jeder Gruppe in zufälliger Reihenfolge. Wenn jeder passende Runner offline oder
ausgelastet ist, wartet der Task in der Warteschlange; wenn überhaupt kein Runner passt, schlägt
der Task fehl.

## Name {#runner-name}

Pflichtfeld. Der Name, unter dem der Runner auf der Seite Runner aufgeführt wird, zum Beispiel
der Host, auf dem er läuft. Er ändert nicht, welche Tasks der Runner erhält.

## Tags {#runner-tags}

Verfügbar in Semaphore Pro; in anderen Editionen ist das Feld gesperrt.

Die Tags, die die Tasks dieses Runners auswählen. Ein Task mit einem Runner-Tag geht nur an die
Runner, die diesen Tag tragen; der Tag der Task-Vorlage hat Vorrang vor dem Tag des Inventorys.
Geben Sie einen Tag ein und drücken Sie Enter, oder wählen Sie einen, den andere Runner
verwenden. Leere und doppelte Tags werden beim Speichern entfernt.

Ein Projekt-Runner braucht mindestens einen Tag, es sei denn, **Is default** ist angehakt.

## Is default {#runner-isDefault}

Ein Task ohne Runner-Tag geht nur an die Runner, die als Standard markiert sind. Entfernen Sie den
Haken für einen Runner, der nur Tasks mit einem seiner Tags übernehmen soll.

Bei einem neuen globalen Runner ist das Kästchen angehakt, bei einem neuen Projekt-Runner nicht.

## Register {#runner-registered}

Wird angezeigt, wenn Sie einen Runner erstellen.

- Angehakt (die Voreinstellung): Der Server erstellt den Runner mit seinem Token. Nach
  **Erstellen** zeigt der Dialog das Token und die Befehle, die den Runner damit starten.
- Nicht angehakt: Der Runner wird ohne Token erstellt. Klicken Sie später in der Zeile des Runners
  auf **Regenerate registration token** (das Symbol mit zwei Pfeilen), um ein einmaliges
  Registrierungstoken zu erhalten, das eine Stunde gültig ist, und registrieren Sie den Runner mit
  `semaphore runner register`.

Ein nicht registrierter Runner erhält keine Tasks.

## Webhook {#runner-webhook}

Optional. Eine URL, die der Server aufruft, wenn er diesem Runner einen Task übergibt, zum
Beispiel um den Runner bei Bedarf zu starten. Der Server sendet eine `POST`-Anfrage mit einem
JSON-Body:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

Die URL muss innerhalb von 30 Sekunden mit dem Status 200 oder 204 antworten; andernfalls schlägt
der Task fehl. Ein Runner mit einem Webhook gilt immer als online, auch wenn er den Server in
letzter Zeit nicht abgefragt hat. Lassen Sie das Feld leer für einen Runner, der ständig läuft.

## Max. Anzahl paralleler Tasks {#runner-maxParallelTasks}

Die Anzahl der Tasks, die der Runner gleichzeitig ausführen darf. Ist das Limit erreicht, geht
der nächste Task an einen anderen passenden Runner oder wartet in der Warteschlange. Lassen Sie
das Feld leer oder setzen Sie `0` für kein Limit.

## Aktiviert {#runner-active}

Nur ein aktivierter Runner erhält neue Tasks. Entfernen Sie den Haken, um den Runner außer
Betrieb zu nehmen, ohne ihn zu löschen; der Schalter in der Liste der Runner bewirkt dasselbe.

Bei einem neuen Runner ist das Kästchen nicht angehakt: Haken Sie es an oder aktivieren Sie den
Runner später in der Liste.
