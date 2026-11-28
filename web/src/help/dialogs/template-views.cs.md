---
form: src/components/EditViewsForm.vue
---

# Upravit zobrazení

Zobrazení jsou karty nad seznamem šablon úloh. Zobrazení ukazuje šablony, které do něj zařadíte,
nebo všechny šablony projektu. Dialog ukládá každou změnu okamžitě: pro celý dialog není žádné
tlačítko Uložit. Pořadí karet změníte přetažením zobrazení za jeho úchyt.

## Název {#viewsForm-title}

Název karty. Napište ho a potvrďte zaškrtnutím vedle pole; křížek vrátí uložený název. Nové
zobrazení se uloží, jakmile má název a vy ho potvrdíte nebo změníte některé z jeho nastavení.
Zavřením dialogu se nepotvrzený název neuloží.

Ozubené kolo otevře nastavení zobrazení pod jeho názvem.

## Odstranit {#viewsForm-viewDelete}

Ikona koše zobrazení okamžitě odstraní, bez dotazu. Šablony zobrazení se neodstraní: zůstanou
v projektu a už nejsou v žádném zobrazení.

## Type {#viewsForm-type}

Které šablony karta ukazuje:

- **Custom**: šablony, jejichž pole **Zobrazení** v dialogu šablony uvádí toto zobrazení. Toto je
  výchozí volba.
- **All**: všechny šablony projektu bez ohledu na jejich zobrazení.

Člen vidí v každém zobrazení pouze šablony, které smí vidět.

## Sort by {#viewsForm-sortColumn}

Sloupec, podle kterého se šablony karty řadí; jediný je **Name**. Karta bez něj se také řadí
podle názvu, takže pole má význam spolu s **Reverse**.

## Reverse {#viewsForm-sortReverse}

Seřadí kartu sestupně. Funguje pouze tehdy, když je nastaveno **Sort by**.

## Hidden {#viewsForm-hidden}

Skryje kartu před členy, kteří nemohou spravovat prostředky projektu. Ti, kdo to mohou, ji dál
vidí, přeškrtnutou. Skrytí karty neskryje její šablony: zobrazení typu **All** je stále uvádí.

## Show task counts {#viewsForm-showTemplateStats}

Zobrazí na kartě, kolik jejích šablon je v jakém stavu, jako jeden štítek z barevných segmentů:
zelený pro poslední úlohu, která uspěla, červený pro neúspěšnou, šedý pro zastavenou nebo pro
šablonu, která nebyla nikdy spuštěna, oranžový pro úlohu čekající na potvrzení, modrý pro úlohu ve
frontě nebo běžící. Ve výchozím stavu je vypnuto a počty se počítají jen pro zobrazení, která je
mají zapnuté.

## Přidat zobrazení {#viewsForm-add}

Přidá zobrazení na konec seznamu, s prázdným názvem k vyplnění.
