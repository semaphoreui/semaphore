---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Dopasowacz to warunek nałożony na przychodzące żądanie. Dopasowacze decydują, którą integrację
uruchamia żądanie wysłane na alias projektu, dzięki czemu jeden adres URL może obsługiwać kilka
integracji: integracja jest uruchamiana tylko wtedy, gdy spełnione są wszystkie jej dopasowacze.
Są używane tylko wtedy, gdy dla integracji włączona jest opcja **Użyj aliasu projektu**; gdy jest
włączona, integracja jest osiągalna wyłącznie przez alias projektu.

Integracja znaleziona przez alias projektu, która nie ma dopasowaczy, nigdy się nie uruchamia:
dodaj co najmniej jeden.

## Name {#integrationMatcherForm-name}

Nazwa, pod którą dopasowacz widnieje na stronie integracji. Nie zmienia tego, co sprawdza
dopasowacz. Pole jest wymagane.

## Match on {#integrationMatcherForm-matchType}

Część żądania, którą odczytuje dopasowacz:

- **Header**: wartość nagłówka żądania.
- **Body**: treść żądania.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Wyświetlane, gdy **Match on** to **Body**.

Sposób odczytu treści:

- **JSON**: treść jest parsowana jako JSON, a **Key** to ścieżka do jednej wartości w niej.
- **String**: cała treść jest porównywana jako tekst, a **Key** nie jest używane.

## Key {#integrationMatcherForm-key}

Co odczytać z żądania. Pole jest wymagane.

- Dla **Header**: nazwa nagłówka, na przykład `X-GitHub-Event`. Wielkość liter nie ma znaczenia.
- Dla **Body** w formacie **JSON**: ścieżka do wartości, której części są połączone kropkami;
  element tablicy zapisuje się jako `.[0]`, na przykład `repository.full_name` lub
  `commits.[0].id`.
- Dla **Body** w formacie **String**: nieużywane, ale formularz wymaga wartości; wpisz cokolwiek.

## Comparison Method {#integrationMatcherForm-method}

Jak wartość odczytana z żądania jest porównywana z **Value**:

- **==**: obie są równe.
- **!=**: obie się różnią. Nagłówek lub klucz, którego brakuje w żądaniu, liczy się jako różny.
- **Contains**: wartość odczytana z żądania zawiera **Value**.

## Value {#integrationMatcherForm-value}

Tekst, z którym następuje porównanie, na przykład `push` lub `refs/heads/main`. Porównanie jest
dokładne i uwzględnia wielkość liter. Liczba lub wartość logiczna z treści JSON jest porównywana w
postaci tekstowej: `42`, `true`. Pole jest wymagane.
