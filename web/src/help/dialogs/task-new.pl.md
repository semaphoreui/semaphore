---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nowe zadanie

Uruchamia zadanie szablonu: jedno wykonanie jego playbooka, kodu Terraform lub skryptu. To szablon
decyduje, które pola wyświetla to okno dialogowe, a każda sekcja poniżej mówi, kiedy jej pole jest
widoczne. Przycisk na dole to **Uruchom**, **Buduj** lub **Wdrażaj**, zależnie od typu szablonu.

Gdy szablon ma ankietę, jej pytania są polami po **Wiadomość**. Pole oznaczone `*` jest wymagane,
pole Integer przyjmuje tylko cyfry, a odpowiedzi domyślne są wypełnione. Semaphore przekazuje
odpowiedzi do zadania jako zmienne: dodatkowe zmienne dla Ansible, `-var` dla Terraform, argumenty
`name=value` dla skryptu lub zmienne środowiskowe dla pytania tak ustawionego. Odpowiedź zastępuje
zmienną o tej samej nazwie w grupach zmiennych szablonu. Tajna odpowiedź jest ukrywana podczas
wpisywania i nie jest przechowywana z zadaniem.

## Commit {#newTask-commit}

Wyświetlane, gdy ponownie uruchamiasz wcześniejsze zadanie, a to zadanie zostało wykonane na znanym
commicie.

Zadanie pobiera ten commit repozytorium zamiast najnowszego commita gałęzi. Zamknij to pole, aby
uruchomić najnowszy commit.

Commit jest używany tylko wtedy, gdy szablon zezwala na nadpisanie gałęzi, i nigdy w przypadku
repozytorium oznaczonego jako monorepo. W przeciwnym razie zadanie uruchamia najnowszy commit.

## Wersja kompilacji {#newTask-buildTaskId}

Wyświetlane, gdy typ szablonu to **Wdrażaj**.

Kompilacja do wdrożenia: jedno z ostatnich 20 udanych zadań szablonu budowy, wyświetlanych z wersją
i wiadomością. Wybrana jest najnowsza. Pole jest wymagane.

Wersja kompilacji trafia do zadania: jako `semaphore_vars.task_details.incoming_version` w
Ansible, jako zmienna środowiskowa `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` w skrypcie. Lista
jest pusta, gdy nie możesz widzieć szablonu budowy.

## Wiadomość {#newTask-message}

Opcjonalne. Notatka o tym, dlaczego zadanie jest uruchamiane, na przykład
`Hotfix for the login page`. Jest wyświetlana z zadaniem w historii zadań i niczego nie zmienia w
jego wykonaniu.

