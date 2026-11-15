---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Extrahovaná hodnota vezme část příchozího požadavku, například větev události push, a předá ji
úloze, kterou integrace spustí. Každá hodnota integrace se bere z každého požadavku, který úlohu
spustí.

## Name {#integrationValueForm-name}

Název, pod kterým je hodnota uvedena na stránce integrace. Nemění to, co se bere ani kam to jde.
Pole je povinné.

## Source of the Value {#integrationValueForm-valueSource}

Část požadavku, ze které se hodnota bere:

- **Header**: hodnota hlavičky požadavku.
- **Body**: tělo požadavku.

## Data Type of Body {#integrationValueForm-bodyDataType}

Zobrazí se, když je **Source of the Value** nastaveno na **Body**.

Jak se tělo čte:

- **JSON**: tělo se zpracuje jako JSON a **Key** je cesta k jedné hodnotě v něm. Hodnota si
  ponechá svůj typ JSON: číslo, seznam nebo objekt se k úloze dostanou jako takové.
- **String**: celé tělo se vezme jako text a **Key** se nepoužije.

## Key {#integrationValueForm-key}

Co se má z požadavku vzít. Pole je povinné.

- Pro **Header**: název hlavičky, například `X-GitHub-Event`. Na velikosti písmen nezáleží.
  Hlavička, která v požadavku chybí, dá prázdnou hodnotu.
- Pro **Body** ve formátu **JSON**: cesta k hodnotě, jejíž části jsou spojené tečkami; položka
  pole se zapisuje jako `.[0]`, například `repository.full_name` nebo `commits.[0].id`. Cesta,
  která v těle chybí, nedá žádnou hodnotu a úloha si ponechá tu nastavenou v integraci.
- Pro **Body** ve formátu **String**: nepoužije se, ale formulář vyžaduje hodnotu; napište
  cokoli.

## Variable Usage {#integrationValueForm-variableType}

Kam se hodnota v úloze šablony dostane:

- **Variables**: doplňková proměnná úlohy, stejně jako se předává odpověď na vstupní proměnnou.
  Nahradí odpověď nebo proměnnou ze skupiny proměnných se stejným názvem.
- **Task Params**: parametr úlohy, stejně jako se ukládají volby dialogu úlohy. Název
  `git_branch` nastaví větev úlohy, když šablona dovoluje větev přepsat; úvodní `refs/heads/` se
  odstraní.

Integrace, která spouští workflow, způsob použití ignoruje: každá hodnota se stane odpovědí na
vstupní proměnnou workflow uvedenou v **Variable**.

## Variable {#integrationValueForm-variable}

Název, pod kterým se hodnota předá, například `branch` pro doplňkovou proměnnou, kterou playbook
čte jako `{{ branch }}`. Pole je povinné.
