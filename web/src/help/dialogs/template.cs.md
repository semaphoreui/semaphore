---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Upravit šablonu 'Ansible Playbook'

Šablona úlohy říká, co a jak Semaphore spouští: repozitář a soubor v něm, inventář a proměnné
a které volby smí uživatel při spuštění úlohy změnit. Každá úloha se spouští ze šablony. Pole
dialogu závisí na aplikaci šablony (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python,
PowerShell a další): pole, které mají jen některé aplikace, to uvádí. Změny platí pro úlohy
spuštěné po stisknutí **Uložit**.

## Typ šablony {#templateForm-typeTabs}

Karty nahoře určují, jakého druhu šablona je:

- **Úloha**: spouští kód. Většina šablon je tohoto druhu.
- **Sestavení**: vytváří artefakt. Každá úloha sestavení dostane verzi, počítanou od **Počáteční
  verze**.
- **Nasazení**: nasazuje artefakt vytvořený sestavením. Při spuštění vyberete úlohu sestavení,
  jejíž verzi nasazuje.

Úloha sestavení nebo nasazení dostane ve svém prostředí proměnné `SEMAPHORE_TASK_TYPE`,
`SEMAPHORE_TASK_TARGET_VERSION` (sestavení) a `SEMAPHORE_TASK_INCOMING_VERSION` (nasazení).
Ansible dostane stejné hodnoty v doplňkové proměnné `semaphore_vars.task_details`.

## Počáteční verze {#templateForm-startVersion}

Zobrazí se, když je typ **Sestavení**. Povinné.

Verze první úlohy sestavení, například `1.0.0` nebo `v0.1`. Každé další sestavení zvýší poslední
číslo předchozí verze o jedna a zachová text kolem něj i jeho doplnění nulami: `1.0.0`, `1.0.1`,
`1.0.2`. Chcete-li poskočit dopředu, zvyšte zde číslo nad poslední sestavení, například na
`2.0.0`; verze bez čísla se nikdy nemění.

## Šablona sestavení {#templateForm-buildTemplateId}

Zobrazí se, když je typ **Nasazení**. Povinné.

Šablona, jejíž úlohy vytvářejí artefakty, které tato šablona nasazuje. Seznam obsahuje šablony
sestavení projektu a pod nimi šablony nasazení: nasazení může následovat po jiném nasazení
a dostane pak verzi, kterou toto nasazení dostalo od svého sestavení.

## Automatické spuštění {#templateForm-autorun}

Zobrazí se, když je typ **Nasazení**.

Spustí úlohu této šablony sama pokaždé, když úloha **Šablona sestavení** uspěje, s verzí, kterou
tato úloha vytvořila. Neúspěšné sestavení nic nespustí.

## Název {#templateForm-name}

Název, pod kterým je šablona uvedena na stránce Šablony úloh. Pole je povinné.

## Repozitář {#templateForm-repositoryId}

Repozitář projektu, který obsahuje kód ke spuštění. Pole je povinné. Před každou úlohou Semaphore
aktualizuje svou kopii repozitáře a spustí kód z ní.

Popisek vpravo od pole ukazuje větev, na kterou se úlohy přepínají. Kliknutím na něj otevřete pole
**Větev / tag**.

## Větev / tag {#templateForm-gitBranch}

Zobrazí se po kliknutí na popisek větve vpravo od pole **Repozitář**.

Větev nebo tag, na které se úlohy této šablony přepínají, místo větve nastavené v repozitáři.
Ponechte pole prázdné, chcete-li použít větev repozitáře. Když Semaphore dokáže načíst větve
repozitáře, pole je nabízí.

Pole je neaktivní, když je repozitář monorepo: všechny šablony sdílejí jednu jeho kopii, na větvi
repozitáře.

## Cesta k souboru playbooku {#templateForm-playbook}

Co se má spustit, jako cesta od kořene repozitáře, například `deploy/site.yml`. Cesta nesmí být
absolutní a nesmí vést mimo repozitář. Seznam nabízí soubory repozitáře; můžete také napsat cestu,
která v seznamu není. Tlačítko vpravo od pole seznam znovu načte.

