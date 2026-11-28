---
form: src/components/ProxyForm.vue
---

# Bewerken Proxy

Een proxy is een host waarlangs de taken van het project servers bereiken waarmee ze niet direct
kunnen verbinden. Kies hem in het veld **Proxy** van een repository, om de repository erlangs op te
halen, of van een inventaris, om Ansible erlangs met de hosts te verbinden.

## Naam {#proxyForm-name}

Verplicht en uniek in het project. Repositories en inventarissen tonen de proxy onder deze naam.

## Type {#proxyForm-type}

- **SSH (jump host)**: Semaphore maakt via SSH verbinding met de host en opent van daaruit de
  verbinding met de server. Alleen dit type kan een andere proxy vereisen.
- **SOCKS5**, **HTTP**, **HTTPS**: een proxyserver die de verbinding draagt. Git gebruikt hem zelf
  voor een `https`-repository; Semaphore leidt een SSH-verbinding erdoor, dus er hoeft geen ander
  hulpmiddel geïnstalleerd te zijn.

## Host {#proxyForm-host}

Verplicht. Een hostnaam of een IP-adres, zonder gebruiker, poort of schema. Een hostnaam mag alleen
letters, cijfers, punten, koppeltekens en underscores bevatten.

## Port {#proxyForm-port}

Optioneel, van 1 tot 65535. Leeg betekent `22` voor een SSH jump host. Geef de poort op bij de
andere types: zonder poort gaan Git en SSH-verbindingen van verschillende poorten uit.

## User {#proxyForm-user}

Alleen SSH: de login op de jump host. Leeg betekent de gebruiker waaronder Semaphore draait.
Letters, cijfers, punten, koppeltekens en underscores, niet beginnend met een koppelteken. Een
SOCKS5- of HTTP-proxy neemt zijn login in plaats daarvan uit de **Toegangssleutel**.

## Toegangssleutel {#proxyForm-sshKeyId}

Optioneel. De sleutel uit de sleutelopslag waarmee wordt ingelogd op de proxy zelf, meestal niet de
sleutel van de servers erachter:

- een **SSH Sleutel** voor een SSH jump host;
- een **Inloggen met wachtwoord** voor SOCKS5, HTTP en HTTPS. Het wachtwoord gaat via de omgeving,
  nooit via een opdrachtregel.

## Requires Proxy {#proxyForm-requiresProxyId}

Alleen SSH, optioneel. Een andere SSH-proxy van het project waar eerst doorheen moet worden gegaan,
voor een jump host die alleen vanaf een andere bereikbaar is. Een keten bevat hoogstens 5 proxy's
en kan niet terugkeren naar een proxy die er al in zit.
