---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Edytuj grupę zmiennych

Grupa zmiennych zawiera dodatkowe zmienne, zmienne środowiskowe i sekrety, które Semaphore
przekazuje każdemu zadaniu szablonu zadania używającego tej grupy. Ansible otrzymuje dodatkowe
zmienne jako `--extra-vars`, Terraform i OpenTofu jako `-var`; zmienne środowiskowe są ustawiane
dla procesu zadania. Szablon może używać kilku grup: gdy dwie z nich definiują tę samą nazwę,
wygrywa grupa, która występuje w szablonie później.

## Nazwa grupy {#varGroup-name}

Nazwa, pod którą grupa widnieje na stronie Grupy zmiennych i na liście wyboru grupy zmiennych w
szablonie zadania. Pole jest wymagane.

## Secret storage {#varGroup-secretStorageId}

Wyświetlane, gdy grupa jest nowa, a serwer obsługuje magazyny sekretów. Opcjonalne.

Magazyn sekretów, w którym przechowywane są sekrety tej grupy, na przykład HashiCorp Vault.
Pozostaw pole puste, aby przechowywać sekrety w bazie danych Semaphore, zaszyfrowane.

Magazyn wybiera się raz, przy tworzeniu grupy: istniejąca grupa zachowuje swój magazyn, a pole nie
jest już wyświetlane. Magazyn oznaczony jako **Read only** nie przyjmuje nowych sekretów: zapisanie
w nim sekretu kończy się niepowodzeniem.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Wyświetlane, gdy grupa jest nowa, a serwer obsługuje magazyny sekretów. Dostępne po wybraniu
**Secret storage**.

Początek ścieżki, którą każdy sekret tej grupy otrzymuje w magazynie. Semaphore dodaje do niego 10
losowych znaków: z przedrostkiem `semaphore/prod/` sekret jest przechowywany jako
`semaphore/prod/`, po którym następują te znaki. Przedrostek jest przyjmowany w niezmienionej
postaci, więc zakończ go znakiem `/`, gdy sekrety mają się znaleźć pod nim jak w folderze.

Przedrostek ustawia się raz, przy tworzeniu grupy. Karta **Secrets** pokazuje go jako wzorzec
ścieżki źródłowej.

## Table, JSON lub YAML {#varGroup-extraVarsMode}

Sposób edycji dodatkowych zmiennych:

- **Table**: jeden wiersz na zmienną, każdy z typem. Dodaj wiersz przyciskiem **+** obok
  przełącznika.
- **JSON**: wszystkie zmienne jako jeden obiekt JSON.
- **YAML**: wszystkie zmienne jako jedno mapowanie YAML.

Okno dialogowe otwiera się w trybie **Table**. Przełączenie konwertuje to, co wpisałeś; gdy nie da
się tego odczytać, okno dialogowe pozostaje w poprzednim trybie i wyświetla błąd. Niezależnie od
trybu zmienne są zapisywane jako jeden obiekt nazw i wartości.

## Dodatkowe zmienne w JSON {#varGroup-extraVarsJson}

Wyświetlane, gdy dodatkowe zmienne są edytowane jako **JSON**.

