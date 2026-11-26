---
form: src/components/RepositoryForm.vue
---

# Edytuj Repository

Repozytorium to miejsce, z którego pochodzą playbooki, kod Terraform lub skrypty szablonu zadania:
repozytorium Git lub katalog na maszynie, która wykonuje zadania. Przed każdym zadaniem Semaphore
klonuje repozytorium Git lub aktualizuje wcześniej sklonowaną kopię, przełącza się na gałąź i
wykonuje zadanie w tej kopii.

## Nazwa {#repository-name}

Nazwa, pod którą repozytorium widnieje na stronie Repozytoria i na liście wyboru repozytorium w
szablonie zadania i w inwentarzu. Pole jest wymagane.

## URL lub ścieżka {#repository-gitUrl}

Gdzie znajduje się kod. Pole jest wymagane. Chipy pod polem pokazują rodzaj wpisanego adresu, a
kliknięcie chipa przepisuje adres na ten rodzaj.

- **ssh**: `git@github.com:org/repo.git` lub `ssh://git@github.com/org/repo.git`. Git loguje się
  jako użytkownik z adresu za pomocą klucza **Klucz SSH** wybranego w polu **Klucz dostępu**.
- **https** i **http**: `https://github.com/org/repo.git`. Prywatne repozytorium wymaga klucza
  **Logowanie z hasłem**; publiczne wymaga klucza **Brak**. Unikaj `http`: login i hasło są wtedy
  przesyłane bez szyfrowania.
- **git**: `git://host/org/repo.git`, protokół Git tylko do odczytu, bez logowania.
- **file**: `file:///srv/git/repo.git`, repozytorium Git na dysku maszyny, która wykonuje
  zadanie, klonowane jak repozytorium zdalne.
- **absolutna ścieżka**: ścieżka bezwzględna, taka jak `/srv/playbooks` lub `D:\playbooks`.
  Semaphore jej nie klonuje: zadanie jest wykonywane bezpośrednio w tym katalogu, który musi
  istnieć na maszynie wykonującej zadanie (serwerze Semaphore lub uruchamiaczu). Gałąź jej nie
  dotyczy.

Zmiana adresu usuwa kopie, które serwer Semaphore sklonował ze starego adresu.

## Gałąź {#repository-gitBranch}

Gałąź, którą Semaphore klonuje i aktualizuje, na przykład `main`. Pole jest wymagane, z wyjątkiem
ścieżki bezwzględnej, dla której jest zablokowane.

Szablon zadania może wskazać własną gałąź, która ma pierwszeństwo przed tą, a szablon może
pozwolić swoim zadaniom wybrać gałąź przy uruchomieniu. Tag również działa, ale tagu nie można
pobrać przez pull, więc Semaphore klonuje repozytorium od nowa przed każdym zadaniem.

## Single checkout for all templates {#repository-monorepo}

Domyślnie wyłączone: każdy szablon zadania otrzymuje własną kopię repozytorium. Zaznacz pole, aby
sklonować repozytorium raz i współdzielić tę kopię między wszystkimi szablonami, które go używają,
w przypadku repozytorium zbyt dużego, by trzymać kopię dla każdego szablonu.

Współdzielona kopia pozostaje na gałęzi tego repozytorium: szablony, które jej używają, nie mogą
mieć własnej gałęzi, a ich zadania nie mogą wybrać innej gałęzi ani commita. Zablokowane dla
ścieżki bezwzględnej.

## Klucz dostępu {#repository-sshKeyId}

Klucz z Magazynu kluczy, za pomocą którego Git loguje się do repozytorium. Pole jest wymagane.
Oferuje klucze trzech typów, a wybór zależy od adresu URL:

- **Klucz SSH** dla adresu `ssh`. Semaphore wczytuje klucz prywatny do agenta SSH na czas
  klonowania lub aktualizacji. Użytkownikiem jest ten z adresu URL, jak `git` w
  `git@github.com:org/repo.git`.
- **Logowanie z hasłem** dla adresu `https`: login i hasło lub token dostępu serwera Git. Przy
  pustym loginie hasło jest wysyłane samo, jako nazwa użytkownika, czego niektóre serwery
  oczekują w przypadku tokenu. Te same dane logowania są przekazywane do `ansible-galaxy` dla ról
  i kolekcji z tego samego serwera.
- **Brak** dla publicznego repozytorium, adresu `git` lub `file` albo ścieżki bezwzględnej.

Klucz, który nie pasuje do adresu URL, nie przekazuje Git żadnych danych logowania, a klonowanie
prywatnego repozytorium kończy się wtedy niepowodzeniem.
