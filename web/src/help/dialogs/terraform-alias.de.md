---
form: src/components/TerraformAliasForm.vue
---

# Bearbeiten Key

Ein Alias ist eine Adresse des HTTP-Backends, das Semaphore für diesen Workspace bereitstellt:
Terraform oder OpenTofu, auf diese Adresse ausgerichtet, liest, schreibt, sperrt und entsperrt
den State des Workspace in Semaphore. Semaphore erzeugt die Adresse, wenn Sie den Alias
erstellen; in diesem Dialog wählen Sie nur den Schlüssel, der sie schützt. Aliase sind eine
Funktion von Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

Der Schlüssel, der den Alias schützt. Die Liste bietet nur Schlüssel vom Typ
**Login mit Passwort** aus dem Schlüsselspeicher des Projekts an; erstellen Sie dort zuerst einen,
wenn die Liste leer ist.

Ein Client des Alias sendet den Login und das Passwort dieses Schlüssels. Kopieren Sie nach dem
Speichern die Adresse aus der Liste der Aliase und tragen Sie sie mit dem Login und dem Passwort
des Schlüssels in das `http`-Backend Ihrer Konfiguration ein:

```hcl
terraform {
  backend "http" {
    address        = "ALIAS_URL"
    lock_address   = "ALIAS_URL"
    unlock_address = "ALIAS_URL"
    username       = "LOGIN"
    password       = "PASSWORD"
  }
}
```

Die Adresse beantwortet die Standardmethoden des `http`-Backends, sodass eine Adresse den State,
das Sperren und das Entsperren bedient. Die Wahl eines anderen Schlüssels für einen bestehenden
Alias behält seine Adresse bei: Clients brauchen dann den Login und das Passwort des neuen
Schlüssels.
