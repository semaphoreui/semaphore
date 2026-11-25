---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Edytuj szablon 'Ansible Playbook'

Szablon zadania określa, co i jak uruchamia Semaphore: repozytorium i plik w nim, inwentarz i
zmienne oraz to, które opcje użytkownik może zmienić przy uruchamianiu zadania. Każde zadanie jest
uruchamiane z szablonu. Pola okna dialogowego zależą od aplikacji szablonu (Ansible, Terraform,
OpenTofu, Terragrunt, Bash, Python, PowerShell i inne): pole, które mają tylko niektóre aplikacje,
mówi o tym. Zmiany obowiązują dla zadań uruchomionych po **Zapisz**.

## Typ szablonu {#templateForm-typeTabs}

Karty u góry określają rodzaj szablonu:

- **Zadanie**: uruchamia kod. Większość szablonów jest tego rodzaju.
- **Buduj**: tworzy artefakt. Każde zadanie budowy otrzymuje wersję, liczoną od
  **Wersja początkowa**.
- **Wdrażaj**: wdraża artefakt utworzony przez budowę. Przy uruchamianiu wybierasz zadanie budowy,
  którego wersję wdraża.

Zadanie budowy lub wdrożenia otrzymuje w swoim środowisku zmienne `SEMAPHORE_TASK_TYPE`,
`SEMAPHORE_TASK_TARGET_VERSION` (budowa) i `SEMAPHORE_TASK_INCOMING_VERSION` (wdrożenie). Ansible
otrzymuje te same wartości w dodatkowej zmiennej `semaphore_vars.task_details`.

## Wersja początkowa {#templateForm-startVersion}

Wyświetlane, gdy typ to **Buduj**. Wymagane.

Wersja pierwszego zadania budowy, na przykład `1.0.0` lub `v0.1`. Każda kolejna budowa zwiększa o
jeden ostatnią liczbę poprzedniej wersji i zachowuje otaczający ją tekst oraz dopełnienie zerami:
`1.0.0`, `1.0.1`, `1.0.2`. Podnieś tu liczbę powyżej ostatniej budowy, aby przeskoczyć dalej, na
przykład do `2.0.0`; wersja bez liczby nigdy się nie zmienia.

## Szablon budowy {#templateForm-buildTemplateId}

Wyświetlane, gdy typ to **Wdrażaj**. Wymagane.

Szablon, którego zadania tworzą artefakty wdrażane przez ten szablon. Lista zawiera szablony budowy
projektu, a pod nimi szablony wdrożeń: wdrożenie może następować po innym wdrożeniu i otrzymuje
wtedy wersję, którą tamto wdrożenie otrzymało od swojej budowy.

## Automatyczne uruchamianie {#templateForm-autorun}

Wyświetlane, gdy typ to **Wdrażaj**.

Samoczynnie uruchamia zadanie tego szablonu za każdym razem, gdy zadanie z **Szablon budowy**
zakończy się powodzeniem, z wersją, którą to zadanie utworzyło. Nieudana budowa niczego nie
uruchamia.

## Nazwa {#templateForm-name}

Nazwa, pod którą szablon widnieje na stronie Szablony zadań. Pole jest wymagane.

## Repozytorium {#templateForm-repositoryId}

Repozytorium projektu, które zawiera kod do uruchomienia. Pole jest wymagane. Przed każdym zadaniem
Semaphore aktualizuje swoją kopię repozytorium i uruchamia z niej kod.

Etykieta po prawej stronie pola pokazuje gałąź, którą pobierają zadania. Kliknij ją, aby otworzyć
pole **Gałąź**.

## Gałąź {#templateForm-gitBranch}

Wyświetlane po kliknięciu etykiety gałęzi po prawej stronie pola **Repozytorium**.

Gałąź lub tag, które pobierają zadania tego szablonu, zamiast gałęzi ustawionej w repozytorium.
Pozostaw pole puste, aby użyć gałęzi repozytorium. Gdy Semaphore może odczytać gałęzie
repozytorium, pole je wyświetla.

Pole jest wyłączone, gdy repozytorium jest monorepo: wszystkie szablony współdzielą jedną jego
kopię, na gałęzi repozytorium.

## Ścieżka do pliku playbook {#templateForm-playbook}

