---
form: src/components/ChangePasswordForm.vue
---

# Passwort ändern

Legt ein neues Passwort für die Anmeldung bei Semaphore mit Benutzername und Passwort fest. Das
neue Passwort wird wirksam, wenn Sie auf **Speichern** klicken. Ein externer Benutzer, der sich
über LDAP oder OpenID Connect anmeldet, hat hier kein Passwort, das er ändern könnte.

## Current password {#changePassword-currentPassword}

Wird angezeigt, wenn Sie Ihr eigenes Passwort ändern.

Ihr aktuelles Passwort. Das Feld ist ein Pflichtfeld: Semaphore prüft es, bevor es das neue
festlegt, damit eine gestohlene Sitzung oder ein gestohlenes API-Token nicht in Ihr Passwort
umgewandelt werden kann. Der Server lehnt die Änderung ab, wenn das Passwort nicht übereinstimmt.

Ein Administrator oder ein Benutzerverwalter, der das Passwort eines anderen Benutzers
zurücksetzt, sieht dieses Feld nicht.

## Passwort {#changePassword-password}

Das neue Passwort. Das Feld ist ein Pflichtfeld. Klicken Sie auf das Augensymbol, um anzuzeigen,
was Sie eingegeben haben.

Semaphore speichert nur einen Hash des Passworts.
