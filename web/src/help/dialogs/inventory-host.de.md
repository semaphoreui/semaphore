---
form: src/components/InventoryHostForm.vue
---

# Host bearbeiten

Ein Host eines verwalteten Inventorys. Eine Aufgabe, die das Inventory verwendet, läuft auf
denjenigen Hosts, die in der Hostliste eingeschaltet sind.

## Name {#inventoryHostForm-name}

Der Name, unter dem Ansible den Host kennt, und die Adresse, mit der es sich verbindet, sofern die
Variablen `ansible_host` nicht setzen. Das Feld ist erforderlich. Ein Name ist innerhalb seines
Inventorys eindeutig, Groß- und Kleinschreibung zählt: `web1` und `Web1` sind zwei Hosts. Er darf
keine Leerzeichen, Kommas, eckigen Klammern oder Doppelpunkte enthalten und ist höchstens 250
Zeichen lang.

## Beschreibung {#inventoryHostForm-description}

Optional. Eine Notiz für Menschen, die in der Hostliste angezeigt wird. Ansible erhält sie nicht.

## Variablen {#inventoryHostForm-variables}

Optional. Variablen dieses Hosts als YAML- oder JSON-Mapping:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Eine Variable des Hosts hat Vorrang vor einer gleichnamigen Variable des Inventorys. YAML-Anker,
-Aliase und Merge-Keys werden nicht akzeptiert.
