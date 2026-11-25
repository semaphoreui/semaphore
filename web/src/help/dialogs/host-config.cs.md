---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Mapování říká Semaphore, které přihlašovací údaje z Úložiště klíčů má použít pro hosta SSH nebo
pro URL Gitu. Repozitář nese jeden klíč, ale úloha se může připojovat i k dalším serverům:
submoduly, role a kolekce z `requirements.yml`, moduly Terraformu, repozitáře inventářů.
Mapování projektu platí pro každé připojení Git a SSH jeho úloh a pro příkazy Git, které
Semaphore pro projekt spouští, například kontrolu nových commitů v repozitáři. Žádný soubor ve
vašem repozitáři se nezmění.

## Typ {#hostConfigForm-type}

Čemu mapování odpovídá:

- **Host**: název hosta SSH. Každé připojení SSH úlohy k tomuto hostovi použije klíč mapování:
  Git přes SSH i Ansible, který se připojuje k hostovi inventáře s tímto názvem. Mapování má
  u stejného hosta přednost před konfigurací SSH serveru.
- **URL**: URL Gitu nebo jeho začátek. Git se ke každému repozitáři pod touto URL připojí
  s přihlašovacími údaji mapování.

Přepnutí na **Host** vymaže **Credential**, který mapování hosta nemůže použít.

## Host nebo URL {#hostConfigForm-host}

Pole je povinné. Pro každého hosta nebo URL může existovat jedno mapování daného typu.

Pro **Host**: název hosta přesně tak, jak se k němu Git nebo Ansible připojují, například
`github.com` nebo `10.0.0.21`. Pouze písmena, číslice, tečky, spojovníky a podtržítka: žádný
uživatel, žádný port, žádné zástupné znaky.

Pro **URL**: adresa, která začíná na `https://` nebo `http://`, bez přihlašovacího jména a bez
mezer, uvozovek nebo `=`. Zadejte celý repozitář, nebo předponu, která končí na `/`, aby
pokryla skupinu repozitářů:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Když odpovídá několik mapování, vyhraje nejdelší: mapování repozitáře výše přepíše pro tento
jeden repozitář mapování `acme`.

S SSH klíčem způsobí mapování URL, že se Git připojí přes SSH k hostovi z URL na portu `22`, ať
URL uvádí jakýkoli port.

## Credential {#hostConfigForm-sshKeyId}

Klíč z Úložiště klíčů, který se použije pro připojení, kterým mapování odpovídá. Pole je
povinné.

- **Host** přijímá pouze **SSH klíč**. Když má klíč přihlašovací jméno, SSH se připojí jako tento
  uživatel.
- **URL** přijímá **SSH klíč** nebo **Přihlášení heslem**. SSH klíč se připojí pod svým
  přihlašovacím jménem, nebo jako `git`, když žádné nemá. Přihlášení heslem funguje pouze
  s URL `https://`; pro token vložte token do hesla a přihlašovací jméno ponechte prázdné.

Přihlášení heslem se ke Gitu dostane přes jeho prostředí, ne přes příkazový řádek, a log úlohy
ukazuje pouze původní URL.
