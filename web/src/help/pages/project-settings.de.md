---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Projekteinstellungen

Die Einstellungen des Projekts als Ganzes: sein Name, wie viele seiner Tasks gleichzeitig laufen
dürfen und ob es Benachrichtigungen versendet. Die Gefahrenzone unterhalb des Formulars enthält
die Aktionen, die das gesamte Projekt betreffen: sein Backup, seinen Cache und seine Löschung.
Änderungen im Formular werden nach **Speichern** wirksam.

## Projektname {#newProject-name}

Der Name, unter dem das Projekt im Projektmenü oben in der Seitenleiste aufgeführt wird. Das Feld
ist ein Pflichtfeld.

## Max. Anzahl paralleler Tasks {#newProject-maxParallelTasks}

Wie viele Tasks dieses Projekts gleichzeitig laufen dürfen. Ein Task, der über dem Limit
gestartet wird, wartet in der Warteschlange, bis einer der laufenden Tasks endet.

Lassen Sie das Feld leer oder setzen Sie es auf `0`, um kein eigenes Limit festzulegen. Das vom
Administrator festgelegte Limit des gesamten Servers gilt in beiden Fällen.

## Telegram Chat ID {#newProject-tg}

Optional. Der Telegram-Chat, der die Benachrichtigungen dieses Projekts empfängt. Ist das Feld
leer, gehen die Benachrichtigungen an den Chat, der für den gesamten Server festgelegt ist.

Das Feld wirkt nur, wenn der Administrator Telegram-Benachrichtigungen für den Server aktiviert
hat und Benachrichtigungen für dieses Projekt erlaubt sind.

## Benachrichtigungen für dieses Projekt erlauben {#newProject-alert}

Aktiviert Benachrichtigungen für die Tasks dieses Projekts. Wohin die Benachrichtigungen gehen,
legt der Administrator für den gesamten Server fest: E-Mail, Telegram, Slack, Rocket.Chat,
Microsoft Teams, DingTalk oder Gotify.

- Eine E-Mail wird versendet, wenn ein Task fehlschlägt, und zwar an die Mitglieder des Projekts,
  die den Task sehen dürfen und in ihrem Konto Benachrichtigungen aktiviert haben.
- Eine Nachricht an die Messenger wird versendet, wenn ein Task erfolgreich ist, fehlschlägt oder
  auf eine Bestätigung wartet.

Eine Task-Vorlage kann verhindern, dass ihre erfolgreichen oder ihre fehlgeschlagenen Tasks
Benachrichtigungen versenden.

## Test Alerts {#settings-testAlerts}

Sendet eine Testbenachrichtigung über jeden Kanal, den der Server aktiviert hat, sodass Sie
prüfen können, ob die Benachrichtigungen dieses Projekts ankommen.

Die Schaltfläche funktioniert nur, wenn **Benachrichtigungen für dieses Projekt erlauben**
angehakt und gespeichert ist. Haken Sie die Option an und klicken Sie zuerst auf **Speichern**.

## Projekt-Backup {#settings-exportProject}

Lädt eine JSON-Datei mit der Konfiguration des Projekts herunter: Task-Vorlagen, Inventorys,
Repositories, Variablengruppen, Schlüssel, Zeitpläne, Integrationen und alles andere, was darin
eingerichtet ist. Die Datei kann mit **Projekt wiederherstellen** im Projektmenü oben in der
Seitenleiste wieder in ein Projekt umgewandelt werden.

Die Datei lässt die Secrets der Schlüssel und die Secrets auf dem Tab **Secrets** der
Variablengruppen weg: Diese müssen nach einer Wiederherstellung erneut eingegeben werden. Alles
andere wird so kopiert, wie es ist, auch ein Passwort, das direkt in ein anderes Feld eingegeben
wurde: die Adresse eines Repositorys mit einem Passwort darin, der Text eines statischen Inventorys,
die Werte gewöhnlicher Variablen und von Umgebungsvariablen. Prüfen Sie die Datei auf solche Werte,
bevor Sie sie weitergeben. Der Verlauf der Tasks ist nicht Teil des Backups.

## Cache leeren {#settings-clearCache}

Löscht die Dateien, die Semaphore für dieses Projekt auf dem Server vorhält, etwa seine Kopien der
Repositories. Sie werden erneut heruntergeladen, wenn der nächste Task startet.

Verwenden Sie diese Aktion, wenn ein Task an einer veralteten oder beschädigten Kopie eines
Repositorys scheitert. Die Aktion kann nicht rückgängig gemacht werden, und sie lässt die
Einstellungen des Projekts unberührt.

## Projekt löschen {#settings-deleteProject}

Löscht das Projekt mit allem, was darin enthalten ist: Task-Vorlagen, den Verlauf der Tasks,
Inventorys, Repositories, Variablengruppen, Schlüssel und Zeitpläne. Semaphore fragt zuvor nach
einer Bestätigung.

Ein gelöschtes Projekt kann nicht wiederhergestellt werden. Laden Sie ein Backup herunter, bevor
Sie ein Projekt löschen, das Sie möglicherweise noch einmal benötigen.
