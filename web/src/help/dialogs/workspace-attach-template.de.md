---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Ordnet diesen Workspace einer Vorlage zu, sodass er unter den Workspaces der Vorlage mit seinem
State und seinen Aliasen erscheint. Ein Workspace gehört jeweils zu einer Vorlage: Einmal
zugeordnet, verlässt er die Seite Inventory und wird in der Vorlage verwaltet. Das Zuordnen macht
ihn nicht zum Standard-Workspace der Vorlage: Öffnen Sie dafür die Vorlage und klicken Sie auf
**Make default**.

## Template {#inventoryAttach-templateId}

Die Vorlage, der der Workspace zugeordnet wird. Das Feld ist ein Pflichtfeld. Die Liste zeigt die
Vorlagen des Projekts, die für dieselbe App wie der Workspace gemacht sind, Terraform oder
OpenTofu.

Der Workspace behält seinen Namen, seinen Schlüssel und seinen State. Um ihn später in eine andere
Vorlage zu verschieben, klicken Sie in der Vorlage, der er zugeordnet ist, auf **Detach**: Er
kehrt auf die Seite Inventory zurück. Der Standard-Workspace einer Vorlage kann nicht gelöst
werden.
