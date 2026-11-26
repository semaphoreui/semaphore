---
form: src/components/EditTemplatePermissionForm.vue
---

# Workflow permissions

Nadaje roli dodatkowe uprawnienia do tego jednego przepływu pracy. Członek, który ma tę rolę,
otrzymuje je oprócz uprawnień, które rola ma w całym projekcie: nadanie uprawnień tutaj tylko je
dodaje, nigdy nie odbiera żadnego uprawnienia. Członek z kilkoma rolami otrzymuje każde uprawnienie
nadane którejkolwiek z nich.

## Rola {#templatePerm-roleSlug}

Rola, która otrzymuje uprawnienia. Lista pokazuje role tego projektu, role współdzielone i role
wbudowane. Pole jest wymagane, a rola może mieć tylko jedno nadanie uprawnień dla przepływu pracy:
serwer odrzuca drugie.

Roli zapisanego nadania nie można zmienić. Serwer zachowuje rolę i zapisuje tylko uprawnienia,
niezależnie od tego, co tu wybierzesz. Aby przekazać uprawnienia innej roli, usuń to nadanie z
listy i dodaj nowe.

## Permissions {#templatePerm-permissionsTitle}

Co członek, który ma tę rolę, może robić z tym przepływem pracy:

- **Can see this workflow**: przepływ pracy jest widoczny na liście przepływów pracy wraz ze
  swoimi przebiegami.
- **Can run this workflow**: uruchamianie przebiegów przepływu pracy.
- **Can edit this workflow**: zmiana i usuwanie przepływu pracy oraz zarządzanie tymi
  uprawnieniami.
- **Can approve this workflow**: zatwierdzanie lub odrzucanie kroków zatwierdzenia jego
  przebiegów.

Uruchamianie, edycja lub zatwierdzanie obejmują widzenie przepływu pracy: gdy zaznaczysz
którekolwiek z nich, **Can see this workflow** również zostaje zaznaczone i nie można go odznaczyć.
Pozostaw wszystkie pola odznaczone, aby zapisać nadanie, które nic nie daje.
