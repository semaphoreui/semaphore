---
form: src/components/ProxyForm.vue
---

# Bearbeiten Proxy

Ein Proxy ist ein Host, über den die Aufgaben des Projekts Server erreichen, mit denen sie sich
nicht direkt verbinden können. Wählen Sie ihn im Feld **Proxy** eines Repositorys, um das
Repository darüber abzurufen, oder eines Inventars, um Ansible darüber mit den Hosts zu verbinden.

## Name {#proxyForm-name}

Pflichtfeld, eindeutig im Projekt. Repositorys und Inventare zeigen den Proxy unter diesem Namen.

## Typ {#proxyForm-type}

- **SSH (jump host)**: Semaphore verbindet sich per SSH mit dem Host und öffnet von dort die
  Verbindung zum Server. Nur dieser Typ kann einen anderen Proxy benötigen.
- **SOCKS5**, **HTTP**, **HTTPS**: ein Proxy-Server, der die Verbindung trägt. Git nutzt ihn für
  ein `https`-Repository selbst; eine SSH-Verbindung leitet Semaphore durch ihn, es muss also kein
  weiteres Werkzeug installiert sein.

## Host {#proxyForm-host}

Pflichtfeld. Ein Hostname oder eine IP-Adresse, ohne Benutzer, Port oder Schema. Ein Hostname darf
nur Buchstaben, Ziffern, Punkte, Bindestriche und Unterstriche enthalten.

## Port {#proxyForm-port}

Optional, von 1 bis 65535. Leer bedeutet `22` für einen SSH-Jump-Host. Geben Sie den Port bei den
anderen Typen an: ohne ihn nehmen Git und SSH-Verbindungen unterschiedliche Ports an.

## User {#proxyForm-user}

Nur SSH: der Login auf dem Jump-Host. Leer bedeutet der Benutzer, unter dem Semaphore läuft.
Buchstaben, Ziffern, Punkte, Bindestriche und Unterstriche, nicht mit einem Bindestrich am Anfang.
Ein SOCKS5- oder HTTP-Proxy nimmt seinen Login stattdessen aus dem **Zugangsschlüssel**.

## Zugangsschlüssel {#proxyForm-sshKeyId}

Optional. Der Schlüssel aus dem Schlüsselspeicher, mit dem man sich am Proxy selbst anmeldet,
meist nicht der Schlüssel der Server dahinter:

- ein **SSH-Schlüssel** für einen SSH-Jump-Host;
- ein **Login mit Passwort** für SOCKS5, HTTP und HTTPS. Das Passwort wird in der Umgebung
  übergeben, nie auf einer Befehlszeile.

## Requires Proxy {#proxyForm-requiresProxyId}

Nur SSH, optional. Ein anderer SSH-Proxy des Projekts, der zuerst durchlaufen wird, für einen
Jump-Host, der nur von einem anderen aus erreichbar ist. Eine Kette umfasst höchstens 5 Proxys und
darf nicht zu einem Proxy zurückführen, der schon in ihr ist.
