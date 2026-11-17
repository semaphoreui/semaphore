---
form: src/views/Cluster.vue
---

# Clear tasks from Redis

Węzły klastra dzielą stan puli zadań w Redis: które zadania czekają, które są wykonywane i który
węzeł jest właścicielem każdego z nich. To okno dialogowe usuwa z Redis zaznaczone grupy rekordów,
aby przywrócić działanie klastra, który utknął, na przykład gdy zakończone zadanie nadal jest
wyświetlane jako wykonywane. Nic nie jest usuwane z bazy danych: zadania zachowują swój status,
swój dziennik i swoją historię.

Niezależnie od tego, co zaznaczysz, każdy węzeł porzuca również zadania przechowywane w pamięci i
odczytuje je ponownie z Redis i bazy danych. Po otwarciu okna zaznaczona jest tylko grupa
**Queue**; zaznaczaj tylko te grupy, które utknęły. Wynik pokazuje, ile rekordów usunięto.

## Queue {#clusterClearDialog-queue}

Zadania, które czekają, aż węzeł je uruchomi. Wyczyszczone zadania nie są uruchamiane: pozostają w
bazie danych ze statusem Waiting. Zatrzymaj je lub ponownie uruchom ich szablony.

## Running {#clusterClearDialog-running}

Zadania, które klaster uznaje za wykonywane. Wyczyszczenie ich nie zatrzymuje: zadanie nadal
działa na swoim uruchamiaczu. Klaster jednak traci je z oczu: zadanie przestaje się wliczać do
limitów równoległych zadań serwera i swojego uruchamiacza, więc może się uruchomić więcej zadań,
niż pozwalają limity, a zadanie, którego uruchamiacz jeszcze nie odebrał, nie jest mu już
przekazywane.

## Active by project {#clusterClearDialog-active}

Niezakończone zadania każdego projektu. Służą one do egzekwowania limitu równoległych zadań
projektu i nie pozwalają uruchomić drugiego zadania szablonu, gdy jedno jest wykonywane, jeśli
szablon nie zezwala na równoległe zadania. Wyczyść tę grupę, gdy zadania projektu czekają bez
końca za zadaniami, które już się zakończyły. Zadania wykonywane w tym momencie przestają się
wliczać, więc obok nich mogą się uruchomić nowe zadania projektu.

## Aliases {#clusterClearDialog-aliases}

Losowe nazwy, które zadania Terraform, OpenTofu i Terragrunt otrzymują na czas swojego życia,
przypisane do zadania. Wykonywane zadanie dociera pod tą nazwą do swojego stanu Terraform na
serwerze: po wyczyszczeniu wykonywanego zadania tych aplikacji nie da się już pod nią odnaleźć.

## Claims {#clusterClearDialog-claims}

Który węzeł jest właścicielem każdego zadania. Węzeł co minutę odnawia roszczenia swoich zadań;
roszczenie, które nie zostało odnowione, wygasa po pięciu minutach. Zadanie bez roszczenia uznaje
się za osierocone: zadanie, które nie dotarło jeszcze do uruchamiacza, wraca do kolejki dla innego
węzła, a zadanie, które uruchamiacz już wykonuje, może się tam zakończyć.

## Runtime fields {#clusterClearDialog-runtimeFields}

Szczegóły każdego zadania, których nie przechowuje baza danych: użytkownik, który je uruchomił,
wersja, z którą zostało uruchomione, jego alias, jego uruchamiacz i jego status oraz projekt
każdego zadania. Bez nich węzeł, który nie uruchomił zadania, pokazuje o nim mniej, a mechanizm
czyszczenia osieroconych zadań pomija wykonywane zadanie, więc zablokowany rekord nie jest już
usuwany samoczynnie.
