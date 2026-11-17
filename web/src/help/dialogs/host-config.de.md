---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Eine Zuordnung teilt Semaphore mit, welches Credential aus dem Schlüsselspeicher für einen
SSH-Host oder eine Git-URL verwendet wird. Ein Repository trägt einen Schlüssel, aber ein Task
kann auch andere Server erreichen: Submodule, Rollen und Collections aus `requirements.yml`,
Terraform-Module, Inventory-Repositories. Die Zuordnungen des Projekts gelten für jede Git- und
SSH-Verbindung seiner Tasks und für die Git-Befehle, die Semaphore für das Projekt ausführt, etwa
die Prüfung eines Repositorys auf neue Commits. Keine Datei Ihres Repositorys ändert sich.

## Typ {#hostConfigForm-type}

Worauf die Zuordnung passt:

- **Host**: ein SSH-Hostname. Jede SSH-Verbindung eines Tasks zu diesem Host verwendet den
  Schlüssel der Zuordnung: Git über SSH und Ansible bei der Verbindung zu einem Inventory-Host
  dieses Namens. Die Zuordnung hat für denselben Host Vorrang vor der SSH-Konfiguration des
  Servers.
- **URL**: eine Git-URL oder ihr Anfang. Git verbindet sich mit jedem Repository unter dieser URL
  mit dem Credential der Zuordnung.

Der Wechsel zu **Host** leert ein **Credential**, das eine Host-Zuordnung nicht verwenden kann.

## Host oder URL {#hostConfigForm-host}

Das Feld ist ein Pflichtfeld. Für jeden Host und jede URL darf es eine Zuordnung pro Typ geben.

Für **Host**: der Hostname genau so, wie Git oder Ansible sich mit ihm verbinden, zum Beispiel
`github.com` oder `10.0.0.21`. Nur Buchstaben, Ziffern, Punkte, Bindestriche und Unterstriche:
kein Benutzer, kein Port, keine Platzhalter.

Für **URL**: eine Adresse, die mit `https://` oder `http://` beginnt, ohne Login darin und ohne
Leerzeichen, Anführungszeichen oder `=`. Geben Sie ein ganzes Repository an oder ein Präfix, das
mit `/` endet, um eine Gruppe von Repositories abzudecken:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Wenn mehrere Zuordnungen passen, gewinnt die längste: Die obige Repository-Zuordnung überschreibt
für dieses eine Repository die Zuordnung von `acme`.

Mit einem SSH-Schlüssel lässt eine URL-Zuordnung Git über SSH auf Port `22` mit dem Host der URL
verbinden, unabhängig davon, welchen Port die URL nennt.

## Credential {#hostConfigForm-sshKeyId}

Der Schlüssel aus dem Schlüsselspeicher, der für die Verbindungen verwendet wird, auf die die
Zuordnung passt. Das Feld ist ein Pflichtfeld.

- **Host** akzeptiert nur einen **SSH-Schlüssel**. Hat der Schlüssel einen Login, verbindet sich
  SSH als dieser Benutzer.
- **URL** akzeptiert einen **SSH-Schlüssel** oder einen **Login mit Passwort**. Ein SSH-Schlüssel
  verbindet sich als sein Login oder als `git`, wenn er keinen hat. Ein Login mit Passwort
  funktioniert nur mit einer `https://`-URL; für ein Token tragen Sie das Token als Passwort ein
  und lassen den Login leer.

Ein Login mit Passwort erreicht Git über dessen Umgebung, nicht über dessen Befehlszeile, und das
Task-Log zeigt nur die ursprüngliche URL.