Popisek a význam se mění podle aplikace:

- **Ansible**: playbook. Je povinný; seznam nabízí soubory `.yml` a `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, adresář s kódem. Ponechte
  pole prázdné, chcete-li spustit kód v kořeni repozitáře. Napište `/`, chcete-li zobrazit obsah
  adresáře.
- **Bash**, **Python**, **PowerShell** a další aplikace: **Script Filename**, skript ke spuštění.
  Je povinný. Skript běží v kořeni repozitáře.

## Different working directory {#templateForm-workingDirectoryToggle}

Zobrazí se u šablon Ansible.

Spustí `ansible-playbook` v jiném adresáři než v kořeni repozitáře. Zaškrtnutím zobrazíte pole
**Working directory**. Zrušením zaškrtnutí se znovu spouští v kořeni.

## Working directory {#templateForm-workingDirectory}

Zobrazí se, když je zaškrtnuto **Different working directory**. Povinné.

Adresář, ve kterém běží `ansible-playbook` a `ansible-galaxy`, jako cesta od kořene repozitáře,
například `ansible`. Použijte ho, když je `ansible.cfg` projektu v podadresáři, protože Ansible
čte tento soubor z aktuálního adresáře. Cesta k playbooku zůstává relativní ke kořeni
repozitáře. Cesta nesmí být absolutní a nesmí vést mimo repozitář.

## Inventář {#templateForm-inventoryId}

Zobrazí se u šablon Ansible, Terraform, OpenTofu a Terragrunt.

- **Ansible**: inventář, proti kterému playbook běží, s přihlašovacími údaji pro hosty. Pole je
  povinné.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, workspace, ve kterém kód běží.
  Semaphore ho vybere před plánem a vytvoří ho, když neexistuje. Nová šablona uložená bez
  workspace dostane workspace s názvem `default`.

Seznam obsahuje inventáře projektu určené pro aplikaci šablony.

## Skupiny proměnných {#templateForm-environmentIds}

Volitelné. Skupiny proměnných, jejichž proměnné, proměnné prostředí a tajné údaje dostane každá
úloha této šablony. Můžete jich vybrat několik: sloučí se v pořadí pole a pozdější skupina vyhraje
nad dřívější, když obě nastavují stejný název.

Jak se proměnné dostanou ke kódu, závisí na aplikaci: doplňkové proměnné pro Ansible, `-var` pro
Terraform a OpenTofu, argumenty `name=value` pro skript. Proměnné prostředí se nastavují pro
každou aplikaci.

## Zobrazení {#templateForm-viewId}

Volitelné. Zobrazení, ve kterém je šablona uvedena: zobrazení jsou karty stránky Šablony úloh,
které seskupují šablony projektu. Ponechte pole prázdné, chcete-li šablonu nezařadit do žádného
zobrazení.

## Tag runneru {#templateForm-runnerTag}

Zobrazí se, když jsou k dispozici runnery projektu.

Úlohy této šablony běží pouze na runneru, který má tento tag, i když by je server mohl spustit
sám. Tag přepisuje tag runneru inventáře. Ponechte pole prázdné, chcete-li použít tag inventáře,
nebo žádný.

## Executor image {#templateForm-executorImage}

Image kontejneru, který runner spouští pro úlohy této šablony, například
`semaphoreui/job:latest`. Ponechte pole prázdné, chcete-li použít image z konfigurace runneru.

Úlohy v kontejneru spouštějí pouze executory Docker a Kubernetes, takže pole je aktivní jen
tehdy, když je jeden z nich k dispozici. Úloha, která běží bez kontejneru, ho ignoruje.

## Vstupní proměnné {#surveyVars}

Hodnoty, na které se uživatel ptá při spuštění úlohy této šablony. Kliknutím na **+** přidáte
proměnnou, kliknutím na štítek ji změníte, křížkem na štítku ji odstraníte; přetažením štítků
nastavíte pořadí, ve kterém se otázky kladou.

Odpovědi se ke kódu dostanou jako proměnné skupiny proměnných, nebo jako proměnné prostředí,
podle nastavení každé proměnné. Vyhrávají nad skupinou proměnných, která nastavuje stejný název.

## Povolit paralelní úlohy {#templateForm-allowParallelTasks}

Dovolí úlohám této šablony běžet současně. Bez toho úloha šablony čeká ve frontě, dokud běžící
úloha neskončí. Limity paralelních úloh projektu a serveru platí v obou případech.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Zobrazí se, když správce zapnul pro server vydávání JWT.

Dá každé úloze této šablony JSON Web Token podepsaný Semaphore, v proměnné prostředí
`SEMAPHORE_JWT`. Kód může token ukázat jiné službě, aby prokázal, o kterou úlohu jde: token uvádí
úlohu, projekt, šablonu a uživatele, který ji spustil. Služba ověří podpis pomocí veřejných klíčů
na adrese `/.well-known/jwks.json` serveru Semaphore.

## JWT audience {#templateForm-jwtAudience}

Zobrazí se, když je zaškrtnuto **Issue JWT to task runner**. Volitelné.

Služby, pro které je token určen, zapsané do jeho claimu `aud`, například
`https://vault.example.com`. Po každé hodnotě stiskněte Enter. Nejvýše 32 hodnot; pro token bez
audience ponechte pole prázdné.

