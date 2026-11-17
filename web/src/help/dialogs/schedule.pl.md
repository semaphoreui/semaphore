---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Edytuj harmonogram

Harmonogram uruchamia zadanie szablonu zadania bez naciskania **Uruchom** przez kogokolwiek:
wielokrotnie według wyrażenia cron lub jednorazowo w ustalonym momencie. Zadanie otrzymuje
parametry zapisane w tym oknie dialogowym. Wszystkie czasy są podawane w strefie czasowej serwera
Semaphore, wyświetlanej obok pól czasu: `UTC`, chyba że administrator ustawi `schedule.timezone`
lub `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

Nazwa, pod którą harmonogram widnieje na stronie Harmonogram. Pole jest wymagane. Nie zmienia
tego, co robi harmonogram.

## Template {#scheduleForm-templateId}

Szablon zadania, który uruchamia harmonogram. Pole jest wymagane. Zadanie jest wykonywane tak,
jakby zostało uruchomione z szablonu, z inwentarzem, repozytorium, grupą zmiennych i pozostałymi
ustawieniami szablonu oraz z poniższymi parametrami tam, gdzie szablon pozwala zadaniu je zmienić.

## Parametry zadania {#scheduleForm-taskParams}

Wyświetlane, gdy wybrano szablon.

Parametry każdego zadania uruchamianego przez harmonogram. To, które pola się pojawiają, zależy od
szablonu: od jego zmiennych ankiety i od opcji, które pozwala zmienić zadaniu, takich jak gałąź,
inwentarz, **Limit**, **Tagi**, **Pomiń tagi** lub **CLI args**, oraz od opcji
Ansible lub Terraform, OpenTofu i Terragrunt.

Serwer sprawdza szablon przy każdym wyzwoleniu harmonogramu: zapisana tu wartość jest ignorowana,
gdy szablon przestaje pozwalać zadaniu ją zmieniać.

## Wiadomość {#taskParams-message}

Opcjonalne. Notatka przechowywana z każdym zadaniem uruchamianym przez harmonogram i wyświetlana w
historii zadań, na przykład `Nightly deploy`. Nie zmienia sposobu wykonania zadania.

## Zmienne ankiety {#taskParams-survey}

Wyświetlane, gdy szablon ma zmienne ankiety.

Wartości zmiennych ankiety szablonu dla zadań z harmonogramu. Docierają do zadania w ten sam
sposób, co przy ręcznym wypełnieniu ankiety. Wymagana zmienna musi mieć wartość, zanim harmonogram
będzie można zapisać.

Wartości domyślne szablonu są kopiowane przy otwarciu okna dialogowego i zapisywane z
harmonogramem. Wartość domyślna zmieniona później w szablonie nie trafia do harmonogramu zapisanego
wcześniej.

Wartość zmiennej typu secret nie jest zapisywana z harmonogramem: zadanie z harmonogramu jest
wykonywane bez niej.

## Gałąź {#taskParams-gitBranch}

Wyświetlane, gdy szablon pozwala zadaniu nadpisać gałąź.

Gałąź lub tag repozytorium, które pobiera zadanie z harmonogramu. Pole początkowo zawiera gałąź
szablonu. Serwer używa tej wartości tylko wtedy, gdy szablon zezwala na nadpisanie.

## Inwentarz {#taskParams-inventoryId}

Wyświetlane, gdy szablon pozwala zadaniu wybrać inwentarz.

Inwentarz, na którym jest wykonywane zadanie z harmonogramu, zamiast inwentarza szablonu. Dla
szablonu Terraform lub OpenTofu pole nazywa się **Workspace**, ale serwer stosuje wybór tylko do
szablonów Ansible, a dla pozostałych zachowuje workspace szablonu.

## Debug {#taskParams-debug}

Wyświetlane dla szablonu Ansible, który zezwala na debugowanie.

Uruchamia `ansible-playbook` z bardziej szczegółowymi danymi wyjściowymi: `-v` powtórzone tyle
razy, ile ustawia **Poziom debugowania**.

## Poziom debugowania {#taskParams-debugLevel}

Wyświetlane dla szablonu Ansible, który zezwala na debugowanie; aktywne, gdy zaznaczono **Debug**.

Szczegółowość danych wyjściowych, od `1` (`-v`) do `6` (`-vvvvvv`). Wartość domyślna to `4`.

## Symulacja {#taskParams-dryRun}

Wyświetlane dla szablonu Ansible, chyba że szablon je ukrywa.

Uruchamia playbook z `--check`: Ansible zgłasza, co by zmienił, i niczego nie zmienia.

## Różnice {#taskParams-diff}

Wyświetlane dla szablonu Ansible, chyba że szablon je ukrywa.

Uruchamia playbook z `--diff`: Ansible pokazuje zmiany, które wprowadza w plikach i szablonach.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Wyświetlane dla szablonu Ansible, który pozwala zadaniu to zmienić.

Pomija instalację ról i kolekcji z plików `requirements.yml` przed uruchomieniem playbooka.
Zaznacz, gdy wymagania są już zainstalowane, a zadanie nie powinno łączyć się z Ansible Galaxy.

## Plan {#taskParams-plan}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Wykonuje tylko `plan`. Zadanie kończy się po planie i niczego nie stosuje.

## Destroy {#taskParams-destroy}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Dodaje `-destroy` do planu i apply: zadanie niszczy zasoby workspace zamiast je tworzyć lub
aktualizować.

## Auto Approve {#taskParams-autoApprove}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Stosuje plan bez czekania na potwierdzenie. Działa tylko wtedy, gdy szablon zezwala na
automatyczne zatwierdzanie; szablon ustawiony na zatwierdzanie każdego zadania stosuje plan w obu
przypadkach.

Bez automatycznego zatwierdzania zadanie z harmonogramu zatrzymuje się po planie i czeka, aż ktoś
je potwierdzi lub odrzuci na stronie zadania.

## Upgrade {#taskParams-upgrade}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Uruchamia `init` z `-upgrade`: dostawcy (providers) i moduły są aktualizowani do najnowszych
wersji, na które pozwala konfiguracja.

## Reconfigure {#taskParams-reconfigure}

Wyświetlane dla szablonu Terraform, OpenTofu lub Terragrunt.

Uruchamia `init` z `-reconfigure` zamiast `-migrate-state`: backend jest konfigurowany od nowa, a
istniejący stan nie jest migrowany.

## Run at {#scheduleForm-runAt}

Wyświetlane dla harmonogramu jednorazowego.

Data i godzina rozpoczęcia zadania, w strefie czasowej serwera wyświetlanej na końcu pola. Nowy
harmonogram zaczyna się od najbliższej pełnej godziny. Moment musi leżeć w przyszłości: serwer
odrzuca czas, który już minął.

Po rozpoczęciu zadania harmonogram jest wyłączany lub usuwany, gdy zaznaczono **Delete after
run**. Gdy Semaphore nie działa w tym momencie, zadanie nie jest uruchamiane później: przy
następnym uruchomieniu serwera harmonogram jest wyłączany lub usuwany bez wykonania.

## Delete after run {#scheduleForm-deleteAfterRun}

Wyświetlane dla harmonogramu jednorazowego.

Usuwa harmonogram po rozpoczęciu jego zadania. Pozostaw pole odznaczone, aby zachować harmonogram,
wyłączony, tak by można było ustawić nowy czas i ponownie go włączyć.

## Show cron format {#scheduleForm-rawCron}

Wyświetlane dla harmonogramu cron.

Przełącza między wpisywaniem wyrażenia cron (włączone) a budowaniem go z **Timing** i pól wyboru
czasu (wyłączone). Oba sposoby edytują to samo wyrażenie. Przeglądarka zapamiętuje wybór.

Przełącznik jest zablokowany we włączonej pozycji, gdy wyrażenia nie da się przedstawić za pomocą
pól wyboru: gdy jest nieprawidłowe lub używa formy takiej jak `@daily` czy `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Wyświetlane dla harmonogramu cron, gdy **Show cron format** jest włączone.