Dodatkowe zmienne jako jeden obiekt JSON. Nazwa nie może być pusta; wartość może być ciągiem
znaków, liczbą, listą lub obiektem:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Dodatkowe zmienne w YAML {#varGroup-extraVarsYaml}

Wyświetlane, gdy dodatkowe zmienne są edytowane jako **YAML**.

Dodatkowe zmienne jako jedno mapowanie YAML. Tekst jest odczytywany z typami, które zna JSON: data
bez cudzysłowów, taka jak `2024-01-01`, pozostaje ciągiem znaków, a `.inf` lub `.nan` są odrzucane.
Lista lub wartość skalarna na najwyższym poziomie również są odrzucane.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Nazwa zmiennej {#varGroup-varName}

Wyświetlane, gdy dodatkowe zmienne są edytowane jako **Table**, w każdym wierszu.

Nazwa zmiennej taka, jaką widzi playbook lub kod Terraform, na przykład `app_version`.

## Typ zmiennej {#varGroup-varType}

Wyświetlane, gdy dodatkowe zmienne są edytowane jako **Table**, w każdym wierszu.

Sposób odczytu wartości w wierszu:

- **String**: tekst w niezmienionej postaci.
- **Number**: liczba, na przykład `42`.
- **List**: tablica JSON, na przykład `["a", "b"]`.
- **Dict**: obiekt JSON, na przykład `{"key": "value"}`.

Wartość, która nie pasuje do swojego typu, zatrzymuje zapis z podaniem nazwy zmiennej.

## Wartość zmiennej {#varGroup-varValue}

Wyświetlane, gdy dodatkowe zmienne są edytowane jako **Table**, w każdym wierszu.

Wartość zmiennej w formacie jej typu. Dla **List** lub **Dict** przycisk na końcu pola otwiera
większy edytor.

Semaphore przekazuje wartości do Ansible jako dane: wyrażenie `{{ … }}` wewnątrz wartości nie jest
obliczane, playbook otrzymuje tekst w niezmienionej postaci.

## Nazwa zmiennej środowiskowej {#varGroup-envName}

Na karcie **Variables**, w sekcji **Zmienne środowiskowe**, w każdym wierszu. Dodaj wiersz
przyciskiem **+** tej sekcji.

Nazwa zmiennej ustawianej w środowisku procesu, który wykonuje zadanie, na przykład `AWS_REGION`.

## Wartość zmiennej środowiskowej {#varGroup-envValue}

Na karcie **Variables**, w sekcji **Zmienne środowiskowe**, w każdym wierszu.

Wartość zmiennej, jako tekst. Jest przechowywana jawnym tekstem i wyświetlana każdemu, kto otworzy
grupę: hasła i tokeny umieszczaj na karcie **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Wyświetlane na karcie **Secrets** grupy, która ma magazyn sekretów.

Włącz, aby importować sekrety z magazynu do tej grupy ze ścieżek ustawionych w **Sync paths**.
Import jest wykonywany po naciśnięciu przycisku synchronizacji grupy na stronie Grupy zmiennych
oraz samoczynnie w odstępie ustawionym w **Sync paths**. Gdy opcja jest wyłączona, nic nie jest
importowane.

## Sync paths {#varGroup-syncPaths}

Wyświetlane, gdy **Sync keys enabled** jest włączone.

Otwiera okno dialogowe ze ścieżkami magazynu, z których odbywa się import, oraz odstępem
automatycznego importu. Liczba na przycisku to liczba ścieżek. Przycisk synchronizacji grupy na
stronie Grupy zmiennych pozostaje wyłączony, dopóki grupa nie ma co najmniej jednej ścieżki.

Okno dialogowe ma tylko przycisk **Zamknij**: to, co w nim zmienisz, jest zapisywane z grupą
przyciskiem **Zapisz** tego okna dialogowego.

## Nazwa tajnej zmiennej {#varGroup-secretVarName}

Na karcie **Secrets**, w sekcji **Dodatkowe zmienne**, w każdym wierszu. Dodaj wiersz przyciskiem
**+** tej sekcji.

Nazwa dodatkowej zmiennej, której wartość jest sekretem. Zadanie otrzymuje ją tak samo jak
dodatkowe zmienne z karty **Variables**; sekret ma pierwszeństwo przed dodatkową zmienną o tej
samej nazwie. Ansible otrzymuje sekrety w pliku zaszyfrowanym za pomocą Ansible Vault, Terraform i
OpenTofu jako `-var`. Jak ostrzega okno dialogowe, sekret przekazany w ten sposób może mimo to
pojawić się jawnym tekstem w dzienniku Ansible.

## Wartość tajnej zmiennej {#varGroup-secretVarValue}

Na karcie **Secrets**, w sekcji **Dodatkowe zmienne**, w każdym wierszu.

Wartość sekretu. Semaphore szyfruje ją lub zapisuje w magazynie sekretów grupy i nigdy ponownie jej
nie wyświetla: gdy później otworzysz grupę, pole jest puste. Pozostaw je puste, aby zachować
zapisaną wartość, lub wpisz nową, aby ją zastąpić. Usunięcie wiersza usuwa sekret przy zapisie
grupy.

## Nazwa tajnej zmiennej środowiskowej {#varGroup-secretEnvName}

Na karcie **Secrets**, w sekcji **Zmienne środowiskowe**, w każdym wierszu. Dodaj wiersz
przyciskiem **+** tej sekcji.

Nazwa zmiennej środowiskowej, której wartość jest sekretem, na przykład `DB_PASSWORD`. Jest
ustawiana w środowisku procesu, który wykonuje zadanie, tak jak zmienne środowiskowe z karty
**Variables**.

## Wartość tajnej zmiennej środowiskowej {#varGroup-secretEnvValue}

Na karcie **Secrets**, w sekcji **Zmienne środowiskowe**, w każdym wierszu.

Wartość sekretu. Podobnie jak w przypadku tajnej dodatkowej zmiennej, nigdy nie jest ponownie
wyświetlana: pozostaw pole puste, aby zachować zapisaną wartość, lub wpisz nową, aby ją zastąpić.
