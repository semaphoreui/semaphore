---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Upravit endpoint

Endpoint je vstupní bod pro webhooky: adresy URL, které volají jiné služby, a způsob, jakým se
jejich požadavky kontrolují. Integrace se připojují k endpointu a provádějí svá spuštění při
požadavcích, které endpoint přijme. Adresy URL endpointu, jeho aliasy, se přidávají na kartě
„Endpointy“.

## Název {#endpointForm-name}

Název, pod kterým je endpoint uveden na kartě „Endpointy“ a ve formuláři integrace. Pole je povinné.

## Popis {#endpointForm-description}

Nepovinné. Poznámka o tom, k čemu endpoint slouží; zobrazuje se vedle jeho názvu.

## Metoda autorizace {#integrationAuthForm-authMethod}

Jak Semaphore ověřuje, že požadavek pochází od služby, kterou očekáváte. Požadavek, který
kontrolou neprojde, nic nespustí a webhook na něj odpoví stejně jako na kterýkoli jiný
požadavek.

- **Žádný**: žádná kontrola. Úlohu může spustit kdokoli, kdo zná URL.
- **GitHub Webhooks**: hlavička `X-Hub-Signature-256` musí obsahovat podpis HMAC-SHA256 těla
  vytvořený pomocí tajného údaje, tak jak ho posílá GitHub.
- **Bitbucket Webhooks**: stejná kontrola hlavičky `X-Hub-Signature`, tak jak ji posílá
  Bitbucket.
- **Token**: hlavička uvedená v **Autorizační hlavička** musí obsahovat samotný tajný údaj.
- **HMAC** a **HMAC (SHA-512)**: hlavička uvedená v **Autorizační hlavička** musí obsahovat HMAC-SHA256
  nebo HMAC-SHA512 těla vytvořený pomocí tajného údaje, šestnáctkově malými písmeny, bez předpony
  jako `sha256=`.
- **BasicAuth**: požadavek musí v hlavičce `Authorization` nést přihlašovací jméno a heslo klíče
  z **Autentizační tajemství**.

## Autorizační hlavička {#integrationAuthForm-authHeader}

Zobrazí se, když je **Metoda autorizace** nastaveno na **Token**, **HMAC** nebo **HMAC (SHA-512)**.

Název hlavičky požadavku, která nese token nebo podpis, například `X-Webhook-Token`. Na velikosti
písmen v názvu nezáleží. Požadavek bez této hlavičky je odmítnut.

## Autentizační tajemství {#integrationAuthForm-authSecretId}

Zobrazí se, když **Metoda autorizace** není **Žádný**.

Klíč z Úložiště klíčů, proti kterému se požadavek ověřuje. Pro **BasicAuth** se hodí pouze klíč
typu **Přihlášení heslem**. Pro ostatní metody se hodí i klíč typu **Řetězec**: kontrola použije
jeho hodnotu, nebo heslo klíče **Přihlášení heslem**.

Pole potřebuje každá metoda kromě **Žádný**: bez klíče, nebo s prázdným tajným údajem, je každý
požadavek odmítnut.
