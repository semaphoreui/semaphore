---
form: src/components/SurveyVars.vue
---

# Vstupní proměnná

Vstupní proměnná je otázka, kterou dialog úlohy položí při každém spuštění šablony. Odpověď se
k úloze dostane jako proměnná: jako doplňková proměnná pro Ansible, jako `-var` pro Terraform
a OpenTofu, jako argument pro skript, nebo jako proměnná prostředí. Proměnná se uloží se
šablonou, když uložíte dialog šablony. Stejný dialog upravuje vstupní proměnné workflow, kde se
typ **Secret** nenabízí.

## Název {#surveyVarDialog-name}

Název, pod kterým úloha odpověď dostane: název doplňkové proměnné, proměnné Terraformu, argumentu
`name=value` skriptu nebo proměnné prostředí. Mezery na začátku a na konci se odstraní. Pole je
povinné.

Použijte název, který nemá žádná jiná vstupní proměnná a který nástroj přijme, například
`app_version`. Odpověď nahradí proměnnou se stejným názvem ze skupiny proměnných šablony.

## Nadpis {#surveyVarDialog-title}

Popisek pole v dialogu úlohy, například `Version to deploy`. Je to také text štítku proměnné
v dialogu šablony. Pole je povinné.

## Popis {#surveyVarDialog-description}

Volitelné. Nápověda zobrazená pod polem v dialogu úlohy, když má pole fokus: co zadat, nebo
příklad.

## Typ {#surveyVarDialog-type}

Jakou odpověď dialog úlohy požaduje:

- **String**: jeden řádek textu. Toto je výchozí volba.
- **Integer**: pouze číslice; dialog úlohy odmítne cokoli jiného, i znaménko minus. Odpověď se
  k úloze přesto dostane jako text: převeďte ji v playbooku, například pomocí `| int`.
- **Secret**: text, který dialog úlohy maskuje. Odpověď se s úlohou neukládá, takže opakované
  spuštění se na ni zeptá znovu, a Ansible ji dostane v souboru zašifrovaném pomocí vaultu. Tajná
  hodnota nemá výchozí hodnotu.
- **Enum**: jedna hodnota vybraná ze seznamu, který definujete níže.
- **Text**: několik řádků textu.
- **Select**: několik hodnot vybraných ze seznamu, který definujete níže. Úloha je dostane jako
  seznam; v argumentu, ve `-var` nebo v proměnné prostředí se seznam zapíše jako JSON, například
  `["a","b"]`.

Změna typu vymaže výchozí hodnotu a typ bez seznamu vymaže hodnoty.

## Pass variable as {#surveyVarDialog-target}

Jak se odpověď k úloze dostane:

- **Extra variable**: obvyklý způsob daného nástroje. Ansible ji dostane jako doplňkovou
  proměnnou, Terraform a OpenTofu jako `-var name=value`, skript Bash, PowerShell nebo Python
  jako argument `name=value` za skriptem.
- **Environment variable**: proměnná prostředí procesu, `name=value`, pro každý druh šablony.
  Proměnná se pak nepředává tím druhým způsobem.

## Klíč {#surveyVarDialog-valueName}

Zobrazí se, když je typ **Enum** nebo **Select**.

Text, který dialog úlohy ukazuje pro tuto hodnotu v seznamu. Každá hodnota potřebuje klíč a klíče
se musí lišit: jinak dialog odmítne proměnnou uložit.

## Hodnota {#surveyVarDialog-valueValue}

Zobrazí se, když je typ **Enum** nebo **Select**.

Co úloha dostane, když je vybrána tato hodnota. Může se lišit od svého klíče: klíč `Production`
může předat `prod`.

## Add Value {#surveyVarDialog-addValue}

Zobrazí se, když je typ **Enum** nebo **Select**.

Přidá do seznamu hodnot prázdný řádek. **Enum** i **Select** potřebují alespoň jednu hodnotu.
Ikona koše na řádku ho odstraní.

## Výchozí hodnota {#surveyVarDialog-defaultValue}

Zobrazí se u každého typu kromě **Secret**.

Odpověď, kterou dialog úlohy vyplní při otevření; můžete ji tam změnit. U **Enum** vyberte jednu
z hodnot, u **Select** libovolný počet. Ponechte pole prázdné, chcete-li začít s prázdnou
odpovědí.

Výchozí hodnotu vyplňuje pouze dialog úlohy. Úloha spuštěná jinak, například přes API, žádnou
výchozí hodnotu nedostane.

## Povinné {#surveyVarDialog-required}

Dialog úlohy nespustí úlohu, dokud proměnná nemá odpověď. Stejně jako výchozí hodnotu kontroluje
to pouze dialog úlohy, ne server.
