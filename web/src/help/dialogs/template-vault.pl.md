---
form: src/components/TemplateVaults.vue
---

# Skarbiec

Skarbiec przekazuje Ansible hasło do plików i zmiennych zaszyfrowanych za pomocą Ansible Vault w
repozytorium playbooka. Każdy skarbiec szablonu staje się jedną opcją `--vault-id` polecenia
`ansible-playbook`. Skarbiec jest zapisywany z szablonem, gdy zapisujesz okno dialogowe szablonu.

## ID skarbca {#templateVaultDialog-name}

Opcjonalne. Vault ID, którego dotyczy hasło, czyli etykieta, z którą zaszyfrowano zawartość
(`ansible-vault encrypt --vault-id prod@prompt`). Pozostaw pole puste dla vault ID `default`.

ID jest zapisywane małymi literami. Każdy skarbiec szablonu musi mieć własne ID i tylko jedno może
być puste.

## Typ {#templateVaultDialog-type}

Skąd pochodzi hasło:

- **Hasło**: klucz z Magazynu kluczy. Semaphore wpisuje jego hasło, gdy Ansible o nie prosi, więc
  nigdy nie jest ono zapisywane do pliku.
- **Skrypt klienta**: skrypt klienta haseł skarbca, który wypisuje hasło; Ansible go uruchamia.

## Hasło skarbca {#templateVaultDialog-vaultKeyId}

Wyświetlane, gdy typ to **Hasło**.

Klucz, który przechowuje hasło skarbca. Wyświetlane są tylko klucze typu **Logowanie z hasłem**;
Semaphore używa ich hasła i ignoruje login. Pole jest wymagane.

## Ścieżka do skryptu {#templateVaultDialog-script}

Wyświetlane, gdy typ to **Skrypt klienta**.

Ścieżka do skryptu, na przykład `scripts/vault-pass-client.py`. Ścieżka względna jest liczona od
katalogu, w którym działa Ansible: katalogu głównego repozytorium lub katalogu roboczego szablonu,
gdy szablon go ma. Pole jest wymagane.

Nazwa pliku musi kończyć się na `-client`, z rozszerzeniem lub bez: tylko wtedy Ansible uruchamia
plik jako skrypt klienta i przekazuje mu vault ID za pomocą `--vault-id`. Skrypt działa na maszynie,
która wykonuje zadanie, jako użytkownik, którym działa Semaphore.