## JWT TTL {#templateForm-jwtTtl}

Zobrazí se, když je zaškrtnuto **Issue JWT to task runner**. Volitelné.

Jak dlouho token zůstává platný od začátku úlohy, jako doba trvání, například `30s`, `15m` nebo
`1h`. Ponechte pole prázdné, chcete-li použít výchozí hodnotu serveru, jednu hodinu, pokud správce
nenastavil jinou. Hodnota nesmí překročit maximum serveru, 24 hodin, pokud není nastaveno jinak.

## Automaticky spustit úlohu, pokud byl nalezen nový commit v Gitu {#templateForm-cronVisible}

Spustí úlohu této šablony, když se v repozitáři objeví nový commit. Semaphore se ve zvoleném
intervalu ptá repozitáře na jeho poslední commit a spustí úlohu pouze tehdy, když se commit od
poslední kontroly změnil.

Zaškrtnutím políčka a uložením se vytvoří harmonogram kontroly; zrušením zaškrtnutí a uložením se
tento harmonogram odstraní. Ostatní harmonogramy šablony zůstanou zachovány.

## Repozitář {#templateForm-cronRepositoryId}

Zobrazí se, když je zaškrtnuto **Automaticky spustit úlohu, pokud byl nalezen nový commit
v Gitu**. Povinné.

Repozitář, ve kterém se sledují nové commity. Semaphore sleduje větev nastavenou v tomto
repozitáři, ne **Větev / tag** šablony.

## Interval kontroly {#templateForm-cronFormat}

Zobrazí se, když je zaškrtnuto **Automaticky spustit úlohu, pokud byl nalezen nový commit
v Gitu**.

Jak často Semaphore kontroluje repozitář, zda nemá nový commit: od každé minuty po jednou denně.
Výchozí je každou minutu.

## Potlačit oznámení o úspěchu {#templateForm-suppressSuccessAlerts}

Zabrání úlohám této šablony, které uspějí, odesílat upozornění. Upozornění na neúspěšné úlohy se
odesílají dál. Má účinek pouze tehdy, když jsou pro projekt zapnutá upozornění.

## Potlačit oznámení o chybě {#templateForm-suppressErrorAlerts}

Zabrání úlohám této šablony, které selžou, odesílat upozornění. Upozornění na úspěšné úlohy se
odesílají dál. Má účinek pouze tehdy, když jsou pro projekt zapnutá upozornění.

## CLI args {#templateForm-args}

