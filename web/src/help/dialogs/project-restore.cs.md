---
form: src/components/RestoreProjectForm.vue
---

# Obnovit projekt...

Vytvoří nový projekt ze souboru zálohy vytvořeného pomocí **Zálohovat projekt** na stránce
Nastavení projektu. Nový projekt dostane šablony úloh, inventáře, repozitáře, skupiny proměnných,
klíče, harmonogramy, integrace a další věci nastavené v záloze a vy se stanete jeho vlastníkem.
Projekt může obnovit pouze administrátor, pokud správce serveru nepovolí vytvářet projekty všem
uživatelům.

## Název projektu {#restoreProject-name}

Název nového projektu. Pole je povinné a nahrazuje název zapsaný v souboru zálohy.

Název se musí lišit od názvů všech projektů na serveru: server odmítne obnovit projekt pod
názvem, který je už obsazený.

## Soubor zálohy {#restoreProject-file}

Soubor JSON se zálohou. Vyberte soubor dříve, než stisknete **Restore**.

Záloha vynechává tajné údaje klíčů a tajné údaje na kartě **Secrets** skupin proměnných: po obnovení
vám Semaphore sdělí, kolik klíčů je prázdných a vyžaduje znovu zadat tajné údaje, a tajné údaje
skupin proměnných je také nutné zadat znovu. Heslo zapsané přímo do jiného pole, například do adresy
repozitáře, se vrátí tak, jak je v souboru. Členové projektu a historie jeho úloh nejsou součástí
zálohy. Záloha zapsaná novější verzí Semaphore je odmítnuta.
