---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Upravit harmonogram

Harmonogram spouští úlohu šablony úlohy, aniž by kdokoli stiskl **Spustit**: opakovaně podle
výrazu cron, nebo jednou v nastaveném okamžiku. Úloha dostane parametry uložené v tomto dialogu.
Všechny časy jsou v časovém pásmu serveru Semaphore, které je uvedeno vedle časových polí: `UTC`,
pokud správce nenastaví `schedule.timezone` nebo `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

Název, pod kterým je harmonogram uveden na stránce Harmonogram. Pole je povinné. Nemění to, co
harmonogram dělá.

## Template {#scheduleForm-templateId}

Šablona úlohy, kterou harmonogram spouští. Pole je povinné. Úloha běží, jako by byla spuštěna ze
šablony, s vlastním inventářem, repozitářem, skupinou proměnných a dalšími nastaveními šablony
a s parametry níže tam, kde šablona dovoluje úloze je změnit.

## Parametry úlohy {#scheduleForm-taskParams}

Zobrazí se, když je vybrána šablona.

Parametry každé úlohy, kterou harmonogram spustí. Která pole se objeví, závisí na šabloně: její
vstupní proměnné a volby, které dovoluje úloze změnit, například větev, inventář, **Limit**,
**Tagy**, **Přeskočit tagy** nebo **CLI args**, a volby Ansiblu nebo Terraformu, OpenTofu
a Terragruntu.

Server šablonu kontroluje při každém spuštění harmonogramu: hodnota uložená zde se ignoruje,
jakmile šablona úloze už nedovoluje ji změnit.

## Zpráva {#taskParams-message}

Volitelné. Poznámka uložená s každou úlohou, kterou harmonogram spustí, a zobrazená v historii
úloh, například `Nightly deploy`. Nemění to, jak úloha běží.

## Vstupní proměnné {#taskParams-survey}

Zobrazí se, když má šablona vstupní proměnné.

Hodnoty vstupních proměnných šablony pro naplánované úlohy. K úloze se dostanou stejně, jako když
je někdo vyplní ručně. Povinná proměnná musí mít hodnotu, než lze harmonogram uložit.

Výchozí hodnoty šablony se zkopírují při otevření dialogu a uloží se s harmonogramem. Výchozí
hodnota později změněná v šabloně se do dříve uloženého harmonogramu nedostane.

Hodnota proměnné typu secret se s harmonogramem neukládá: naplánovaná úloha běží bez ní.

## Větev / tag {#taskParams-gitBranch}

Zobrazí se, když šablona dovoluje úloze přepsat větev.

Větev nebo tag repozitáře, na které se naplánovaná úloha přepne. Pole začíná větví šablony.
Server hodnotu použije pouze tehdy, dokud šablona přepsání dovoluje.

## Inventář {#taskParams-inventoryId}

Zobrazí se, když šablona dovoluje úloze vybrat inventář.

Inventář, proti kterému naplánovaná úloha běží, místo inventáře šablony. U šablony Terraform nebo
OpenTofu se pole jmenuje **Workspace**, ale server volbu použije pouze u šablon Ansible a u
ostatních ponechá workspace šablony.

## Ladění {#taskParams-debug}

Zobrazí se u šablony Ansible, která dovoluje ladění.

Spustí `ansible-playbook` s podrobnějším výstupem: `-v` opakované tolikrát, kolik nastavuje
**Úroveň ladění**.

## Úroveň ladění {#taskParams-debugLevel}

Zobrazí se u šablony Ansible, která dovoluje ladění; aktivní, když je zaškrtnuto **Ladění**.

Jak podrobný je výstup, od `1` (`-v`) do `6` (`-vvvvvv`). Výchozí je `4`.

## Zkušební běh {#taskParams-dryRun}

Zobrazí se u šablony Ansible, pokud ho šablona neskrývá.

Spustí playbook s `--check`: Ansible ohlásí, co by změnil, a nic nezmění.

## Rozdíl {#taskParams-diff}

Zobrazí se u šablony Ansible, pokud ho šablona neskrývá.

Spustí playbook s `--diff`: Ansible ukáže změny, které provádí v souborech a šablonách.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Zobrazí se u šablony Ansible, která dovoluje úloze tuto volbu změnit.

Přeskočí instalaci rolí a kolekcí ze souborů `requirements.yml` před spuštěním playbooku.
Zaškrtněte, když jsou požadavky už nainstalované a úloha se nemá připojovat k Ansible Galaxy.

## Plan {#taskParams-plan}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Spustí pouze `plan`. Úloha po plánu skončí a nic neaplikuje.

## Destroy {#taskParams-destroy}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Přidá `-destroy` k plánu a aplikaci: úloha prostředky workspace zničí, místo aby je vytvořila
nebo aktualizovala.

## Auto Approve {#taskParams-autoApprove}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Aplikuje plán bez čekání na potvrzení. Funguje pouze tehdy, když šablona dovoluje automatické
schválení; šablona nastavená na schvalování každé úlohy plán aplikuje v obou případech.

Bez automatického schválení se naplánovaná úloha po plánu zastaví a čeká, dokud ji někdo na
stránce úlohy nepotvrdí nebo nezamítne.

## Upgrade {#taskParams-upgrade}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Spustí `init` s `-upgrade`: poskytovatelé a moduly se aktualizují na nejnovější verze, které
konfigurace dovoluje.

## Reconfigure {#taskParams-reconfigure}

Zobrazí se u šablony Terraform, OpenTofu nebo Terragrunt.

Spustí `init` s `-reconfigure` místo `-migrate-state`: backend se nastaví znovu a existující stav
se nemigruje.

## Run at {#scheduleForm-runAt}

Zobrazí se u harmonogramu, který se spouští jednou.

Datum a čas spuštění úlohy v časovém pásmu serveru uvedeném na konci pole. Nový harmonogram
začíná nejbližší celou hodinou. Okamžik musí být v budoucnosti: server odmítne čas, který už
uplynul.

Po spuštění úlohy se harmonogram vypne, nebo odstraní, když je zaškrtnuto **Delete after run**.
Když Semaphore v tu chvíli neběží, úloha se nespustí později: při dalším startu se harmonogram
vypne, nebo odstraní, bez spuštění.

## Delete after run {#scheduleForm-deleteAfterRun}

Zobrazí se u harmonogramu, který se spouští jednou.

Odstraní harmonogram poté, co se jeho úloha spustila. Ponechte nezaškrtnuté, chcete-li
harmonogram zachovat, vypnutý, abyste mohli nastavit nový čas a znovu ho zapnout.

## Show cron format {#scheduleForm-rawCron}

Zobrazí se u harmonogramu cron.

Přepíná mezi psaním výrazu cron (zapnuto) a jeho sestavením pomocí **Timing** a časových
zaškrtávacích políček (vypnuto). Oba způsoby upravují stejný výraz. Prohlížeč si volbu pamatuje.

Přepínač je uzamčen v zapnuté poloze, když výraz nelze zobrazit jako zaškrtávací políčka: když je
neplatný, nebo používá tvar jako `@daily` nebo `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Zobrazí se u harmonogramu cron, když je zapnuto **Show cron format**.

