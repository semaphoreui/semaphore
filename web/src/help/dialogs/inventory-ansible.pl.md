---
form: src/components/InventoryForm.vue
---

# Edytuj Ansible Inventory

Inwentarz wskazuje Ansible, na których hostach uruchomić playbook, i przechowuje dane logowania,
za pomocą których Semaphore łączy się z tymi hostami. Szablon zadania odwołuje się do inwentarza,
a każde zadanie tego szablonu jest uruchamiane na nim. Format inwentarza opisuje
[przewodnik po inwentarzu Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Nazwa {#inventoryForm-name}

Nazwa, pod którą inwentarz widnieje na stronie Inwentarz i na liście wyboru inwentarza w
szablonie zadania. Wybierz nazwę, która odróżnia hosty, na przykład `Production web servers`.

## Tag uruchamiacza {#inventoryForm-runnerTag}

Wyświetlane, gdy dostępne są uruchamiacze projektu.

Zadania korzystające z tego inwentarza są wykonywane na uruchamiaczu, który ma ten tag. Tag
uruchamiacza ustawiony w szablonie zadania ma pierwszeństwo przed tagiem inwentarza. Pozostaw pole
puste, aby zadania mógł przejąć dowolny uruchamiacz.

## Dane logowania użytkownika {#inventoryForm-sshKeyId}

Klucz z Magazynu kluczy, którym Ansible loguje się do hostów. Pole jest wymagane i przyjmuje trzy
typy kluczy:

- **Klucz SSH**: Ansible łączy się za pomocą klucza prywatnego oraz, jeśli klucz go zawiera,
  jego loginu.
- **Logowanie z hasłem**: Ansible łączy się za pomocą loginu i hasła.
- **Brak**: Semaphore nie przekazuje żadnych danych logowania. Użyj tej opcji dla hostów, które
  ich nie wymagają, takich jak `localhost`, lub gdy sam inwentarz określa sposób połączenia.

## Dane logowania sudo {#inventoryForm-becomeKeyId}

Opcjonalne. Klucz typu **Logowanie z hasłem** z Magazynu kluczy, używany, gdy play podnosi
uprawnienia za pomocą `become`: jego login to użytkownik, którym staje się Ansible, a jego hasło
to hasło, którym Ansible odpowiada na monit.

Pozostaw pole puste, gdy żaden play nie podnosi uprawnień lub gdy powyższy użytkownik nie
potrzebuje do tego hasła.

## Typ {#inventoryForm-type}

Skąd pochodzi inwentarz:

- **Static**: inwentarz jest wpisywany w tym oknie dialogowym w formacie INI.
- **Static YAML**: inwentarz jest wpisywany w tym oknie dialogowym w formacie YAML. Użyj go także
  do konfiguracji wtyczki inwentarza.
- **File**: inwentarz jest plikiem w repozytorium Git.
- **Managed**: hosty i ich zmienne są przechowywane jako rekordy. Otwórz inwentarz z listy, aby je
  edytować.

Pola poniżej typu zmieniają się w zależności od niego.

## Ścieżka do pliku inwentarza {#inventoryForm-inventoryPath}

Wyświetlane, gdy typ to **File**.

Ścieżka do pliku inwentarza względem katalogu głównego jego repozytorium, na przykład
`inventory/production.yml`. Repozytorium to repozytorium wybrane w polu Repozytorium lub, gdy to
pole jest puste, repozytorium szablonu zadania, który używa inwentarza.

## Repozytorium {#inventoryForm-repositoryId}

Wyświetlane, gdy typ to **File**. Opcjonalne.

Repozytorium zawierające plik inwentarza. Semaphore aktualizuje swoją kopię repozytorium przed
każdym zadaniem. Pozostaw pole puste, gdy plik znajduje się w tym samym repozytorium co playbook:
ścieżka jest wtedy wyszukiwana w repozytorium szablonu zadania.

## Inwentarz {#inventoryForm-inventoryCode}

Wyświetlane, gdy typ to **Static** lub **Static YAML**.

Treść inwentarza. Semaphore przechowuje ją i dla każdego zadania zapisuje do pliku tymczasowego.
Statyczny inwentarz w formacie INI:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Te same hosty jako **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
