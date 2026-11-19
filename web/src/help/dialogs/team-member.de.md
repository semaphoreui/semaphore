---
form: src/components/TeamMemberForm.vue
---

# Neu Teammitglied

Fügt dem Projekt einen Benutzer mit einer Rolle hinzu. Die Rolle bestimmt, was das Mitglied im
Projekt sehen und tun darf; sonst ändert sich nichts an seinem Konto. Hat der Administrator
Einladungen eingeschaltet, erhält der Benutzer stattdessen eine Einladung und wird Mitglied,
sobald er sie annimmt.

## E-Mail oder Benutzername {#teamMember-inviteType}

Wird angezeigt, wenn Benutzer sowohl per E-Mail als auch per Benutzername eingeladen werden
können.

Wie Sie die Person auswählen: **Email**, um eine E-Mail-Adresse einzuladen, **Username**, um
einen Benutzer auszuwählen, der bereits ein Konto auf diesem Semaphore-Server hat.

## E-Mail {#teamMember-email}

Wird angezeigt, wenn die Person per E-Mail eingeladen wird.

Die E-Mail-Adresse der einzuladenden Person. Sie wird Mitglied des Projekts, sobald sie die
Einladung annimmt.

## Benutzer {#teamMember-userId}

Wird angezeigt, wenn die Person per Benutzername ausgewählt wird.

Der hinzuzufügende Benutzer, aufgeführt mit Benutzername und vollständigem Namen. Das Feld ist ein
Pflichtfeld. Die Liste enthält jeden Benutzer des Semaphore-Servers, der noch kein Mitglied des
Projekts ist: Eine Person ohne Konto muss zuerst eines von einem Administrator erhalten.

## Rolle {#teamMember-role}

Was das Mitglied im Projekt tun darf. Das Feld ist ein Pflichtfeld. Die eingebauten Rollen:

- **Owner**: alles, einschließlich der Einstellungen des Projekts, seines Teams und seiner
  Löschung.
- **Manager**: führt Tasks und Workflows aus, gibt Workflows frei und erstellt, ändert und löscht
  Task-Vorlagen, Inventorys, Schlüssel, Repositories, Variablengruppen, Zeitpläne, Integrationen
  und Workflows. Ein Manager kann die Einstellungen des Projekts und sein Team nicht ändern.
- **Task Runner**: sieht das Projekt, führt Tasks und Workflows aus und gibt Workflows frei. Ein
  Task Runner kann im Projekt nichts ändern.
- **Guest**: sieht das Projekt, seine Tasks und deren Logs und ändert nichts.

Die Liste enthält außerdem die für dieses Projekt erstellten Rollen (**Projektrollen**) und die
von jedem Projekt gemeinsam genutzten Rollen (**Shared roles**). Eine solche Rolle gewährt die
Berechtigungen, die auf der Seite Rollen für sie festgelegt sind. Eine Einladung kann nur eine
eingebaute Rolle enthalten.

Ein Mitglied, das das Team verwalten darf, kann jede Rolle vergeben, auch **Owner**. Der Dialog
vergibt eine Rolle; fügen Sie danach in der Spalte **Rolle** der Teamliste weitere hinzu.