Co uruchomić, jako ścieżka od katalogu głównego repozytorium, na przykład `deploy/site.yml`.
Ścieżka nie może być bezwzględna i nie może wychodzić poza repozytorium. Lista oferuje pliki
repozytorium; możesz też wpisać ścieżkę, której nie ma na liście. Przycisk po prawej stronie pola
ponownie wczytuje listę.

Etykieta i znaczenie zmieniają się w zależności od aplikacji:

- **Ansible**: playbook. Pole jest wymagane; lista oferuje pliki `.yml` i `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, katalog z kodem. Pozostaw
  pole puste, aby uruchomić kod w katalogu głównym repozytorium. Wpisz `/`, aby wyświetlić
  zawartość katalogu.
- **Bash**, **Python**, **PowerShell** i inne aplikacje: **Script Filename**, skrypt do
  uruchomienia. Pole jest wymagane. Skrypt działa w katalogu głównym repozytorium.

## Different working directory {#templateForm-workingDirectoryToggle}

Wyświetlane dla szablonów Ansible.

Uruchamia `ansible-playbook` w katalogu innym niż katalog główny repozytorium. Zaznacz, aby
wyświetlić pole **Working directory**. Odznacz, aby ponownie uruchamiać w katalogu głównym.

## Working directory {#templateForm-workingDirectory}

Wyświetlane, gdy zaznaczono **Different working directory**. Wymagane.

Katalog, w którym działają `ansible-playbook` i `ansible-galaxy`, jako ścieżka od katalogu
głównego repozytorium, na przykład `ansible`. Użyj go, gdy plik `ansible.cfg` projektu znajduje się
w podkatalogu, ponieważ Ansible odczytuje ten plik z bieżącego katalogu. Ścieżka playbooka pozostaje
względna wobec katalogu głównego repozytorium. Ścieżka nie może być bezwzględna i nie może wychodzić
poza repozytorium.

## Inwentarz {#templateForm-inventoryId}

Wyświetlane dla szablonów Ansible, Terraform, OpenTofu i Terragrunt.

- **Ansible**: inwentarz, na którym jest uruchamiany playbook, wraz z danymi logowania do hostów.
  Pole jest wymagane.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, workspace, w którym działa kod.
  Semaphore wybiera go przed planem i tworzy, gdy nie istnieje. Nowy szablon zapisany bez
  workspace otrzymuje workspace o nazwie `default`.

Lista zawiera inwentarze projektu utworzone dla aplikacji szablonu.

## Grupy zmiennych {#templateForm-environmentIds}

Opcjonalne. Grupy zmiennych, których zmienne, zmienne środowiskowe i sekrety otrzymuje każde
zadanie tego szablonu. Możesz wybrać kilka: są scalane w kolejności w polu, a późniejsza grupa ma
pierwszeństwo przed wcześniejszą, gdy obie ustawiają tę samą nazwę.

Sposób, w jaki zmienne docierają do kodu, zależy od aplikacji: dodatkowe zmienne dla Ansible,
`-var` dla Terraform i OpenTofu, argumenty `name=value` dla skryptu. Zmienne środowiskowe są
ustawiane dla każdej aplikacji.

## Widok {#templateForm-viewId}

Opcjonalne. Widok, w którym widnieje szablon: widoki to karty strony Szablony zadań, które grupują
szablony projektu. Pozostaw pole puste, aby nie umieszczać szablonu w żadnym widoku.

## Tag uruchamiacza {#templateForm-runnerTag}

Wyświetlane, gdy dostępne są uruchamiacze projektu.

Zadania tego szablonu są wykonywane tylko na uruchamiaczu, który ma ten tag, nawet gdy serwer mógłby
wykonać je sam. Tag zastępuje tag uruchamiacza inwentarza. Pozostaw pole puste, aby użyć tagu
inwentarza lub żadnego.

## Executor image {#templateForm-executorImage}

Obraz kontenera, który uruchamiacz uruchamia dla zadań tego szablonu, na przykład
`semaphoreui/job:latest`. Pozostaw pole puste, aby użyć obrazu z konfiguracji uruchamiacza.

Tylko wykonawcy (executors) Docker i Kubernetes uruchamiają zadania w kontenerze, więc pole jest
włączone tylko wtedy, gdy jeden z nich jest dostępny. Zadanie wykonywane bez kontenera je ignoruje.

## Zmienne ankiety {#surveyVars}

Wartości, o które użytkownik jest pytany przy uruchamianiu zadania tego szablonu. Kliknij **+**,
aby dodać zmienną, chip, aby ją zmienić, krzyżyk na chipie, aby ją usunąć; przeciągaj chipy, aby
ustalić kolejność zadawania pytań.

Odpowiedzi docierają do kodu jak zmienne grupy zmiennych lub jako zmienne środowiskowe, zależnie od
ustawienia każdej zmiennej. Mają pierwszeństwo przed grupą zmiennych, która ustawia tę samą nazwę.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Pozwala zadaniom tego szablonu działać jednocześnie. Bez tej opcji zadanie szablonu czeka w
kolejce, aż wykonywane zadanie się zakończy. Limity równoległych zadań projektu i serwera
obowiązują w obu przypadkach.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Wyświetlane, gdy administrator włączył wydawanie JWT na serwerze.

Przekazuje każdemu zadaniu tego szablonu JSON Web Token podpisany przez Semaphore, w zmiennej
środowiskowej `SEMAPHORE_JWT`. Kod może przedstawić token innej usłudze, aby udowodnić, którym jest
zadaniem: token wskazuje zadanie, projekt, szablon i użytkownika, który je uruchomił. Usługa
sprawdza podpis za pomocą kluczy publicznych pod adresem `/.well-known/jwks.json` serwera
Semaphore.

## JWT audience {#templateForm-jwtAudience}

Wyświetlane, gdy zaznaczono **Issue JWT to task runner**. Opcjonalne.

Usługi, dla których przeznaczony jest token, zapisywane w jego roszczeniu `aud`, na przykład
`https://vault.example.com`. Naciśnij Enter po każdej wartości. Do 32 wartości; pozostaw pole
puste dla tokenu bez odbiorcy.

