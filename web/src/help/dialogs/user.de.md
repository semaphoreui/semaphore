---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Benutzer bearbeiten

Ein Benutzer ist ein Konto, das sich bei Semaphore anmeldet. Der Tab **Settings** enthält das
Profil des Kontos und seine Optionen; **Speichern** schreibt sie. Der Tab **Security** enthält das
Passwort, die Zwei-Faktor-Authentifizierung und die verknüpften Anmeldekonten, und der Tab
**Roles** die globalen Rollen des Benutzers: Beide wirken sofort, ohne **Speichern**. Was ein
Benutzer innerhalb eines Projekts tun darf, wird auf der Seite **Team** dieses Projekts
festgelegt, nicht hier.

## Name {#user-name}

Der vollständige Name des Benutzers, angezeigt in der Seitenleiste und in den Listen der Benutzer
und der Projektmitglieder. Das Feld ist ein Pflichtfeld.

Bei einem Benutzer mit **Extern** übernimmt Semaphore den Namen bei der Anmeldung vom
Identity-Provider, wenn der Provider einen sendet, sodass eine hier vorgenommene Änderung die
nächste Anmeldung möglicherweise nicht überdauert.

## Benutzername {#user-username}

Der Login des Benutzers. Das Feld ist ein Pflichtfeld. Der Benutzer meldet sich mit dem
Benutzernamen oder der E-Mail und dem Passwort an.

Der Benutzername eines Benutzers mit **Extern** kann nicht geändert werden: Der Server lehnt das
Speichern ab.

## E-Mail {#user-email}

Die E-Mail-Adresse des Benutzers. Das Feld ist ein Pflichtfeld. Semaphore sendet die
Task-Benachrichtigungen des Benutzers an sie, und eine Authenticator-App zeigt sie als Namen des
Kontos an. Sie ist privat: Nur der Benutzer und diejenigen, die Benutzer verwalten dürfen, sehen
sie.

Bei einem Benutzer mit **Extern** stammt die E-Mail vom Identity-Provider: Das Feld ist
schreibgeschützt, und Semaphore aktualisiert die Adresse bei jeder Anmeldung.

## Passwort {#user-password}

Wird angezeigt, wenn ein neuer Benutzer erstellt wird.

Das Passwort, mit dem sich der Benutzer anmeldet. Das Feld ist ein Pflichtfeld, außer bei einem
Benutzer mit **Extern**, der in Semaphore kein Passwort hat. Um das Passwort eines bestehenden
Benutzers zu ändern, verwenden Sie **Change password** auf dem Tab **Security**.

## Benachrichtigungen senden {#user-alert}

Sendet dem Benutzer eine E-Mail, wenn ein Task in einem Projekt fehlschlägt, in dem der Benutzer
den Task sehen darf. Das Projekt muss Benachrichtigungen erlauben, und der Administrator des
Servers muss E-Mail-Benachrichtigungen eingeschaltet haben. Eine Task-Vorlage kann verhindern,
dass ihre fehlgeschlagenen Tasks Benachrichtigungen versenden.

Die Option wirkt sich nicht auf die Benachrichtigungen eines Projekts aus, die an Telegram, Slack
und die anderen Messenger gesendet werden.

## Admin-Benutzer {#user-admin}

Wird angezeigt, wenn Sie Administrator sind.

Ein Administrator darf auf dem Server alles: die Benutzer, die globalen Rollen und die
Einstellungen des Servers verwalten und jedes Projekt mit vollen Rechten öffnen.

Sie können Ihrem eigenen Konto diese Option nicht entziehen, und der Server behält mindestens
einen Administrator: Er lehnt es ab, sie dem letzten zu entziehen.

## Pro user {#user-pro}

Wird in der Pro-Edition angezeigt. Nur ein Administrator kann die Option ändern.

Gibt dem Benutzer einen Platz der Pro-Subscription. Die Subscription hat eine feste Anzahl von
Pro-Benutzern: Ist sie erreicht, lehnt der Server das Speichern der Option ab. Ohne aktive
Subscription wirkt die Option nicht.

## Extern {#user-external}

Wird angezeigt, wenn Sie Administrator sind. Kann nur beim Erstellen des Benutzers gesetzt
werden.

