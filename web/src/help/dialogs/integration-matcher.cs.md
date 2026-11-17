---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Matcher je podmínka pro příchozí požadavek. Matchery rozhodují, kterou integraci požadavek na
alias projektu spustí, takže jedna URL může sloužit několika integracím: integrace se spustí,
jen když platí všechny její matchery. Používají se pouze tehdy, když je u integrace zapnutá volba
**Použít alias projektu**; když je zapnutá, integrace je dostupná pouze přes alias projektu.

Integrace nalezená přes alias projektu, která nemá žádné matchery, se nikdy nespustí: přidejte
alespoň jeden.

## Name {#integrationMatcherForm-name}

Název, pod kterým je matcher uveden na stránce integrace. Nemění to, co matcher kontroluje. Pole
je povinné.

## Match on {#integrationMatcherForm-matchType}

Část požadavku, kterou matcher čte:

- **Header**: hodnota hlavičky požadavku.
- **Body**: tělo požadavku.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Zobrazí se, když je **Match on** nastaveno na **Body**.

Jak se tělo čte:

- **JSON**: tělo se zpracuje jako JSON a **Key** je cesta k jedné hodnotě v něm.
- **String**: celé tělo se porovná jako text a **Key** se nepoužije.

## Key {#integrationMatcherForm-key}

Co se má z požadavku přečíst. Pole je povinné.

- Pro **Header**: název hlavičky, například `X-GitHub-Event`. Na velikosti písmen nezáleží.
- Pro **Body** ve formátu **JSON**: cesta k hodnotě, jejíž části jsou spojené tečkami; položka
  pole se zapisuje jako `.[0]`, například `repository.full_name` nebo `commits.[0].id`.
- Pro **Body** ve formátu **String**: nepoužije se, ale formulář vyžaduje hodnotu; napište
  cokoli.

## Comparison Method {#integrationMatcherForm-method}

Jak se hodnota přečtená z požadavku porovná s **Value**:

- **==**: obě hodnoty jsou stejné.
- **!=**: obě hodnoty se liší. Hlavička nebo klíč, které v požadavku chybí, se počítají jako
  odlišné.
- **Contains**: hodnota přečtená z požadavku obsahuje **Value**.

## Value {#integrationMatcherForm-value}

Text, se kterým se porovnává, například `push` nebo `refs/heads/main`. Porovnání je přesné
a rozlišuje velikost písmen. Číslo nebo logická hodnota z těla JSON se porovnává ve své textové
podobě: `42`, `true`. Pole je povinné.