Playbook może ją odczytać jako `semaphore_vars.task_details.message`, a skrypt jako zmienną
środowiskową `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Gałąź {#newTask-gitBranch}

Wyświetlane, gdy szablon zezwala na nadpisanie gałęzi, chyba że repozytorium jest monorepo.

Gałąź lub tag repozytorium, które pobiera zadanie. Pole początkowo zawiera gałąź szablonu. Gdy jest
puste, zadanie używa gałęzi szablonu lub, gdy szablon jej nie ma, gałęzi repozytorium. Nazwa, która
nie jest prawidłową nazwą gałęzi Git, jest odrzucana.

## Inwentarz {#newTask-inventoryId}

Wyświetlane, gdy szablon zezwala na wybór inwentarza w zadaniu i możesz widzieć projekt. Dla
szablonu Terraform, OpenTofu lub Terragrunt pole nosi nazwę **Workspace**, ale serwer używa wyboru
tylko dla szablonów Ansible: zadanie Terraform, OpenTofu lub Terragrunt działa w workspace szablonu,
niezależnie od tego, co tu wybierzesz.

Inwentarz, na którym jest wykonywane zadanie. Pole początkowo zawiera inwentarz szablonu i
wyświetla inwentarze projektu utworzone dla aplikacji szablonu. Gdy wybrany inwentarz zniknie przed
rozpoczęciem zadania, zadanie jest wykonywane na inwentarzu szablonu.

## Limit {#taskParams-limit}

Wyświetlane dla szablonu Ansible, który zezwala na nadpisanie limitu.

Hosty lub grupy inwentarza, na których jest uruchamiany playbook, jeden wzorzec na chip, na
przykład `web` lub `10.0.0.11`. Semaphore przekazuje je jako `--limit`, połączone przecinkami.
Lista początkowo zawiera limit szablonu. Usuń wszystkie chipy, aby uruchomić playbook na wszystkich
hostach inwentarza.

## Tagi {#taskParams-tags}

Wyświetlane dla szablonu Ansible, który zezwala na nadpisanie tagów.

Wykonywane są tylko zadania i role playbooka, które mają jeden z tych tagów: Semaphore przekazuje
je jako `--tags`, jeden tag na chip. Lista początkowo zawiera tagi szablonu. Usuń wszystkie chipy,
aby uruchomić cały playbook.

## Pomiń tagi {#taskParams-skipTags}

Wyświetlane dla szablonu Ansible, który zezwala na nadpisanie pomijanych tagów.

Zadania i role playbooka, które mają jeden z tych tagów, nie są wykonywane: Semaphore przekazuje
je jako `--skip-tags`, jeden tag na chip. Lista początkowo zawiera pomijane tagi szablonu.

## Debug {#taskParams-debug}

Wyświetlane dla szablonu Ansible, który zezwala na debugowanie.

Uruchamia `ansible-playbook` z flagą szczegółowości, dzięki czemu dziennik zadania pokazuje więcej
szczegółów. Etykieta pokazuje flagę: `-v` tyle razy, ile wskazuje **Poziom debugowania**.

## Poziom debugowania {#taskParams-debugLevel}

Wyświetlane razem z **Debug** i włączone, gdy zaznaczono **Debug**.

Szczegółowość dziennika, od 1 (`-v`) do 6 (`-vvvvvv`). Wartość domyślna to 4 (`-vvvv`), która
pokazuje także szczegóły połączeń z hostami.

## Symulacja {#taskParams-dryRun}

Wyświetlane dla szablonu Ansible, chyba że szablon je ukrywa.

Uruchamia playbook z `--check`: Ansible zgłasza, co zmieniłby na hostach, i niczego nie zmienia.

## Różnice {#taskParams-diff}

Wyświetlane dla szablonu Ansible, chyba że szablon je ukrywa.

Uruchamia playbook z `--diff`: dziennik pokazuje zmiany wprowadzone w plikach i szablonach na
hostach. Zaznacz także **Symulacja**, aby zobaczyć zmiany bez ich wprowadzania.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Wyświetlane dla szablonu Ansible, który zezwala na nadpisanie instalacji Galaxy.

Przed uruchomieniem playbooka Semaphore instaluje role i kolekcje wymienione w plikach
`requirements.yml` repozytorium, gdy któryś plik zmienił się od ostatniej instalacji. Zaznacz pole,
aby pominąć ten krok dla tego zadania.

Pole jest początkowo odznaczone niezależnie od ustawienia szablonu, a odznaczone pole instaluje
wymagania.

## Plan {#taskParams-plan}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Wykonuje `plan` i na tym kończy zadanie: nic nie jest stosowane. Użyj tego, aby najpierw przejrzeć
zmiany. Bez tej opcji zadanie przechodzi do zastosowania planu, gdy zawiera on zmiany.

## Destroy {#taskParams-destroy}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Dodaje `-destroy` do `plan` i `apply`: zadanie planuje, a następnie stosuje usunięcie każdego
zasobu w stanie workspace. Zaznacz także **Plan**, aby zobaczyć, co zostałoby zniszczone, bez
niszczenia.

## Auto Approve {#taskParams-autoApprove}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Stosuje plan zaraz po jego utworzeniu. Bez tej opcji zadanie zatrzymuje się po planie i czeka, aż
ktoś je potwierdzi lub odrzuci na stronie zadania.

Pole działa tylko wtedy, gdy szablon zezwala na automatyczne zatwierdzanie. Szablon ustawiony na
zatwierdzanie zawsze stosuje plan bez pytania, niezależnie od stanu pola.

## Upgrade {#taskParams-upgrade}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Uruchamia `init` z `-upgrade`: dostawcy (providers) i moduły są aktualizowani do najnowszych
wersji, na które pozwalają ich ograniczenia wersji.

## Reconfigure {#taskParams-reconfigure}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Uruchamia `init` z `-reconfigure`: backend jest konfigurowany od nowa, a zapisana konfiguracja
backendu jest ignorowana. Bez tej opcji Semaphore uruchamia `init -migrate-state`, co przenosi
istniejący stan do zmienionego backendu.

## CLI args {#newTask-args}

Wyświetlane, gdy szablon zezwala na nadpisanie argumentów CLI.

Dodatkowe argumenty wiersza poleceń zadania, jeden na chip: `--forks=10` to jeden chip, a
`--forks` i `10` to dwa. Kliknij chip, aby go edytować, przeciągnij go, aby zmienić kolejność, a
ostatniego chipa użyj, aby dodać nowy.

Lista początkowo zawiera argumenty szablonu. Semaphore zawsze przekazuje argumenty szablonu i
dodaje tę listę po nich, więc gdy zmienisz listę, pozostawione w niej argumenty szablonu są
przekazywane dwukrotnie. Usuń je i zostaw tylko argumenty, które dodajesz.
