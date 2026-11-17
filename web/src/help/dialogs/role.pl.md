---
form: src/components/EditRoleForm.vue
---

# Edit Role

Rola to nazwany zestaw uprawnień wewnątrz projektu. Nadajesz role członkom projektu, a członek może
robić to, na co pozwala którakolwiek z jego ról. Otwarte ze strony Roles projektu okno dialogowe
edytuje rolę tylko tego projektu; otwarte ze strony Roles poza projektem edytuje rolę
współdzieloną, której może używać każdy projekt. Wbudowanych ról Owner, Manager, Task Runner i
Guest nie można zmienić.

## Nazwa {#role-name}

Nazwa, pod którą rola widnieje na stronie Roles i wszędzie tam, gdzie wybiera się rolę, na przykład
przy dodawaniu członka do projektu. Pole jest wymagane. Nazwę roli można zmienić w dowolnym
momencie.

## Slug {#role-slug}

Identyfikator roli. Semaphore wypełnia go na podstawie nazwy podczas jej wpisywania, dopóki sam
nie zmienisz sluga. Składa się z 1 do 100 liter, cyfr, podkreśleń i myślników, na przykład
`release-managers`.

Sluga nie można zmienić po zapisaniu roli: podczas edycji roli pole jest zablokowane.

- Slug musi być unikalny, a sprawdzenie nie uwzględnia wielkości liter. Slug roli współdzielonej
  musi się różnić od każdej innej roli; slug roli projektu musi się różnić od ról współdzielonych
  i od pozostałych ról swojego projektu.
- `owner`, `manager`, `task_runner`, `guest` i `all` są zarezerwowane, niezależnie od wielkości
  liter.

## Permissions {#role-permissionsTitle}

Co członkowie mający tę rolę mogą robić w projekcie. Nowa rola nie ma zaznaczonych żadnych
uprawnień. Uprawnienia wszystkich ról członka sumują się; administrator Semaphore może robić
wszystko w każdym projekcie niezależnie od ról.

- **Can read project**: wgląd w projekt: jego szablony zadań, zadania i ich dzienniki, inwentarze,
  klucze, repozytoria, grupy zmiennych, harmonogramy i członków. Każdy odczyt w projekcie wymaga
  tego uprawnienia i żadne inne go nie obejmuje, więc zaznacz je niemal dla każdej roli. Bez niego
  członek widzi tylko te szablony zadań i przepływy pracy, które same przyznają roli dostęp, wraz z
  ich zadaniami.
- **Can run project tasks**: uruchamianie zadań z szablonów zadań, ich zatrzymywanie oraz
  potwierdzanie lub odrzucanie zadania, które czeka na potwierdzenie.
- **Can update project**: zmiana ustawień projektu i jego usunięcie.
- **Can manage project resources**: tworzenie, zmiana i usuwanie szablonów zadań, inwentarzy,
  kluczy, repozytoriów, grup zmiennych, harmonogramów, widoków, integracji, magazynów sekretów,
  uruchamiaczy projektu i ról projektu; usuwanie zadań; określanie, które role mogą używać szablonu
  zadania.
- **Can manage project users**: dodawanie członków do projektu, zmiana ich ról i ich usuwanie.
- **Can run workflows**: uruchamianie przepływu pracy i zatrzymywanie jego przebiegów.
- **Can edit workflows**: tworzenie, zmiana i usuwanie przepływów pracy oraz określanie, które
  role mogą używać przepływu pracy.
- **Can approve workflows**: zatwierdzanie lub odrzucanie kroku przebiegu przepływu pracy, który
  czeka na zatwierdzenie.
