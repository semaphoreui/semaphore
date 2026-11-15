---
form: src/components/TerraformAliasForm.vue
---

# Editar Key

Un alias es una dirección del backend HTTP que Semaphore sirve para este workspace: Terraform u
OpenTofu, apuntados a ella, leen, escriben, bloquean y desbloquean el estado del workspace en
Semaphore. Semaphore crea la dirección cuando usted crea el alias; en este diálogo solo elige la
clave que la protege. Los alias son una función de Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

La clave que protege el alias. La lista ofrece solo claves del tipo **Iniciar sesión con
contraseña** del Almacén de Claves del proyecto; cree una allí primero cuando la lista esté vacía.

Un cliente del alias envía el usuario y la contraseña de esta clave. Después de guardar, copie la
dirección de la lista de alias y póngala en el backend `http` de su configuración, con el usuario y
la contraseña de la clave:

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

La dirección responde a los métodos predeterminados del backend `http`, así que una sola dirección
sirve para el estado, el bloqueo y el desbloqueo. Elegir otra clave para un alias existente conserva
su dirección: los clientes necesitan entonces el usuario y la contraseña de la nueva clave.