Wyrażenie cron w pięciu polach: minuta, godzina, dzień miesiąca, miesiąc, dzień tygodnia (`0` to
niedziela). Pole jest wymagane, a serwer sprawdza wyrażenie podczas wpisywania. Nowy harmonogram
zaczyna się od `* * * * *`, czyli co minutę.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

Serwer akceptuje także `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` i `@every 1h30m`
oraz przedrostek taki jak `CRON_TZ=Europe/Berlin` dla innej strefy czasowej. Podgląd poniżej nie
potrafi pokazać następnego uruchomienia dla takich wyrażeń.

## Timing {#scheduleForm-timing}

Wyświetlane dla harmonogramu cron, gdy **Show cron format** jest wyłączone.

Jak często jest uruchamiane zadanie. Pola wyboru poniżej zmieniają się w zależności od tego
ustawienia:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** i **Minutes**.
- **Weekly**: **Weekdays**, **Hours** i **Minutes**.
- **Monthly**: **Days**, **Hours** i **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** i **Minutes**.

## Months {#scheduleForm-months}

Wyświetlane, gdy **Timing** to **Yearly**.

Miesiące, w których jest uruchamiane zadanie. Zaznacz kilka, aby uruchamiać je w każdym z nich.
Gdy żaden nie jest zaznaczony, zadanie jest uruchamiane w każdym miesiącu.

