---
form: src/components/InventoryGroupForm.vue
---

# Edytuj grupę

Grupa hostów zarządzanego inwentarza. Playbook jest uruchamiany na części hostów po nazwie grupy, a
zmienną grupy ustawia się raz dla wszystkich jej hostów. Hosty dodaje się do grupy na karcie Hosts,
w oknie hosta lub działaniami na zaznaczonych hostach.

Grupa zaimportowana ze źródła inwentarza jest zarządzana przez to źródło. Wszystkie jej pola są tylko
do odczytu i są aktualizowane przy każdej synchronizacji.

## Nazwa {#inventoryGroupForm-name}

Nazwa, pod którą Ansible zna grupę, na przykład w `hosts: web` playbooka. Pole jest wymagane. Nazwa
jest unikalna w obrębie inwentarza, a wielkość liter ma znaczenie: `web` i `Web` to dwie grupy.
Nazwa nie może zawierać spacji, przecinków, nawiasów kwadratowych ani dwukropków, ma najwyżej 250
znaków i nie może być `all` ani `ungrouped`. Ansible ostrzega przed nazwami ze znakami innymi niż
litery, cyfry i podkreślenia, na przykład `web-eu`, ale taka nazwa nadal działa. Niedozwolone są też
znaki `*`, `?`, `!`, `&`, `~` i `@`, ponieważ Ansible czyta je jako wzorzec.

## Opis {#inventoryGroupForm-description}

Opcjonalne. Notatka dla ludzi, widoczna na liście grup. Ansible jej nie otrzymuje.

## Child groups {#inventoryGroupForm-children}

Opcjonalne. Grupy wewnątrz tej grupy. Hosty grupy podrzędnej należą też do tej grupy. Grupa nie może
być własnym przodkiem.

## Zmienne {#inventoryGroupForm-variables}

Opcjonalne. Zmienne tej grupy zapisane jako mapowanie YAML lub JSON:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Otrzymuje je każdy host grupy. Zmienna hosta ma pierwszeństwo przed zmienną grupy o tej samej
nazwie, a zmienna grupy ma pierwszeństwo przed zmienną inwentarza. Kotwice, aliasy i klucze scalania
YAML nie są akceptowane.
