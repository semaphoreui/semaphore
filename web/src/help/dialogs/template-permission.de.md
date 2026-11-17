---
form: src/components/EditTemplatePermissionForm.vue
---

# Vorlagenrechte

Gibt einer Rolle zusätzliche Berechtigungen für genau diese Vorlage. Ein Mitglied mit dieser Rolle
erhält sie zusätzlich zu den Berechtigungen, die die Rolle im gesamten Projekt hat: Eine Freigabe
hier fügt nur hinzu, sie entzieht nie eine Berechtigung. Ein Mitglied mit mehreren Rollen erhält
jede Berechtigung, die einer von ihnen gewährt wird.

## Rolle {#templatePerm-roleSlug}

Die Rolle, die die Berechtigungen erhält. Die Liste zeigt die Rollen dieses Projekts, die
gemeinsamen Rollen und die eingebauten Rollen. Das Feld ist ein Pflichtfeld, und eine Rolle kann
für eine Vorlage nur eine Freigabe haben: Der Server lehnt eine zweite ab.

Die Rolle einer gespeicherten Freigabe kann nicht geändert werden. Der Server behält die Rolle bei
und speichert nur die Berechtigungen, was immer Sie hier wählen. Um die Berechtigungen einer
anderen Rolle zu geben, entfernen Sie diese Freigabe aus der Liste und fügen Sie eine neue hinzu.

## Berechtigungen {#templatePerm-permissionsTitle}

Was ein Mitglied mit dieser Rolle mit dieser Vorlage tun darf:

- **Can see this template**: Die Vorlage steht in der Liste der Vorlagen, mit ihren Tasks und
  deren Logs.
- **Can run this template**: Tasks aus der Vorlage starten und ihre Tasks stoppen, bestätigen
  oder ablehnen.
- **Can edit this template**: die Vorlage ändern und löschen, alle ihre Tasks stoppen und diese
  Berechtigungen verwalten.

Ausführen oder Bearbeiten schließt das Sehen der Vorlage ein: Wenn Sie eines davon anhaken, wird
**Can see this template** ebenfalls angehakt und kann nicht entfernt werden. Lassen Sie alle
Kästchen leer, um eine Freigabe zu speichern, die nichts gewährt.
