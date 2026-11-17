---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Edytuj użytkownika

Użytkownik to konto, które loguje się do Semaphore. Karta **Settings** zawiera profil konta i jego
flagi; **Zapisz** je zapisuje. Karta **Security** zawiera hasło, uwierzytelnianie dwuskładnikowe i
powiązane konta logowania, a karta **Roles** role globalne użytkownika: obie działają od razu, bez
**Zapisz**. To, co użytkownik może robić wewnątrz projektu, ustawia się na stronie **Zespół** tego
projektu, a nie tutaj.

## Nazwa {#user-name}

Pełne imię i nazwisko użytkownika, wyświetlane na pasku bocznym oraz na listach użytkowników i
członków projektów. Pole jest wymagane.

W przypadku użytkownika **Zewnętrzny** Semaphore pobiera nazwę od dostawcy tożsamości przy
logowaniu, gdy dostawca ją przesyła, więc zmiana wprowadzona tutaj może nie przetrwać następnego
logowania.

## Nazwa użytkownika {#user-username}

Login użytkownika. Pole jest wymagane. Użytkownik loguje się nazwą użytkownika lub adresem e-mail
oraz hasłem.

Nazwy użytkownika **Zewnętrzny** nie można zmienić: serwer odmawia jej zapisania.

## Email {#user-email}

Adres e-mail użytkownika. Pole jest wymagane. Semaphore wysyła na niego alerty zadań użytkownika, a
aplikacja uwierzytelniająca wyświetla go jako nazwę konta. Jest prywatny: widzą go tylko sam
użytkownik i osoby, które mogą zarządzać użytkownikami.

W przypadku użytkownika **Zewnętrzny** adres e-mail pochodzi od dostawcy tożsamości: pole jest tylko
do odczytu, a Semaphore aktualizuje adres przy każdym logowaniu.

## Hasło {#user-password}

Wyświetlane, gdy tworzony jest nowy użytkownik.

Hasło, którym loguje się użytkownik. Pole jest wymagane, z wyjątkiem użytkownika **Zewnętrzny**,
który nie ma hasła w Semaphore. Aby zmienić hasło istniejącego użytkownika, użyj
**Change password** na karcie **Security**.

## Wyślij alerty {#user-alert}

Wysyła użytkownikowi e-mail, gdy zadanie zakończy się niepowodzeniem w projekcie, w którym
użytkownik może widzieć to zadanie. Projekt musi zezwalać na alerty, a administrator serwera musi
mieć włączone alerty e-mail. Szablon zadania może zablokować wysyłanie alertów przez swoje zadania
zakończone niepowodzeniem.

Flaga nie wpływa na alerty projektu wysyłane do Telegram, Slack i innych komunikatorów.

## Użytkownik administratora {#user-admin}

Wyświetlane, gdy jesteś administratorem.

Administrator może robić wszystko na serwerze: zarządzać użytkownikami, rolami globalnymi i
ustawieniami serwera oraz otwierać każdy projekt z pełnymi uprawnieniami.

Nie możesz odebrać tej flagi własnemu kontu, a serwer zachowuje co najmniej jednego
administratora: odmawia odebrania flagi ostatniemu z nich.

## Pro user {#user-pro}

Wyświetlane w wersji Pro. Tylko administrator może zmienić tę flagę.

Przydziela użytkownikowi miejsce w subskrypcji Pro. Subskrypcja ma stałą liczbę użytkowników Pro:
po jej osiągnięciu serwer odmawia zapisania flagi. Bez aktywnej subskrypcji flaga nie działa.

## Zewnętrzny {#user-external}

Wyświetlane, gdy jesteś administratorem. Flagę można ustawić tylko przy tworzeniu użytkownika.

Użytkownik zewnętrzny loguje się przez LDAP lub OpenID Connect, a nie hasłem Semaphore: pole hasła
jest wyłączone, a logowanie hasłem jest odrzucane. Nazwa i adres e-mail są pobierane od dostawcy
tożsamości przy każdym logowaniu. Pozostaw pole odznaczone dla użytkownika, który loguje się
hasłem.

## Change password {#user-changePassword}

Wyświetlane na karcie **Security** istniejącego użytkownika, który nie jest **Zewnętrzny**, gdy
serwer zezwala na logowanie hasłem.

Otwiera okno dialogowe, które ustawia nowe hasło. Aby zmienić własne hasło, musisz podać obecne;
aby zresetować hasło innego użytkownika, nie musisz.

## Time-based one-time password {#user-totp}

Wyświetlane na karcie **Security**, gdy administrator włączył na serwerze uwierzytelnianie
dwuskładnikowe.

Włącza uwierzytelnianie dwuskładnikowe dla tego użytkownika. Po włączeniu przełącznika Semaphore od
razu tworzy sekret i wyświetla go jako kod QR: zeskanuj go aplikacją uwierzytelniającą. Od tej pory
logowanie wymaga po haśle kodu z aplikacji.

Gdy serwer zezwala na odzyskiwanie, pod kodem QR wyświetlany jest kod odzyskiwania. Jest
wyświetlany tylko raz: skopiuj go i przechowuj w bezpiecznym miejscu. Wyłączenie przełącznika od
razu usuwa sekret.

## Dostawca LDAP {#user-ldapProvider}

Wyświetlane na karcie **Security** twojego własnego konta, gdy serwer ma dostawcę LDAP, z którym
twoje konto nie jest jeszcze powiązane.

Serwer LDAP, z którym ma zostać powiązane twoje konto, gdy jest ich kilka.

## LDAP username {#user-ldapUsername}

Wyświetlane razem z **Dostawca LDAP**.

Twoja nazwa użytkownika na serwerze LDAP. Powiązanie pozwala logować się do tego konta także przez
LDAP.

## LDAP password {#user-ldapPassword}

Wyświetlane razem z **Dostawca LDAP**.

Twoje hasło na serwerze LDAP. Semaphore używa go jednorazowo, aby sprawdzić, czy konto LDAP należy
do ciebie, i go nie przechowuje.

## Link {#user-ldapLink}

Wyświetlane razem z **Dostawca LDAP**.

Wiąże twoje konto z powyższym kontem LDAP. Powiązanie jest odrzucane, gdy dane logowania są
nieprawidłowe, gdy konto LDAP jest powiązane z innym użytkownikiem lub gdy twoje konto jest już
powiązane z tym dostawcą. Powiązane konto pojawia się w sekcji **Linked accounts**; zamknij jego
chip, aby usunąć powiązanie.

## Global role {#global-role-select}

Wyświetlane na karcie **Roles** istniejącego użytkownika, w wersji Pro, gdy możesz zarządzać rolami
globalnymi.

Rola globalna do nadania użytkownikowi, spośród ról, których użytkownik jeszcze nie ma. Rola
globalna przyznaje uprawnienia do całego serwera: do zarządzania użytkownikami, rolami globalnymi
lub ustawieniami serwera albo do odczytu dziennika audytu. Same role globalne tworzy się na karcie
**Role globalne** strony Użytkownicy.

## Assign {#globalRoleAssign-assign}

Wyświetlane razem z **Global role**.

Od razu nadaje użytkownikowi wybraną rolę. Role, które ma użytkownik, są wymienione poniżej, każda
z przyciskiem do jej odebrania. **Effective global permissions** pokazuje, co ostatecznie może robić
użytkownik i skąd pochodzi każde uprawnienie.