Argumenty přidané k příkazu každé úlohy. Každý štítek je jeden argument, předaný tak, jak je, bez
shellu: `--forks=10` dejte do jednoho štítku, nebo `--forks` a `10` do dvou. Přetažením štítků
změníte jejich pořadí. Tlačítko vpravo od seznamu otevře argumenty jako pole JSON:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: přidají se k `ansible-playbook`, před playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: přidají se k `plan` a `apply`, ne k `init`.
- Skript: přidají se za název skriptu.

## Argumenty CLI {#templateForm-allowOverrideArgsInTask}

V části **Parametry při spuštění**. Dovolí uživateli přidat při spuštění úlohy vlastní argumenty.
Přidají se za argumenty šablony.

## Větev {#templateForm-allowOverrideBranchInTask}

V části **Parametry při spuštění**. Dovolí uživateli při spuštění úlohy vybrat jinou větev, tag
nebo commit místo větve šablony. Bez toho se úloha vždy přepne na větev šablony.

Políčko je neaktivní, když je repozitář monorepo.

## Inventář {#templateForm-allowOverrideInventory}

V části **Parametry při spuštění**. Zobrazí se u šablon Ansible, Terraform, OpenTofu
a Terragrunt.

Dovolí uživateli při spuštění úlohy vybrat jiný inventář. Inventář tam vybraný používají pouze
šablony Ansible: úloha Terraform, OpenTofu nebo Terragrunt běží ve workspace šablony, i když je
políčko zaškrtnuté.

## Limit {#templateForm-limit}

Zobrazí se u šablon Ansible.

Hosté nebo skupiny z inventáře, na kterých se playbook spustí, předané jako `--limit`. Každý
štítek je jeden vzor, například `web` nebo `db01`; ponechte seznam prázdný, chcete-li spustit na
všech hostech.

## Tagy {#templateForm-tags}

Zobrazí se u šablon Ansible.

Spustí pouze úkoly a role s těmito tagy, předané jako `--tags`. Ponechte seznam prázdný, chcete-li
spustit všechno.

## Přeskočit tagy {#templateForm-skipTags}

Zobrazí se u šablon Ansible.

Přeskočí úkoly a role s těmito tagy, předané jako `--skip-tags`.

## Vaulty {#templateVaults}

Zobrazí se u šablon Ansible.

Hesla Ansible Vault, která playbook potřebuje k dešifrování svých souborů. Kliknutím na **+**
jedno přidáte, kliknutím na štítek ho změníte, křížkem na štítku ho odstraníte. Každý vault má
název, své vault ID, a poskytuje heslo z klíče Úložiště klíčů nebo z klientského skriptu
repozitáře.

## Automaticky schválit {#templateForm-autoApprove}

Zobrazí se u šablon Terraform, OpenTofu a Terragrunt.

Aplikuje plán bez dotazu. Bez toho se úloha po plánu zastaví a čeká, dokud ji uživatel nepotvrdí
nebo nezamítne; plán bez změn ukončí úlohu, aniž by bylo co aplikovat.

## Přepsat nastavení backendu {#templateForm-overrideBackend}

Zobrazí se u šablon Terraform, OpenTofu a Terragrunt, když Semaphore může uchovávat stav
Terraformu.

Uchovává stav kódu v Semaphore. Před `init` Semaphore zapíše do adresáře s kódem soubor, který
nastavuje backend `http`, a po úloze ho odstraní:

```hcl
terraform {
  backend "http" {
  }
}
```

## Název souboru backendu {#templateForm-backendFilename}

Zobrazí se spolu s **Přepsat nastavení backendu**; aktivní, když je toto políčko zaškrtnuté.
Volitelné.

Název souboru, do kterého Semaphore zapíše backend, například `semaphore_backend.tf`. Výchozí je
`backend.tf`. Název musí končit na `.tf` a obsahovat pouze písmena, číslice, `_`, `-` a `.`.
Zvolte název, který nemá žádný soubor vašeho kódu: Semaphore tento soubor přepíše a po úloze ho
odstraní.

