---
form: src/components/AppForm.vue
---

# Edit App

Aplikace je nástroj, kterým Semaphore spouští úlohy: Ansible, Terraform, OpenTofu, Terragrunt,
Bash, Python, PowerShell nebo vlastní příkaz. Každá šablona úlohy je určena pro jednu aplikaci
a tento dialog nastavuje, jak ji server spouští a jak se aplikace zobrazuje v rozhraní.

Nastavení se ukládá do databáze a platí okamžitě na serveru, který ho uloží; ostatní uzly
clusteru ho načtou při svém restartu. Pole, které vyprázdníte, si ponechá původní hodnotu až do
restartu serveru.

## ID {#appForm-id}

Povinné. Klíč, pod kterým je aplikace uložena; šablona úlohy odkazuje na svou aplikaci tímto
klíčem. Vestavěné aplikace jsou `ansible`, `terraform`, `tofu`, `terragrunt`, `pulumi`, `bash`,
`python` a `powershell`. Jakékoli jiné ID přidá vlastní aplikaci: Semaphore ji spouští jako
skript, příkazem uvedeným v **Path**, nebo, když je **Path** prázdné, příkazem pojmenovaným
stejně jako ID.

Použijte ID bez teček. Změna ID existující aplikace ji nepřejmenuje: nastavení se uloží jako nová
aplikace pod novým ID a původní aplikace zůstane.

## Icon {#appForm-icon}

Volitelné. Název ikony ze sady Material Design Icons bez předpony `mdi-`, například
`language-python`. Ponechte pole prázdné, chcete-li zachovat vestavěnou ikonu aplikace;
vlastní aplikace pak ukazuje otazník.

## Název {#appForm-title}

Volitelné. Název aplikace v seznamu aplikací a v nabídce **Nová šablona**. Ponechte pole prázdné,
chcete-li zachovat vestavěný název, například `Ansible Playbook`; vlastní aplikace pak nemá žádný
název, takže jí nějaký dejte.

## Path {#appForm-path}

Volitelné. Příkaz, který Semaphore spustí místo výchozího, buď název hledaný v `PATH` serveru,
nebo úplná cesta, například `/usr/local/bin/terraform-1.9`. Výchozí jsou `terraform`, `tofu`,
`terragrunt`, `bash`, `python3` a `powershell`; vlastní aplikace spouští příkaz pojmenovaný
stejně jako její ID.

Ansible toto pole ignoruje: Semaphore vždy spouští `ansible-playbook` a `ansible-galaxy`.

## Priority {#appForm-priority}

Místo aplikace v seznamu aplikací a v nabídce **Nová šablona**: aplikace s vyšším číslem je
první. Vestavěné aplikace začínají na 1000 pro Ansible a klesají až na 500 pro Python. Vestavěná
aplikace s prioritou `0` nebo nižší dostane při startu serveru zpět svou vestavěnou prioritu.

## Args {#appForm-args}

Volitelné. Argumenty vložené do příkazového řádku hned za příkaz, jeden na štítek. Kliknutím na
**+** jeden přidáte, kliknutím na štítek ho změníte, přetažením štítků změníte jejich pořadí.

- Terraform, OpenTofu a Terragrunt: argumenty jsou před argumenty, které Semaphore přidává ke
  každému příkazu, například `init` a `apply`.
- Skriptové aplikace: argumenty jsou před skriptem šablony a nahrazují výchozí: PowerShell se ve
  výchozím stavu spouští s `-File`, takže `-File` dejte ve svém seznamu na konec.
- Ansible toto pole ignoruje.

## Active {#appForm-active}

Aktivní aplikace se nabízí v nabídce **Nová šablona**. Zrušte zaškrtnutí, chcete-li aplikaci
v této nabídce skrýt: šablony, které s ní už byly vytvořeny, si svou aplikaci ponechají a dál se
spouštějí.

U nové aplikace je políčko nezaškrtnuté: zaškrtněte ho, aby byla aplikace k dispozici.
