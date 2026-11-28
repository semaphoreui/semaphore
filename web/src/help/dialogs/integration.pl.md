---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Edytuj uruchomienie

Uruchomienie to to, co integracja uruchamia, gdy inna usługa wywołuje adres URL webhooka Semaphore,
na przykład przy zdarzeniu push w GitHub: zadanie szablonu zadania albo workflow. Adres URL i
sprawdzanie żądania należą do endpointu integracji, a nie do uruchomienia. Strona uruchomienia
zawiera jego matchery i wartości, które pobiera z żądania.

## Nazwa {#integrationForm-name}

Nazwa, pod którą integracja widnieje na stronie Integracje. Zadania uruchomione przez integrację
są jej przypisywane w dzienniku audytu. Pole jest wymagane.

## Integracja {#integrationForm-groupId}

Integracja, do której należy uruchomienie. Uruchomienie startuje przy żądaniach, które trafiają do
endpointu tej integracji i przechodzą jej autoryzację. Pole jest wymagane.

## Cel {#integrationForm-targetToggle}

Wyświetlane, gdy serwer obsługuje przepływy pracy.

Co uruchamia integracja: **Template** uruchamia zadanie szablonu zadania, **Workflow** uruchamia
przebieg przepływu pracy. Przełączenie na **Workflow** usuwa poniższe parametry zadania.

## Workflow to run {#integrationForm-workflowId}

Wyświetlane, gdy **Cel** to **Workflow**.

Przepływ pracy uruchamiany przez żądanie. Każda wartość, którą integracja pobiera z żądania, jest
przekazywana do przepływu pracy jako odpowiedź na jego zmienną ankiety o tej samej nazwie.
Integracje przepływów pracy wymagają wersji Pro: bez niej pole jest wyłączone, a serwer odrzuca
zmianę.

## Szablon zadania do uruchomienia {#integrationForm-templateId}

Wyświetlane, gdy **Cel** to **Template**.

Szablon zadania, którego zadanie uruchamia żądanie. Pole jest wymagane. Po wybraniu szablonu pod
nim pojawiają się parametry jego zadań.

## Parametry zadania {#integrationForm-taskParams}

Wyświetlane, gdy wybrano szablon zadania.

Parametry, które otrzymuje każde zadanie uruchomione przez tę integrację, tak jakby uruchomić je
samodzielnie z szablonu. Znajdują się tu tylko opcje, które szablon pozwala nadpisać w zadaniu.
Wartości pobrane z żądania mają przed nimi pierwszeństwo: wartość typu **Variables** zastępuje
odpowiedź o tej samej nazwie, a wartość typu **Task Params** parametr o tej samej nazwie.

Gdy szablon zezwala na argumenty CLI w zadaniu, **CLI args** zawiera dodatkowe argumenty wiersza
poleceń zadania.

## Wiadomość {#taskParams-message}

Opcjonalne. Wiadomość każdego zadania uruchomionego przez integrację, wyświetlana w historii zadań.

## Zmienne ankiety {#taskParams-survey}

Wyświetlane, gdy szablon ma zmienne ankiety.

Odpowiedzi na ankietę szablonu, takie same dla każdego zadania. Wartość typu **Variables** pobrana
z żądania zastępuje odpowiedź o tej samej nazwie.

Odpowiedzi na zmienne typu secret nie są zapisywane z integracją: przekaż sekret przez grupę
zmiennych szablonu.

## Gałąź {#taskParams-gitBranch}

Wyświetlane, gdy szablon pozwala nadpisać gałąź w zadaniu.

Gałąź lub tag repozytorium, które pobierają zadania. Pozostaw bez zmian, aby użyć gałęzi szablonu.
Wartość o nazwie `git_branch` typu **Task Params** pobrana z żądania ma przed nią pierwszeństwo;
początkowe `refs/heads/` jest z tej wartości usuwane.

## Inwentarz {#taskParams-inventoryId}

Wyświetlane, gdy szablon pozwala nadpisać inwentarz w zadaniu.

Inwentarz, a w przypadku Terraform workspace, na którym są wykonywane zadania zamiast inwentarza
szablonu.

## Opcje Ansible {#taskParams-ansible}

Wyświetlane, gdy szablon uruchamia playbook Ansible.

Opcje `ansible-playbook` dla każdego zadania, każda wyświetlana, gdy szablon na nią zezwala:
**Limit**, **Tagi**, **Pomiń tagi**, **Debug**, **Symulacja** (`--check`), **Różnice** i **Skip
Galaxy install**.

## Opcje Terraform {#taskParams-terraform}

Wyświetlane, gdy szablon uruchamia Terraform, OpenTofu lub Terragrunt.

- **Plan**: zadanie wykonuje plan i zatrzymuje się, nic nie jest stosowane.
- **Destroy**: plan i apply niszczą zasoby.
- **Auto Approve**: plan jest stosowany bez potwierdzenia, gdy szablon zezwala na automatyczne
  zatwierdzanie. Bez tego zadanie czeka, aż ktoś je potwierdzi w Semaphore.
- **Upgrade** i **Reconfigure**: `init` jest uruchamiane z `-upgrade` i `-reconfigure`.
