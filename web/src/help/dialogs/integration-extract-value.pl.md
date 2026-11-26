---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Wyodrębniona wartość pobiera fragment przychodzącego żądania, na przykład gałąź ze zdarzenia push,
i przekazuje go do zadania uruchamianego przez integrację. Każda wartość integracji jest pobierana
z każdego żądania, które uruchamia zadanie.

## Name {#integrationValueForm-name}

Nazwa, pod którą wartość widnieje na stronie integracji. Nie zmienia tego, co jest pobierane ani
dokąd trafia. Pole jest wymagane.

## Source of the Value {#integrationValueForm-valueSource}

Część żądania, z której pobierana jest wartość:

- **Header**: wartość nagłówka żądania.
- **Body**: treść żądania.

## Data Type of Body {#integrationValueForm-bodyDataType}

Wyświetlane, gdy **Source of the Value** to **Body**.

Sposób odczytu treści:

- **JSON**: treść jest parsowana jako JSON, a **Key** to ścieżka do jednej wartości w niej.
  Wartość zachowuje swój typ JSON: liczba, lista lub obiekt docierają do zadania jako takie.
- **String**: cała treść jest pobierana jako tekst, a **Key** nie jest używane.

## Key {#integrationValueForm-key}

Co pobrać z żądania. Pole jest wymagane.

- Dla **Header**: nazwa nagłówka, na przykład `X-GitHub-Event`. Wielkość liter nie ma znaczenia.
  Nagłówek, którego brakuje w żądaniu, daje pustą wartość.
- Dla **Body** w formacie **JSON**: ścieżka do wartości, której części są połączone kropkami;
  element tablicy zapisuje się jako `.[0]`, na przykład `repository.full_name` lub
  `commits.[0].id`. Ścieżka, której brakuje w treści, nie daje żadnej wartości, a zadanie
  zachowuje wartość ustawioną w integracji.
- Dla **Body** w formacie **String**: nieużywane, ale formularz wymaga wartości; wpisz cokolwiek.

## Variable Usage {#integrationValueForm-variableType}

Dokąd trafia wartość w zadaniu szablonu:

- **Variables**: dodatkowa zmienna zadania, przekazywana tak jak odpowiedź na zmienną ankiety.
  Zastępuje odpowiedź lub zmienną grupy zmiennych o tej samej nazwie.
- **Task Params**: parametr zadania, przechowywany tak jak opcje okna dialogowego zadania. Nazwa
  `git_branch` ustawia gałąź zadania, gdy szablon pozwala nadpisać gałąź; początkowe
  `refs/heads/` jest usuwane.

Integracja, która uruchamia przepływ pracy, ignoruje to ustawienie: każda wartość staje się
odpowiedzią na zmienną ankiety przepływu pracy o nazwie podanej w **Variable**.

## Variable {#integrationValueForm-variable}

Nazwa, pod którą wartość jest przekazywana, na przykład `branch` dla dodatkowej zmiennej, którą
playbook odczytuje jako `{{ branch }}`. Pole jest wymagane.
