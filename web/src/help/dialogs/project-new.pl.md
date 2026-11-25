---
form: src/components/ProjectForm.vue
---

# Nowy projekt

Projekt zawiera szablony zadań, inwentarze, repozytoria, grupy zmiennych i klucze jednego
przedsięwzięcia oraz osoby, które nad nim pracują. Stajesz się właścicielem utworzonego przez
siebie projektu. Tylko administrator może utworzyć projekt, chyba że administrator serwera zezwala
na to wszystkim użytkownikom. Każde z poniższych ustawień można później zmienić na stronie
Ustawienia projektu.

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

## Demo {#newProject-demo}

Wypełnia nowy projekt przykładami do wypróbowania Semaphore: publicznym repozytorium
`https://github.com/semaphoreui/semaphore-demo.git`, inwentarzami, kluczem z hasłem skarbca oraz
szablonami zadań dla Ansible, OpenTofu, Terragrunt, Bash i PowerShell, podzielonymi na widoki
**Build**, **Deploy** i **Tools**.

Pozostaw przełącznik wyłączony, aby utworzyć projekt bez przykładów: otrzyma on tylko klucz
**None** i widok **All**. Przykłady są zwykłymi elementami projektu: edytuj
je lub usuwaj według potrzeb.
