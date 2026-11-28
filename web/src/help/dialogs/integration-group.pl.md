---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Edytuj integrację

Integracja to zestaw uruchomień, które startują z jednego endpointu. Gdy żądanie wysłane na adres
URL tego endpointu przejdzie autoryzację, startuje każde uruchomienie integracji, którego matchery
akceptują żądanie.

## Nazwa {#integrationGroupForm-name}

Nazwa, pod którą integracja jest widoczna na stronie „Integracje”. Pole jest wymagane.

## Endpoint {#integrationGroupForm-endpointId}

Endpoint, którego adresy URL uruchamiają uruchomienia tej integracji. Z jednego endpointu może
korzystać kilka integracji. Pole jest wymagane.

## Nadpisz autoryzację {#integrationGroupForm-authOverride}

Pozostaw wyłączone, aby żądania były sprawdzane tak jak w endpoincie. Włącz, aby sprawdzać żądania
tej integracji własną metodą i własnym sekretem, ustawianymi w polach, które pojawią się poniżej;
ustawienia endpointu nie mają wtedy zastosowania do tej integracji.

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
