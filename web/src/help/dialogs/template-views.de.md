---
form: src/components/EditViewsForm.vue
---

# Ansichten bearbeiten

Ansichten sind die Tabs über der Liste der Task-Vorlagen. Eine Ansicht zeigt die Vorlagen, die Sie
ihr zuordnen, oder jede Vorlage des Projekts. Der Dialog speichert jede Änderung sofort: Es gibt
keine Speichern-Schaltfläche für den gesamten Dialog. Ziehen Sie eine Ansicht an ihrem Griff, um
die Reihenfolge der Tabs zu ändern.

## Titel {#viewsForm-title}

Der Name des Tabs. Geben Sie ihn ein und bestätigen Sie ihn mit dem Häkchen neben dem Feld; das
Kreuz stellt den gespeicherten Titel wieder her. Eine neue Ansicht wird gespeichert, sobald sie
einen Titel hat und Sie ihn bestätigen oder eine ihrer Einstellungen ändern. Das Schließen des
Dialogs speichert keinen Titel, den Sie nicht bestätigt haben.

Das Zahnrad öffnet die Einstellungen der Ansicht unter ihrem Titel.

## Löschen {#viewsForm-viewDelete}

Das Papierkorbsymbol löscht die Ansicht sofort, ohne Nachfrage. Die Vorlagen der Ansicht werden
nicht gelöscht: Sie bleiben im Projekt und sind in keiner Ansicht mehr.

## Type {#viewsForm-type}

Welche Vorlagen der Tab zeigt:

- **Custom**: die Vorlagen, deren Feld **Ansicht** im Vorlagendialog diese Ansicht nennt. Das ist
  die Voreinstellung.
- **All**: jede Vorlage des Projekts, unabhängig von ihrer Ansicht.

Ein Mitglied sieht in jeder Ansicht nur die Vorlagen, die es sehen darf.

## Sort by {#viewsForm-sortColumn}

Die Spalte, nach der die Vorlagen des Tabs sortiert werden; **Name** ist die einzige. Ein Tab
ohne diese Angabe wird ebenfalls nach Namen sortiert, daher ist das Feld zusammen mit
**Reverse** von Bedeutung.

## Reverse {#viewsForm-sortReverse}

Sortiert den Tab in absteigender Reihenfolge. Das funktioniert nur, wenn **Sort by** gesetzt ist.

## Hidden {#viewsForm-hidden}

Blendet den Tab für Mitglieder aus, die die Ressourcen des Projekts nicht verwalten dürfen.
Diejenigen, die es dürfen, sehen ihn weiterhin, durchgestrichen. Das Ausblenden eines Tabs blendet
seine Vorlagen nicht aus: Eine Ansicht vom Typ **All** listet sie weiterhin auf.

## Ansicht hinzufügen {#viewsForm-add}

Fügt am Ende der Liste eine Ansicht mit einem leeren Titel zum Eintippen hinzu.
