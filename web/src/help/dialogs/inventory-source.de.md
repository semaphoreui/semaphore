---
form: src/components/InventorySourceForm.vue
---

# Quelle bearbeiten

Eine Quelle importiert die Hosts und Gruppen eines verwalteten Inventorys aus einer Datei in einem Git-Repository. Die Datei wird bei jedem Klick auf Sync gelesen, und die Hosts und Gruppen daraus bleiben mit ihr abgeglichen. Der Server braucht Ansible und Zugriff auf das Repository.

## Name {#inventorySourceForm-name}

Der Name der Quelle, er steht auf ihrer Karte und bei den Hosts, die sie liefert. Das Feld ist erforderlich.

## Repository {#inventorySourceForm-repository}

Das Repository des Projekts, aus dem die Datei gelesen wird. Das Feld ist erforderlich. Verwendet wird sein Branch.

## Pfad {#inventorySourceForm-path}

Der Pfad einer Inventory-Datei im Repository, zum Beispiel `inventory/prod.yml`, oder eines Verzeichnisses mit Inventory-Dateien. Das Feld ist erforderlich.

Dateien werden als statisches YAML oder INI gelesen, Skripte und Konfigurationsdateien von Plugins werden abgelehnt. In einem Verzeichnis muss jede Datei, die Ansible nicht überspringt, ein gültiges Inventory sein, sonst schlägt die Synchronisierung fehl und nichts ändert sich. Ansible überspringt in einem Verzeichnis einige Endungen, zum Beispiel `.cfg`, `.md` und `.txt`. Ein Repository mit einem Link, der aus ihm hinausführt, wird abgelehnt.