## Weekdays {#scheduleForm-weekdays}

Wyświetlane, gdy **Timing** to **Weekly**.

Dni tygodnia, w które jest uruchamiane zadanie. Gdy żaden nie jest zaznaczony, zadanie jest
uruchamiane codziennie.

## Days {#scheduleForm-days}

Wyświetlane, gdy **Timing** to **Monthly** lub **Yearly**.

Dni miesiąca, w które jest uruchamiane zadanie. Miesiąc bez danego dnia jest pomijany: zadanie
ustawione na 31. dzień nie jest uruchamiane w kwietniu. Gdy żaden nie jest zaznaczony, zadanie jest
uruchamiane codziennie.

## Hours {#scheduleForm-hours}

Wyświetlane, gdy **Timing** to **Daily**, **Weekly**, **Monthly** lub **Yearly**.

Godziny, o których jest uruchamiane zadanie, w strefie czasowej wyświetlanej po prawej stronie. Gdy
żadna nie jest zaznaczona, zadanie jest uruchamiane co godzinę.

## Minutes {#scheduleForm-minutes}

Wyświetlane dla harmonogramu cron, gdy **Show cron format** jest wyłączone.

Minuty godziny, w których jest uruchamiane zadanie, co pięć minut. Gdy żadna nie jest zaznaczona,
zadanie jest uruchamiane co minutę: zaznacz co najmniej jedną, na przykład `:00`, dla zadania,
które ma być uruchamiane raz na godzinę lub raz dziennie.

## Next run time {#scheduleForm-nextRun}

Kiedy rozpoczęłoby się następne zadanie według harmonogramu w obecnej postaci formularza: w strefie
czasowej serwera i w czasie lokalnym twojej przeglądarki. Kreska oznacza, że wyrażenie jest
nieprawidłowe lub podgląd nie potrafi go odczytać, jak w przypadku `@daily`.

## Włączony {#scheduleForm-active}

Czy harmonogram uruchamia zadania. Odznacz pole, aby zachować harmonogram bez jego uruchamiania.
Przełącznik na liście harmonogramów działa tak samo. Nowy harmonogram jest włączony.
