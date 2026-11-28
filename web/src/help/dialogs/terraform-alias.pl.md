---
form: src/components/TerraformAliasForm.vue
---

# Edytuj Key

Alias to adres backendu HTTP, który Semaphore udostępnia dla tego workspace: Terraform lub OpenTofu
skierowany na niego odczytuje, zapisuje, blokuje i odblokowuje stan workspace w Semaphore.
Semaphore tworzy adres podczas tworzenia aliasu; w tym oknie dialogowym wybierasz tylko klucz,
który go chroni. Aliasy są funkcją Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

Klucz, który strzeże aliasu. Lista oferuje tylko klucze typu **Logowanie z hasłem** z Magazynu
kluczy projektu; gdy lista jest pusta, najpierw utwórz tam taki klucz.

Klient aliasu wysyła login i hasło tego klucza. Po zapisaniu skopiuj adres z listy aliasów i umieść
go w backendzie `http` swojej konfiguracji wraz z loginem i hasłem klucza:

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

Adres obsługuje domyślne metody backendu `http`, więc jeden adres służy do stanu, blokady i
odblokowania. Wybranie innego klucza dla istniejącego aliasu zachowuje jego adres: klienci
potrzebują wtedy loginu i hasła nowego klucza.
