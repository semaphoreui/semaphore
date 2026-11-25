---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Ordnet dieser Vorlage einen Workspace zu, der im Projekt bereits existiert, sodass er unter den
Workspaces der Vorlage mit seinem State und seinen Aliasen erscheint. Ein Workspace gehört
jeweils zu einer Vorlage. Das Zuordnen macht den Workspace nicht zum Standard: Tasks laufen
weiterhin im Standard-Workspace, bis Sie auf **Make default** klicken.

## Workspace {#workspaceSelect-inventoryId}

Der zuzuordnende Workspace. Das Feld ist ein Pflichtfeld. Die Liste zeigt die Workspaces des
Projekts, die für die App dieser Vorlage (Terraform, OpenTofu oder Terragrunt) gemacht sind und
keiner Vorlage zugeordnet sind; ein mit **Detach** von seiner Vorlage gelöster Workspace erscheint
hier wieder.

Der Workspace behält seinen Namen, seinen Schlüssel und seinen State. Der Server lehnt einen
Workspace ab, den inzwischen eine andere Vorlage übernommen hat.
