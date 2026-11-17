---
form: src/components/KeyForm.vue
---

# Upravit Key

Klíč z Úložiště klíčů obsahuje přihlašovací údaje, které projekt používá, aniž by je ukazoval: SSH
klíč nebo heslo, kterým se Ansible přihlašuje k hostům, přihlašovací údaje repozitáře Git, heslo
vaultu Ansible, tajný údaj integrace. Inventáře, repozitáře, šablony úloh a integrace odkazují na
klíč jeho názvem. Po uložení se tajný údaj už nikdy neposílá zpět do prohlížeče.

## Název klíče {#key-name}

Název, pod kterým je klíč uveden v Úložišti klíčů a ve výběrech klíče v inventářích,
repozitářích, šablonách úloh a integracích. Pole je povinné. Zvolte název, který říká, co klíč
otevírá, například `Deploy key of the web servers`.

## Zdroj {#key-source}

Kde je tajný údaj klíče uložen. Karty:

- **Local**: tajný údaj se zadá do tohoto dialogu a uloží se do databáze Semaphore, zašifrovaný,
  když správce nastavil šifrovací klíč pro přístupové klíče. Toto je výchozí volba.
- **Storage**: tajný údaj je uložen v externím úložišti tajných údajů projektu, vybraném níže
  v poli **Storage**. Karta je k dispozici pouze v edici Pro. Samotná úložiště se nastavují na
  kartě **Storages** Úložiště klíčů.
- **Env**: tajný údaj se čte z proměnné prostředí procesu serveru Semaphore. Název proměnné
  zadejte do pole pod kartami.
- **File**: tajný údaj se čte ze souboru na serveru Semaphore. Absolutní cestu k souboru zadejte
  do pole pod kartami. Soubor musí být uvnitř adresáře tajných údajů serveru (`dirs.secrets` v jeho
  konfiguraci, ve výchozím stavu `/tmp/semaphore`); cesta mimo něj, nebo cesta obsahující `..`,
  selže při použití klíče. Koncové zalomení řádku v souboru se odstraní.

U **Env** a **File** Semaphore tajný údaj nikdy nezapisuje: hodnotu čte pokaždé, když úloha klíč
použije, takže tajný údaj měníte tam, kde je uložen. Hodnota musí odpovídat poli **Typ** klíče.
Klíč **Řetězec** přebírá hodnotu tak, jak je. Klíč **SSH klíč** nebo **Přihlášení heslem** očekává
JSON s poli daného typu, například:

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Při úpravě klíče jsou karty uzamčené, dokud není zaškrtnuto **Přepsat**. Klíč synchronizovaný
z úložiště tajných údajů si ponechává svůj zdroj a jeho název a typ nelze změnit.

## Storage {#key-sourceStorageId}

Zobrazí se, když je vybrána karta **Storage**.

Úložiště tajných údajů projektu, které tajný údaj obsahuje: HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager nebo Azure Key Vault, tak jak je nastaveno na kartě **Storages**
Úložiště klíčů.

- Úložiště, které není jen pro čtení, přijme tajný údaj, který zadáte v tomto dialogu: Semaphore
  ho tam zapíše při uložení. Odstranění klíče odstraní tajný údaj i z úložiště. Klíč uložený
  v takovém úložišti nelze později přesunout do jiného úložiště.
- Úložiště jen pro čtení se pouze čte. Dialog pak neukazuje žádná pole pro tajný údaj a Semaphore
  ho čte z úložiště pokaždé, když úloha klíč použije.

## Source Key {#key-sourceStorageKey}

Zobrazí se, když je vybráno **Storage**.

Kde je tajný údaj v tomto úložišti, například cesta k tajnému údaji v HashiCorp Vault.

## Typ {#key-type}

Jaký druh přihlašovacích údajů klíč obsahuje. Pole je povinné a pole pod ním se mění podle typu.

- **SSH klíč**: soukromý klíč s volitelným přihlašovacím jménem a heslem k němu. Ansible se jím
  připojuje k hostům a Git jím klonuje repozitář s URL `ssh`.
- **Přihlášení heslem**: přihlašovací jméno a heslo. Ansible se jimi připojuje k hostům nebo se
  jimi stává jiným uživatelem (**Přihlašovací údaje sudo** inventáře); Git jimi klonuje repozitář
  `https`; heslo otevírá vault Ansible šablony úlohy.
- **Řetězec**: jediná tajná hodnota bez přihlašovacího jména: token, API klíč, heslo. Použijte ho
  pro tajný údaj integrace. Nemůže sloužit jako SSH klíč, přihlašovací údaje sudo, heslo vaultu
  ani přístupový klíč repozitáře.
- **Žádný**: žádné přihlašovací údaje. Použijte ho pro repozitář, který nepotřebuje přihlášení
  (veřejný repozitář `https`, URL `file` nebo místní cesta), a pro inventář, jehož hosté
  nepotřebují žádné přihlašovací údaje nebo si připojení definují sami.

