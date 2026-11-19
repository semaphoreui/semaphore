---
form: src/components/KeyForm.vue
---

# Edytuj Key

Klucz z Magazynu kluczy przechowuje poświadczenie, którego projekt używa bez jego ujawniania: klucz
SSH lub hasło, za pomocą których Ansible loguje się do hostów, dane logowania do repozytorium Git,
hasło skarbca Ansible, sekret integracji. Inwentarze, repozytoria, szablony zadań i integracje
odwołują się do klucza po jego nazwie. Po zapisaniu sekret nigdy nie jest odsyłany do
przeglądarki.

## Nazwa klucza {#key-name}

Nazwa, pod którą klucz widnieje w Magazynie kluczy i na listach wyboru kluczy w inwentarzach,
repozytoriach, szablonach zadań i integracjach. Pole jest wymagane. Wybierz nazwę, która mówi, co
klucz otwiera, na przykład `Deploy key of the web servers`.

## Źródło {#key-source}

Gdzie przechowywany jest sekret klucza. Karty:

- **Local**: sekret jest wpisywany w tym oknie dialogowym i przechowywany w bazie danych
  Semaphore, zaszyfrowany, gdy administrator ustawił klucz szyfrowania dla kluczy dostępu. To
  ustawienie domyślne.
- **Storage**: sekret jest przechowywany w zewnętrznym magazynie sekretów projektu, wybranym
  poniżej w polu **Storage**. Ta karta jest dostępna tylko w wersji Pro. Same magazyny
  konfiguruje się na karcie **Storages** Magazynu kluczy.
- **Env**: sekret jest odczytywany ze zmiennej środowiskowej procesu serwera Semaphore. Wpisz
  nazwę zmiennej w polu pod kartami.
- **File**: sekret jest odczytywany z pliku na serwerze Semaphore. Wpisz bezwzględną ścieżkę do
  pliku w polu pod kartami. Plik musi znajdować się w katalogu sekretów serwera (`dirs.secrets` w
  jego konfiguracji, domyślnie `/tmp/semaphore`); ścieżka spoza niego lub zawierająca `..`
  powoduje błąd przy użyciu klucza. Końcowy znak nowego wiersza pliku jest pomijany.

Przy **Env** i **File** Semaphore nigdy nie zapisuje sekretu: odczytuje wartość za każdym razem,
gdy zadanie używa klucza, więc sekret zmieniasz tam, gdzie się znajduje. Wartość musi pasować do
pola **Typ** klucza. Klucz **Ciąg znaków** przyjmuje wartość taką, jaka jest. Klucz **Klucz SSH**
lub **Logowanie z hasłem** oczekuje JSON z polami danego typu, na przykład:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Podczas edycji klucza karty są zablokowane, dopóki nie zaznaczysz **Nadpisz**. Klucz
zsynchronizowany z magazynu sekretów zachowuje swoje źródło, a jego nazwy i typu nie można zmienić.

## Storage {#key-sourceStorageId}

Wyświetlane, gdy wybrano kartę **Storage**.

Magazyn sekretów projektu, który przechowuje sekret: HashiCorp Vault, OpenBao, Devolutions Server,
AWS Secrets Manager lub Azure Key Vault, skonfigurowany na karcie **Storages** Magazynu kluczy.

- Magazyn, który nie jest tylko do odczytu, otrzymuje sekret wpisany w tym oknie dialogowym:
  Semaphore zapisuje go tam przy zapisie. Usunięcie klucza usuwa również sekret z magazynu. Klucza
  przechowywanego w takim magazynie nie można później przenieść do innego magazynu.
- Magazyn tylko do odczytu jest wyłącznie odczytywany. Okno dialogowe nie wyświetla wtedy pól
  sekretu, a Semaphore odczytuje go z magazynu za każdym razem, gdy zadanie używa klucza.

## Source Key {#key-sourceStorageKey}

Wyświetlane, gdy wybrano **Storage**.

Miejsce sekretu w tym magazynie, na przykład ścieżka do sekretu w HashiCorp Vault.

## Typ {#key-type}

Jakiego rodzaju poświadczenie przechowuje klucz. Pole jest wymagane, a pola poniżej zmieniają się w
zależności od typu.

- **Klucz SSH**: klucz prywatny z opcjonalnym loginem i hasłem klucza. Ansible łączy się za jego
  pomocą z hostami, a Git klonuje nim repozytorium o adresie `ssh`.
- **Logowanie z hasłem**: login i hasło. Ansible łączy się za ich pomocą z hostami lub staje się
  za ich pomocą innym użytkownikiem (**Dane logowania sudo** inwentarza); Git klonuje nimi
  repozytorium `https`; hasło otwiera skarbiec Ansible szablonu zadania.
- **Ciąg znaków**: pojedyncza wartość tajna bez loginu: token, klucz API, hasło. Używaj go jako
  sekretu integracji. Nie może być kluczem SSH, danymi logowania sudo, hasłem skarbca ani kluczem
  dostępu repozytorium.
- **Brak**: brak poświadczenia. Używaj go dla repozytorium, które nie wymaga logowania (publiczne
  repozytorium `https`, adres `file` lub ścieżka lokalna), oraz dla inwentarza, którego hosty nie
  wymagają danych logowania lub same określają sposób połączenia.

