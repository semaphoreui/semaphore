---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Eine globale Rolle gewährt den Benutzern, denen sie zugewiesen ist, Rechte über die gesamte
Semaphore-Instanz; Sie weisen sie im Dialog eines Benutzers zu. Sie gibt keinen Zugriff innerhalb
von Projekten: Dieser ergibt sich nur aus den Rollen, die ein Benutzer als Mitglied eines Projekts
hat. Ein Administrator hat jedes globale Recht, unabhängig davon, welche Rollen zugewiesen sind.
Globale Rollen sind eine Pro-Funktion.

## Name {#globalRole-name}

Der Name, unter dem die Rolle auf dem Tab Globale Rollen der Seite Benutzer und im Dialog eines
Benutzers aufgeführt wird. Das Feld ist ein Pflichtfeld. Sie können eine Rolle jederzeit
umbenennen.

## Slug {#globalRole-slug}

Die Kennung der Rolle. Semaphore füllt sie aus dem Namen aus, während Sie den Namen eingeben, bis
Sie den Slug selbst bearbeiten. Er besteht aus 1 bis 100 Buchstaben, Ziffern, Unterstrichen und
Bindestrichen, zum Beispiel `user-admins`.

Der Slug muss unter den globalen Rollen eindeutig sein; eine Projektrolle darf denselben Slug
haben. Er kann nach dem Speichern der Rolle nicht mehr geändert werden: Das Feld ist gesperrt,
wenn Sie eine Rolle bearbeiten.

## Global permissions {#globalRole-permissionsTitle}

Was ein Benutzer mit dieser Rolle in der gesamten Instanz tun darf. Bei einer neuen Rolle ist
nichts angehakt. Die Rechte aller globalen Rollen eines Benutzers addieren sich.

- **Can manage users**: Benutzer erstellen, ändern und löschen und ihre Details in der Liste der
  Benutzer sehen. Ein solcher Benutzer kann keinen Benutzer zum Administrator oder Pro-Benutzer
  machen und kann weder einen Administrator noch einen Benutzer ändern oder löschen, der ein
  globales Recht hat, das er selbst nicht hat.
- **Can manage global roles**: die globalen Rollen und die gemeinsamen Projektrollen erstellen,
  ändern und löschen und festlegen, welche Benutzer welche globale Rolle haben. Geben Sie dieses
  Recht nur Personen, denen Sie die gesamte Instanz anvertrauen.
- **Can manage system settings**: die Einstellungen der Instanz, die Anwendungen, die globalen
  Runner und die Gruppenzuordnungen ändern; den Cluster und die Informationen über den Server
  sehen; die Tasks aller Projekte sehen und entfernen; den Cache des Servers leeren.
- **Can read the audit log**: das Audit-Log der Instanz lesen und exportieren.
