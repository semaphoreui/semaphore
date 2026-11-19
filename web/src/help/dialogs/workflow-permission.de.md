---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Gibt einer Rolle zusätzliche Berechtigungen für genau diesen Workflow. Ein Mitglied mit dieser
Rolle erhält sie zusätzlich zu den Berechtigungen, die die Rolle im gesamten Projekt hat: Eine
Freigabe hier fügt nur hinzu, sie entzieht nie eine Berechtigung. Ein Mitglied mit mehreren
Rollen erhält jede Berechtigung, die einer von ihnen gewährt wird.

## Rolle {#templatePerm-roleSlug}

Die Rolle, die die Berechtigungen erhält. Die Liste zeigt die Rollen dieses Projekts, die
gemeinsamen Rollen und die eingebauten Rollen. Das Feld ist ein Pflichtfeld, und eine Rolle kann
für einen Workflow nur eine Freigabe haben: Der Server lehnt eine zweite ab.

Die Rolle einer gespeicherten Freigabe kann nicht geändert werden. Der Server behält die Rolle bei
und speichert nur die Berechtigungen, was immer Sie hier wählen. Um die Berechtigungen einer
anderen Rolle zu geben, entfernen Sie diese Freigabe aus der Liste und fügen Sie eine neue hinzu.

## Berechtigungen {#templatePerm-permissionsTitle}

Was ein Mitglied mit dieser Rolle mit diesem Workflow tun darf:

- **Can see this workflow**: Der Workflow steht in der Liste der Workflows, mit seinen Läufen.
- **Can run this workflow**: Läufe des Workflows starten.
- **Can edit this workflow**: den Workflow ändern und löschen und diese Berechtigungen verwalten.
- **Can approve this workflow**: die Freigabeschritte seiner Läufe freigeben oder ablehnen.

Ausführen, Bearbeiten oder Freigeben schließt das Sehen des Workflows ein: Wenn Sie eines davon
anhaken, wird **Can see this workflow** ebenfalls angehakt und kann nicht entfernt werden. Lassen
Sie alle Kästchen leer, um eine Freigabe zu speichern, die nichts gewährt.
