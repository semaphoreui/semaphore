---
form: src/components/TerraformAliasForm.vue
---

# Edit Key

An alias is an address of the HTTP backend that Semaphore serves for this workspace: Terraform or
OpenTofu pointed at it reads, writes, locks and unlocks the state of the workspace in Semaphore.
Semaphore makes the address when you create the alias; in this dialog you only choose the key
that protects it. Aliases are a feature of Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

The key that guards the alias. The list offers only keys of the type **Login with password** from
the Key Store of the project; create one there first when the list is empty.

A client of the alias sends the login and the password of this key. After you save, copy the
address from the list of aliases and put it into the `http` backend of your configuration, with
the login and the password of the key:

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

The address answers the default methods of the `http` backend, so one address serves the state,
the lock and the unlock. Choosing another key for an existing alias keeps its address: clients
then need the login and the password of the new key.
