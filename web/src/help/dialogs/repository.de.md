---
form: src/components/RepositoryForm.vue
---

# Bearbeiten Repository

Ein Repository ist der Ort, aus dem die Playbooks, der Terraform-Code oder die Skripte einer
Task-Vorlage stammen: ein Git-Repository oder ein Verzeichnis auf dem Rechner, der die Tasks
ausführt. Vor jedem Task klont Semaphore das Git-Repository oder aktualisiert die zuvor geklonte
Kopie, checkt den Branch aus und führt den Task in dieser Kopie aus.

## Name {#repository-name}

Der Name, unter dem das Repository auf der Seite Repositories und in der Repository-Auswahl einer
Task-Vorlage und eines Inventorys aufgeführt wird. Das Feld ist ein Pflichtfeld.

## URL oder Pfad {#repository-gitUrl}

Wo der Code liegt. Das Feld ist ein Pflichtfeld. Die Chips darunter zeigen, welche Art von Adresse
Sie eingegeben haben, und ein Klick auf einen Chip schreibt die Adresse in diese Art um.

- **ssh**: `git@github.com:org/repo.git` oder `ssh://git@github.com/org/repo.git`. Git meldet
  sich mit dem Benutzer in der Adresse und dem in **Zugangsschlüssel** gewählten
  **SSH-Schlüssel** an.
- **https** und **http**: `https://github.com/org/repo.git`. Ein privates Repository braucht einen
  Schlüssel vom Typ **Login mit Passwort**; ein öffentliches braucht **Keine**. Vermeiden Sie
  `http`: Login und Passwort werden dann unverschlüsselt übertragen.
- **git**: `git://host/org/repo.git`, das schreibgeschützte Git-Protokoll, das keinen Login hat.
- **file**: `file:///srv/git/repo.git`, ein Git-Repository auf der Festplatte des Rechners, der
  den Task ausführt, das wie ein entferntes Repository geklont wird.
- **Absoluter Pfad**: ein absoluter Pfad wie `/srv/playbooks` oder `D:\playbooks`. Semaphore
  klont ihn nicht: Der Task läuft direkt in diesem Verzeichnis, das auf dem Rechner existieren
  muss, der den Task ausführt (der Semaphore-Server oder ein Runner). Der Branch gilt dafür nicht.

Das Ändern der Adresse entfernt die Kopien, die der Semaphore-Server von der alten Adresse
geklont hat.

## Branch {#repository-gitBranch}

Der Branch, den Semaphore klont und aktuell hält, zum Beispiel `main`. Das Feld ist ein
Pflichtfeld, außer bei einem absoluten Pfad, für den es gesperrt ist.

Eine Task-Vorlage kann einen eigenen Branch nennen, der Vorrang vor diesem hat, und eine Vorlage
kann ihre Tasks beim Start einen Branch wählen lassen. Ein Tag funktioniert auch, aber ein Tag
kann nicht gepullt werden, daher klont Semaphore das Repository vor jedem Task neu.

## Single checkout for all templates {#repository-monorepo}

Standardmäßig aus: Jede Task-Vorlage erhält ihre eigene Kopie des Repositorys. Haken Sie das
Kästchen an, um das Repository einmal zu klonen und diese Kopie zwischen allen Vorlagen zu teilen,
die es verwenden, für ein Repository, das zu groß ist, um eine Kopie pro Vorlage vorzuhalten.

Die gemeinsame Kopie bleibt auf dem Branch dieses Repositorys: Vorlagen, die sie verwenden,
können keinen eigenen Branch haben, und ihre Tasks können keinen anderen Branch und keinen
anderen Commit wählen. Bei einem absoluten Pfad gesperrt.

## Zugangsschlüssel {#repository-sshKeyId}

Der Schlüssel aus dem Schlüsselspeicher, mit dem Git sich am Repository anmeldet. Das Feld ist ein
Pflichtfeld. Es bietet Schlüssel von drei Typen an; welcher zu wählen ist, hängt von der URL ab:

- **SSH-Schlüssel** für eine `ssh`-URL. Semaphore lädt den privaten Schlüssel für die Dauer des
  Klonens oder der Aktualisierung in einen SSH-Agenten. Der Benutzer ist der in der URL, wie
  `git` in `git@github.com:org/repo.git`.
- **Login mit Passwort** für eine `https`-URL: der Login und das Passwort oder das Zugangstoken
  des Git-Servers. Bei leerem Login wird das Passwort allein als Benutzername gesendet, was
  manche Server bei einem Token erwarten. Dieselben Credentials erhält `ansible-galaxy` für Rollen
  und Collections auf demselben Server.
- **Keine** für ein öffentliches Repository, eine `git`- oder `file`-URL oder einen absoluten
  Pfad.

Ein Schlüssel, der nicht zur URL passt, gibt Git keine Credentials, und ein privates Repository
lässt sich dann nicht klonen.