## JWT TTL {#templateForm-jwtTtl}

Wyświetlane, gdy zaznaczono **Issue JWT to task runner**. Opcjonalne.

Jak długo token jest ważny od rozpoczęcia zadania, jako czas trwania, taki jak `30s`, `15m` lub
`1h`. Pozostaw pole puste, aby użyć wartości domyślnej serwera, jednej godziny, chyba że
administrator ustawił inną. Wartość nie może przekraczać maksimum serwera, 24 godzin, chyba że
ustawiono inaczej.

## Chcę uruchomić zadanie przez cron tylko dla nowych commitów z niektórego repozytorium {#templateForm-cronVisible}

Uruchamia zadanie tego szablonu, gdy w repozytorium pojawi się nowy commit. Semaphore w wybranym
odstępie pyta repozytorium o jego ostatni commit i uruchamia zadanie tylko wtedy, gdy commit
zmienił się od ostatniego sprawdzenia.

Zaznaczenie pola i zapisanie tworzy harmonogram sprawdzania; odznaczenie i zapisanie usuwa ten
harmonogram. Pozostałe harmonogramy szablonu są zachowywane.

## Repozytorium {#templateForm-cronRepositoryId}

Wyświetlane, gdy zaznaczono **Chcę uruchomić zadanie przez cron tylko dla nowych commitów z
niektórego repozytorium**. Wymagane.

Repozytorium obserwowane pod kątem nowych commitów. Semaphore obserwuje gałąź ustawioną w tym
repozytorium, a nie **Gałąź** szablonu.

## interwałSprawdzania {#templateForm-cronFormat}

Wyświetlane, gdy zaznaczono **Chcę uruchomić zadanie przez cron tylko dla nowych commitów z
niektórego repozytorium**.

Jak często Semaphore sprawdza repozytorium pod kątem nowego commita: od co minutę do raz dziennie.
Wartość domyślna to co minutę.

## Tłumienie alertów o sukcesie {#templateForm-suppressSuccessAlerts}

Sprawia, że zadania tego szablonu zakończone powodzeniem nie wysyłają alertów. Alerty zadań
zakończonych niepowodzeniem są nadal wysyłane. Działa tylko wtedy, gdy alerty są włączone dla
projektu.

## Tłumienie alertów o błędach {#templateForm-suppressErrorAlerts}

Sprawia, że zadania tego szablonu zakończone niepowodzeniem nie wysyłają alertów. Alerty zadań
zakończonych powodzeniem są nadal wysyłane. Działa tylko wtedy, gdy alerty są włączone dla
projektu.

