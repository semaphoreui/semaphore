---
form: src/components/TerraformInventoryForm.vue
---

# Nowy Terraform Workspace

Workspace to inwentarz szablonu Terraform, OpenTofu lub Terragrunt: wskazuje workspace, który
zadanie wybiera po `init`, oraz klucz, za pomocą którego `init` pobiera prywatne moduły. Utworzony
tutaj workspace jest dołączany do tego szablonu i wyświetlany nad jego stanem. Zadania szablonu są
wykonywane w jego domyślnym workspace, oznaczonym znacznikiem wyboru; **Make default** go zmienia.
To samo okno dialogowe edytuje workspace, gdy otworzysz go przyciskiem z ołówkiem.

## Workspace name {#tfWorkspace-inventory}

Nazwa workspace Terraform, na przykład `production`. Pole jest wymagane.

Przed planem Semaphore uruchamia `workspace select -or-create=true` z tą nazwą, więc workspace jest
tworzony, gdy jeszcze nie istnieje. Gdy `workspace list` kończy się niepowodzeniem, na przykład
dlatego, że backend nie obsługuje workspace, Semaphore pomija wybór.

Semaphore przechowuje workspace jako `TEMPLATE - NAME`, czyli nazwę szablonu i tę nazwę: tę pełną
nazwę widzisz na stronie Inwentarz i podczas dołączania workspace do szablonu. Zmiana nazwy
workspace nie zmienia jej w Terraform: następne zadanie wybiera lub tworzy workspace o nowej
nazwie.

## SSH key for private modules {#tfWorkspace-sshKeyId}

Klucz, którego `init` używa do pobierania modułów z prywatnych repozytoriów Git przez SSH. W tym
oknie dialogowym pole jest wymagane. Lista oferuje klucze typu **Klucz SSH**, **Logowanie z
hasłem** i **Brak** z Magazynu kluczy.

Działa tylko **Klucz SSH**: Semaphore wczytuje go do agenta SSH na czas `init`, a Git używa go dla
źródeł modułów takich jak `git::ssh://git@example.com/modules/vpc.git`. Klucz typu **Logowanie z
hasłem** nie przekazuje Git żadnych danych logowania. Wybierz **Brak**, gdy moduły nie wymagają
klucza.
