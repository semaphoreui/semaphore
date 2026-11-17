---
form: src/components/WorkflowRunDialog.vue
---

# Workflow ausführen

Startet einen Lauf des Workflows, sobald Sie seinen Survey beantwortet haben. Der Dialog öffnet
sich nur bei einem Workflow mit Survey: Ein Workflow ohne Survey startet sofort. Nach
**Ausführen** öffnet sich die Seite des neuen Laufs; lehnt der Server die Werte ab, bleibt der
Dialog mit Ihren Antworten offen.

## Variablen {#workflowRunDialog-survey}

Ein Feld pro Variable des Surveys des Workflows, in der Reihenfolge des Surveys, mit eingetragenem
Standardwert. Ein mit `*` markiertes Feld ist ein Pflichtfeld, ein Integer-Feld nimmt nur Ziffern
an, und ein Listenfeld nimmt einen Wert oder, wenn die Variable mehrere erlaubt, beliebig viele
an. **Ausführen** startet nichts, bis jedes Feld gültig ist.

Die Werte werden mit dem Lauf gespeichert und auf seiner Seite unter **Variables** angezeigt. Ein
Task des Workflows erhält sie nicht von selbst: Eine Variable seiner Vorlage erhält nur dann einen
Wert, wenn ihr Knoten sie an eine Variable des Workflows bindet. Der Name allein liefert nichts.

Ein Workflow-Survey enthält keine Secrets, da der Lauf seine Werte im Klartext aufbewahrt. Legen
Sie Passwörter und Tokens stattdessen in einem Schlüssel oder einer Variablengruppe ab.
