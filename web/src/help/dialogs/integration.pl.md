---
form: src/components/IntegrationForm.vue, src/components/TaskParamsForm.vue
---

# Edytuj integrację

Integracja uruchamia szablon zadania lub przepływ pracy, gdy inna usługa wywoła adres URL webhooka
Semaphore, na przykład przy zdarzeniu push z GitHub. Adres URL to alias: `/api/integrations/` i
losowy kod. Aliasy projektu znajdują się na stronie Integracje, aliasy pojedynczej integracji na
jej własnej stronie. Ta strona zawiera również dopasowacze i wartości, które integracja pobiera z
żądania.

## Name {#integrationForm-name}

Nazwa, pod którą integracja widnieje na stronie Integracje. Zadania uruchomione przez integrację
są jej przypisywane w dzienniku audytu. Pole jest wymagane.

## Runs {#integrationForm-targetToggle}

Wyświetlane, gdy serwer obsługuje przepływy pracy.

Co uruchamia integracja: **Template** uruchamia zadanie szablonu zadania, **Workflow** uruchamia
przebieg przepływu pracy. Przełączenie na **Workflow** usuwa poniższe parametry zadania.

## Workflow to run {#integrationForm-workflowId}

Wyświetlane, gdy **Runs** to **Workflow**.

Przepływ pracy uruchamiany przez żądanie. Każda wartość, którą integracja pobiera z żądania, jest
przekazywana do przepływu pracy jako odpowiedź na jego zmienną ankiety o tej samej nazwie.
Integracje przepływów pracy wymagają wersji Pro: bez niej pole jest wyłączone, a serwer odrzuca
zmianę.

## Task Template to run {#integrationForm-templateId}

Wyświetlane, gdy **Runs** to **Template**.

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

## Auth method {#integrationForm-authMethod}

Jak Semaphore sprawdza, czy żądanie pochodzi od oczekiwanej usługi. Żądanie, które nie przejdzie
sprawdzenia, niczego nie uruchamia, a webhook odpowiada na nie tak samo jak na każde inne żądanie.

- **None**: brak sprawdzania. Każdy, kto zna adres URL, może uruchomić zadanie.
- **GitHub Webhooks**: nagłówek `X-Hub-Signature-256` musi zawierać podpis HMAC-SHA256 treści
  utworzony za pomocą sekretu, tak jak wysyła go GitHub.
- **Bitbucket Webhooks**: to samo sprawdzenie nagłówka `X-Hub-Signature`, tak jak wysyła go
  Bitbucket.
- **Token**: nagłówek podany w **Auth header** musi zawierać sam sekret.
- **HMAC** i **HMAC (SHA-512)**: nagłówek podany w **Auth header** musi zawierać HMAC-SHA256 lub
  HMAC-SHA512 treści utworzony za pomocą sekretu, zapisany małymi literami szesnastkowo, bez
  przedrostka takiego jak `sha256=`.
- **BasicAuth**: żądanie musi zawierać w nagłówku `Authorization` login i hasło klucza wybranego w
  **Sekret uwierzytelniania**.

## Auth header {#integrationForm-authHeader}

Wyświetlane, gdy **Auth method** to **Token**, **HMAC** lub **HMAC (SHA-512)**.

Nazwa nagłówka żądania, który zawiera token lub podpis, na przykład `X-Webhook-Token`. Wielkość
liter w nazwie nie ma znaczenia. Żądanie bez tego nagłówka jest odrzucane.

## Sekret uwierzytelniania {#integrationForm-authSecretId}

Wyświetlane, gdy **Auth method** jest inne niż **None**.

Klucz z Magazynu kluczy, względem którego sprawdzane jest żądanie. Dla **BasicAuth** pasuje tylko
klucz typu **Logowanie z hasłem**. Dla pozostałych metod pasuje także klucz typu **Ciąg znaków**:
sprawdzenie używa jego wartości lub hasła klucza **Logowanie z hasłem**.

Pole jest potrzebne przy każdej metodzie oprócz **None**: bez klucza lub z pustym sekretem każde
żądanie jest odrzucane.
