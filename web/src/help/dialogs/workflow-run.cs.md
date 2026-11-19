---
form: src/components/WorkflowRunDialog.vue
---

# Spustit workflow

Spustí běh workflow, jakmile vyplníte jeho vstupní proměnné. Dialog se otevře pouze u workflow se
vstupními proměnnými: workflow bez nich se spustí hned. Po stisknutí **Spustit** se otevře stránka
nového běhu; když server hodnoty odmítne, dialog zůstane otevřený s vašimi odpověďmi.

## Proměnné {#workflowRunDialog-survey}

Jedno pole pro každou vstupní proměnnou workflow, v pořadí, v jakém jsou uvedené, s vyplněnou
výchozí hodnotou. Pole označené `*` je povinné, pole typu Integer přijímá pouze číslice a pole se
seznamem přijímá jednu hodnotu, nebo, když proměnná dovoluje víc hodnot, libovolný počet.
**Spustit** nic nespustí, dokud nejsou všechna pole platná.

Hodnoty se uloží s během a zobrazí se na jeho stránce v části **Variables**. Úloha workflow je
sama nedostane: proměnná její šablony dostane hodnotu jen tehdy, když ji uzel naváže na proměnnou
workflow. Samotný název nic nepředá.

Vstupní proměnné workflow neobsahují žádné tajné údaje, protože běh uchovává jejich hodnoty jako
prostý text. Hesla a tokeny místo toho uložte do klíče nebo do skupiny proměnných.