## CLI args {#templateForm-args}

Argumenty dodawane do polecenia każdego zadania. Każdy chip to jeden argument, przekazywany w
niezmienionej postaci, bez powłoki: umieść `--forks=10` w jednym chipie albo `--forks` i `10` w
dwóch. Przeciągaj chipy, aby zmienić ich kolejność. Przycisk po prawej stronie listy otwiera
argumenty jako tablicę JSON:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: dodawane do `ansible-playbook`, przed playbookiem.
- **Terraform**, **OpenTofu**, **Terragrunt**: dodawane do `plan` i `apply`, nie do `init`.
- Skrypt: dodawane po nazwie skryptu.

## Argumenty CLI {#templateForm-allowOverrideArgsInTask}

W sekcji **Podpowiedzi**. Pozwala użytkownikowi dodać własne argumenty przy uruchamianiu zadania.
Są one dodawane po argumentach szablonu.

## Gałąź {#templateForm-allowOverrideBranchInTask}

W sekcji **Podpowiedzi**. Pozwala użytkownikowi wybrać inną gałąź, tag lub commit przy
uruchamianiu zadania zamiast gałęzi szablonu. Bez tej opcji zadanie zawsze pobiera gałąź szablonu.

Pole jest wyłączone, gdy repozytorium jest monorepo.

## Inwentarz {#templateForm-allowOverrideInventory}

W sekcji **Podpowiedzi**. Wyświetlane dla szablonów Ansible, Terraform, OpenTofu i Terragrunt.

Pozwala użytkownikowi wybrać inny inwentarz przy uruchamianiu zadania. Tylko szablony Ansible
używają wybranego tam inwentarza: zadanie Terraform, OpenTofu lub Terragrunt działa w workspace
szablonu, nawet gdy pole jest zaznaczone.

## Limit {#templateForm-limit}

Wyświetlane dla szablonów Ansible.

Hosty lub grupy z inwentarza, na których jest uruchamiany playbook, przekazywane jako `--limit`.
Każdy chip to jeden wzorzec, na przykład `web` lub `db01`; pozostaw listę pustą, aby uruchomić
playbook na każdym hoście.

## Tagi {#templateForm-tags}

Wyświetlane dla szablonów Ansible.

Wykonuje tylko zadania i role z tymi tagami, przekazywanymi jako `--tags`. Pozostaw listę pustą,
aby wykonać wszystko.

## Pomiń tagi {#templateForm-skipTags}

Wyświetlane dla szablonów Ansible.

Pomija zadania i role z tymi tagami, przekazywanymi jako `--skip-tags`.

## Skarbce {#templateVaults}

Wyświetlane dla szablonów Ansible.

Hasła Ansible Vault, których playbook potrzebuje do odszyfrowania swoich plików. Kliknij **+**, aby
dodać hasło, chip, aby je zmienić, krzyżyk na chipie, aby je usunąć. Każdy skarbiec ma nazwę, swoje
vault ID, i dostarcza hasło z klucza Magazynu kluczy lub ze skryptu klienta z repozytorium.

## Automatyczne zatwierdzenie {#templateForm-autoApprove}

Wyświetlane dla szablonów Terraform, OpenTofu i Terragrunt.

Stosuje plan bez pytania. Bez tej opcji zadanie zatrzymuje się po planie i czeka, aż użytkownik je
potwierdzi lub odrzuci; plan bez zmian kończy zadanie bez niczego do zastosowania.

## Override backend settings {#templateForm-overrideBackend}

Wyświetlane dla szablonów Terraform, OpenTofu i Terragrunt, gdy Semaphore może przechowywać stan
Terraform.

Przechowuje stan kodu w Semaphore. Przed `init` Semaphore zapisuje w katalogu kodu plik ustawiający
backend `http` i usuwa go po zadaniu:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Wyświetlane razem z **Override backend settings**; włączone, gdy to pole jest zaznaczone.
Opcjonalne.

Nazwa pliku, do którego Semaphore zapisuje backend, na przykład `semaphore_backend.tf`. Wartość
domyślna to `backend.tf`. Nazwa musi kończyć się na `.tf` i zawierać tylko litery, cyfry, `_`, `-`
i `.`. Wybierz nazwę, której nie ma żaden plik twojego kodu: Semaphore nadpisuje ten plik i usuwa
go po zadaniu.

