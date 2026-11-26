---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Mapowanie wskazuje Semaphore, którego poświadczenia z Magazynu kluczy użyć dla hosta SSH lub
adresu URL Git. Repozytorium ma jeden klucz, ale zadanie może łączyć się także z innymi serwerami:
submodułami, rolami i kolekcjami z `requirements.yml`, modułami Terraform, repozytoriami
inwentarzy. Mapowania projektu dotyczą każdego połączenia Git i SSH jego zadań oraz poleceń Git,
które Semaphore wykonuje dla projektu, takich jak sprawdzanie repozytorium pod kątem nowych
commitów. Żaden plik twojego repozytorium się nie zmienia.

## Typ {#hostConfigForm-type}

Co dopasowuje mapowanie:

- **Host**: nazwa hosta SSH. Każde połączenie SSH zadania z tym hostem używa klucza mapowania:
  Git przez SSH oraz Ansible łączący się z hostem inwentarza o tej nazwie. Mapowanie ma
  pierwszeństwo przed konfiguracją SSH serwera dla tego samego hosta.
- **URL**: adres URL Git lub jego początek. Git łączy się z każdym repozytorium pod tym adresem
  za pomocą poświadczenia mapowania.

Przełączenie na **Host** czyści pole **Credential**, jeśli mapowanie hosta nie może użyć wybranego
klucza.

## Host or URL {#hostConfigForm-host}

Pole jest wymagane. Dla każdego hosta lub adresu URL może istnieć jedno mapowanie danego typu.

Dla **Host**: nazwa hosta dokładnie taka, pod jaką łączą się z nim Git lub Ansible, na przykład
`github.com` lub `10.0.0.21`. Tylko litery, cyfry, kropki, myślniki i podkreślenia: bez
użytkownika, bez portu, bez symboli wieloznacznych.

Dla **URL**: adres zaczynający się od `https://` lub `http://`, bez loginu oraz bez spacji,
cudzysłowów i `=`. Podaj całe repozytorium albo przedrostek zakończony `/`, aby objąć grupę
repozytoriów:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Gdy pasuje kilka mapowań, wygrywa najdłuższe: powyższe mapowanie repozytorium zastępuje mapowanie
`acme` dla tego jednego repozytorium.

Z kluczem SSH mapowanie URL sprawia, że Git łączy się przez SSH z hostem z adresu URL na porcie
`22`, niezależnie od portu podanego w adresie.

## Credential {#hostConfigForm-sshKeyId}

Klucz z Magazynu kluczy używany do połączeń dopasowanych przez mapowanie. Pole jest wymagane.

- **Host** przyjmuje tylko **Klucz SSH**. Gdy klucz ma login, SSH łączy się jako ten użytkownik.
- **URL** przyjmuje **Klucz SSH** lub **Logowanie z hasłem**. Klucz SSH łączy się jako swój login
  lub jako `git`, gdy loginu nie ma. Logowanie z hasłem działa tylko z adresem `https://`; w
  przypadku tokenu wpisz token w hasło i pozostaw login pusty.

Logowanie z hasłem trafia do Git przez jego środowisko, a nie wiersz poleceń, a dziennik zadania
pokazuje tylko oryginalny adres URL.
