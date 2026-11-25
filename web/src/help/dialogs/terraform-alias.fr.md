---
form: src/components/TerraformAliasForm.vue
---

# Modifier Key

Un alias est une adresse du backend HTTP que Semaphore fournit pour ce workspace : Terraform ou
OpenTofu configuré pour l'utiliser lit, écrit, verrouille et déverrouille l'état du workspace dans
Semaphore. Semaphore crée l'adresse lorsque vous créez l'alias ; dans cette boîte de dialogue, vous
choisissez seulement la clé qui la protège. Les alias sont une fonctionnalité de Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

La clé qui protège l'alias. La liste ne propose que les clés de type **Connexion avec mot de
passe** du Magasin de clés du projet ; créez-en une d'abord lorsque la liste est vide.

Un client de l'alias envoie l'identifiant et le mot de passe de cette clé. Après l'enregistrement,
copiez l'adresse depuis la liste des alias et placez-la dans le backend `http` de votre
configuration, avec l'identifiant et le mot de passe de la clé :

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

L'adresse répond aux méthodes par défaut du backend `http`, de sorte qu'une seule adresse sert
pour l'état, le verrouillage et le déverrouillage. Choisir une autre clé pour un alias existant
conserve son adresse : les clients ont alors besoin de l'identifiant et du mot de passe de la
nouvelle clé.
