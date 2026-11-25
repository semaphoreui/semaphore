---
form: src/components/ChangePasswordForm.vue
---

# Zmień hasło

Ustawia nowe hasło do logowania do Semaphore za pomocą nazwy użytkownika i hasła. Nowe hasło
zaczyna obowiązywać po naciśnięciu **Zapisz**. Użytkownik zewnętrzny, który loguje się przez LDAP
lub OpenID Connect, nie ma tutaj hasła do zmiany.

## Current password {#changePassword-currentPassword}

Wyświetlane, gdy zmieniasz własne hasło.

Twoje obecne hasło. Pole jest wymagane: Semaphore sprawdza je przed ustawieniem nowego, aby
skradziona sesja lub token API nie mogły zostać zamienione na twoje hasło. Serwer odrzuca zmianę,
gdy hasło się nie zgadza.

Administrator lub menedżer użytkowników, który resetuje hasło innego użytkownika, nie widzi tego
pola.

## Hasło {#changePassword-password}

Nowe hasło. Pole jest wymagane. Kliknij ikonę oka, aby wyświetlić to, co wpisałeś.

Semaphore przechowuje tylko skrót (hash) hasła.
