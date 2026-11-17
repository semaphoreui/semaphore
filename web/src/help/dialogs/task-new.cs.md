---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nová úloha

Spustí úlohu šablony: jeden běh jejího playbooku, kódu Terraformu nebo skriptu. Šablona rozhoduje
o tom, která pole tento dialog ukazuje, a každá část níže uvádí, kdy je její pole k dispozici.
Tlačítko dole je **Spuštění**, **Sestavení** nebo **Nasazení**, podle typu šablony.

Když má šablona vstupní proměnné, jejich otázky jsou pole za polem **Zpráva**. Pole označené `*`
je povinné, pole typu Integer přijímá pouze číslice a výchozí odpovědi jsou vyplněné. Semaphore
předá odpovědi úloze jako proměnné: jako doplňkové proměnné pro Ansible, `-var` pro Terraform,
argumenty `name=value` pro skript, nebo jako proměnné prostředí u otázky takto nastavené.
Odpověď nahradí proměnnou se stejným názvem ve skupinách proměnných šablony. Tajná odpověď je při
psaní skrytá a s úlohou se neukládá.

## Commit {#newTask-commit}

Zobrazí se, když znovu spouštíte dřívější úlohu a ta úloha běžela na známém commitu.

Úloha se přepne na tento commit repozitáře místo nejnovějšího commitu větve. Zavřete políčko,
chcete-li spustit nejnovější commit.

Commit se použije pouze tehdy, když šablona dovoluje přepsat větev, a nikdy u repozitáře
označeného jako monorepo. Jinak úloha spustí nejnovější commit.

## Verze sestavení {#newTask-buildTaskId}

Zobrazí se, když je typ šablony **Nasazení**.

Sestavení, které se má nasadit: jedna z posledních 20 úspěšných úloh šablony sestavení, uvedených
podle verze a zprávy. Vybrané je nejnovější. Pole je povinné.

Verze sestavení se dostane do úlohy: jako `semaphore_vars.task_details.incoming_version`
v Ansiblu, jako proměnná prostředí `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` ve skriptu. Seznam
je prázdný, když šablonu sestavení nesmíte vidět.

## Zpráva {#newTask-message}

Volitelné. Poznámka, proč úloha běží, například `Hotfix for the login page`. Zobrazuje se
s úlohou v historii úloh a na běhu nic nemění.

