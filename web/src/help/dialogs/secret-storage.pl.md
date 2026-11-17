---
form: src/components/SecretStorageForm.vue, src/components/SecretSourceToggle.vue
skip: secretStorageSync-syncInterval
---

# Edytuj Secret Storage

Magazyn sekretów to zewnętrzny menedżer sekretów, w którym Semaphore przechowuje sekrety i z
którego je odczytuje: HashiCorp Vault, OpenBao, Devolutions Server, AWS Secrets Manager lub Azure
Key Vault. Grupa zmiennych utworzona z magazynem przechowuje w nim swoje sekrety, a magazyn z
włączoną synchronizacją importuje swoje sekrety do Magazynu kluczy projektu. Rodzaj magazynu
wybiera się w menu **New Storage** i nie można go zmienić w tym oknie dialogowym; pola poniżej
zależą od niego.

## Nazwa {#secretStorage-name}

Nazwa, pod którą magazyn widnieje na karcie **Storages** Magazynu kluczy i na liście wyboru
**Secret storage** nowej grupy zmiennych. Pole jest wymagane.

## Server URL {#secretStorage-vaultURL}

Wyświetlane, gdy magazyn to HashiCorp Vault, OpenBao lub Devolutions Server.

Adres serwera wraz ze schematem, na przykład `https://vault.example.com:8200`. Pole jest wymagane.

## Mount {#secretStorage-vaultMount}

Wyświetlane, gdy magazyn to HashiCorp Vault lub OpenBao. Opcjonalne.

Ścieżka montowania silnika sekretów, który przechowuje sekrety. Pozostaw pole puste, aby użyć
domyślnej wartości `secret`.

## Namespace {#secretStorage-vaultNamespace}

Wyświetlane, gdy magazyn to HashiCorp Vault lub OpenBao. Opcjonalne.

Przestrzeń nazw, w której znajdują się sekrety. Przestrzenie nazw istnieją w Vault Enterprise i
HCP Vault Dedicated oraz w OpenBao od wersji 2.3. Pozostaw pole puste dla serwera bez przestrzeni
nazw.

## Źródło sekretu {#secretSource-options}

Wyświetlane dla każdego rodzaju magazynu, nad sekretem, którym loguje się Semaphore: **Token**
Vault lub OpenBao, sekretem aplikacji Devolutions Server, kluczem tajnym AWS, sekretem klienta
Azure.

Skąd Semaphore pobiera ten sekret:

- **Store in DB**: wpisujesz sekret w tym oknie dialogowym; Semaphore szyfruje go i przechowuje w
  swojej bazie danych.
- **From ENV**: wpisujesz nazwę zmiennej środowiskowej serwera Semaphore; sekret jest z niej
  odczytywany przy każdym użyciu magazynu.
- **From File**: wpisujesz bezwzględną ścieżkę do pliku na serwerze Semaphore; sekret jest z niego
  odczytywany za każdym razem, bez końcowego znaku nowego wiersza. Plik musi znajdować się w
  katalogu sekretów serwera (`dirs.secrets` w konfiguracji, domyślnie `/tmp/semaphore`), a
  ścieżka zawierająca `..` jest odrzucana.

Zmiana źródła czyści pole pod nim.

## Token {#secretStorage-vaultToken}

Wyświetlane, gdy magazyn to HashiCorp Vault lub OpenBao, a źródło sekretu to **Store in DB**.

Token, którym Semaphore uwierzytelnia się na serwerze. Musi mieć uprawnienia do odczytu sekretów
oraz, o ile magazyn nie jest **Read only**, do ich zapisu i usuwania.

Pole jest wymagane dla nowego magazynu. Gdy otwierasz istniejący magazyn, jest puste: pozostaw je
puste, aby zachować zapisany token, lub wpisz nowy, aby go zastąpić.

## Env var name / Path to the file {#secretStorage-vaultTokenSource}

Wyświetlane, gdy magazyn to HashiCorp Vault lub OpenBao, a źródło sekretu to **From ENV** lub
**From File**.

