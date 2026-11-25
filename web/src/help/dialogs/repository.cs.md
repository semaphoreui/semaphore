---
form: src/components/RepositoryForm.vue
---

# Upravit Repository

Repozitář je místo, odkud pocházejí playbooky, kód Terraformu nebo skripty šablony úlohy: repozitář
Git, nebo adresář na stroji, který úlohy spouští. Před každou úlohou Semaphore naklonuje repozitář
Git, nebo aktualizuje kopii, kterou naklonoval dříve, přepne se na větev a spustí úlohu v této
kopii.

## Název {#repository-name}

Název, pod kterým je repozitář uveden na stránce Repozitáře a ve výběru repozitáře v šabloně úlohy
a v inventáři. Pole je povinné.

## URL nebo cesta {#repository-gitUrl}

Kde je kód. Pole je povinné. Štítky pod polem ukazují druh adresy, kterou jste zadali, a kliknutím
na štítek se adresa přepíše na tento druh.

- **ssh**: `git@github.com:org/repo.git` nebo `ssh://git@github.com/org/repo.git`. Git se
  přihlásí uživatelem z adresy a klíčem **SSH klíč** vybraným v poli **Přístupový klíč**.
- **https** a **http**: `https://github.com/org/repo.git`. Soukromý repozitář potřebuje klíč
  **Přihlášení heslem**; veřejný potřebuje **Žádný**. Vyhněte se `http`: přihlašovací jméno
  a heslo pak putují nešifrovaně.
- **git**: `git://host/org/repo.git`, protokol Git pouze pro čtení, který nemá přihlášení.
- **file**: `file:///srv/git/repo.git`, repozitář Git na disku stroje, který úlohu spouští,
  klonovaný stejně jako vzdálený.
- **absolutní cesta**: absolutní cesta jako `/srv/playbooks` nebo `D:\playbooks`. Semaphore ji
  neklonuje: úloha běží přímo v tomto adresáři, který musí existovat na stroji, který úlohu
  spouští (na serveru Semaphore nebo na runneru). Větev se na ni nevztahuje.

Změna adresy odstraní kopie, které server Semaphore naklonoval z původní adresy.

## Větev / tag {#repository-gitBranch}

Větev, kterou Semaphore klonuje a udržuje aktuální, například `main`. Pole je povinné, kromě
absolutní cesty, pro kterou je uzamčené.

Šablona úlohy může uvést vlastní větev, která má přednost před touto, a šablona může svým úlohám
dovolit vybrat větev při spuštění. Funguje i tag, ale tag nelze aktualizovat pomocí pull, takže
Semaphore před každou úlohou naklonuje repozitář znovu.

## Single checkout for all templates {#repository-monorepo}

Ve výchozím stavu vypnuto: každá šablona úlohy dostane vlastní kopii repozitáře. Zaškrtněte
políčko, chcete-li repozitář naklonovat jednou a sdílet tuto kopii mezi všemi šablonami, které ho
používají, u repozitáře příliš velkého na to, aby měla každá šablona vlastní kopii.

Sdílená kopie zůstává na větvi tohoto repozitáře: šablony, které ji používají, nemohou mít vlastní
větev a jejich úlohy nemohou vybrat jinou větev ani commit. U absolutní cesty je uzamčeno.

## Přístupový klíč {#repository-sshKeyId}

Klíč z Úložiště klíčů, kterým se Git k repozitáři přihlašuje. Pole je povinné. Nabízí klíče tří
typů; který zvolit, závisí na URL:

- **SSH klíč** pro URL `ssh`. Semaphore načte soukromý klíč do SSH agenta po dobu klonování nebo
  aktualizace. Uživatel je ten z URL, jako `git` v `git@github.com:org/repo.git`.
- **Přihlášení heslem** pro URL `https`: přihlašovací jméno a heslo, nebo přístupový token
  serveru Git. S prázdným přihlašovacím jménem se heslo pošle samo jako uživatelské jméno, což
  některé servery u tokenu očekávají. Stejné přihlašovací údaje dostane `ansible-galaxy` pro role
  a kolekce na stejném serveru.
- **Žádný** pro veřejný repozitář, URL `git` nebo `file`, nebo absolutní cestu.

Klíč, který neodpovídá URL, nedá Gitu žádné přihlašovací údaje a soukromý repozitář se pak
nepodaří naklonovat.