## Limit {#templateForm-allowOverrideLimit}

V části **Parametry Ansible při spuštění**. Dovolí uživateli při spuštění úlohy nastavit hosty, na
kterých se spustí, místo **Limit** šablony.

## Tagy {#templateForm-allowOverrideTags}

V části **Parametry Ansible při spuštění**. Dovolí uživateli při spuštění úlohy nastavit tagy,
místo **Tagy** šablony.

## Přeskočit tagy {#templateForm-allowOverrideSkipTags}

V části **Parametry Ansible při spuštění**. Dovolí uživateli při spuštění úlohy nastavit tagy,
které se mají přeskočit, místo **Přeskočit tagy** šablony.

## Ladění {#templateForm-allowDebug}

V části **Parametry Ansible při spuštění**. Dovolí uživateli při spuštění úlohy zapnout podrobný
výstup, od `-v` do `-vvvvvv`.

## Zkušební běh {#templateForm-showDryRun}

V části **Parametry Ansible při spuštění**. Ve výchozím stavu zaškrtnuto. Nabízí při spuštění
úlohy volbu **Zkušební běh**, která spustí playbook s `--check`: Ansible ohlásí, co by změnil,
a nic nezmění. Zrušením zaškrtnutí volbu odeberete.

## Rozdíl {#templateForm-showDiff}

V části **Parametry Ansible při spuštění**. Ve výchozím stavu zaškrtnuto. Nabízí při spuštění
úlohy volbu **Rozdíl**, která spustí playbook s `--diff` a ukáže změny provedené v souborech.
Zrušením zaškrtnutí volbu odeberete.

## Automaticky schválit {#templateForm-allowAutoApprove}

V části parametrů Terraform, OpenTofu a Terragrunt při spuštění. Dovolí uživateli při spuštění
úlohy zaškrtnout **Auto Approve**, aby se plán aplikoval bez potvrzení. Dialog úlohy tuto volbu
nabízí v každém případě, ale bez tohoto políčka ji Semaphore ignoruje: úloha po plánu čeká na
potvrzení, pokud není zaškrtnuto **Automaticky schválit** šablony.

## Galaxy install options {#templateForm-galaxy}

Zobrazí se u šablon Ansible. Kliknutím zobrazíte volby instalace z Galaxy; popisek počítá volby,
které se liší od výchozích.

Před spuštěním playbooku Semaphore pomocí `ansible-galaxy` nainstaluje role a kolekce, které
repozitář požaduje. Hledá `requirements.yml`, `roles/requirements.yml`
a `collections/requirements.yml` vedle playbooku a v kořeni repozitáře a instalaci spustí pouze
tehdy, když se některý soubor od poslední úlohy změnil.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

V části **Galaxy install options**. Přeskočí instalaci z Galaxy: playbook běží s rolemi
a kolekcemi nainstalovanými dříve. Použijte to, když se požadavky instalují jinak nebo je
instalace pomalá a není potřeba.

## Role install args {#templateForm-galaxyRoleArgs}

V části **Galaxy install options**. Argumenty přidané k `ansible-galaxy role install`, jeden na
štítek, například `--ignore-certs` nebo `--server=https://galaxy.example.com`. Semaphore přijímá
pouze volby instalace, které je bezpečné předat: neznámá volba, `--token`, `--api-key`, `-p`
nebo `--roles-path` způsobí, že šablonu nelze uložit.

## Collection install args {#templateForm-galaxyCollectionArgs}

V části **Galaxy install options**. Argumenty přidané k `ansible-galaxy collection install`,
jeden na štítek, například `--pre` nebo `--upgrade`. Platí stejné pravidlo jako u **Role install
args**: `--token`, `--api-key`, `-p` a `--collections-path` nejsou přijímány.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

V části **Galaxy install options**, v **Prompts**. Dovolí uživateli při spuštění úlohy zvolit, zda
přeskočit instalaci z Galaxy. Volba úlohy vyhrává nad **Skip Galaxy install** šablony.
