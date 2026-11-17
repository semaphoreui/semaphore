---
form: src/components/TerraformAliasForm.vue
---

# Upravit Key

Alias je adresa backendu HTTP, který Semaphore pro tento workspace poskytuje: Terraform nebo
OpenTofu nasměrovaný na ni čte, zapisuje, zamyká a odemyká stav workspace v Semaphore. Semaphore
adresu vytvoří, když alias vytvoříte; v tomto dialogu volíte jen klíč, který ji chrání. Aliasy
jsou funkcí Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

Klíč, který alias chrání. Seznam nabízí pouze klíče typu **Přihlášení heslem** z Úložiště klíčů
projektu; když je seznam prázdný, nejprve tam nějaký vytvořte.

Klient aliasu posílá přihlašovací jméno a heslo tohoto klíče. Po uložení zkopírujte adresu ze
seznamu aliasů a vložte ji do backendu `http` své konfigurace spolu s přihlašovacím jménem
a heslem klíče:

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

Adresa odpovídá na výchozí metody backendu `http`, takže jedna adresa slouží pro stav, zámek
i odemčení. Výběr jiného klíče pro existující alias jeho adresu zachová: klienti pak potřebují
přihlašovací jméno a heslo nového klíče.
