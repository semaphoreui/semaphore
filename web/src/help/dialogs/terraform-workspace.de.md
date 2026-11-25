---
form: src/components/TerraformInventoryForm.vue
---

# Neu Terraform Workspace

Ein Workspace ist das Inventory einer Terraform-, OpenTofu- oder Terragrunt-Vorlage: Er nennt den
Workspace, den ein Task nach `init` auswählt, und den Schlüssel, mit dem `init` private Module
herunterlädt. Der Workspace, den Sie hier erstellen, wird dieser Vorlage zugeordnet und über ihrem
State aufgeführt. Tasks der Vorlage laufen in ihrem Standard-Workspace, dem mit einem Häkchen
markierten; **Make default** ändert ihn. Derselbe Dialog bearbeitet einen Workspace, wenn Sie ihn
mit der Stift-Schaltfläche öffnen.

## Workspace name {#tfWorkspace-inventory}

Der Name des Terraform-Workspace, zum Beispiel `production`. Das Feld ist ein Pflichtfeld.

Vor einem Plan führt Semaphore `workspace select -or-create=true` mit diesem Namen aus, sodass der
Workspace erstellt wird, wenn er noch nicht existiert. Wenn `workspace list` fehlschlägt, zum
Beispiel weil das Backend keine Workspaces unterstützt, überspringt Semaphore die Auswahl.

Semaphore speichert den Workspace als `TEMPLATE - NAME`, den Namen der Vorlage und diesen Namen:
Diesen vollständigen Namen sehen Sie auf der Seite Inventory und wenn Sie den Workspace einer
Vorlage zuordnen. Das Umbenennen eines Workspace benennt ihn in Terraform nicht um: Der nächste
Task wählt einen Workspace mit dem neuen Namen aus oder erstellt ihn.

## SSH key for private modules {#tfWorkspace-sshKeyId}

Der Schlüssel, mit dem `init` Module aus privaten Git-Repositories über SSH herunterlädt. Das Feld
ist in diesem Dialog ein Pflichtfeld. Die Liste bietet die Schlüssel vom Typ **SSH-Schlüssel**,
**Login mit Passwort** und **Keine** aus dem Schlüsselspeicher an.

Nur ein **SSH-Schlüssel** wirkt: Semaphore lädt ihn für die Dauer von `init` in einen SSH-Agenten,
und Git verwendet ihn für Modulquellen wie `git::ssh://git@example.com/modules/vpc.git`. Ein
Schlüssel vom Typ **Login mit Passwort** übergibt Git keine Credentials. Wählen Sie **Keine**, wenn
die Module keinen Schlüssel brauchen.
