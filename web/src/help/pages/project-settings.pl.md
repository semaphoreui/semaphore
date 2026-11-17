---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Ustawienia projektu

Ustawienia projektu jako całości: jego nazwa, liczba jego zadań, które mogą być wykonywane
jednocześnie, oraz to, czy wysyła alerty. Strefa zagrożenia pod formularzem zawiera akcje
dotyczące całego projektu: jego kopię zapasową, pamięć podręczną i usunięcie. Zmiany w formularzu
zaczynają obowiązywać po kliknięciu **Zapisz**.

## Nazwa projektu {#newProject-name}

Nazwa, pod którą projekt widnieje w menu projektów u góry paska bocznego. Pole jest wymagane.

## Maksymalna liczba równoległych zadań {#newProject-maxParallelTasks}

Ile zadań tego projektu może być wykonywanych jednocześnie. Zadanie uruchomione ponad limit czeka
w kolejce, aż jedno z wykonywanych zadań się zakończy.

Pozostaw pole puste lub wpisz w nim `0`, aby nie stosować własnego limitu. Limit całego serwera,
ustawiony przez jego administratora, obowiązuje w obu przypadkach.

## ID czatu Telegram {#newProject-tg}

Opcjonalne. Czat Telegram, który otrzymuje alerty tego projektu. Gdy pole jest puste, alerty
trafiają do czatu ustawionego dla całego serwera.

Pole działa tylko wtedy, gdy administrator włączył alerty Telegram na serwerze, a alerty są
dozwolone dla tego projektu.

## Zezwól na alerty dla tego projektu {#newProject-alert}

Włącza alerty dla zadań tego projektu. Miejsce, do którego trafiają alerty, ustawia dla całego
serwera jego administrator: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams, DingTalk lub
Gotify.

- E-mail jest wysyłany, gdy zadanie zakończy się niepowodzeniem, do tych członków projektu,
  którzy mogą zobaczyć zadanie i mają włączone alerty na swoim koncie.
- Wiadomość do komunikatorów jest wysyłana, gdy zadanie zakończy się powodzeniem,
  niepowodzeniem lub czeka na potwierdzenie.

Szablon zadania może zablokować wysyłanie alertów przez swoje zadania zakończone powodzeniem lub
niepowodzeniem.

## Test Alerts {#settings-testAlerts}

Wysyła alert testowy każdym kanałem włączonym na serwerze, aby można było sprawdzić, czy alerty
tego projektu docierają.

Przycisk działa tylko wtedy, gdy opcja **Zezwól na alerty dla tego projektu** jest zaznaczona i
zapisana. Najpierw ją zaznacz i kliknij **Zapisz**.

## Kopia zapasowa projektu {#settings-exportProject}

Pobiera plik JSON z konfiguracją projektu: szablonami zadań, inwentarzami, repozytoriami, grupami
zmiennych, kluczami, harmonogramami, integracjami i pozostałymi elementami w nim skonfigurowanymi.
Plik można z powrotem zamienić w projekt za pomocą **Przywróć projekt** w menu projektów u góry
paska bocznego.

Plik pomija sekrety kluczy oraz sekrety z karty **Secrets** grup zmiennych: trzeba je wprowadzić
ponownie po przywróceniu. Wszystko inne jest kopiowane bez zmian, także hasło wpisane bezpośrednio w
inne pole: adres repozytorium zawierający hasło, tekst statycznego inwentarza, wartości zwykłych
zmiennych i zmiennych środowiskowych. Przed udostępnieniem pliku sprawdź, czy nie zawiera takich
wartości. Historia zadań nie jest częścią kopii zapasowej.

## Wyczyść pamięć podręczną {#settings-clearCache}

Usuwa pliki, które Semaphore przechowuje na serwerze dla tego projektu, na przykład swoje kopie
repozytoriów. Są one pobierane ponownie przy uruchomieniu następnego zadania.

Użyj tej akcji, gdy zadanie kończy się niepowodzeniem z powodu nieaktualnej lub uszkodzonej kopii
repozytorium. Akcji nie można cofnąć i nie zmienia ona ustawień projektu.

## Usuń projekt {#settings-deleteProject}

Usuwa projekt wraz ze wszystkim, co zawiera: szablonami zadań, historią zadań, inwentarzami,
repozytoriami, grupami zmiennych, kluczami i harmonogramami. Semaphore najpierw prosi o
potwierdzenie.

Usuniętego projektu nie można przywrócić. Pobierz kopię zapasową, zanim usuniesz projekt, który
może być jeszcze potrzebny.
