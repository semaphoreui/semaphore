---
form: src/views/Tokens.vue
---

# New Token

API token umožňuje skriptu nebo jiné službě volat API Semaphore vaším jménem, se stejnými
oprávněními, jaká má váš účet. Pošlete ho v hlavičce `Authorization` požadavku: `Bearer` a token.
Token se zobrazí jen jednou, hned po vytvoření: tehdy si ho zkopírujte.

## Název tokenu {#token-name}

Volitelné. Název, který token odliší v seznamu tokenů, například název skriptu, který ho používá.
Není součástí tokenu a nemění, co token smí dělat.

## Expires {#token-expires}

Kdy token přestane fungovat. Zvolte období počítané od teď, **Never** pro token, který funguje,
dokud ho neodstraníte, nebo **Custom**, chcete-li okamžik zvolit sami.

Token, jehož platnost nikdy nevyprší, funguje, dokud ho neodstraníte ze seznamu. Pro token předaný
skriptu nebo jiné službě dejte přednost období.

## Expiration date {#token-expiresAt}

Zobrazí se, když je **Expires** nastaveno na **Custom**.

Datum a čas, kdy token přestane fungovat, v časovém pásmu vašeho prohlížeče. Okamžik musí být
v budoucnosti: server odmítne token, jehož platnost by už vypršela.
