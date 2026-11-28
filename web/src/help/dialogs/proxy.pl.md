---
form: src/components/ProxyForm.vue
---

# Edytuj Proxy

Proxy to host, przez który zadania projektu docierają do serwerów, z którymi nie mogą połączyć się
bezpośrednio. Wybierz je w polu **Proxy** repozytorium, aby pobierać repozytorium przez nie, lub
inwentarza, aby Ansible łączył się przez nie z hostami.

## Nazwa {#proxyForm-name}

Wymagane i unikalne w projekcie. Repozytoria i inwentarze pokazują proxy pod tą nazwą.

## Typ {#proxyForm-type}

- **SSH (jump host)**: Semaphore łączy się z hostem przez SSH i stamtąd otwiera połączenie
  z serwerem. Tylko ten typ może wymagać innego proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: serwer proxy, który przenosi połączenie. Git sam go używa dla
  repozytorium `https`; połączenie SSH Semaphore przeprowadza przez niego, więc nie trzeba
  instalować żadnego innego narzędzia.

## Host {#proxyForm-host}

Wymagane. Nazwa hosta lub adres IP, bez użytkownika, portu i schematu. Nazwa hosta może zawierać
tylko litery, cyfry, kropki, myślniki i podkreślenia.

## Port {#proxyForm-port}

Opcjonalne, od 1 do 65535. Puste oznacza `22` dla jump hosta SSH. Dla pozostałych typów podaj port:
bez niego Git i połączenia SSH zakładają różne porty.

## User {#proxyForm-user}

Tylko SSH: login na jump hoście. Puste oznacza użytkownika, jako który działa Semaphore. Litery,
cyfry, kropki, myślniki i podkreślenia, bez myślnika na początku. Proxy SOCKS5 lub HTTP bierze login
z pola **Klucz dostępu**.

## Klucz dostępu {#proxyForm-sshKeyId}

Opcjonalne. Klucz z magazynu kluczy, którym loguje się do samego proxy, zwykle inny niż klucz
serwerów za nim:

- **Klucz SSH** dla jump hosta SSH;
- **Logowanie z hasłem** dla SOCKS5, HTTP i HTTPS. Hasło jest przekazywane w środowisku, nigdy
  w wierszu poleceń.

## Requires Proxy {#proxyForm-requiresProxyId}

Tylko SSH, opcjonalne. Inne proxy SSH projektu, przez które trzeba przejść najpierw, dla jump hosta
osiągalnego tylko z innego. Łańcuch ma najwyżej 5 proxy i nie może wracać do proxy, które już
w nim jest.
