---
form: src/components/ChangePasswordForm.vue
---

# Změnit heslo

Nastaví nové heslo pro přihlašování do Semaphore uživatelským jménem a heslem. Nové heslo začne
platit po stisknutí **Save**. Externí uživatel, který se přihlašuje přes LDAP nebo OpenID
Connect, zde žádné heslo ke změně nemá.

## Current password {#changePassword-currentPassword}

Zobrazí se, když měníte své vlastní heslo.

Vaše současné heslo. Pole je povinné: Semaphore ho ověří dříve, než nastaví nové, aby ukradenou
relaci nebo API token nebylo možné proměnit ve vaše heslo. Když heslo nesouhlasí, server změnu
odmítne.

Administrátor nebo správce uživatelů, který obnovuje heslo jiného uživatele, toto pole nevidí.

## Heslo {#changePassword-password}

Nové heslo. Pole je povinné. Kliknutím na ikonu oka zobrazíte, co jste napsali.

Semaphore ukládá pouze hash hesla.
