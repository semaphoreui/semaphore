---
form: src/components/TerraformAliasForm.vue
---

# Bewerken Key

Een alias is een adres van de HTTP-backend die Semaphore voor deze workspace aanbiedt: Terraform of
OpenTofu dat ernaar verwijst, leest, schrijft, vergrendelt en ontgrendelt de state van de
workspace in Semaphore. Semaphore maakt het adres aan wanneer u de alias aanmaakt; in dit
dialoogvenster kiest u alleen de sleutel die het beschermt. Aliassen zijn een functie van
Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

De sleutel die de alias bewaakt. De lijst biedt alleen sleutels van het type **Inloggen met
wachtwoord** uit de Sleutel Opslag van het project aan; maak er daar eerst een aan wanneer de lijst
leeg is.

Een client van de alias verstuurt de login en het wachtwoord van deze sleutel. Kopieer na het
opslaan het adres uit de lijst met aliassen en zet het in de `http`-backend van uw configuratie,
met de login en het wachtwoord van de sleutel:

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

Het adres beantwoordt de standaardmethoden van de `http`-backend, dus één adres bedient de state,
het vergrendelen en het ontgrendelen. Een andere sleutel kiezen voor een bestaande alias behoudt
het adres: clients hebben dan de login en het wachtwoord van de nieuwe sleutel nodig.