Nazwa zmiennej środowiskowej lub ścieżka do pliku, który zawiera token. Pole jest wymagane dla
nowego magazynu.

## Skip TLS certificate verification {#secretStorage-dvlsInsecureTls}

Wyświetlane, gdy magazyn to Devolutions Server.

Semaphore łączy się z serwerem bez sprawdzania jego certyfikatu TLS. Każdy, kto znajduje się
między Semaphore a serwerem, może wtedy odczytać sekrety. Pozostaw tę opcję wyłączoną i zamiast
tego wyposaż serwer w certyfikat, któremu Semaphore ufa.

## Vault ID {#secretStorage-dvlsKey}

Wyświetlane, gdy magazyn to Devolutions Server.

ID skarbca Devolutions Server, który przechowuje sekrety. Pole jest wymagane dla nowego magazynu.

## App Key {#secretStorage-dvlsAppKey}

Wyświetlane, gdy magazyn to Devolutions Server.

Klucz aplikacji tożsamości aplikacji, którą loguje się Semaphore, tej, której sekret trafia do
pola **Secret**. Pole jest wymagane dla nowego magazynu.

## Secret {#secretStorage-dvlsSecret}

Wyświetlane, gdy magazyn to Devolutions Server, a źródło sekretu to **Store in DB**.

Sekret aplikacji tożsamości aplikacji z pola **App Key**. Pole jest wymagane dla nowego magazynu.
Gdy otwierasz istniejący magazyn, jest puste: pozostaw je puste, aby zachować zapisany sekret, lub
wpisz nowy, aby go zastąpić.

## Env var name / Path to the file {#secretStorage-dvlsEnv}

Wyświetlane, gdy magazyn to Devolutions Server, a źródło sekretu to **From ENV** lub
**From File**.

Nazwa zmiennej środowiskowej lub ścieżka do pliku, który zawiera sekret aplikacji. Pole jest
wymagane dla nowego magazynu.

## Region {#secretStorage-awsRegion}

Wyświetlane, gdy magazyn to AWS Secrets Manager.

Region AWS, w którym znajdują się sekrety, na przykład `us-east-1`. Pole jest wymagane.

## Endpoint URL {#secretStorage-awsEndpointURL}

Wyświetlane, gdy magazyn to AWS Secrets Manager. Opcjonalne.

Adres API Secrets Manager, który ma być wywoływany zamiast domyślnego punktu końcowego regionu, na
przykład punkt końcowy VPC. Pozostaw pole puste, aby użyć domyślnego punktu końcowego AWS.

## Use IAM Role / Instance Profile {#secretStorage-awsUseIamRole}

Wyświetlane, gdy magazyn to AWS Secrets Manager.

Gdy opcja jest włączona, Semaphore nie przechowuje klucza dostępu i pobiera poświadczenia AWS ze
środowiska, w którym działa serwer, na przykład z roli IAM jego instancji. Pola klucza dostępu
poniżej są wtedy wyłączone.

Włączenie opcji usuwa zapisany klucz dostępu przy zapisie. Ponowne jej wyłączenie wymaga nowego
**Secret Access Key**.

## Access Key ID {#secretStorage-awsAccessKeyId}

Wyświetlane, gdy magazyn to AWS Secrets Manager. Wyłączone, gdy
**Use IAM Role / Instance Profile** jest włączone.

ID klucza dostępu, którym loguje się Semaphore, na przykład `AKIAIOSFODNN7EXAMPLE`. Pole jest
wymagane, gdy rola IAM jest wyłączona. ID nie jest tajne: jest przechowywane i wyświetlane w
niezmienionej postaci.

## Secret Access Key {#secretStorage-awsSecretKey}

Wyświetlane, gdy magazyn to AWS Secrets Manager, a źródło sekretu to **Store in DB**.
Wyłączone, gdy **Use IAM Role / Instance Profile** jest włączone.

