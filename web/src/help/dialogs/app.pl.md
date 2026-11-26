---
form: src/components/AppForm.vue
---

# Edit App

Aplikacja to narzędzie, za pomocą którego Semaphore uruchamia zadania: Ansible, Terraform, OpenTofu,
Terragrunt, Bash, Python, PowerShell lub własne polecenie. Każdy szablon zadania jest tworzony dla
jednej aplikacji, a to okno dialogowe określa, jak serwer ją uruchamia i jak aplikacja jest
wyświetlana w interfejsie.

Ustawienia są przechowywane w bazie danych i zaczynają obowiązywać od razu na serwerze, który je
zapisuje; pozostałe węzły klastra odczytują je przy ponownym uruchomieniu. Pole, które opróżnisz,
zachowuje swoją poprzednią wartość do ponownego uruchomienia serwera.

## ID {#appForm-id}

Wymagane. Klucz, pod którym aplikacja jest przechowywana; szablon zadania odwołuje się do swojej
aplikacji za pomocą tego klucza. Wbudowane aplikacje to `ansible`, `terraform`, `tofu`,
`terragrunt`, `pulumi`, `bash`, `python` i `powershell`. Każde inne ID dodaje własną aplikację:
Semaphore uruchamia ją jak skrypt, poleceniem podanym w polu **Path** lub, gdy **Path** jest puste,
poleceniem o nazwie takiej jak ID.

Użyj ID bez kropek. Zmiana ID istniejącej aplikacji nie zmienia jej nazwy: ustawienia są
zapisywane jako nowa aplikacja pod nowym ID, a stara aplikacja pozostaje.

## Icon {#appForm-icon}

Opcjonalne. Nazwa ikony z zestawu Material Design Icons bez przedrostka `mdi-`, na przykład
`language-python`. Pozostaw pole puste, aby zachować wbudowaną ikonę aplikacji;
własna aplikacja wyświetla wtedy znak zapytania.

## Nazwa {#appForm-title}

Opcjonalne. Nazwa aplikacji na liście aplikacji i w menu **Nowy szablon**. Pozostaw pole puste,
aby zachować wbudowaną nazwę, taką jak `Ansible Playbook`; własna aplikacja nie ma wtedy nazwy,
więc nadaj jej nazwę.

## Path {#appForm-path}

Opcjonalne. Polecenie, które Semaphore uruchamia zamiast domyślnego: nazwa wyszukiwana w `PATH`
serwera albo pełna ścieżka, na przykład `/usr/local/bin/terraform-1.9`. Domyślne polecenia to
`terraform`, `tofu`, `terragrunt`, `bash`, `python3` i `powershell`; własna aplikacja uruchamia
polecenie o nazwie takiej jak jej ID.

Ansible ignoruje to pole: Semaphore zawsze uruchamia `ansible-playbook` i `ansible-galaxy`.

## Priority {#appForm-priority}

Miejsce aplikacji na liście aplikacji i w menu **Nowy szablon**: aplikacja z wyższą liczbą jest
wyżej. Wbudowane aplikacje zaczynają się od 1000 dla Ansible i schodzą do 500 dla Pythona.
Wbudowana aplikacja z priorytetem `0` lub niższym odzyskuje swój wbudowany priorytet przy
uruchomieniu serwera.

## Args {#appForm-args}

Opcjonalne. Argumenty umieszczane w wierszu poleceń zaraz po poleceniu, jeden na chip. Kliknij
**+**, aby dodać argument, kliknij chip, aby go zmienić, przeciągaj chipy, aby zmienić ich
kolejność.

- Terraform, OpenTofu i Terragrunt: argumenty stoją przed argumentami, które Semaphore dodaje do
  każdego polecenia, takimi jak `init` i `apply`.
- Aplikacje skryptowe: argumenty stoją przed skryptem szablonu i zastępują argumenty domyślne:
  PowerShell jest domyślnie uruchamiany z `-File`, więc umieść `-File` na końcu swojej listy.
- Ansible ignoruje to pole.

## Active {#appForm-active}

Aktywna aplikacja jest oferowana w menu **Nowy szablon**. Odznacz pole, aby ukryć aplikację w tym
menu: szablony już z nią utworzone zachowują swoją aplikację i nadal działają.

Dla nowej aplikacji pole jest odznaczone: zaznacz je, aby aplikacja była dostępna.