## Nazwa użytkownika {#key-loginPasswordLogin}

Wyświetlane, gdy typ to **Logowanie z hasłem**. Opcjonalne.

Nazwa użytkownika poświadczenia.

- Ansible łączy się z hostami jako ten użytkownik (`--user`); jako **Dane logowania sudo**
  inwentarza jest to użytkownik, którym staje się Ansible (`--become-user`). Gdy pole jest puste,
  Ansible używa użytkownika wskazanego przez inwentarz lub własną konfigurację.
- W przypadku repozytorium `https` Git loguje się jako ten użytkownik z hasłem. Pozostaw pole
  puste dla tokenu, który serwer Git przyjmuje samodzielnie: hasło jest wtedy wysyłane jako nazwa
  użytkownika.

## Hasło {#key-loginPasswordPassword}

Wyświetlane, gdy typ to **Logowanie z hasłem**. Wymagane.

Hasło poświadczenia. Semaphore odpowiada nim na monity o hasło Ansible: hasło SSH (`--ask-pass`)
dla danych logowania użytkownika, hasło `become` (`--ask-become-pass`) dla danych logowania sudo
i hasło skarbca dla skarbca Ansible. W przypadku repozytorium `https` jest to hasło lub token
dostępu serwera Git.

## Wartość {#key-stringValue}

Wyświetlane, gdy typ to **Ciąg znaków**. Wymagane.

Sama wartość tajna, na przykład token, względem którego integracja sprawdza otrzymywane żądania.
Jest przechowywana tak, jak została wpisana.

## Nazwa użytkownika {#key-sshLogin}

Wyświetlane, gdy typ to **Klucz SSH**. Opcjonalne.

Użytkownik, jako który Ansible łączy się z hostami (`--user`). Gdy pole jest puste, Ansible używa
użytkownika wskazanego przez inwentarz lub własną konfigurację.

W przypadku repozytorium Git umieść użytkownika w adresie URL repozytorium, jak w
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Wyświetlane, gdy typ to **Klucz SSH**. Opcjonalne.

Hasło, którym zaszyfrowany jest klucz prywatny. Pozostaw pole puste dla klucza bez hasła.
Nieprawidłowe hasło powoduje, że każde zadanie używające klucza kończy się niepowodzeniem przy
wczytywaniu klucza. Klucz wygenerowany przez Semaphore nie ma hasła: zawartość tego pola jest
wtedy odrzucana.

## Generate SSH Key {#key-generateSshKey}

Wyświetlane, gdy typ to **Klucz SSH**.

Semaphore tworzy nową parę kluczy przy zapisie: 2048-bitowy klucz prywatny RSA, który pozostaje w
kluczu, oraz jego klucz publiczny, który jest wyświetlany raz w oknie dialogowym po zapisaniu, a
następnie w sekcji **Public key** przy każdej edycji klucza. Dodaj klucz publiczny do
`~/.ssh/authorized_keys` na hostach lub jako klucz wdrożeniowy (deploy key) repozytorium Git.

Pole klucza prywatnego jest zablokowane, gdy to pole jest zaznaczone. W przypadku istniejącego
klucza najpierw zaznacz **Nadpisz**: stara para kluczy zostanie zastąpiona. Klucza, którego sekret
jest odczytywany z **Env**, **File** lub magazynu tylko do odczytu, nie można wygenerować, ponieważ
Semaphore nie ma go gdzie zapisać.

## Klucz prywatny {#key-sshPrivateKey}

Wyświetlane, gdy typ to **Klucz SSH**. Wymagane, chyba że zaznaczono **Generate SSH Key**.

Klucz prywatny w formacie OpenSSH lub PEM, wraz z wierszami `BEGIN` i `END`. Na potrzeby zadania
Semaphore wczytuje go do własnego agenta SSH, z którego Ansible i Git korzystają przez cały czas
trwania zadania.

## Public key {#key-publicKey}

Wyświetlane, gdy edytujesz klucz SSH wygenerowany przez Semaphore, a **Nadpisz** nie jest
zaznaczone.

Publiczna połowa wygenerowanej pary kluczy. **Show** rozwija go, przycisk kopiowania umieszcza go w
schowku. Dodaj go do `~/.ssh/authorized_keys` na hostach lub jako klucz wdrożeniowy (deploy key)
repozytorium Git.

Jest to klucz publiczny pary, którą Semaphore wygenerował ostatnio. Po ręcznym zastąpieniu klucza
prywatnego za pomocą **Nadpisz** wyświetlany tu klucz już do niego nie pasuje.

## Nadpisz {#key-overrideSecret}

Wyświetlane, gdy edytujesz istniejący klucz.

Semaphore nigdy ponownie nie wyświetla zapisanego sekretu, więc pola sekretu są puste i
zablokowane. Zaznacz pole, aby zastąpić sekret: źródło, typ i pola sekretu odblokowują się, a
wszystko, co się w nich znajduje, jest zapisywane od nowa, więc wpisz ponownie cały sekret, a nie
tylko zmienianą część. Bez tego zapis zmienia tylko nazwę klucza.

Zapisanie klucza usuwa również kopie repozytoriów, które go używają, przechowywane przez serwer
Semaphore: następne zadanie sklonuje je ponownie.
