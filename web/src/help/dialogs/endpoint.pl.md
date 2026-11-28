---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Edytuj endpoint

Endpoint to punkt wejścia dla webhooków: adresy URL wywoływane przez inne usługi oraz sposób
sprawdzania ich żądań. Integracje podłącza się do endpointu, a ich uruchomienia startują przy
żądaniach, które on przyjmuje. Adresy URL endpointu, czyli jego aliasy, dodaje się na karcie
„Endpointy”.

## Nazwa {#endpointForm-name}

Nazwa, pod którą endpoint jest widoczny na karcie „Endpointy” i w formularzu integracji. Pole jest
wymagane.

## Opis {#endpointForm-description}

Opcjonalne. Notatka o tym, do czego służy endpoint; wyświetlana obok jego nazwy.

## Metoda autoryzacji {#integrationAuthForm-authMethod}

Jak Semaphore sprawdza, czy żądanie pochodzi od oczekiwanej usługi. Żądanie, które nie przejdzie
sprawdzenia, niczego nie uruchamia, a webhook odpowiada na nie tak samo jak na każde inne żądanie.

- **Brak**: brak sprawdzania. Każdy, kto zna adres URL, może uruchomić zadanie.
- **GitHub Webhooks**: nagłówek `X-Hub-Signature-256` musi zawierać podpis HMAC-SHA256 treści
  utworzony za pomocą sekretu, tak jak wysyła go GitHub.
- **Bitbucket Webhooks**: to samo sprawdzenie nagłówka `X-Hub-Signature`, tak jak wysyła go
  Bitbucket.
- **Token**: nagłówek podany w **Nagłówek autoryzacji** musi zawierać sam sekret.
- **HMAC** i **HMAC (SHA-512)**: nagłówek podany w **Nagłówek autoryzacji** musi zawierać HMAC-SHA256 lub
  HMAC-SHA512 treści utworzony za pomocą sekretu, zapisany małymi literami szesnastkowo, bez
  przedrostka takiego jak `sha256=`.
- **BasicAuth**: żądanie musi zawierać w nagłówku `Authorization` login i hasło klucza wybranego w
  **Sekret uwierzytelniania**.

## Nagłówek autoryzacji {#integrationAuthForm-authHeader}

Wyświetlane, gdy **Metoda autoryzacji** to **Token**, **HMAC** lub **HMAC (SHA-512)**.

Nazwa nagłówka żądania, który zawiera token lub podpis, na przykład `X-Webhook-Token`. Wielkość
liter w nazwie nie ma znaczenia. Żądanie bez tego nagłówka jest odrzucane.

## Sekret uwierzytelniania {#integrationAuthForm-authSecretId}

Wyświetlane, gdy **Metoda autoryzacji** jest inne niż **Brak**.

Klucz z Magazynu kluczy, względem którego sprawdzane jest żądanie. Dla **BasicAuth** pasuje tylko
klucz typu **Logowanie z hasłem**. Dla pozostałych metod pasuje także klucz typu **Ciąg znaków**:
sprawdzenie używa jego wartości lub hasła klucza **Logowanie z hasłem**.

Pole jest potrzebne przy każdej metodzie oprócz **Brak**: bez klucza lub z pustym sekretem każde
żądanie jest odrzucane.
