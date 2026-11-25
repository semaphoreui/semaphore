---
form: src/components/SurveyVars.vue
---

# Zmienna ankiety

Zmienna ankiety to pytanie, które okno dialogowe zadania zadaje przy każdym uruchomieniu szablonu.
Odpowiedź trafia do zadania jako zmienna: dodatkowa zmienna dla Ansible, `-var` dla Terraform i
OpenTofu, argument dla skryptu lub zmienna środowiskowa. Zmienna jest zapisywana z szablonem, gdy
zapisujesz okno dialogowe szablonu. To samo okno dialogowe edytuje ankietę przepływu pracy, w
której typ **Secret** nie jest oferowany.

## Nazwa {#surveyVarDialog-name}

Nazwa, pod którą zadanie otrzymuje odpowiedź: nazwa dodatkowej zmiennej, zmiennej Terraform,
argumentu `name=value` skryptu lub zmiennej środowiskowej. Spacje na początku i na końcu są
usuwane. Pole jest wymagane.

Użyj nazwy, której nie ma żadna inna zmienna ankiety i którą akceptuje narzędzie, na przykład
`app_version`. Odpowiedź zastępuje zmienną o tej samej nazwie z grupy zmiennych szablonu.

## Tytuł {#surveyVarDialog-title}

Etykieta pola w oknie dialogowym zadania, na przykład `Version to deploy`. Jest to także tekst
chipa zmiennej w oknie dialogowym szablonu. Pole jest wymagane.

## Opis {#surveyVarDialog-description}

Opcjonalne. Podpowiedź wyświetlana pod polem okna dialogowego zadania, gdy pole ma fokus: co
wpisać lub przykład.

## Typ {#surveyVarDialog-type}

Jakiego rodzaju odpowiedzi wymaga okno dialogowe zadania:

- **String**: jeden wiersz tekstu. To ustawienie domyślne.
- **Integer**: tylko cyfry; okno dialogowe zadania odrzuca wszystko inne, także znak minus.
  Odpowiedź nadal trafia do zadania jako tekst: przekształć ją w playbooku, na przykład za pomocą
  `| int`.
- **Secret**: tekst maskowany przez okno dialogowe zadania. Odpowiedź nie jest zapisywana z
  zadaniem, więc ponowne uruchomienie znów o nią pyta, a Ansible otrzymuje ją w pliku
  zaszyfrowanym za pomocą Ansible Vault. Sekret nie ma wartości domyślnej.
- **Enum**: jedna wartość wybrana z listy zdefiniowanej poniżej.
- **Text**: kilka wierszy tekstu.
- **Select**: kilka wartości wybranych z listy zdefiniowanej poniżej. Zadanie otrzymuje je jako
  listę; w argumencie, w `-var` lub w zmiennej środowiskowej lista jest zapisywana jako JSON, na
  przykład `["a","b"]`.

Zmiana typu czyści wartość domyślną, a typ bez listy czyści wartości.

## Pass variable as {#surveyVarDialog-target}

Jak odpowiedź trafia do zadania:

- **Extra variable**: zwykły sposób danego narzędzia. Ansible otrzymuje ją jako dodatkową
  zmienną, Terraform i OpenTofu jako `-var name=value`, skrypt Bash, PowerShell lub Python jako
  argument `name=value` po skrypcie.
- **Environment variable**: zmienna środowiskowa procesu, `name=value`, dla każdego rodzaju
  szablonu. Zmienna nie jest wtedy przekazywana w drugi sposób.

## Klucz {#surveyVarDialog-valueName}

Wyświetlane, gdy typ to **Enum** lub **Select**.

Tekst, który okno dialogowe zadania wyświetla dla tej wartości na liście. Każda wartość wymaga
klucza, a klucze muszą się różnić: w przeciwnym razie okno dialogowe odmawia zapisania zmiennej.

## Wartość {#surveyVarDialog-valueValue}

Wyświetlane, gdy typ to **Enum** lub **Select**.

Co otrzymuje zadanie, gdy wybrano tę wartość. Może się różnić od swojego klucza: klucz
`Production` może przekazywać `prod`.

## Add Value {#surveyVarDialog-addValue}

Wyświetlane, gdy typ to **Enum** lub **Select**.

Dodaje pusty wiersz do listy wartości. **Enum** lub **Select** wymaga co najmniej jednej wartości.
Ikona kosza w wierszu go usuwa.

## Default value {#surveyVarDialog-defaultValue}

Wyświetlane dla każdego typu oprócz **Secret**.

Odpowiedź, którą okno dialogowe zadania wypełnia przy otwarciu; możesz ją tam zmienić. Dla
**Enum** wybierz jedną z wartości, dla **Select** dowolną ich liczbę. Pozostaw pole puste, aby
zacząć od pustej odpowiedzi.

Wartość domyślną wypełnia tylko okno dialogowe zadania. Zadanie uruchomione w inny sposób, na
przykład przez API, nie otrzymuje wartości domyślnej.

## Wymagane {#surveyVarDialog-required}

Okno dialogowe zadania nie uruchamia zadania, dopóki zmienna nie ma odpowiedzi. Podobnie jak
wartość domyślna, sprawdzenie wykonuje tylko okno dialogowe zadania, a nie serwer.
