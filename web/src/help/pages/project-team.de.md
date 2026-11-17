---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Team

Die Mitglieder des Projekts und ihre Rollen. Eine Rolle legt fest, was ein Mitglied im Projekt
darf: Tasks ausführen, Vorlagen und andere Ressourcen ändern, das Team verwalten. Ein Mitglied kann
mehrere Rollen haben und erhält jede Berechtigung, die eine davon gewährt.

- **Neues Teammitglied** fügt einen Benutzer mit einer Rolle zum Projekt hinzu. Die Schaltfläche
  sehen Mitglieder, die das Team verwalten dürfen.
- Die Spalte **Rolle** ändert die Rollen eines Mitglieds sofort, ohne Speichern. Ein Mitglied, das
  das Team nicht verwalten darf, sieht die Rollen, kann sie aber nicht ändern.
- Eine Rolle mit Schloss stammt aus einer Gruppe des Benutzers in LDAP oder OIDC, die der
  Administrator dieser Rolle zugeordnet hat. Sie lässt sich hier nicht entfernen: Sie entfällt, wenn
  der Benutzer die Gruppe verlässt oder die Zuordnung entfernt wird.
- Das Projekt braucht mindestens einen Owner. Semaphore verweigert es, den letzten Owner zu
  entfernen oder ihm die Rolle Owner zu nehmen, außer ein Administrator von Semaphore tut es.
- **Projekt verlassen** nimmt Sie aus dem Projekt. Die Schaltfläche erscheint nur, wenn der
  Administrator Mitgliedern das Verlassen von Projekten erlaubt; auch der letzte Owner kann das
  Projekt nicht verlassen.
