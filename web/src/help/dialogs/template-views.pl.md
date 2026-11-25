---
form: src/components/EditViewsForm.vue
---

# Edytuj widoki

Widoki to karty nad listą szablonów zadań. Widok pokazuje szablony, które do niego dodasz, albo
wszystkie szablony projektu. Okno dialogowe zapisuje każdą zmianę od razu: nie ma przycisku zapisu
dla całego okna. Przeciągnij widok za uchwyt, aby zmienić kolejność kart.

## Tytuł {#viewsForm-title}

Nazwa karty. Wpisz ją i potwierdź znacznikiem wyboru obok pola; krzyżyk przywraca zapisany tytuł.
Nowy widok jest zapisywany, gdy ma tytuł i go potwierdzisz lub zmienisz jedno z jego ustawień.
Zamknięcie okna dialogowego nie zapisuje niepotwierdzonego tytułu.

Koło zębate otwiera ustawienia widoku pod jego tytułem.

## Usuń {#viewsForm-viewDelete}

Ikona kosza od razu usuwa widok, bez pytania. Szablony widoku nie są usuwane: pozostają w projekcie
i nie należą już do żadnego widoku.

## Type {#viewsForm-type}

Które szablony pokazuje karta:

- **Custom**: szablony, których pole **Widok** w oknie dialogowym szablonu wskazuje ten widok. To
  ustawienie domyślne.
- **All**: wszystkie szablony projektu, niezależnie od ich widoku.

Członek widzi tylko szablony, które może widzieć, w każdym widoku.

## Sort by {#viewsForm-sortColumn}

Kolumna, według której sortowane są szablony karty; **Name** to jedyna dostępna. Karta bez tego
ustawienia również jest sortowana według nazwy, więc pole ma znaczenie razem z **Reverse**.

## Reverse {#viewsForm-sortReverse}

Sortuje kartę w kolejności malejącej. Działa tylko wtedy, gdy ustawiono **Sort by**.

## Hidden {#viewsForm-hidden}

Ukrywa kartę przed członkami, którzy nie mogą zarządzać zasobami projektu. Ci, którzy mogą, nadal
ją widzą, przekreśloną. Ukrycie karty nie ukrywa jej szablonów: widok typu **All** nadal je
wyświetla.

## Dodaj widok {#viewsForm-add}

Dodaje widok na końcu listy z pustym tytułem do wpisania.