Sekret klucza dostępu. Pole jest wymagane dla nowego magazynu i po wyłączeniu roli IAM. W
pozostałych przypadkach jest puste po otwarciu magazynu: pozostaw je puste, aby zachować zapisany
klucz, lub wpisz nowy, aby go zastąpić.

## Env var name / Path to the file {#secretStorage-awsSecretKeySource}

Wyświetlane, gdy magazyn to AWS Secrets Manager, a źródło sekretu to **From ENV** lub
**From File**. Wyłączone, gdy **Use IAM Role / Instance Profile** jest włączone.

Nazwa zmiennej środowiskowej lub ścieżka do pliku, który zawiera tajny klucz dostępu. Pole jest
wymagane w tych samych przypadkach co **Secret Access Key**.

## Vault URL {#secretStorage-azureVaultURL}

Wyświetlane, gdy magazyn to Azure Key Vault.

Adres magazynu kluczy, na przykład `https://my-vault.vault.azure.net`. Pole jest wymagane.

## Tenant ID {#secretStorage-azureTenantId}

Wyświetlane, gdy magazyn to Azure Key Vault.

ID dzierżawy Microsoft Entra aplikacji, jako która loguje się Semaphore. Pole jest wymagane.

## Client ID {#secretStorage-azureClientId}

Wyświetlane, gdy magazyn to Azure Key Vault.

ID aplikacji (klienta) tej aplikacji. Musi ona mieć dostęp do sekretów magazynu kluczy. Pole jest
wymagane.

## Client Secret {#secretStorage-azureClientSecret}

Wyświetlane, gdy magazyn to Azure Key Vault, a źródło sekretu to **Store in DB**.

Sekret klienta aplikacji. Pole jest wymagane dla nowego magazynu. Gdy otwierasz istniejący
magazyn, jest puste: pozostaw je puste, aby zachować zapisany sekret, lub wpisz nowy, aby go
zastąpić.

## Env var name / Path to the file {#secretStorage-azureClientSecretSource}

Wyświetlane, gdy magazyn to Azure Key Vault, a źródło sekretu to **From ENV** lub
**From File**.

Nazwa zmiennej środowiskowej lub ścieżka do pliku, który zawiera sekret klienta. Pole jest
wymagane dla nowego magazynu.

## Read only {#secretStorage-readonly}

Gdy opcja jest włączona, Semaphore tylko odczytuje dane z magazynu. Sekretu grupy zmiennych nie
można w nim utworzyć ani zmienić: zapis kończy się niepowodzeniem. Usunięcie sekretu, klucza lub
grupy zmiennych w Semaphore pozostawia jego sekret w magazynie.

Gdy opcja jest wyłączona, Semaphore zapisuje w magazynie sekrety grup zmiennych, które go używają,
i usuwa je z niego razem z grupami.

## Sync keys enabled {#secretStorage-syncEnabled}

Włącz, aby importować sekrety z magazynu do Magazynu kluczy projektu ze ścieżek ustawionych w
**Sync paths**. Import jest wykonywany po naciśnięciu przycisku synchronizacji magazynu na karcie
**Storages** oraz samoczynnie w odstępie ustawionym w **Sync paths**.

Zaimportowany klucz jest oznaczony jako zsynchronizowany: nie można zmienić jego nazwy ani typu, a
usunięcie go w Semaphore pozostawia sekret w magazynie. Usunięcie magazynu przy włączonej
synchronizacji usuwa również zaimportowane przez niego klucze.

## Sync paths {#secretStorage-syncPaths}

Wyświetlane, gdy **Sync keys enabled** jest włączone.

Otwiera okno dialogowe ze ścieżkami magazynu, z których odbywa się import, oraz odstępem
automatycznego importu. Liczba na przycisku to liczba ścieżek. Przycisk synchronizacji magazynu
pozostaje wyłączony, dopóki magazyn nie ma co najmniej jednej ścieżki.

Okno dialogowe ma tylko przycisk **Zamknij**: to, co w nim zmienisz, jest zapisywane z magazynem
przyciskiem **Zapisz** tego okna dialogowego (**Utwórz** dla nowego magazynu).
