---
form: src/components/RunnerForm.vue
---

# Upravit runner

Runner je samostatný proces `semaphore runner`, který spouští úlohy na vlastním stroji místo na
serveru Semaphore. Úloha jde na runner, když je server nastaven tak, aby spouštěl úlohy na
runnerech, nebo když má šablona úlohy či její inventář tag runneru. Server pak vybere runner,
který je povolený, registrovaný, online a nedosáhl svého limitu paralelních úloh: nejprve runnery
projektu, potom globální runnery, v rámci každé skupiny v náhodném pořadí. Když jsou všechny
odpovídající runnery offline nebo zaneprázdněné, úloha čeká ve frontě; když neodpovídá vůbec
žádný runner, úloha selže.

## Název {#runner-name}

Povinné. Název, pod kterým je runner uveden na stránce Runnery, například host, na kterém běží.
Nemění to, které úlohy runner dostává.

## Tags {#runner-tags}

K dispozici v Semaphore Pro; v ostatních edicích je pole uzamčené.

Tagy, které vybírají úlohy tohoto runneru. Úloha s tagem runneru jde pouze na runnery, které tento
tag mají; tag šablony úlohy má přednost před tagem inventáře. Napište tag a stiskněte Enter, nebo
vyberte tag, který používají jiné runnery. Prázdné a opakované tagy se při uložení odstraní.

Runner projektu potřebuje alespoň jeden tag, pokud není zaškrtnuto **Is default**.

## Is default {#runner-isDefault}

Úloha bez tagu runneru jde pouze na runnery označené jako výchozí. Zrušte zaškrtnutí u runneru,
který má přebírat pouze úlohy s některým ze svých tagů.

U nového globálního runneru je políčko zaškrtnuté a u nového runneru projektu nezaškrtnuté.

## Registrovat {#runner-registered}

Zobrazí se, když vytváříte runner.

- Zaškrtnuto (výchozí): server vytvoří runner s jeho tokenem. Po stisknutí **Vytvořit** dialog
  ukáže token a příkazy, které s ním runner spustí.
- Nezaškrtnuto: runner se vytvoří bez tokenu. Později klikněte na **Znovu vygenerovat registrační
  token** (ikona dvou šipek) na řádku runneru, abyste získali jednorázový registrační token
  platný jednu hodinu, a zaregistrujte runner pomocí `semaphore runner register`.

Neregistrovaný runner nedostává žádné úlohy.

## Webhook {#runner-webhook}

Volitelné. URL, kterou server zavolá, když tomuto runneru předá úlohu, například aby runner
spustil na vyžádání. Server pošle požadavek `POST` s tělem JSON:

```json
{"action": "start", "project_id": 1, "task_id": 42, "template_id": 7, "runner_id": 3}
```

URL musí do 30 sekund odpovědět stavem 200 nebo 204; jinak úloha selže. Runner s webhookem se
vždy považuje za online, i když se serveru v poslední době nedotazoval. Pro runner, který běží
neustále, pole ponechte prázdné.

## Maximální počet paralelních úloh {#runner-maxParallelTasks}

Počet úloh, které smí runner spouštět současně. Po dosažení limitu jde další úloha na jiný
odpovídající runner, nebo čeká ve frontě. Ponechte pole prázdné nebo nastavte `0`, pokud nemá být
žádný limit.

## Povoleno {#runner-active}

Nové úlohy dostává pouze povolený runner. Zrušte zaškrtnutí, chcete-li runner vyřadit z provozu
bez jeho odstranění; přepínač v seznamu runnerů dělá totéž.

U nového runneru je políčko nezaškrtnuté: zaškrtněte ho, nebo runner povolte později
v seznamu.