## Uživatelské jméno {#key-loginPasswordLogin}

Zobrazí se, když je typ **Přihlášení heslem**. Volitelné.

Uživatelské jméno přihlašovacích údajů.

- Ansible se k hostům připojuje jako tento uživatel (`--user`); jako **Přihlašovací údaje sudo**
  inventáře je to uživatel, kterým se Ansible stane (`--become-user`). Když je pole prázdné,
  Ansible použije uživatele, kterého uvádí inventář nebo jeho vlastní konfigurace.
- U repozitáře `https` se Git přihlásí tímto uživatelem a heslem. Ponechte pole prázdné pro token,
  který server Git přijímá samotný: heslo se pak pošle jako uživatelské jméno.

## Heslo {#key-loginPasswordPassword}

Zobrazí se, když je typ **Přihlášení heslem**. Povinné.

Heslo přihlašovacích údajů. Semaphore jím odpovídá na výzvy Ansiblu k zadání hesla: heslo SSH
(`--ask-pass`) u přihlašovacích údajů uživatele, heslo pro `become` (`--ask-become-pass`)
u přihlašovacích údajů sudo a heslo vaultu u vaultu Ansible. U repozitáře `https` je to heslo
nebo přístupový token serveru Git.

## Hodnota {#key-stringValue}

Zobrazí se, když je typ **Řetězec**. Povinné.

Samotná tajná hodnota, například token, proti kterému integrace ověřuje požadavky, které přijímá.
Ukládá se tak, jak je zadána.

## Uživatelské jméno {#key-sshLogin}

Zobrazí se, když je typ **SSH klíč**. Volitelné.

Uživatel, jako který se Ansible připojuje k hostům (`--user`). Když je pole prázdné, Ansible
použije uživatele, kterého uvádí inventář nebo jeho vlastní konfigurace.

U repozitáře Git místo toho uveďte uživatele v URL repozitáře, jako v
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Zobrazí se, když je typ **SSH klíč**. Volitelné.

Heslo, kterým je soukromý klíč zašifrován. Pro klíč bez hesla pole ponechte prázdné. Špatné heslo
způsobí, že každá úloha, která klíč používá, selže při načítání klíče. Klíč vygenerovaný
Semaphore nemá žádné heslo: obsah tohoto pole se pak zahodí.

## Generate SSH Key {#key-generateSshKey}

Zobrazí se, když je typ **SSH klíč**.

Semaphore při uložení vytvoří nový pár klíčů: 2048bitový soukromý klíč RSA, který zůstane v klíči,
a jeho veřejný klíč, který se po uložení jednou zobrazí v dialogu a poté pod **Public key**
pokaždé, když klíč upravujete. Přidejte veřejný klíč do `~/.ssh/authorized_keys` na hostech, nebo
jako deploy key repozitáře Git.

Pole soukromého klíče je uzamčené, dokud je políčko zaškrtnuté. U existujícího klíče nejprve
zaškrtněte **Přepsat**: původní pár klíčů se nahradí. Klíč, jehož tajný údaj se čte z **Env**,
**File** nebo z úložiště jen pro čtení, nelze vygenerovat, protože Semaphore ho nemá kam zapsat.

## Soukromý klíč {#key-sshPrivateKey}

Zobrazí se, když je typ **SSH klíč**. Povinné, pokud není zaškrtnuto **Generate SSH Key**.

Soukromý klíč ve formátu OpenSSH nebo PEM, včetně řádků `BEGIN` a `END`. Pro úlohu ho Semaphore
načte do vlastního SSH agenta, kterého Ansible a Git používají po dobu úlohy.

## Public key {#key-publicKey}

Zobrazí se, když upravujete SSH klíč, který vygeneroval Semaphore, dokud není zaškrtnuto
**Přepsat**.

Veřejná polovina vygenerovaného páru klíčů. **Show** ji rozbalí, tlačítko kopírování ji vloží do
schránky. Přidejte ji do `~/.ssh/authorized_keys` na hostech, nebo jako deploy key repozitáře
Git.

Je to veřejný klíč páru, který Semaphore vygeneroval naposledy. Když soukromý klíč ručně nahradíte
pomocí **Přepsat**, klíč zobrazený zde mu už neodpovídá.

## Přepsat {#key-overrideSecret}

Zobrazí se, když upravujete existující klíč.

Semaphore uložený tajný údaj už nikdy neukáže, takže pole tajného údaje jsou prázdná a uzamčená.
Zaškrtněte políčko, chcete-li tajný údaj nahradit: zdroj, typ a pole tajného údaje se odemknou
a vše v nich se uloží znovu, takže zadejte celý tajný údaj znovu, ne jen část, která se mění. Bez
toho uložení změní pouze název klíče.

Uložení klíče také odstraní kopie repozitářů, které ho používají, uchovávané serverem Semaphore:
další úloha je naklonuje znovu.