## Limit {#templateForm-allowOverrideLimit}

W sekcji **Ansible podpowiedzi**. Pozwala użytkownikowi ustawić hosty przy uruchamianiu zadania
zamiast **Limit** szablonu.

## Tagi {#templateForm-allowOverrideTags}

W sekcji **Ansible podpowiedzi**. Pozwala użytkownikowi ustawić tagi przy uruchamianiu zadania
zamiast **Tagi** szablonu.

## Pomiń tagi {#templateForm-allowOverrideSkipTags}

W sekcji **Ansible podpowiedzi**. Pozwala użytkownikowi ustawić pomijane tagi przy uruchamianiu
zadania zamiast **Pomiń tagi** szablonu.

## Debug {#templateForm-allowDebug}

W sekcji **Ansible podpowiedzi**. Pozwala użytkownikowi włączyć szczegółowe dane wyjściowe przy
uruchamianiu zadania, od `-v` do `-vvvvvv`.

## Symulacja {#templateForm-showDryRun}

W sekcji **Ansible podpowiedzi**. Domyślnie zaznaczone. Oferuje opcję **Symulacja** przy
uruchamianiu zadania, która uruchamia playbook z `--check`: Ansible zgłasza, co by zmienił, i
niczego nie zmienia. Odznacz pole, aby usunąć tę opcję.

## Różnice {#templateForm-showDiff}

W sekcji **Ansible podpowiedzi**. Domyślnie zaznaczone. Oferuje opcję **Różnice** przy
uruchamianiu zadania, która uruchamia playbook z `--diff`, aby pokazać zmiany wprowadzone w
plikach. Odznacz pole, aby usunąć tę opcję.

## Automatyczne zatwierdzenie {#templateForm-allowAutoApprove}

W sekcji podpowiedzi Terraform, OpenTofu i Terragrunt. Pozwala użytkownikowi zaznaczyć
**Auto Approve** przy uruchamianiu zadania, aby zastosować plan bez potwierdzenia. Okno dialogowe
zadania oferuje tę opcję w każdym przypadku, ale bez tego pola Semaphore ją ignoruje: zadanie czeka
na potwierdzenie po planie, chyba że zaznaczono **Automatyczne zatwierdzenie** szablonu.

## Galaxy install options {#templateForm-galaxy}

Wyświetlane dla szablonów Ansible. Kliknij, aby wyświetlić opcje instalacji Galaxy; etykieta
pokazuje liczbę opcji różniących się od domyślnych.

Przed uruchomieniem playbooka Semaphore instaluje za pomocą `ansible-galaxy` role i kolekcje,
których wymaga repozytorium. Szuka plików `requirements.yml`, `roles/requirements.yml` i
`collections/requirements.yml` obok playbooka i w katalogu głównym repozytorium i wykonuje
instalację tylko wtedy, gdy któryś plik zmienił się od ostatniego zadania.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

W sekcji **Galaxy install options**. Pomija instalację Galaxy: playbook działa z rolami i
kolekcjami zainstalowanymi wcześniej. Użyj tej opcji, gdy wymagania są instalowane w inny sposób
lub instalacja jest wolna i niepotrzebna.

## Role install args {#templateForm-galaxyRoleArgs}

W sekcji **Galaxy install options**. Argumenty dodawane do `ansible-galaxy role install`, jeden na
chip, na przykład `--ignore-certs` lub `--server=https://galaxy.example.com`. Semaphore akceptuje
tylko te flagi instalacji, które można bezpiecznie przekazać: nieznana flaga, `--token`,
`--api-key`, `-p` lub `--roles-path` uniemożliwia zapisanie szablonu.

## Collection install args {#templateForm-galaxyCollectionArgs}

W sekcji **Galaxy install options**. Argumenty dodawane do `ansible-galaxy collection install`,
jeden na chip, na przykład `--pre` lub `--upgrade`. Obowiązuje ta sama zasada co dla
**Role install args**: `--token`, `--api-key`, `-p` i `--collections-path` nie są akceptowane.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

W sekcji **Galaxy install options**, w części **Prompts**. Pozwala użytkownikowi wybrać przy
uruchamianiu zadania, czy pominąć instalację Galaxy. Wybór zadania ma pierwszeństwo przed
**Skip Galaxy install** szablonu.
