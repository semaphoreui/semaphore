---
form: src/components/ProxyForm.vue
---

# Upravit Proxy

Proxy je hostitel, přes kterého úlohy projektu procházejí k serverům, ke kterým se nemohou
připojit přímo. Vyberte ji v poli **Proxy** repozitáře, aby se repozitář stahoval přes ni, nebo
inventáře, aby se přes ni Ansible připojoval k hostům.

## Název {#proxyForm-name}

Povinné a v projektu jedinečné. Repozitáře a inventáře zobrazují proxy pod tímto názvem.

## Typ {#proxyForm-type}

- **SSH (jump host)**: Semaphore se k hostiteli připojí přes SSH a odtud otevře spojení
  k serveru. Jen tento typ může vyžadovat jinou proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: proxy server, který spojení přenáší. Git jej sám použije pro
  repozitář `https`; SSH spojení přes něj vede Semaphore, takže není třeba instalovat žádný další
  nástroj.

## Host {#proxyForm-host}

Povinné. Název hostitele nebo IP adresa, bez uživatele, portu a schématu. Název hostitele smí
obsahovat jen písmena, číslice, tečky, pomlčky a podtržítka.

## Port {#proxyForm-port}

Nepovinné, od 1 do 65535. Prázdné pole znamená `22` pro SSH jump host. U ostatních typů port
uveďte: bez něj Git a SSH spojení předpokládají různé porty.

## User {#proxyForm-user}

Jen pro SSH: přihlašovací jméno na jump hostu. Prázdné pole znamená uživatele, pod kterým běží
Semaphore. Písmena, číslice, tečky, pomlčky a podtržítka, bez pomlčky na začátku. Proxy SOCKS5
nebo HTTP bere přihlašovací jméno z pole **Přístupový klíč**.

## Přístupový klíč {#proxyForm-sshKeyId}

Nepovinné. Klíč z úložiště klíčů, kterým se přihlašuje k samotné proxy, obvykle jiný než klíč
serverů za ní:

- **SSH klíč** pro SSH jump host;
- **Přihlášení heslem** pro SOCKS5, HTTP a HTTPS. Heslo se předává v prostředí, nikdy na
  příkazové řádce.

## Requires Proxy {#proxyForm-requiresProxyId}

Jen pro SSH, nepovinné. Jiná SSH proxy projektu, přes kterou se musí projít nejdřív, pro jump
host dostupný jen z jiného. Řetězec má nejvýše 5 proxy a nesmí se vracet k proxy, která v něm už
je.
