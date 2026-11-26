---
form: src/components/ProjectForm.vue
---

# Neues Projekt

Ein Projekt enthält die Task-Vorlagen, Inventorys, Repositories, Variablengruppen und Schlüssel
eines Arbeitsbereichs sowie die Personen, die daran arbeiten. Sie werden Eigentümer des Projekts,
das Sie erstellen. Nur ein Administrator kann ein Projekt erstellen, es sei denn, der
Administrator des Servers erlaubt es jedem Benutzer. Jede der folgenden Einstellungen kann später
auf der Seite Einstellungen des Projekts geändert werden.

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

## Demo {#newProject-demo}

Füllt das neue Projekt mit Beispielen, an denen Sie Semaphore ausprobieren können: das
öffentliche Repository `https://github.com/semaphoreui/semaphore-demo.git`, Inventorys, ein
Schlüssel mit einem Vault-Passwort und Task-Vorlagen für Ansible, OpenTofu, Terragrunt, Bash und
PowerShell, einsortiert in die Ansichten **Build**, **Deploy** und **Tools**.

Lassen Sie den Schalter aus für ein Projekt ohne Beispiele: Es erhält nur den Schlüssel **None**
und die Ansicht **All**. Die Beispiele sind gewöhnliche Elemente des Projekts: Bearbeiten oder
löschen Sie sie nach Belieben.
