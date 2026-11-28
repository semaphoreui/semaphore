---
form: src/components/InventoryGroupForm.vue
---

# Gruppe bearbeiten

Eine Gruppe von Hosts eines verwalteten Inventorys. Ein Playbook läuft über den Namen der Gruppe auf
einem Teil der Hosts, und eine Variable der Gruppe wird einmal für alle ihre Hosts gesetzt. Hosts
kommen auf dem Tab Hosts in eine Gruppe, im Host-Dialog oder mit den Aktionen für die Auswahl.

Eine aus einer Inventarquelle importierte Gruppe wird von dieser Quelle verwaltet. Alle ihre Felder
sind schreibgeschützt und werden bei jeder Synchronisierung aktualisiert.

## Name {#inventoryGroupForm-name}

Der Name, unter dem Ansible die Gruppe kennt, zum Beispiel in `hosts: web` eines Playbooks. Das Feld
ist erforderlich. Ein Name ist innerhalb seines Inventorys eindeutig, Groß- und Kleinschreibung
zählt: `web` und `Web` sind zwei Gruppen. Er darf keine Leerzeichen, Kommas, eckigen Klammern oder
Doppelpunkte enthalten, ist höchstens 250 Zeichen lang und darf nicht `all` oder `ungrouped` sein.
Ansible warnt bei Namen mit anderen Zeichen als Buchstaben, Ziffern und Unterstrichen, etwa
`web-eu`, der Name funktioniert aber trotzdem. Die Zeichen `*`, `?`, `!`, `&`, `~` und `@` sind
ebenfalls nicht erlaubt, weil Ansible sie als Muster liest.

## Beschreibung {#inventoryGroupForm-description}

Optional. Eine Notiz für Menschen, die in der Gruppenliste angezeigt wird. Ansible erhält sie nicht.

## Child groups {#inventoryGroupForm-children}

Optional. Gruppen innerhalb dieser Gruppe. Die Hosts einer untergeordneten Gruppe gehören auch zu
dieser Gruppe. Eine Gruppe kann nicht ihr eigener Vorfahr sein.

## Variablen {#inventoryGroupForm-variables}

Optional. Variablen dieser Gruppe als YAML- oder JSON-Mapping:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Jeder Host der Gruppe erhält sie. Eine Variable eines Hosts hat Vorrang vor einer gleichnamigen
Variable der Gruppe, und eine Variable der Gruppe hat Vorrang vor einer Variable des Inventorys.
YAML-Anker, -Aliase und Merge-Keys werden nicht akzeptiert.
