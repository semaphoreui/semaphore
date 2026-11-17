---
form: src/components/InventoryForm.vue
---

# Bearbeiten Ansible Inventory

Ein Inventory legt fest, auf welchen Hosts Ansible ein Playbook ausführt, und enthält die
Credentials, mit denen Semaphore sich mit ihnen verbindet. Eine Task-Vorlage verweist auf ein
Inventory, und jeder Task dieser Vorlage läuft gegen dieses Inventory. Das Format eines Inventorys
ist im [Ansible-Leitfaden zu Inventorys](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html)
beschrieben.

## Name {#inventoryForm-name}

Der Name, unter dem das Inventory auf der Seite Inventory und in der Inventory-Auswahl einer
Task-Vorlage aufgeführt wird. Wählen Sie einen Namen, der die Hosts unterscheidbar macht, zum
Beispiel `Production web servers`.

## Runner-Tag {#inventoryForm-runnerTag}

Wird angezeigt, wenn Projekt-Runner verfügbar sind.

Tasks, die dieses Inventory verwenden, laufen auf einem Runner, der diesen Tag trägt. Ein in der
Task-Vorlage gesetzter Runner-Tag hat Vorrang vor dem Tag des Inventorys. Lassen Sie das Feld
leer, damit jeder Runner die Tasks übernehmen kann.

## Benutzer-Credentials {#inventoryForm-sshKeyId}

Der Schlüssel aus dem Schlüsselspeicher, mit dem Ansible sich an den Hosts anmeldet. Das Feld ist
ein Pflichtfeld und akzeptiert drei Arten von Schlüsseln:

- **SSH-Schlüssel**: Ansible verbindet sich mit dem privaten Schlüssel und, falls der Schlüssel
  einen hat, mit dessen Login.
- **Login mit Passwort**: Ansible verbindet sich mit dem Login und dem Passwort.
- **Keine**: Semaphore übergibt keine Credentials. Verwenden Sie diese Option für Hosts, die
  keine benötigen, wie `localhost`, oder wenn das Inventory selbst festlegt, wie verbunden wird.

## Sudo-Credentials {#inventoryForm-becomeKeyId}

Optional. Ein Schlüssel vom Typ **Login mit Passwort** aus dem Schlüsselspeicher, der verwendet
wird, wenn ein Play mit `become` seine Rechte erweitert: Sein Login ist der Benutzer, zu dem
Ansible wird, und sein Passwort ist das Passwort, mit dem Ansible die Abfrage beantwortet.

Lassen Sie das Feld leer, wenn die Plays keine Rechte erweitern oder wenn der oben angegebene
Benutzer dafür kein Passwort benötigt.

## Typ {#inventoryForm-type}

Woher das Inventory stammt:

- **Static**: Das Inventory wird in diesem Dialog im INI-Format eingegeben.
- **Static YAML**: Das Inventory wird in diesem Dialog im YAML-Format eingegeben. Verwenden Sie
  diesen Typ auch für die Konfiguration eines Inventory-Plugins.
- **File**: Das Inventory ist eine Datei in einem Git-Repository.
- **Managed**: Die Hosts und ihre Variablen werden als Datensätze gespeichert. Öffnen Sie das
  Inventory in der Liste, um sie zu bearbeiten.

Die Felder unterhalb des Typs ändern sich mit ihm.

## Pfad zur Inventory-Datei {#inventoryForm-inventoryPath}

Wird angezeigt, wenn der Typ **File** ist.

Der Pfad zur Inventory-Datei ab dem Stammverzeichnis ihres Repositorys, zum Beispiel
`inventory/production.yml`. Das Repository ist das im Feld Repository ausgewählte oder, wenn
dieses Feld leer ist, das Repository der Task-Vorlage, die das Inventory verwendet.

## Repository {#inventoryForm-repositoryId}

Wird angezeigt, wenn der Typ **File** ist. Optional.

Das Repository, das die Inventory-Datei enthält. Semaphore aktualisiert seine Kopie des
Repositorys vor jedem Task. Lassen Sie das Feld leer, wenn die Datei im selben Repository wie das
Playbook liegt: Der Pfad wird dann im Repository der Task-Vorlage gesucht.

## Inventory {#inventoryForm-inventoryCode}

Wird angezeigt, wenn der Typ **Static** oder **Static YAML** ist.

Der Text des Inventorys. Semaphore speichert ihn und schreibt ihn für jeden Task in eine
temporäre Datei. Ein statisches Inventory im INI-Format:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Dieselben Hosts als **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
