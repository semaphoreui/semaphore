---
form: src/components/EditGlobalRoleForm.vue
---

# Edit Global Role

Rola globalna nadaje uprawnienia do całej instancji Semaphore użytkownikom, którym jest
przypisana; przypisujesz ją w oknie dialogowym użytkownika. Nie daje żadnego dostępu wewnątrz
projektów: ten wynika wyłącznie z ról, które użytkownik ma jako członek projektu. Administrator ma
wszystkie uprawnienia globalne niezależnie od przypisanych ról. Role globalne są funkcją wersji
Pro.

## Nazwa {#globalRole-name}

Nazwa, pod którą rola widnieje na karcie Role globalne strony Użytkownicy i w oknie dialogowym
użytkownika. Pole jest wymagane. Nazwę roli można zmienić w dowolnym momencie.

## Slug {#globalRole-slug}

Identyfikator roli. Semaphore wypełnia go na podstawie nazwy podczas jej wpisywania, dopóki sam
nie zmienisz sluga. Składa się z 1 do 100 liter, cyfr, podkreśleń i myślników, na przykład
`user-admins`.

Slug musi być unikalny wśród ról globalnych; rola projektu może mieć ten sam slug. Nie można go
zmienić po zapisaniu roli: podczas edycji roli pole jest zablokowane.

## Global permissions {#globalRole-permissionsTitle}

Co użytkownik mający tę rolę może robić w całej instancji. Nowa rola nie ma zaznaczonych żadnych
uprawnień. Uprawnienia wszystkich ról globalnych użytkownika sumują się.

- **Can manage users**: tworzenie użytkowników, ich zmiana i usuwanie oraz wgląd w ich dane na
  liście użytkowników. Taki użytkownik nie może uczynić użytkownika administratorem ani
  użytkownikiem Pro i nie może zmienić ani usunąć administratora ani użytkownika, który ma
  uprawnienie globalne, którego on sam nie ma.
- **Can manage global roles**: tworzenie, zmiana i usuwanie ról globalnych i współdzielonych ról
  projektów oraz decydowanie, którzy użytkownicy mają którą rolę globalną. Nadawaj je tylko
  osobom, którym powierzasz całą instancję.
- **Can manage system settings**: zmiana ustawień instancji, aplikacji, globalnych uruchamiaczy i
  mapowań grup; wgląd w klaster i informacje o serwerze; wgląd w zadania wszystkich projektów i ich
  usuwanie; czyszczenie pamięci podręcznej serwera.
- **Can read the audit log**: odczyt dziennika audytu instancji i jego eksport.
