---
form: src/components/InventorySourceForm.vue
---

# Edytuj źródło

Źródło importuje hosty i grupy zarządzanego inwentarza z pliku w repozytorium Git. Plik jest czytany po każdym naciśnięciu Sync, a hosty i grupy, które opisuje, są utrzymywane zgodnie z nim. Serwer potrzebuje Ansible i dostępu do repozytorium.

## Nazwa {#inventorySourceForm-name}

Nazwa źródła, widoczna na jego karcie i przy hostach, które dostarcza. Pole jest wymagane.

## Repozytorium {#inventorySourceForm-repository}

Repozytorium projektu, z którego czytany jest plik. Pole jest wymagane. Używana jest jego gałąź.

## Ścieżka {#inventorySourceForm-path}

Ścieżka pliku inwentarza w repozytorium, na przykład `inventory/prod.yml`, albo katalogu z plikami inwentarza. Pole jest wymagane.

Pliki są czytane jako statyczny YAML lub INI, skrypty i pliki konfiguracji wtyczek są odrzucane. W katalogu każdy plik, którego Ansible nie pomija, musi być poprawnym inwentarzem, inaczej synchronizacja się nie udaje i nic się nie zmienia. Ansible pomija w katalogu niektóre rozszerzenia, na przykład `.cfg`, `.md` i `.txt`. Repozytorium z dowiązaniem prowadzącym poza nie jest odrzucane.