Ein externer Benutzer meldet sich über LDAP oder OpenID Connect an, nicht mit einem Passwort von
Semaphore: Das Passwortfeld ist deaktiviert, und eine Anmeldung mit Passwort wird abgelehnt. Der
Name und die E-Mail werden bei jeder Anmeldung vom Identity-Provider übernommen. Lassen Sie das
Kästchen leer für einen Benutzer, der sich mit einem Passwort anmeldet.

## Change password {#user-changePassword}

Wird auf dem Tab **Security** eines bestehenden Benutzers angezeigt, der nicht **Extern** ist,
wenn der Server die Anmeldung mit Passwort erlaubt.

Öffnet einen Dialog, der ein neues Passwort festlegt. Um Ihr eigenes Passwort zu ändern, müssen
Sie das aktuelle angeben; um das Passwort eines anderen Benutzers zurückzusetzen, nicht.

## Time-based one-time password {#user-totp}

Wird auf dem Tab **Security** angezeigt, wenn der Administrator die Zwei-Faktor-Authentifizierung
für den Server eingeschaltet hat.

Schaltet die Zwei-Faktor-Authentifizierung für diesen Benutzer ein. Wenn Sie den Schalter
einschalten, erzeugt Semaphore sofort ein Secret und zeigt es als QR-Code an: Scannen Sie ihn mit
einer Authenticator-App. Von da an fragt eine Anmeldung nach dem Passwort den Code der App ab.

Wenn der Server eine Wiederherstellung erlaubt, wird unter dem QR-Code ein Wiederherstellungscode
angezeigt. Er wird nur einmal angezeigt: Kopieren Sie ihn und bewahren Sie ihn sicher auf. Das
Ausschalten des Schalters entfernt das Secret sofort.

## LDAP-Provider {#user-ldapProvider}

Wird auf dem Tab **Security** Ihres eigenen Kontos angezeigt, wenn der Server einen LDAP-Provider
hat, mit dem Ihr Konto noch nicht verknüpft ist.

Der LDAP-Server, mit dem Ihr Konto verknüpft werden soll, wenn es mehrere gibt.

## LDAP username {#user-ldapUsername}

Wird mit **LDAP-Provider** angezeigt.

Ihr Benutzername auf dem LDAP-Server. Die Verknüpfung erlaubt Ihnen, sich auch über LDAP bei
diesem Konto anzumelden.

## LDAP password {#user-ldapPassword}

Wird mit **LDAP-Provider** angezeigt.

Ihr Passwort auf dem LDAP-Server. Semaphore verwendet es einmal, um zu prüfen, dass das LDAP-Konto
Ihnen gehört, und speichert es nicht.

## Link {#user-ldapLink}

Wird mit **LDAP-Provider** angezeigt.

Verknüpft Ihr Konto mit dem obigen LDAP-Konto. Die Verknüpfung wird abgelehnt, wenn die
Credentials falsch sind, wenn das LDAP-Konto mit einem anderen Benutzer verknüpft ist oder wenn
Ihr Konto bereits mit diesem Provider verknüpft ist. Ein verknüpftes Konto erscheint unter
**Linked accounts**; schließen Sie seinen Chip, um die Verknüpfung aufzuheben.

## Global role {#global-role-select}

Wird auf dem Tab **Roles** eines bestehenden Benutzers angezeigt, in der Pro-Edition, wenn Sie
globale Rollen verwalten dürfen.

Eine globale Rolle, die dem Benutzer gegeben werden soll, aus den Rollen, die der Benutzer noch
nicht hat. Eine globale Rolle gewährt Rechte über den gesamten Server: Benutzer, globale Rollen
oder die Einstellungen des Servers zu verwalten oder das Audit-Log zu lesen. Die globalen Rollen
selbst werden auf dem Tab **Globale Rollen** der Seite Benutzer erstellt.

## Assign {#globalRoleAssign-assign}

Wird mit **Global role** angezeigt.

Gibt dem Benutzer sofort die ausgewählte Rolle. Die Rollen, die der Benutzer hat, sind darunter
aufgeführt, jede mit einer Schaltfläche, um sie zu entziehen. **Effective global permissions**
zeigt, was der Benutzer am Ende tun darf und woher jedes Recht stammt.
