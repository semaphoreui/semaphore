---
form: src/components/EditTemplatePermissionForm.vue
---

# Template permissions

Nadaje roli dodatkowe uprawnienia do tego jednego szablonu. Członek, który ma tę rolę, otrzymuje je
oprócz uprawnień, które rola ma w całym projekcie: nadanie uprawnień tutaj tylko je dodaje, nigdy
nie odbiera żadnego uprawnienia. Członek z kilkoma rolami otrzymuje każde uprawnienie nadane
którejkolwiek z nich.

## Rola {#templatePerm-roleSlug}

Rola, która otrzymuje uprawnienia. Lista pokazuje role tego projektu, role współdzielone i role
wbudowane. Pole jest wymagane, a rola może mieć tylko jedno nadanie uprawnień dla szablonu: serwer
odrzuca drugie.

Roli zapisanego nadania nie można zmienić. Serwer zachowuje rolę i zapisuje tylko uprawnienia,
niezależnie od tego, co tu wybierzesz. Aby przekazać uprawnienia innej roli, usuń to nadanie z
listy i dodaj nowe.

## Permissions {#templatePerm-permissionsTitle}

Co członek, który ma tę rolę, może robić z tym szablonem:

- **Can see this template**: szablon jest widoczny na liście szablonów wraz ze swoimi zadaniami i
  ich dziennikami.
- **Can run this template**: uruchamianie zadań z szablonu oraz zatrzymywanie, potwierdzanie lub
  odrzucanie jego zadań.
- **Can edit this template**: zmiana i usuwanie szablonu, zatrzymywanie wszystkich jego zadań i
  zarządzanie tymi uprawnieniami.

Uruchamianie lub edycja obejmują widzenie szablonu: gdy zaznaczysz którekolwiek z nich, **Can see
this template** również zostaje zaznaczone i nie można go odznaczyć. Pozostaw wszystkie pola
odznaczone, aby zapisać nadanie, które nic nie daje.
