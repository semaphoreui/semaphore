---
form: src/components/InventoryHostForm.vue
---

# Edytuj host

Host zarządzanego inwentarza. Zadanie, które używa inwentarza, jest uruchamiane na tych hostach,
które są włączone na liście hostów.

Host zaimportowany ze źródła inwentarza jest zarządzany przez to źródło. Jego nazwa i zmienne są tylko
do odczytu i są aktualizowane przy każdej synchronizacji. Opis, grupy niezaimportowane i przełącznik
Enabled nadal można edytować.

## Nazwa {#inventoryHostForm-name}

Nazwa, pod którą Ansible zna host, oraz adres, z którym się łączy, o ile zmienne nie ustawiają
`ansible_host`. Pole jest wymagane. Nazwa jest unikalna w obrębie inwentarza, a wielkość liter ma
znaczenie: `web1` i `Web1` to dwa hosty. Nazwa nie może zawierać spacji, przecinków, nawiasów
kwadratowych ani dwukropków i ma najwyżej 250 znaków.

## Opis {#inventoryHostForm-description}

Opcjonalne. Notatka dla ludzi, widoczna na liście hostów. Ansible jej nie otrzymuje.

## Groups {#inventoryHostForm-groups}

Opcjonalne. Grupy, do których należy ten host. Host może należeć do kilku grup. Grupy tworzy się na
karcie Groups.

## Zmienne {#inventoryHostForm-variables}

Opcjonalne. Zmienne tego hosta zapisane jako mapowanie YAML lub JSON:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Zmienna hosta ma pierwszeństwo przed zmienną inwentarza o tej samej nazwie. Kotwice, aliasy i klucze
scalania YAML nie są akceptowane.
