---
form: src/components/EditRoleForm.vue
---

# Rolle bearbeiten

Eine Rolle ist ein benannter Satz von Berechtigungen innerhalb eines Projekts. Sie vergeben Rollen
an die Mitglieder eines Projekts, und ein Mitglied darf tun, was eine seiner Rollen erlaubt. Von
der Seite Rollen eines Projekts geöffnet, bearbeitet der Dialog nur eine Rolle dieses Projekts;
von der Seite Rollen außerhalb eines Projekts geöffnet, bearbeitet er eine gemeinsame Rolle, die
jedes Projekt verwenden kann. Die eingebauten Rollen Owner, Manager, Task Runner und Guest können
nicht geändert werden.

## Name {#role-name}

Der Name, unter dem die Rolle auf der Seite Rollen und überall dort aufgeführt wird, wo eine Rolle
ausgewählt wird, zum Beispiel wenn Sie einem Projekt ein Mitglied hinzufügen. Das Feld ist ein
Pflichtfeld. Sie können eine Rolle jederzeit umbenennen.

## Slug {#role-slug}

Die Kennung der Rolle. Semaphore füllt sie aus dem Namen aus, während Sie den Namen eingeben, bis
Sie den Slug selbst bearbeiten. Er besteht aus 1 bis 100 Buchstaben, Ziffern, Unterstrichen und
Bindestrichen, zum Beispiel `release-managers`.

Der Slug kann nach dem Speichern der Rolle nicht mehr geändert werden: Das Feld ist gesperrt, wenn
Sie eine Rolle bearbeiten.

- Ein Slug muss eindeutig sein, und die Prüfung ignoriert Groß- und Kleinschreibung. Der Slug
  einer gemeinsamen Rolle muss sich von jeder anderen Rolle unterscheiden; der Slug einer
  Projektrolle muss sich von den gemeinsamen Rollen und von den anderen Rollen seines Projekts
  unterscheiden.
- `owner`, `manager`, `task_runner`, `guest` und `all` sind reserviert, in jeder Schreibweise.

## Berechtigungen {#role-permissionsTitle}

Was die Mitglieder, die die Rolle haben, im Projekt tun dürfen. Bei einer neuen Rolle ist nichts
angehakt. Die Berechtigungen aller Rollen eines Mitglieds addieren sich; ein Administrator von
Semaphore darf in jedem Projekt alles, unabhängig von den Rollen.

- **Can read project**: das Projekt sehen: seine Task-Vorlagen, Tasks und deren Logs, Inventorys,
  Schlüssel, Repositories, Variablengruppen, Zeitpläne und Mitglieder. Jeder Lesezugriff im
  Projekt verlangt diese Berechtigung, und keine andere Berechtigung schließt sie ein, haken Sie
  sie also für fast jede Rolle an. Ohne sie sieht ein Mitglied nur die Task-Vorlagen und
  Workflows, die der Rolle selbst Zugriff gewähren, mit ihren Tasks.
- **Kann Projekt-Tasks ausführen**: Tasks aus Task-Vorlagen starten, sie stoppen und einen Task,
  der auf eine Bestätigung wartet, bestätigen oder ablehnen.
- **Kann Projekt aktualisieren**: die Einstellungen des Projekts ändern und es löschen.
- **Kann Ressourcen verwalten**: Task-Vorlagen, Inventorys, Schlüssel, Repositories,
  Variablengruppen, Zeitpläne, Ansichten, Integrationen, Secret-Speicher, Projekt-Runner und die
  Rollen des Projekts erstellen, ändern und löschen; Tasks löschen; festlegen, welche Rollen eine
  Task-Vorlage verwenden dürfen.
- **Kann Benutzer verwalten**: dem Projekt Mitglieder hinzufügen, ihre Rollen ändern und sie
  entfernen.
- **Can run workflows**: einen Workflow starten und seine Läufe stoppen.
- **Can edit workflows**: Workflows erstellen, ändern und löschen und festlegen, welche Rollen
  einen Workflow verwenden dürfen.
- **Can approve workflows**: einen Schritt eines Workflow-Laufs, der auf eine Freigabe wartet,
  freigeben oder ablehnen.