Playbook ji může číst jako `semaphore_vars.task_details.message` a skript jako proměnnou
prostředí `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Větev / tag {#newTask-gitBranch}

Zobrazí se, když šablona dovoluje přepsat větev, pokud repozitář není monorepo.

Větev nebo tag repozitáře, na které se úloha přepne. Pole začíná větví šablony. Když je prázdné,
úloha použije větev šablony, nebo větev repozitáře, když šablona žádnou nemá. Název, který není
platným názvem větve Git, je odmítnut.

## Inventář {#newTask-inventoryId}

Zobrazí se, když šablona dovoluje v úloze vybrat inventář a vy smíte vidět projekt. U šablony
Terraform, OpenTofu nebo Terragrunt má pole popisek **Workspace**, ale server volbu použije pouze u
šablon Ansible: úloha Terraform, OpenTofu nebo Terragrunt běží ve workspace šablony, ať zde vyberete
cokoli.

Inventář, proti kterému úloha běží. Pole začíná inventářem šablony a nabízí inventáře projektu
určené pro aplikaci šablony. Když vybraný inventář v okamžiku spuštění úlohy už neexistuje, úloha
běží proti inventáři šablony.

## Limit {#taskParams-limit}

Zobrazí se u šablony Ansible, která dovoluje přepsat limit.

Hosté nebo skupiny inventáře, na kterých playbook běží, jeden vzor na štítek, například `web`
nebo `10.0.0.11`. Semaphore je předá jako `--limit` spojené čárkami. Seznam začíná limitem
šablony. Odstraňte všechny štítky, chcete-li spustit na všech hostech inventáře.

## Tagy {#taskParams-tags}

Zobrazí se u šablony Ansible, která dovoluje přepsat tagy.

Spustí se pouze úkoly a role playbooku s některým z těchto tagů: Semaphore je předá jako
`--tags`, jeden tag na štítek. Seznam začíná tagy šablony. Odstraňte všechny štítky, chcete-li
spustit celý playbook.

## Přeskočit tagy {#taskParams-skipTags}

Zobrazí se u šablony Ansible, která dovoluje přepsat přeskočené tagy.

Úkoly a role playbooku s některým z těchto tagů se nespustí: Semaphore je předá jako
`--skip-tags`, jeden tag na štítek. Seznam začíná přeskočenými tagy šablony.

## Ladění {#taskParams-debug}

Zobrazí se u šablony Ansible, která dovoluje ladění.

Spustí `ansible-playbook` s příznakem podrobnosti výpisu, takže log úlohy ukazuje víc
podrobností. Popisek ukazuje příznak: `-v` tolikrát, kolik udává **Úroveň ladění**.

## Úroveň ladění {#taskParams-debugLevel}

Zobrazí se spolu s **Ladění** a je aktivní, když je **Ladění** zaškrtnuto.

Jak podrobný je log, od 1 (`-v`) do 6 (`-vvvvvv`). Výchozí je 4 (`-vvvv`), což ukazuje také
podrobnosti připojení k hostům.

## Zkušební běh {#taskParams-dryRun}

Zobrazí se u šablony Ansible, pokud ho šablona neskrývá.

Spustí playbook s `--check`: Ansible ohlásí, co by na hostech změnil, a nic nezmění.

## Rozdíl {#taskParams-diff}

Zobrazí se u šablony Ansible, pokud ho šablona neskrývá.

Spustí playbook s `--diff`: log ukazuje změny provedené v souborech a šablonách na hostech.
Zaškrtněte také **Zkušební běh**, chcete-li změny vidět, aniž by se provedly.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Zobrazí se u šablony Ansible, která dovoluje přepsat instalaci z Galaxy.

Před spuštěním playbooku Semaphore nainstaluje role a kolekce uvedené v souborech
`requirements.yml` repozitáře, když se některý soubor od poslední instalace změnil. Zaškrtněte
políčko, chcete-li tento krok pro tuto úlohu přeskočit.

Políčko je na začátku nezaškrtnuté bez ohledu na nastavení šablony a nezaškrtnuté políčko
požadavky nainstaluje.

## Plan {#taskParams-plan}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Spustí `plan` a tím úlohu ukončí: nic se neaplikuje. Použijte to, chcete-li si změny nejprve
prohlédnout. Bez toho úloha pokračuje aplikací plánu, když obsahuje změny.

## Destroy {#taskParams-destroy}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Přidá `-destroy` k `plan` a `apply`: úloha naplánuje a pak aplikuje odstranění všech prostředků
ve stavu workspace. Zaškrtněte také **Plan**, chcete-li vidět, co by bylo zničeno, aniž by se to
zničilo.

## Auto Approve {#taskParams-autoApprove}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Aplikuje plán, jakmile je vytvořen. Bez toho se úloha po plánu zastaví a čeká, až ji někdo na
stránce úlohy potvrdí nebo zamítne.

Políčko má účinek pouze tehdy, když šablona dovoluje automatické schválení. Šablona nastavená na
schvalování vždy aplikuje bez dotazu, ať je políčko nastaveno jakkoli.

## Upgrade {#taskParams-upgrade}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Spustí `init` s `-upgrade`: poskytovatelé a moduly se aktualizují na nejnovější verze, které
jejich omezení verzí dovolují.

## Reconfigure {#taskParams-reconfigure}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Spustí `init` s `-reconfigure`: backend se nastaví znovu a uložená konfigurace backendu se
ignoruje. Bez toho Semaphore spouští `init -migrate-state`, které přesune existující stav do
změněného backendu.

## CLI args {#newTask-args}

Zobrazí se, když šablona dovoluje přepsat argumenty CLI.

Doplňkové argumenty pro příkazový řádek úlohy, jeden na štítek: `--forks=10` je jeden štítek,
`--forks` a `10` jsou dva. Kliknutím na štítek ho upravíte, přetažením změníte pořadí a posledním
štítkem nový přidáte.

Seznam začíná argumenty šablony. Semaphore vždy předá argumenty šablony a tento seznam přidá za
ně, takže jakmile seznam změníte, argumenty šablony, které v něm zůstaly, se předají dvakrát.
Odstraňte je a ponechte jen argumenty, které přidáváte.