Výraz cron v pěti polích: minuta, hodina, den v měsíci, měsíc, den v týdnu (`0` je neděle). Pole
je povinné a server výraz kontroluje během psaní. Nový harmonogram začíná výrazem `* * * * *`,
každou minutu.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

Server přijímá také `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` a `@every 1h30m`
a předponu jako `CRON_TZ=Europe/Berlin` pro jiné časové pásmo. Náhled níže u nich nedokáže
ukázat příští spuštění.

## Timing {#scheduleForm-timing}

Zobrazí se u harmonogramu cron, když je **Show cron format** vypnuto.

Jak často úloha běží. Zaškrtávací políčka níže se podle toho mění:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** a **Minutes**.
- **Weekly**: **Weekdays**, **Hours** a **Minutes**.
- **Monthly**: **Days**, **Hours** a **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** a **Minutes**.

## Months {#scheduleForm-months}

Zobrazí se, když je **Timing** nastaveno na **Yearly**.

Měsíce, ve kterých úloha běží. Zaškrtněte několik, chcete-li ji spouštět v každém z nich. Když
není zaškrtnutý žádný, úloha běží v každém měsíci.

## Weekdays {#scheduleForm-weekdays}

Zobrazí se, když je **Timing** nastaveno na **Weekly**.

Dny v týdnu, ve které úloha běží. Když není zaškrtnutý žádný, úloha běží každý den.

## Days {#scheduleForm-days}

Zobrazí se, když je **Timing** nastaveno na **Monthly** nebo **Yearly**.

Dny v měsíci, ve které úloha běží. Měsíc bez tohoto dne se přeskočí: úloha nastavená na 31. den
v dubnu neběží. Když není zaškrtnutý žádný, úloha běží každý den.

## Hours {#scheduleForm-hours}

Zobrazí se, když je **Timing** nastaveno na **Daily**, **Weekly**, **Monthly** nebo **Yearly**.

Hodiny, ve které úloha běží, v časovém pásmu uvedeném vpravo. Když není zaškrtnutá žádná, úloha
běží každou hodinu.

## Minutes {#scheduleForm-minutes}

Zobrazí se u harmonogramu cron, když je **Show cron format** vypnuto.

Minuty hodiny, ve které úloha běží, po pěti minutách. Když není zaškrtnutá žádná, úloha běží
každou minutu: pro úlohu, která má běžet jednou za hodinu nebo jednou denně, zaškrtněte alespoň
jednu, například `:00`.

## Next run time {#scheduleForm-nextRun}

Kdy by se spustila další úloha podle harmonogramu, jak je ve formuláři: v časovém pásmu serveru
a v místním čase vašeho prohlížeče. Pomlčka znamená, že výraz je neplatný, nebo je takový, který
náhled nedokáže přečíst, například `@daily`.

## Povoleno {#scheduleForm-active}

Zda harmonogram spouští úlohy. Zrušte zaškrtnutí, chcete-li harmonogram ponechat, aniž by se
spouštěl. Přepínač v seznamu harmonogramů dělá totéž. Nový harmonogram je povolený.
